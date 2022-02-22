package main

import (
	"math"

	"github.com/gagliardetto/solana-go"
	"go.blockdaemon.com/pyth_exporter/metrics"
	"go.blockdaemon.com/pyth_exporter/pyth"
)

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
