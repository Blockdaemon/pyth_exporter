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
	"math"

	"github.com/gagliardetto/solana-go"
	"go.blockdaemon.com/pyth_exporter/metrics"
	"go.blockdaemon.com/pyth_exporter/pyth"
)

// priceScraper scrapes prices out of the on-chain Pyth price accounts.
type priceScraper struct {
	productKeys []solana.PublicKey
	publishKeys []solana.PublicKey
}

func (p *priceScraper) onUpdate(update pyth.PriceAccountUpdate) {
	if !p.isInteresting(update) {
		return
	}
	decimals := math.Pow10(int(update.Exponent))
	p.aggregate(&update.Product, &update.Agg, decimals)
	// Update price of each publisher.
	for _, publisher := range p.publishKeys {
		comp := update.GetComponent(&publisher)
		if comp != nil {
			p.component(&update.Product, &publisher, comp, decimals)
		}
	}
}

func (p *priceScraper) isInteresting(update pyth.PriceAccountUpdate) bool {
	for _, product := range p.productKeys {
		if product == update.Product {
			return true
		}
	}
	return false
}

// aggregate exports price as aggregated by the smart contract.
func (p *priceScraper) aggregate(product *solana.PublicKey, agg *pyth.PriceInfo, decimals float64) {
	productStr := product.String()
	metrics.AggPrice.
		WithLabelValues(productStr).
		Set(float64(agg.Price) * decimals)
	metrics.AggConf.
		WithLabelValues(productStr).
		Set(float64(agg.Conf) * decimals)
}

// component exports a price component (i.e. a price value published by an individual Pyth publisher).
func (p *priceScraper) component(product *solana.PublicKey, publisher *solana.PublicKey, comp *pyth.PriceComp, decimals float64) {
	productStr := product.String()
	publisherStr := publisher.String()
	metrics.PublisherSlot.
		WithLabelValues(productStr, publisherStr).
		Set(float64(comp.Latest.PubSlot))
	metrics.PublisherPrice.
		WithLabelValues(productStr, publisherStr).
		Set(float64(comp.Latest.Price) * decimals)
	metrics.PublisherConf.
		WithLabelValues(productStr, publisherStr).
		Set(float64(comp.Latest.Conf) * decimals)
}
