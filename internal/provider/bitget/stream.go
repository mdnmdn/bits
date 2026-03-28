package bitget

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/ws"
)

const wsURL = "wss://ws.bitget.com/v2/ws/public"

// bitgetWSMessage represents the common structure of Bitget WS messages.
type bitgetWSMessage struct {
	Action string          `json:"action"`
	Arg    bitgetWSArg     `json:"arg"`
	Data   json.RawMessage `json:"data"`
	Ts     int64           `json:"ts"`
}

type bitgetWSArg struct {
	InstType string `json:"instType"`
	Channel  string `json:"channel"`
	InstID   string `json:"instId"`
}

type bitgetWSTickerData struct {
	InstID    string `json:"instId"`
	LastPr    string `json:"lastPr"`
	Change24h string `json:"change24h"`
}

type bitgetWSDepthData struct {
	Asks [][]string `json:"asks"`
	Bids [][]string `json:"bids"`
	Ts   string     `json:"ts"`
}

// WatchPrices implements provider.PriceStreamProvider.
func (c *Client) WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	updates := make(chan *model.CoinPrice, 100)
	base := ws.NewBaseClient(wsURL)
	base.UserAgent = c.userAgent

	base.OnConnect = func(ctx context.Context, conn *websocket.Conn) error {
		args := make([]map[string]string, 0, len(ids))
		for _, id := range ids {
			args = append(args, map[string]string{
				"instType": "SPOT",
				"channel":  "ticker",
				"instId":   id,
			})
		}
		return base.WriteJSON(map[string]any{
			"op":   "subscribe",
			"args": args,
		})
	}

	base.OnMessage = func(ctx context.Context, raw []byte) error {
		var msg bitgetWSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			return nil
		}
		if msg.Action != "snapshot" && msg.Action != "update" {
			return nil
		}

		var data []bitgetWSTickerData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return nil
		}

		for _, d := range data {
			price, _ := strconv.ParseFloat(d.LastPr, 64)
			changePct, _ := strconv.ParseFloat(d.Change24h, 64)
			changePct *= 100

			select {
			case updates <- &model.CoinPrice{
				ID:        d.InstID,
				Symbol:    d.InstID,
				Price:     price,
				Change24h: &changePct,
			}:
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	}

	if err := base.Connect(ctx); err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		_ = base.Close()
		close(updates)
	}()

	return updates, nil
}

// WatchOrderBook implements provider.OrderBookStreamProvider.
func (c *Client) WatchOrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	updates := make(chan *model.OrderBook, 100)
	base := ws.NewBaseClient(wsURL)
	base.UserAgent = c.userAgent

	instType := "SPOT"
	if market == model.MarketFutures {
		instType = "USDT-FUTURES"
	}

	channel := "depth"
	if depth > 0 && depth <= 15 {
		// Bitget supports depth5, depth15 for aggregated depth
		if depth <= 5 {
			channel = "depth5"
		} else {
			channel = "depth15"
		}
	}

	base.OnConnect = func(ctx context.Context, conn *websocket.Conn) error {
		return base.WriteJSON(map[string]any{
			"op": "subscribe",
			"args": []map[string]string{
				{
					"instType": instType,
					"channel":  channel,
					"instId":   symbol,
				},
			},
		})
	}

	base.OnMessage = func(ctx context.Context, raw []byte) error {
		var msg bitgetWSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			return nil
		}
		if msg.Action != "snapshot" && msg.Action != "update" {
			return nil
		}

		var data []bitgetWSDepthData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return nil
		}

		parseEntries := func(raw [][]string) []model.OrderBookEntry {
			entries := make([]model.OrderBookEntry, 0, len(raw))
			for _, e := range raw {
				if len(e) >= 2 {
					price, _ := strconv.ParseFloat(e[0], 64)
					qty, _ := strconv.ParseFloat(e[1], 64)
					entries = append(entries, model.OrderBookEntry{Price: price, Quantity: qty})
				}
			}
			return entries
		}

		for _, d := range data {
			var ts *time.Time
			if d.Ts != "" {
				if ms, err := strconv.ParseInt(d.Ts, 10, 64); err == nil {
					t := time.UnixMilli(ms)
					ts = &t
				}
			}

			ob := &model.OrderBook{
				Symbol: symbol,
				Market: market,
				Bids:   parseEntries(d.Bids),
				Asks:   parseEntries(d.Asks),
				Time:   ts,
			}

			select {
			case updates <- ob:
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	}

	if err := base.Connect(ctx); err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		_ = base.Close()
		close(updates)
	}()

	return updates, nil
}
