package whitebit

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/ws"
)

const wsURL = "wss://ws.whitebit.com"

// whitebitWSRequest represents a WhiteBit WebSocket request.
type whitebitWSRequest struct {
	ID     int64  `json:"id"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

// whitebitWSResponse represents a WhiteBit WebSocket response.
type whitebitWSResponse struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *whitebitWSError `json:"error"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
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

// WatchPrices implements provider.PriceStreamProvider.
func (c *Client) WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	updates := make(chan *model.CoinPrice, 100)
	base := ws.NewBaseClient(wsURL)
	base.UserAgent = c.userAgent

	base.OnConnect = func(ctx context.Context, conn *websocket.Conn) error {
		params := make([]any, 0, len(ids))
		for _, id := range ids {
			params = append(params, id)
		}
		return base.WriteJSON(whitebitWSRequest{
			ID:     time.Now().UnixNano(),
			Method: "ticker_subscribe",
			Params: params,
		})
	}

	base.OnMessage = func(ctx context.Context, raw []byte) error {
		var resp whitebitWSResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil
		}

		if resp.Method == "ticker_update" {
			var params []json.RawMessage
			if err := json.Unmarshal(resp.Params, &params); err != nil || len(params) < 2 {
				return nil
			}

			var symbol string
			if err := json.Unmarshal(params[0], &symbol); err != nil {
				return nil
			}

			var data whitebitWSTickerData
			if err := json.Unmarshal(params[1], &data); err != nil {
				return nil
			}

			price, _ := strconv.ParseFloat(data.Last, 64)
			changePct, _ := strconv.ParseFloat(data.Change, 64)

			select {
			case updates <- &model.CoinPrice{
				ID:        symbol,
				Symbol:    symbol,
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

	if depth <= 0 {
		depth = 20
	}

	base.OnConnect = func(ctx context.Context, conn *websocket.Conn) error {
		return base.WriteJSON(whitebitWSRequest{
			ID:     time.Now().UnixNano(),
			Method: "depth_subscribe",
			Params: []any{symbol, depth, "0", true},
		})
	}

	base.OnMessage = func(ctx context.Context, raw []byte) error {
		var resp whitebitWSResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil
		}

		if resp.Method == "depth_update" {
			var params []json.RawMessage
			if err := json.Unmarshal(resp.Params, &params); err != nil || len(params) < 3 {
				return nil
			}

			// params[0] is bool (snapshot), params[1] is data, params[2] is symbol
			var data whitebitWSDepthData
			if err := json.Unmarshal(params[1], &data); err != nil {
				return nil
			}

			var sym string
			if err := json.Unmarshal(params[2], &sym); err != nil {
				sym = symbol
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

			ob := &model.OrderBook{
				Symbol: sym,
				Market: market,
				Bids:   parseEntries(data.Bids),
				Asks:   parseEntries(data.Asks),
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
