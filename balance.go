package main

import (
	"context"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"go.blockdaemon.com/pyth_exporter/metrics"
	"go.uber.org/zap"
)

// balanceScraper retrieves the Pyth publisher balances on request.
type balanceScraper struct {
	*prometheus.GaugeVec

	log        *zap.Logger
	rpc        *rpc.Client
	publishers []solana.PublicKey
}

func newBalanceScraper(publishers []solana.PublicKey, rpcURL string, log *zap.Logger) *balanceScraper {
	return &balanceScraper{
		GaugeVec:   metrics.PublisherBalances,
		rpc:        rpc.New(rpcURL),
		log:        log,
		publishers: publishers,
	}
}

// Collect gets invoked by the Prometheus exporter when a new scrape is requested.
func (b *balanceScraper) Collect(metrics chan<- prometheus.Metric) {
	const collectTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), collectTimeout)
	defer cancel()
	b.update(ctx)
	b.GaugeVec.Collect(metrics)
}

func (b *balanceScraper) update(ctx context.Context) {
	res, err := b.rpc.GetMultipleAccounts(ctx, b.publishers...)
	if err != nil {
		b.log.Warn("Failed to check publisher SOL balances", zap.Error(err))
		return
	}
	for i, acc := range res.Value {
		b.GaugeVec.
			WithLabelValues(b.publishers[i].String()).
			Set(float64(acc.Lamports))
	}
}
