package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/ws"
)

type depthStreamData struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

type binanceHandler struct {
	providerID string
}

type tickerStreamData struct {
	Symbol      string `json:"s"`
	LastPrice   string `json:"c"`
	PriceChange string `json:"p"`
	ChangePct   string `json:"P"`
}

func (h *binanceHandler) Handle(ctx context.Context, raw []byte) (any, error) {
	var combined struct {
		Stream string          `json:"stream"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &combined); err != nil {
		return nil, nil
	}

	if strings.Contains(combined.Stream, "@ticker") {
		var data tickerStreamData
		if err := json.Unmarshal(combined.Data, &data); err != nil {
			return nil, nil
		}
		price, _ := strconv.ParseFloat(data.LastPrice, 64)
		changePct, _ := strconv.ParseFloat(data.ChangePct, 64)

		return &model.Response[model.CoinPrice]{
			Kind:     model.KindPrice,
			Provider: h.providerID,
			Data: model.CoinPrice{
				ID:        data.Symbol,
				Symbol:    data.Symbol,
				Price:     price,
				Change24h: &changePct,
			},
		}, nil
	}

	if strings.Contains(combined.Stream, "@depth") {
		var data depthStreamData
		if err := json.Unmarshal(combined.Data, &data); err != nil {
			return nil, nil
		}
		parts := strings.SplitN(combined.Stream, "@", 2)
		sym := strings.ToUpper(parts[0])
		uid := data.LastUpdateID

		return &model.Response[model.OrderBook]{
			Kind:     model.KindOrderBook,
			Provider: h.providerID,
			Data: model.OrderBook{
				Symbol:       sym,
				LastUpdateID: &uid,
				Bids:         parseStringPairs(data.Bids),
				Asks:         parseStringPairs(data.Asks),
			},
		}, nil
	}

	return nil, nil
}

func (h *binanceHandler) OnCommand(ctx context.Context, cmd ws.Command, client *ws.BaseClient) error {
	return nil
}

func (h *binanceHandler) OnPing(ctx context.Context, client *ws.BaseClient) error {
	return nil
}

func (c *Client) Stream(ctx context.Context, cmdChan <-chan ws.Command, streams []string) (<-chan ws.StreamResponse[any], error) {
	baseURL := "wss://stream.binance.com:9443/stream"
	if c.cfg.Futures.UseTestnet {
		baseURL = "wss://testnet.binance.vision/stream"
	}
	url := fmt.Sprintf("%s?streams=%s", baseURL, strings.Join(streams, "/"))

	handler := &binanceHandler{providerID: providerID}
	mgr := ws.NewManager(url, handler)

	go func() {
		for cmd := range cmdChan {
			mgr.Commands() <- cmd
		}
	}()

	return mgr.Start(ctx)
}

// WatchPrices implements provider.PriceStreamProvider.
func (c *Client) WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	streams := make([]string, len(ids))
	for i, id := range ids {
		streams[i] = strings.ToLower(id) + "@ticker"
	}

	cmdChan := make(chan ws.Command, 1)
	outChan, err := c.Stream(ctx, cmdChan, streams)
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
	if depth <= 0 {
		depth = 20
	}
	streamName := fmt.Sprintf("%s@depth%d", strings.ToLower(symbol), depth)

	cmdChan := make(chan ws.Command, 1)
	outChan, err := c.Stream(ctx, cmdChan, []string{streamName})
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
