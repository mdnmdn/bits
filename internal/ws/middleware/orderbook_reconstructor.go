package middleware

import (
	"context"
	"sort"
	"sync"

	"github.com/mdnmdn/bits/model"
)

type OrderBookReconstructor struct {
	mu    sync.RWMutex
	books map[string]*model.OrderBook
}

func NewOrderBookReconstructor() *OrderBookReconstructor {
	return &OrderBookReconstructor{
		books: make(map[string]*model.OrderBook),
	}
}

func (r *OrderBookReconstructor) Middleware(ctx context.Context, msg any, next func(any) (any, error)) (any, error) {
	resp, ok := msg.(*model.Response[model.OrderBook])
	if !ok {
		return next(msg)
	}

	book := &resp.Data
	isSnapshot := book.LastUpdateID == nil

	if isSnapshot {
		r.mu.Lock()
		r.books[book.Symbol] = book
		r.mu.Unlock()
		return next(msg)
	}

	r.mu.RLock()
	local, found := r.books[book.Symbol]
	r.mu.RUnlock()

	if !found {
		return next(msg)
	}

	local.Bids = mergeLevels(local.Bids, book.Bids)
	local.Asks = mergeLevels(local.Asks, book.Asks)
	local.Time = book.Time

	// Return a copy of the reconstructed full book to avoid aliasing issues
	reconstructed := &model.Response[model.OrderBook]{
		Kind:     resp.Kind,
		Provider: resp.Provider,
		Data: model.OrderBook{
			Symbol:       local.Symbol,
			Market:       local.Market,
			Bids:         copyEntries(local.Bids),
			Asks:         copyEntries(local.Asks),
			Time:         local.Time,
			LastUpdateID: local.LastUpdateID,
			Extra:        local.Extra,
		},
	}
	return next(reconstructed)
}

func copyEntries(entries []model.OrderBookEntry) []model.OrderBookEntry {
	if entries == nil {
		return nil
	}
	result := make([]model.OrderBookEntry, len(entries))
	copy(result, entries)
	return result
}

func mergeLevels(existing, updates []model.OrderBookEntry) []model.OrderBookEntry {
	if len(existing) == 0 {
		return updates
	}
	if len(updates) == 0 {
		return existing
	}

	priceMap := make(map[float64]float64, len(existing)+len(updates))
	for _, e := range existing {
		priceMap[e.Price] = e.Quantity
	}
	for _, e := range updates {
		if e.Quantity == 0 {
			delete(priceMap, e.Price)
		} else {
			priceMap[e.Price] = e.Quantity
		}
	}

	if len(priceMap) == 0 {
		return nil
	}

	result := make([]model.OrderBookEntry, 0, len(priceMap))
	for price, qty := range priceMap {
		result = append(result, model.OrderBookEntry{Price: price, Quantity: qty})
	}

	sortByPrice(result)
	return result
}

func sortByPrice(entries []model.OrderBookEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Price > entries[j].Price
	})
}

func (r *OrderBookReconstructor) Reset(symbol string) {
	r.mu.Lock()
	delete(r.books, symbol)
	r.mu.Unlock()
}
