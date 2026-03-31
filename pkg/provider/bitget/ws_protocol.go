package bitget

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/pkg/model"
)

// bitgetProtocol implements ws.Protocol for Bitget public market streams.
type bitgetProtocol struct {
	providerID string
}

// bitgetWSMessage is the common envelope for all Bitget WS messages.
type bitgetWSMessage struct {
	Action string          `json:"action"`
	Arg    bitgetWSArg     `json:"arg"`
	Data   json.RawMessage `json:"data"`
	Ts     int64           `json:"ts"`
	// Error response fields
	Event string `json:"event"`
	Code  string `json:"code"`
	Msg   string `json:"msg"`
	Op    string `json:"op"`
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

func (p *bitgetProtocol) Dial(ctx context.Context, conn *ws.Conn) error {
	// No authentication required for public feeds
	return nil
}

func (p *bitgetProtocol) Ping(ctx context.Context, conn *ws.Conn) error {
	// Bitget spot expects plain string "ping" (not JSON-encoded)
	return conn.WriteMessage(websocket.TextMessage, []byte("ping"))
}

func (p *bitgetProtocol) Subscribe(ctx context.Context, conn *ws.Conn, sub ws.Subscription) error {
	args, ok := sub.Params.([]map[string]string)
	if !ok {
		return &model.ProviderError{
			Kind:            model.ErrKindInvalidRequest,
			ProviderID:      p.providerID,
			ProviderMessage: "invalid subscribe params type",
		}
	}

	return conn.WriteJSON(map[string]any{
		"op":   "subscribe",
		"args": args,
	})
}

func (p *bitgetProtocol) Unsubscribe(ctx context.Context, conn *ws.Conn, sub ws.Subscription) error {
	args, ok := sub.Params.([]map[string]string)
	if !ok {
		// best-effort: silently ignore type errors
		return nil
	}

	return conn.WriteJSON(map[string]any{
		"op":   "unsubscribe",
		"args": args,
	})
}

func (p *bitgetProtocol) Parse(ctx context.Context, raw []byte) (any, error) {
	// Bitget server responds to "ping" keepalive with plain string "pong"
	if string(raw) == "pong" {
		return nil, nil
	}

	var msg bitgetWSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		// Silently ignore parse errors (could be pong or other non-JSON)
		return nil, nil
	}

	// Handle error responses
	if msg.Event == "error" {
		return nil, &model.ProviderError{
			Kind:            model.ErrKindInvalidRequest,
			ProviderID:      p.providerID,
			ProviderCode:    msg.Code,
			ProviderMessage: msg.Msg,
		}
	}

	// Ignore non-snapshot/update messages
	if msg.Action != "snapshot" && msg.Action != "update" {
		return nil, nil
	}

	// Handle ticker channel
	if msg.Arg.Channel == "ticker" {
		var data []bitgetWSTickerData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return nil, nil
		}
		if len(data) == 0 {
			return nil, nil
		}

		d := data[0]
		price, _ := strconv.ParseFloat(d.LastPr, 64)
		changePct, _ := strconv.ParseFloat(d.Change24h, 64)
		changePct *= 100

		return &model.Response[model.CoinPrice]{
			Kind:     model.KindPrice,
			Provider: p.providerID,
			Data: model.CoinPrice{
				ID:        d.InstID,
				Symbol:    d.InstID,
				Price:     price,
				Change24h: &changePct,
			},
		}, nil
	}

	// Handle order book channels: books, books5, books15
	if msg.Arg.Channel == "books" || msg.Arg.Channel == "books5" ||
		msg.Arg.Channel == "books15" {
		var data []bitgetWSDepthData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return nil, nil
		}
		if len(data) == 0 {
			return nil, nil
		}

		d := data[0]

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

		var ts *time.Time
		if d.Ts != "" {
			if ms, err := strconv.ParseInt(d.Ts, 10, 64); err == nil {
				t := time.UnixMilli(ms)
				ts = &t
			}
		}

		return &model.Response[model.OrderBook]{
			Kind:     model.KindOrderBook,
			Provider: p.providerID,
			Data: model.OrderBook{
				Symbol: msg.Arg.InstID,
				Bids:   parseEntries(d.Bids),
				Asks:   parseEntries(d.Asks),
				Time:   ts,
			},
		}, nil
	}

	// Unknown channel, ignore
	return nil, nil
}
