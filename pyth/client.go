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

package pyth

import (
	"context"
	"errors"

	"github.com/cenkalti/backoff/v4"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"go.blockdaemon.com/pyth_exporter/metrics"
	"go.uber.org/zap"
)

type Client struct {
	Opts

	Log          *zap.Logger
	WebSocketURL string
}

type Opts struct {
	ProgramKey solana.PublicKey
}

type PriceAccountUpdate struct {
	Slot uint64
	*PriceAccount
}

// StreamPriceAccounts sends an update to Prometheus any time a Pyth oracle account changes.
func (c *Client) StreamPriceAccounts(ctx context.Context, updates chan<- PriceAccountUpdate) error {
	return backoff.Retry(func() error {
		err := c.streamPriceAccounts(ctx, updates)
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return backoff.Permanent(err)
		default:
			return err
		}
	}, backoff.NewExponentialBackOff())
}

func (c *Client) streamPriceAccounts(ctx context.Context, updates chan<- PriceAccountUpdate) error {
	client, err := ws.Connect(ctx, c.WebSocketURL)
	if err != nil {
		return err
	}
	defer client.Close()

	metrics.WsActiveConns.Inc()
	defer metrics.WsActiveConns.Dec()

	sub, err := client.ProgramSubscribeWithOpts(
		c.Opts.ProgramKey,
		rpc.CommitmentConfirmed,
		solana.EncodingBase64Zstd,
		[]rpc.RPCFilter{
			{
				Memcmp: &rpc.RPCFilterMemcmp{
					Offset: 0,
					Bytes: solana.Base58{
						0xd4, 0xc3, 0xb2, 0xa1, // Magic
						0x02, 0x00, 0x00, 0x00, // V2
					},
				},
			},
		},
	)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	// Terminate subscription when context cancels.
	go func() {
		defer sub.Unsubscribe()
		<-ctx.Done()
	}()

	// Stream updates.
	for {
		update, err := sub.Recv()
		if err != nil {
			return err
		}
		metrics.WsEventsTotal.Inc()
		if update.Value.Account.Owner != c.Opts.ProgramKey {
			continue
		}
		accountData := update.Value.Account.Data.GetBinary()
		if PeekAccount(accountData) != AccountTypePrice {
			continue
		}
		priceAcc := new(PriceAccount)
		if err := priceAcc.UnmarshalBinary(accountData); err != nil {
			c.Log.Warn("Failed to unmarshal priceAcc account", zap.Error(err))
			continue
		}
		updates <- PriceAccountUpdate{
			Slot:         update.Context.Slot,
			PriceAccount: priceAcc,
		}
	}
}
