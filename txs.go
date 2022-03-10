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
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"go.blockdaemon.com/pyth_exporter/metrics"
	"go.uber.org/zap"
)

// TODO(richard): consider using logsSubscribe() instead of polling.

// txScraper polls publisher transactions.
type txScraper struct {
	tailers []*txTailer
	log     *zap.Logger
	rpc     *rpc.Client
}

func newTxScraper(rpcURL string, log *zap.Logger, publishers []solana.PublicKey) *txScraper {
	scraper := &txScraper{
		rpc:     rpc.New(rpcURL),
		log:     log,
		tailers: make([]*txTailer, len(publishers)),
	}

	for i, pubkey := range publishers {
		scraper.tailers[i] = newTxTailer(scraper, pubkey)
	}

	return scraper
}

func (s *txScraper) run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.log.Info("Polling transactions of publishers", zap.Int("num_publishers", len(s.tailers)))

	for {
	wait:
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			break wait
		}

		s.poll(ctx, interval)
	}
}

func (s *txScraper) poll(ctx context.Context, interval time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, interval)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(s.tailers))
	for _, tailer := range s.tailers {
		go s.pollOne(ctx, &wg, tailer)
	}
	wg.Wait()
}

func (s *txScraper) pollOne(ctx context.Context, wg *sync.WaitGroup, tailer *txTailer) {
	defer wg.Done()
	for {
		isEnd, err := tailer.poll(ctx)
		if err != nil {
			s.log.Warn("Failed to poll account txs", zap.Error(err))
		}
		if isEnd {
			break
		}
	}
}

// txTailer "tails" the transaction log of an account.
type txTailer struct {
	*txScraper

	pubkey    solana.PublicKey
	pubkeyStr string
	lastSlot  uint64
	lastSig   solana.Signature
}

func newTxTailer(scraper *txScraper, pubkey solana.PublicKey) *txTailer {
	return &txTailer{
		txScraper: scraper,
		pubkey:    pubkey,
		pubkeyStr: pubkey.String(),
	}
}

func (t *txTailer) refreshLastSig(ctx context.Context) error {
	t.log.Debug("Getting latest sig",
		zap.String("publisher", t.pubkeyStr))

	oneInt := 1
	sigs, err := t.rpc.GetSignaturesForAddressWithOpts(ctx, t.pubkey, &rpc.GetSignaturesForAddressOpts{
		Limit: &oneInt,
	})
	if err != nil {
		return err
	}
	metrics.RpcRequestsTotal.Inc()

	if len(sigs) == 0 {
		t.log.Debug("Publisher has not sent any txs yet",
			zap.String("publisher", t.pubkeyStr))
		return nil // empty account
	}
	t.lastSlot = sigs[0].Slot
	t.lastSig = sigs[0].Signature

	t.log.Debug("Tailing txs starting at",
		zap.String("publisher", t.pubkeyStr),
		zap.Stringer("start_sig", t.lastSig))

	return nil
}

// poll retrieves the latest transactions.
func (t *txTailer) poll(ctx context.Context) (end bool, err error) {
	t.log.Debug("Polling new txs",
		zap.String("publisher", t.pubkeyStr),
		zap.Stringer("last_sig", t.lastSig))

	// Get starting sig if none.
	if t.lastSig.IsZero() {
		return true, t.refreshLastSig(ctx)
	}

	// Get sigs since last check.
	var tailLimit = 100
	sigs, err := t.rpc.GetSignaturesForAddressWithOpts(ctx, t.pubkey, &rpc.GetSignaturesForAddressOpts{
		Limit: &tailLimit,
		Until: t.lastSig,
	})
	if err != nil {
		return true, err
	}
	metrics.RpcRequestsTotal.Inc()

	if len(sigs) == 0 {
		return true, nil
	}

	// Iteration is newest to latest.
	// So write down the first sig as the newest sig, so we can later continue.
	stopSlot := t.lastSlot
	if sigs[0].Slot > t.lastSlot {
		t.lastSlot = sigs[0].Slot
		t.lastSig = sigs[0].Signature
	}

	// If the number of returned sigs matches exactly our requested limit, there is probably more.
	end = len(sigs) != tailLimit

	// Scroll through page.
	for len(sigs) > 0 && sigs[0].Slot > stopSlot {
		sig := sigs[0]

		var status string
		if sig.Err != nil {
			status = metrics.TxStatusFailed
		} else {
			status = metrics.TxStatusSuccess
		}
		metrics.TxCount.WithLabelValues(t.pubkeyStr, status).Inc()

		sigs = sigs[1:]
	}

	return end, nil
}
