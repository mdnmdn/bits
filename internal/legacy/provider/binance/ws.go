package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mdnmdn/bits/internal/legacy/model"
	"github.com/mdnmdn/bits/internal/legacy/ws"
)

// WatchOrderBook implements OrderBookStreamProvider.
// It uses Binance combined streams: wss://stream.binance.com:9443/stream?streams=btcusdt@depth20/ethusdt@depth20
func (c *Client) WatchOrderBook(ctx context.Context, symbols []string, limit int) (<-chan *model.OrderBook, error) {
	if limit <= 0 {
		limit = 20
	}

	// Build streams string: btcusdt@depth20
	streams := make([]string, len(symbols))
	for i, s := range symbols {
		streams[i] = fmt.Sprintf("%s@depth%d", strings.ToLower(s), limit)
	}

	url := fmt.Sprintf("wss://stream.binance.com:9443/stream?streams=%s", strings.Join(streams, "/"))
	if c.config.Futures.UseTestnet {
		url = fmt.Sprintf("wss://testnet.binance.vision/stream?streams=%s", strings.Join(streams, "/"))
	}

	updates := make(chan *model.OrderBook, 100)
	base := ws.NewBaseClient(url)
	base.UserAgent = c.userAgent

	base.OnMessage = func(ctx context.Context, raw []byte) error {
		var combined struct {
			Stream string          `json:"stream"`
			Data   depthStreamData `json:"data"`
		}
		if err := json.Unmarshal(raw, &combined); err != nil {
			return err
		}

		// Binance stream names look like "btcusdt@depth20"
		parts := strings.Split(combined.Stream, "@")
		symbol := strings.ToUpper(parts[0])

		ob := &model.OrderBook{
			Symbol: symbol,
			Bids:   make([]model.OrderBookEntry, 0, len(combined.Data.Bids)),
			Asks:   make([]model.OrderBookEntry, 0, len(combined.Data.Asks)),
		}

		for _, b := range combined.Data.Bids {
			p, _ := strconv.ParseFloat(b[0], 64)
			q, _ := strconv.ParseFloat(b[1], 64)
			ob.Bids = append(ob.Bids, model.OrderBookEntry{Price: p, Quantity: q})
		}
		for _, a := range combined.Data.Asks {
			p, _ := strconv.ParseFloat(a[0], 64)
			q, _ := strconv.ParseFloat(a[1], 64)
			ob.Asks = append(ob.Asks, model.OrderBookEntry{Price: p, Quantity: q})
		}

		select {
		case updates <- ob:
		case <-ctx.Done():
		}
		return nil
	}

	if err := base.Connect(ctx); err != nil {
		return nil, err
	}

	// Close updates channel when base client stops
	go func() {
		<-ctx.Done()
		base.Close()
		close(updates)
	}()

	return updates, nil
}

type depthStreamData struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}
