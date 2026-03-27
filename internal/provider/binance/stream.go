package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/mdnmdn/bits/internal/model"
)

type depthStreamData struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

// WatchOrderBook streams live order book depth updates for a single symbol.
// It connects to the Binance combined stream endpoint using gorilla/websocket.
func (c *Client) WatchOrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	if depth <= 0 {
		depth = 20
	}

	streamName := fmt.Sprintf("%s@depth%d", strings.ToLower(symbol), depth)

	var url string
	if c.cfg.Futures.UseTestnet {
		url = fmt.Sprintf("wss://testnet.binance.vision/stream?streams=%s", streamName)
	} else {
		url = fmt.Sprintf("wss://stream.binance.com:9443/stream?streams=%s", streamName)
	}

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("binance: WatchOrderBook dial %s: %w", url, err)
	}

	updates := make(chan *model.OrderBook, 100)

	go func() {
		defer func() {
			conn.Close()
			close(updates)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			_, raw, err := conn.ReadMessage()
			if err != nil {
				// Connection closed or context cancelled — exit gracefully.
				return
			}

			var combined struct {
				Stream string          `json:"stream"`
				Data   depthStreamData `json:"data"`
			}
			if err := json.Unmarshal(raw, &combined); err != nil {
				continue
			}

			// Stream name looks like "btcusdt@depth20" — extract symbol.
			parts := strings.SplitN(combined.Stream, "@", 2)
			sym := strings.ToUpper(parts[0])

			uid := combined.Data.LastUpdateID
			ob := &model.OrderBook{
				Symbol:       sym,
				Market:       market,
				LastUpdateID: &uid,
				Bids:         parseStringPairs(combined.Data.Bids),
				Asks:         parseStringPairs(combined.Data.Asks),
			}

			select {
			case updates <- ob:
			case <-ctx.Done():
				return
			}
		}
	}()

	return updates, nil
}

func parseStringPairs(pairs [][]string) []model.OrderBookEntry {
	result := make([]model.OrderBookEntry, 0, len(pairs))
	for _, p := range pairs {
		if len(p) < 2 {
			continue
		}
		price, _ := strconv.ParseFloat(p[0], 64)
		qty, _ := strconv.ParseFloat(p[1], 64)
		result = append(result, model.OrderBookEntry{Price: price, Quantity: qty})
	}
	return result
}
