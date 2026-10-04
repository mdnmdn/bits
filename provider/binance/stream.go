package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider"
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
				ID:             data.Symbol,
				Symbol:         model.NormalizeSymbol(data.Symbol),
				OriginalSymbol: data.Symbol,
				Price:          price,
				Change24h:      &changePct,
			},
		}, nil
	}

	if strings.Contains(combined.Stream, "@trade") || strings.Contains(combined.Stream, "@aggTrade") {
		tr, ok := parseTradeEvent(combined.Data)
		if !ok {
			return nil, nil
		}
		return &model.Response[model.Trade]{Kind: model.KindTrade, Provider: h.providerID, Data: tr}, nil
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
				Symbol:         model.NormalizeSymbol(sym),
				OriginalSymbol: sym,
				LastUpdateID:   &uid,
				Bids:           parseStringPairs(data.Bids),
				Asks:           parseStringPairs(data.Asks),
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

// getStreamBaseURL returns the WebSocket base URL for the given market.
func (c *Client) getStreamBaseURL(market model.MarketType) string {
	if market == model.MarketFutures {
		if c.cfg.Futures.UseTestnet {
			return "wss://stream.binancefuture.com/stream"
		}
		return "wss://fstream.binance.com/stream"
	}
	// Spot and margin use the same websocket endpoint
	if c.cfg.Spot.UseTestnet {
		return "wss://testnet.binance.vision/stream"
	}
	return "wss://stream.binance.com:9443/stream"
}

func (c *Client) Stream(ctx context.Context, cmdChan <-chan ws.Command, streams []string, market model.MarketType) (<-chan ws.StreamResponse[any], error) {
	baseURL := c.getStreamBaseURL(market)
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
	outChan, err := c.Stream(ctx, cmdChan, streams, model.MarketSpot)
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
	outChan, err := c.Stream(ctx, cmdChan, []string{streamName}, market)
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

// tradeStreamData is the payload of the spot "trade" and futures "aggTrade"
// events (the fields they share).
type tradeStreamData struct {
	Event      string `json:"e"`
	EventTime  int64  `json:"E"` // declared so "E" does not fold into "e"
	Symbol     string `json:"s"`
	TradeID    int64  `json:"t"` // spot
	AggTradeID int64  `json:"a"` // futures
	Price      string `json:"p"`
	Quantity   string `json:"q"`
	TradeTime  int64  `json:"T"` // ms
	BuyerMaker bool   `json:"m"`
}

// parseTradeEvent converts a trade or aggTrade payload. It reports false for
// a malformed payload.
func parseTradeEvent(raw []byte) (model.Trade, bool) {
	var d tradeStreamData
	if err := json.Unmarshal(raw, &d); err != nil || d.Symbol == "" {
		return model.Trade{}, false
	}
	price, err1 := strconv.ParseFloat(d.Price, 64)
	qty, err2 := strconv.ParseFloat(d.Quantity, 64)
	if err1 != nil || err2 != nil || d.TradeTime <= 0 {
		return model.Trade{}, false
	}
	id := d.TradeID
	if d.Event == "aggTrade" {
		id = d.AggTradeID
	}
	side := model.TradeSideBuy
	if d.BuyerMaker { // the buyer rested, so the taker sold
		side = model.TradeSideSell
	}
	return model.Trade{
		Symbol:         model.NormalizeSymbol(d.Symbol),
		OriginalSymbol: d.Symbol,
		ID:             strconv.FormatInt(id, 10),
		Price:          price,
		Quantity:       qty,
		Side:           side,
		Time:           time.UnixMilli(d.TradeTime).UTC(),
	}, true
}

// WatchTrades streams the public trade tape of symbol: the "trade" stream on
// spot and margin, the "aggTrade" stream on futures (the only trade stream the
// futures host offers).
func (c *Client) WatchTrades(ctx context.Context, symbol string, market model.MarketType) (<-chan *model.Trade, error) {
	name := "@trade"
	if market == model.MarketFutures {
		name = "@aggTrade"
	}
	cmdChan := make(chan ws.Command, 1)
	outChan, err := c.Stream(ctx, cmdChan, []string{strings.ToLower(symbol) + name}, market)
	if err != nil {
		return nil, err
	}

	trades := make(chan *model.Trade, 100)
	go func() {
		defer close(trades)
		for res := range outChan {
			if res.Error != nil || res.Response == nil {
				continue
			}
			if resp, ok := res.Response.(*model.Response[model.Trade]); ok {
				resp.Data.Market = market
				select {
				case trades <- &resp.Data:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return trades, nil
}

// StartTradeStream implements provider.TradeStreamProvider. One symbol per
// stream, like StartOrderBookStream.
func (c *Client) StartTradeStream(ctx context.Context, symbols []string, market model.MarketType) (<-chan *model.Trade, error) {
	if len(symbols) != 1 {
		return nil, fmt.Errorf("binance: trade stream takes exactly one symbol, got %d", len(symbols))
	}
	return c.WatchTrades(ctx, symbols[0], market)
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

func (c *Client) StartPriceStream(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	return nil, nil
}

func (c *Client) SubscribePrice(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	return nil, nil
}

func (c *Client) UnsubscribePrice(ctx context.Context, ids []string) error {
	return nil
}

func (c *Client) SubscribedPrices() []string {
	return nil
}

func (c *Client) StopPriceStream() error {
	return nil
}

func (c *Client) PriceStreamStatus() provider.StreamStatus {
	return provider.StreamStatusStopped
}

func (c *Client) GetLastPrice(id string) (*model.CoinPrice, error) {
	return nil, nil
}

func (c *Client) ReconnectPrice(ctx context.Context) error {
	return nil
}

func (c *Client) GetDataChannelPrice() <-chan *model.CoinPrice {
	return nil
}

func (c *Client) StartOrderBookStream(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	if len(symbols) == 0 {
		return nil, nil
	}
	return c.WatchOrderBook(ctx, symbols[0], market, depth)
}

func (c *Client) SubscribeOrderBook(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	return c.StartOrderBookStream(ctx, symbols, market, depth)
}

func (c *Client) UnsubscribeOrderBook(ctx context.Context, symbols []string) error {
	return nil
}

func (c *Client) SubscribedOrderBooks() []string {
	return nil
}

func (c *Client) StopOrderBookStream() error {
	return nil
}

func (c *Client) OrderBookStreamStatus() provider.StreamStatus {
	return provider.StreamStatusStopped
}

func (c *Client) GetLastOrderBook(symbol string) (*model.OrderBook, error) {
	return nil, nil
}

func (c *Client) ReconnectOrderBook(ctx context.Context) error {
	return nil
}

func (c *Client) GetDataChannelOrderBook() <-chan *model.OrderBook {
	return nil
}
