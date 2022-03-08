package main

import (
	"context"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"go.blockdaemon.com/pyth"
)

type productCacher struct {
	symbols map[solana.PublicKey]string
	client  *pyth.Client
	lock    sync.RWMutex
}

// discoverProducts lists all products in the Pyth oracle. Creates a symbol cache and product list.
func discoverProducts(ctx context.Context, client *pyth.Client) (*productCacher, []pyth.ProductAccountEntry, error) {
	products, err := client.GetAllProductAccounts(ctx, rpc.CommitmentProcessed)
	if err != nil {
		return nil, nil, err
	}

	symbols := make(map[solana.PublicKey]string)
	for _, p := range products {
		symbols[p.Pubkey] = p.Attrs.KVs()["symbol"]
	}

	return &productCacher{symbols: symbols}, products, nil
}

// getSymbol looks up a symbol name from the cache or fetches it from RPC.
func (p *productCacher) getSymbol(product solana.PublicKey) (string, error) {
	sym, ok := p.getSymbolFromCache(product)
	if ok {
		return sym, nil
	}
	return p.fetchProduct(product)
}

func (p *productCacher) getSymbolFromCache(product solana.PublicKey) (string, bool) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	sym, ok := p.symbols[product]
	return sym, ok
}

func (p *productCacher) fetchProduct(product solana.PublicKey) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	account, err := p.client.GetProductAccount(ctx, product, rpc.CommitmentProcessed)
	if err != nil {
		return "", err
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	sym := account.Attrs.KVs()["symbol"]
	p.symbols[product] = sym

	return sym, nil
}
