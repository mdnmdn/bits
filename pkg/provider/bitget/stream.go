package bitget

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/internal/logger"
	"github.com/mdnmdn/bits/pkg/model"
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

type bitgetHandler struct {
	providerID string
	userAgent  string
}

func (h *bitgetHandler) Handle(ctx context.Context, raw []byte) (any, error) {
	var msg bitgetWSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, nil
	}
	if msg.Action != "snapshot" && msg.Action != "update" {
		return nil, nil
	}

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
			Provider: h.providerID,
			Data: model.CoinPrice{
				ID:        d.InstID,
				Symbol:    d.InstID,
				Price:     price,
				Change24h: &changePct,
			},
		}, nil
	}

	if msg.Arg.Channel == "books" || msg.Arg.Channel == "books5" || msg.Arg.Channel == "books15" || msg.Arg.Channel == "book1" {
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
			Provider: h.providerID,
			Data: model.OrderBook{
				Symbol: msg.Arg.InstID,
				Bids:   parseEntries(d.Bids),
				Asks:   parseEntries(d.Asks),
				Time:   ts,
			},
		}, nil
	}

	return nil, nil
}

func (h *bitgetHandler) OnCommand(ctx context.Context, cmd ws.Command, client *ws.BaseClient) error {
	logger.Default.Debug("bitget: sending command", "kind", cmd.Kind, "params", cmd.Params)
	switch cmd.Kind {
	case ws.CommandSubscribe:
		args, ok := cmd.Params.([]map[string]string)
		if !ok {
			return fmt.Errorf("invalid subscribe params")
		}
		return client.WriteJSON(map[string]any{
			"op":   "subscribe",
			"args": args,
		})
	case ws.CommandUnsubscribe:
		args, ok := cmd.Params.([]map[string]string)
		if !ok {
			return fmt.Errorf("invalid unsubscribe params")
		}
		return client.WriteJSON(map[string]any{
			"op":   "unsubscribe",
			"args": args,
		})
	}
	return nil
}

func (h *bitgetHandler) OnPing(ctx context.Context, client *ws.BaseClient) error {
	return client.WriteJSON("ping")
}

func (c *Client) Stream(ctx context.Context, cmdChan <-chan ws.Command) (<-chan ws.StreamResponse[any], error) {
	handler := &bitgetHandler{
		providerID: providerID,
		userAgent:  c.userAgent,
	}
	mgr := ws.NewManager(wsURL, handler)

	go func() {
		for cmd := range cmdChan {
			mgr.Commands() <- cmd
		}
	}()

	return mgr.Start(ctx)
}

// WatchPrices implements provider.PriceStreamProvider.
func (c *Client) WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	cmdChan := make(chan ws.Command, 1)
	args := make([]map[string]string, 0, len(ids))
	for _, id := range ids {
		args = append(args, map[string]string{
			"instType": "SPOT",
			"channel":  "ticker",
			"instId":   id,
		})
	}
	cmdChan <- ws.Command{Kind: ws.CommandSubscribe, Params: args}

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
func (c *Client) WatchOrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	cmdChan := make(chan ws.Command, 1)
	instType := "SPOT"
	if market == model.MarketFutures {
		instType = "USDT-FUTURES"
	}

	channel := "books5"
	if depth > 0 {
		if depth <= 5 {
			channel = "books5"
		} else if depth <= 15 {
			channel = "books15"
		} else {
			channel = "books"
		}
	}

	args := []map[string]string{
		{
			"instType": instType,
			"channel":  channel,
			"instId":   symbol,
		},
	}
	cmdChan <- ws.Command{Kind: ws.CommandSubscribe, Params: args}

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
