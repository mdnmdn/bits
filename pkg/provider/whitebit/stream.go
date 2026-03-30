package whitebit

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/pkg/model"
)

// whitebitWSRequest represents a WhiteBit WebSocket request.
type whitebitWSRequest struct {
	ID     int64  `json:"id"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

// whitebitWSResponse represents a WhiteBit WebSocket response.
type whitebitWSResponse struct {
	ID     int64            `json:"id"`
	Result json.RawMessage  `json:"result"`
	Error  *whitebitWSError `json:"error"`
	Method string           `json:"method"`
	Params json.RawMessage  `json:"params"`
}

type whitebitWSError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type whitebitWSTickerData struct {
	Open   string `json:"open"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Last   string `json:"last"`
	Volume string `json:"volume"`
	Deal   string `json:"deal"`
	Change string `json:"change"`
}

type whitebitWSDepthData struct {
	Asks [][]string `json:"asks"`
	Bids [][]string `json:"bids"`
}

type whitebitHandler struct {
	providerID string
	userAgent  string
}

func (h *whitebitHandler) Handle(ctx context.Context, raw []byte) (any, error) {
	var resp whitebitWSResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, nil
	}

	if resp.Method == "lastprice_update" {
		var params []json.RawMessage
		if err := json.Unmarshal(resp.Params, &params); err != nil || len(params) < 2 {
			return nil, nil
		}

		var symbol string
		if err := json.Unmarshal(params[0], &symbol); err != nil {
			return nil, nil
		}

		var priceStr string
		if err := json.Unmarshal(params[1], &priceStr); err != nil {
			return nil, nil
		}

		price, _ := strconv.ParseFloat(priceStr, 64)

		return &model.Response[model.CoinPrice]{
			Kind:     model.KindPrice,
			Provider: h.providerID,
			Data: model.CoinPrice{
				ID:     symbol,
				Symbol: symbol,
				Price:  price,
			},
		}, nil
	}

	if resp.Method == "depth_update" {
		var params []json.RawMessage
		if err := json.Unmarshal(resp.Params, &params); err != nil || len(params) < 3 {
			return nil, nil
		}

		var data whitebitWSDepthData
		if err := json.Unmarshal(params[1], &data); err != nil {
			return nil, nil
		}

		var symbol string
		if err := json.Unmarshal(params[2], &symbol); err != nil {
			return nil, nil
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

		return &model.Response[model.OrderBook]{
			Kind:     model.KindOrderBook,
			Provider: h.providerID,
			Data: model.OrderBook{
				Symbol: symbol,
				Bids:   parseEntries(data.Bids),
				Asks:   parseEntries(data.Asks),
			},
		}, nil
	}

	return nil, nil
}

func (h *whitebitHandler) OnCommand(ctx context.Context, cmd ws.Command, client *ws.BaseClient) error {
	switch cmd.Kind {
	case ws.CommandSubscribe:
		params, ok := cmd.Params.([]any)
		if !ok {
			return fmt.Errorf("invalid subscribe params")
		}
		method := "lastprice_subscribe"
		if cmd.Method == "depth" {
			method = "depth_subscribe"
			// Add WhiteBit depth_subscribe specific params: [market, limit, interval, snapshot]
			params = []any{params[0], params[1], "0", true}
		}

		return client.WriteJSON(whitebitWSRequest{
			ID:     time.Now().UnixNano(),
			Method: method,
			Params: params,
		})
	case ws.CommandUnsubscribe:
		return nil
	}
	return nil
}

func (h *whitebitHandler) OnPing(ctx context.Context, client *ws.BaseClient) error {
	return client.WriteJSON(whitebitWSRequest{
		ID:     0,
		Method: "ping",
		Params: []any{},
	})
}

// WatchPrices implements provider.PriceStreamProvider using the unified Stream interface.
func (c *Client) WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	cmdChan := make(chan ws.Command, 1)
	params := make([]any, len(ids))
	for i, id := range ids {
		params[i] = id
	}
	cmdChan <- ws.Command{Kind: ws.CommandSubscribe, Method: "ticker", Params: params}

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

// WatchOrderBook implements provider.OrderBookStreamProvider using the unified Stream interface.
func (c *Client) WatchOrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	cmdChan := make(chan ws.Command, 1)
	if market == model.MarketFutures {
		symbol = translateFuturesSymbol(symbol)
	}
	cmdChan <- ws.Command{Kind: ws.CommandSubscribe, Method: "depth", Params: []any{symbol, depth}}

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
				books <- &resp.Data
			}
		}
	}()
	return books, nil
}

// Stream implements a unified streaming interface for WhiteBit.
func (c *Client) Stream(ctx context.Context, cmdChan <-chan ws.Command) (<-chan ws.StreamResponse[any], error) {
	handler := &whitebitHandler{
		providerID: providerID,
		userAgent:  c.userAgent,
	}
	mgr := ws.NewManager("wss://api.whitebit.com/ws", handler)

	go func() {
		for cmd := range cmdChan {
			mgr.Commands() <- cmd
		}
	}()

	return mgr.Start(ctx)
}
