//  Copyright 2022 Blockdaemon Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.blockdaemon.com/pyth"
	"go.blockdaemon.com/pyth_exporter/metrics"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	flagDev      bool
	flagLogLevel = zap.InfoLevel
)

func main() {
	// Define flags.
	flag.BoolVar(&flagDev, "dev", false, "Run in development mode")
	listen := flag.String("listen", ":8080", "Address where to serve debug info and metrics HTTP server")
	flag.Var(&flagLogLevel, "log-level", "Log level")
	envStr := flag.String("env", "mainnet", "Pyth env (devnet, testnet, mainnet)")
	var programKey solana.PublicKey
	flag.Var(&programKey, "program", "Pyth program key (derived from env)")
	var mappingKey solana.PublicKey
	flag.Var(&mappingKey, "mapping", "Pyth mapping key (derived from env)")
	rpcURL := flag.String("rpc", "", "Solana RPC URL")
	wsURL := flag.String("ws", "", "Solana WebSocket RPC URL")
	var productKeys pubkeyList // if empty, assuming all products
	flag.Var(&productKeys, "products", "Pyth product keys (default all)")
	var publishKeys pubkeyList
	flag.Var(&publishKeys, "publishers", "Pyth publishers (default all)")
	flag.Parse()

	log := getLogger()
	defer log.Sync()

	// Check flag values.
	var env pyth.Env
	if programKey.IsZero() {
		switch *envStr {
		case "devnet":
			env = pyth.Devnet
		case "testnet":
			env = pyth.Testnet
		case "mainnet":
			env = pyth.Mainnet
		default:
			log.Fatal("Missing -program or -env flag")
		}
		log.Sugar().Infof("Using network %s", *envStr)
	} else {
		env = pyth.Env{
			Program: programKey,
		}
		log.Sugar().Infof("Using program ID %s", programKey)
	}
	if !mappingKey.IsZero() {
		env.Mapping = mappingKey
	}
	if env.Mapping.IsZero() {
		log.Fatal("Missing mapping key")
	}
	if *rpcURL == "" {
		log.Fatal("Missing -rpc flag")
	}
	if *wsURL == "" {
		if !strings.HasPrefix(*rpcURL, "http://") && !strings.HasPrefix(*rpcURL, "https://") {
			log.Fatal("Missing -ws flag")
		}
		*wsURL = "ws" + strings.TrimPrefix(*rpcURL, "http")
	}

	// Initiate connectivity to Pyth.
	// If running without filters, we might need to list all publishers first (by listing all mappings, products, and prices).
	client := pyth.NewClient(env, *rpcURL, *wsURL)
	allPublishers := publishKeys.pubkeys

	const initTimeout = 30 * time.Second
	initContext, initCancel := context.WithTimeout(context.Background(), initTimeout)

	log.Info("Listing all products")
	productCache, products, err := discoverProducts(initContext, client)
	if err != nil {
		log.Fatal("Failed to list products")
	}
	if len(allPublishers) == 0 {
		log.Info("Listing all publishers")
		allPublishers, err = discoverPublishers(initContext, client, products)
		if err != nil {
			log.Fatal("Failed to discover publishers", zap.Error(err))
		}
	}

	// The main context is composed of a loose set of services.
	// If any of them terminates, the rest will get terminated gracefully through context.
	initCancel()
	ctx := context.Background()
	group, ctx := errgroup.WithContext(ctx)

	// Start HTTP server.
	group.Go(func() error {
		httpLog := log.Named("http")
		errorLog, err := zap.NewStdLogAt(httpLog, zap.ErrorLevel)
		if err != nil {
			return fmt.Errorf("failed to create error log: %w", err)
		}

		// Setup handlers.
		http.HandleFunc("/health", func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodGet {
				http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte("ok"))
		})
		http.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			ErrorLog:          errorLog,
			EnableOpenMetrics: true,
		}))

		server := &http.Server{
			Addr:     *listen,
			ErrorLog: errorLog,
		}

		// Register shutdown handler, allowing clients to gracefully disconnect.
		go func() {
			<-ctx.Done()
			const shutdownGracePeriod = 5 * time.Second
			shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
			defer cancel()
			httpLog.Info("Stopping HTTP server")
			if err := server.Shutdown(shutdownCtx); err != nil {
				httpLog.Error("Error during server shutdown", zap.Error(err))
			}
		}()

		httpLog.Info("Starting HTTP server", zap.String("listen", *listen))
		defer httpLog.Debug("Stopped HTTP server")
		return server.ListenAndServe()
	})

	// Pull price updates from RPC.
	priceStream := client.StreamPriceAccounts()
	go func() {
		defer priceStream.Close()
		<-ctx.Done()
	}()
	group.Go(func() error {
		defer log.Debug("Stopped price streamer")
		return priceStream.Err()
	})

	// Send price updates to Prometheus.
	prices := priceScraper{
		log:          log.Named("scraper"),
		productKeys:  productKeys.pubkeys,
		publishKeys:  publishKeys.pubkeys,
		productCache: productCache,
	}
	group.Go(func() error {
		defer log.Debug("Stopped price scraper")
		for update := range priceStream.Updates() {
			prices.onUpdate(update)
		}
		return nil
	})

	if len(allPublishers) > 0 {
		// Scrape publisher balances.
		balances := newBalanceScraper(publishKeys.pubkeys, *rpcURL, log.Named("balances"))
		metrics.Registry.MustRegister(balances)

		// Create tx tailer.
		txs := newTxScraper(*rpcURL, log.Named("txs"), allPublishers)
		group.Go(func() error {
			defer log.Debug("Stopped tx tailer")
			const scrapeInterval = 5 * time.Second
			txs.run(ctx, scrapeInterval)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		log.Fatal("App crashed", zap.Error(err))
	}
	log.Info("App exiting")
}

type pubkeyList struct {
	pubkeys []solana.PublicKey
}

func (p *pubkeyList) Set(v string) error {
	fields := strings.Fields(v)
	p.pubkeys = make([]solana.PublicKey, len(fields))
	for i, field := range fields {
		if err := p.pubkeys[i].Set(field); err != nil {
			return fmt.Errorf("invalid pubkey %s: %w", field, err)
		}
	}
	return nil
}

func (p *pubkeyList) String() string {
	var builder strings.Builder
	for i, pubkey := range p.pubkeys {
		if i != 0 {
			builder.WriteRune(' ')
		}
		builder.WriteString(pubkey.String())
	}
	return builder.String()
}

// discoverPublishers scrapes all authorized publishers for all products from the Pyth program.
func discoverPublishers(ctx context.Context, client *pyth.Client, products []pyth.ProductAccountEntry) ([]solana.PublicKey, error) {
	priceKeys := make([]solana.PublicKey, 0, len(products))
	for _, product := range products {
		if !product.FirstPrice.IsZero() {
			priceKeys = append(priceKeys, product.FirstPrice)
		}
	}
	entries, err := client.GetPriceAccountsRecursive(ctx, rpc.CommitmentProcessed, priceKeys...)
	if err != nil {
		return nil, err
	}

	publishers := make(map[solana.PublicKey]struct{})
	for _, entry := range entries {
		for _, comp := range entry.Components {
			if !comp.Publisher.IsZero() {
				publishers[comp.Publisher] = struct{}{}
			}
		}
	}

	publishKeys := make([]solana.PublicKey, 0, len(publishers))
	for pub := range publishers {
		publishKeys = append(publishKeys, pub)
	}

	return publishKeys, nil
}
