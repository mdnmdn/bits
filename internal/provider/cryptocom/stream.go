package cryptocom

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/ws"
)

const wsMarketURL = "wss://stream.crypto.com/v2/market"

// cryptocomHandler implements ws.MessageHandler for the Crypto.com market
// WebSocket stream.
//
// Protocol notes:
//   - Subscribe:   send {"id":N,"method":"subscribe","params":{"channels":[...]}}
//   - Data push:   method=="subscribe", result.channel=="ticker"|"book"
//   - Heartbeat:   server sends method=="public/heartbeat" every ~30 s;
//     client must reply with {"id":N,"method":"public/respond-heartbeat"}
//     using the same id.  We buffer the received IDs in a channel and flush
//     them on every OnPing tick (also ~30 s).
type cryptocomHandler struct {
	providerID string
	userAgent  string
	heartbeats chan int64 // buffered; receives heartbeat IDs from Handle
}

func newCryptocomHandler(providerID, userAgent string) *cryptocomHandler {
	return &cryptocomHandler{
		providerID: providerID,
		userAgent:  userAgent,
		heartbeats: make(chan int64, 10),
	}
}

// Handle parses an incoming WebSocket frame and returns a typed model value.
// Returns nil, nil for control messages (heartbeats, confirmations).
func (h *cryptocomHandler) Handle(_ context.Context, raw []byte) (any, error) {
	var msg wsEnvelope
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, nil
	}

	// Heartbeat: queue the ID for OnPing to acknowledge.
	if msg.Method == "public/heartbeat" {
		select {
		case h.heartbeats <- msg.ID:
		default: // drop if buffer full; server will close connection after timeout
		}
		return nil, nil
	}

	// Only "subscribe" method carries data pushes and confirmations.
	if msg.Method != "subscribe" || len(msg.Result) == 0 {
		return nil, nil
	}

	var result wsResult
	if err := json.Unmarshal(msg.Result, &result); err != nil {
		return nil, nil
	}

	switch result.Channel {
	case "ticker":
		return h.handleTicker(result)
	case "book":
		return h.handleBook(result)
	}

	return nil, nil
}

func (h *cryptocomHandler) handleTicker(result wsResult) (any, error) {
	var data []wsTickerData
	if err := json.Unmarshal(result.Data, &data); err != nil || len(data) == 0 {
		return nil, nil
	}
	d := data[0]

	var changePct *float64
	if d.A != 0 { // derive percent change from bid/ask midpoint not available; skip
		// The WebSocket ticker does not carry an open price — use 0-safe guard.
	}
	_ = changePct // explicitly unused; WebSocket ticker omits open price

	return &model.Response[model.CoinPrice]{
		Kind:     model.KindPrice,
		Provider: h.providerID,
		Market:   model.MarketSpot,
		Data: model.CoinPrice{
			ID:     d.I,
			Symbol: d.I,
			Price:  d.C,
		},
	}, nil
}

func (h *cryptocomHandler) handleBook(result wsResult) (any, error) {
	var data []wsBookData
	if err := json.Unmarshal(result.Data, &data); err != nil || len(data) == 0 {
		return nil, nil
	}
	d := data[0]

	parseEntries := func(raw [][]float64) []model.OrderBookEntry {
		entries := make([]model.OrderBookEntry, 0, len(raw))
		for _, e := range raw {
			if len(e) >= 2 {
				entries = append(entries, model.OrderBookEntry{Price: e[0], Quantity: e[1]})
			}
		}
		return entries
	}

	var ts *time.Time
	if d.T > 0 {
		t := time.UnixMilli(d.T)
		ts = &t
	}

	return &model.Response[model.OrderBook]{
		Kind:     model.KindOrderBook,
		Provider: h.providerID,
		Market:   model.MarketSpot,
		Data: model.OrderBook{
			Symbol: result.InstrumentName,
			Market: model.MarketSpot,
			Bids:   parseEntries(d.Bids),
			Asks:   parseEntries(d.Asks),
			Time:   ts,
		},
	}, nil
}

// OnCommand translates a ws.Command into the Crypto.com subscribe/unsubscribe
// wire format.  Params must be []string (channel names).
func (h *cryptocomHandler) OnCommand(_ context.Context, cmd ws.Command, client *ws.BaseClient) error {
	channels, ok := cmd.Params.([]string)
	if !ok {
		return fmt.Errorf("cryptocom: OnCommand expects []string params, got %T", cmd.Params)
	}

	var method string
	switch cmd.Kind {
	case ws.CommandSubscribe:
		method = "subscribe"
	case ws.CommandUnsubscribe:
		method = "unsubscribe"
	default:
		return nil
	}

	return client.WriteJSON(map[string]any{
		"id":     time.Now().UnixMilli(),
		"method": method,
		"params": map[string]any{"channels": channels},
		"nonce":  time.Now().UnixMilli(),
	})
}

// OnPing flushes any buffered heartbeat IDs and sends the required
// public/respond-heartbeat replies.
func (h *cryptocomHandler) OnPing(_ context.Context, client *ws.BaseClient) error {
	for {
		select {
		case id := <-h.heartbeats:
			if err := client.WriteJSON(map[string]any{
				"id":     id,
				"method": "public/respond-heartbeat",
			}); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

// Stream starts the WebSocket manager and returns a channel of raw stream
// responses.  cmdChan is forwarded to the manager's command channel.
func (c *Client) Stream(ctx context.Context, cmdChan <-chan ws.Command) (<-chan ws.StreamResponse[any], error) {
	handler := newCryptocomHandler(providerID, c.userAgent)
	mgr := ws.NewManager(wsMarketURL, handler)

	go func() {
		for cmd := range cmdChan {
			mgr.Commands() <- cmd
		}
	}()

	return mgr.Start(ctx)
}

// WatchPrices implements provider.PriceStreamProvider.
// ids must be Crypto.com instrument names (e.g. "BTC_USDT").
func (c *Client) WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	channels := make([]string, 0, len(ids))
	for _, id := range ids {
		channels = append(channels, "ticker."+id)
	}

	cmdChan := make(chan ws.Command, 1)
	cmdChan <- ws.Command{
		Kind:   ws.CommandSubscribe,
		Method: "ticker",
		Params: channels,
	}

	outChan, err := c.Stream(ctx, cmdChan)
	if err != nil {
		return nil, err
	}

	prices := make(chan *model.CoinPrice, 100)
	go func() {
		defer close(prices)
		for res := range outChan {
			if res.Error != nil || res.Response == nil {
				continue
			}
			if resp, ok := res.Response.(*model.Response[model.CoinPrice]); ok {
				prices <- &resp.Data
			}
		}
	}()
	return prices, nil
}

// WatchOrderBook implements provider.OrderBookStreamProvider.
// depth 0 defaults to 10 levels; max is 150.
func (c *Client) WatchOrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	if depth <= 0 {
		depth = 10
	}

	channel := fmt.Sprintf("book.%s.%d", symbol, depth)
	cmdChan := make(chan ws.Command, 1)
	cmdChan <- ws.Command{
		Kind:   ws.CommandSubscribe,
		Method: "book",
		Params: []string{channel},
	}

	outChan, err := c.Stream(ctx, cmdChan)
	if err != nil {
		return nil, err
	}

	books := make(chan *model.OrderBook, 100)
	go func() {
		defer close(books)
		for res := range outChan {
			if res.Error != nil || res.Response == nil {
				continue
			}
			if resp, ok := res.Response.(*model.Response[model.OrderBook]); ok {
				resp.Data.Market = market
				books <- &resp.Data
			}
		}
	}()
	return books, nil
}
