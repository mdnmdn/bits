package whitebit

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/pkg/model"
)

const wsURL = "wss://api.whitebit.com/ws"

type whitebitSubParams struct {
	Method string
	Args   []any
}

type whitebitProtocol struct {
	providerID string
	requestID  int
}

type whitebitJSONRPCMsg struct {
	ID     *int64          `json:"id"`
	Result any             `json:"result"`
	Error  any             `json:"error"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

func (p *whitebitProtocol) Dial(ctx context.Context, conn *ws.Conn) error {
	return nil
}

func (p *whitebitProtocol) Ping(ctx context.Context, conn *ws.Conn) error {
	p.requestID++
	return conn.WriteJSON(map[string]any{
		"id":     p.requestID,
		"method": "ping",
		"params": []any{},
	})
}

func (p *whitebitProtocol) Subscribe(ctx context.Context, conn *ws.Conn, sub ws.Subscription) error {
	subParams, ok := sub.Params.(whitebitSubParams)
	if !ok {
		return nil
	}

	p.requestID++
	return conn.WriteJSON(map[string]any{
		"id":     p.requestID,
		"method": subParams.Method,
		"params": subParams.Args,
	})
}

func (p *whitebitProtocol) Unsubscribe(ctx context.Context, conn *ws.Conn, sub ws.Subscription) error {
	subParams, ok := sub.Params.(whitebitSubParams)
	if !ok {
		return nil
	}

	p.requestID++
	return conn.WriteJSON(map[string]any{
		"id":     p.requestID,
		"method": subParams.Method + "_unsubscribe",
		"params": subParams.Args,
	})
}

func (p *whitebitProtocol) Parse(ctx context.Context, raw []byte) (any, error) {
	var msg whitebitJSONRPCMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, nil
	}

	if msg.ID != nil {
		return nil, nil
	}

	switch msg.Method {
	case "lastprice_update":
		return p.parseLastPriceUpdate(msg.Params)
	case "market_update":
		return p.parseMarketUpdate(msg.Params)
	case "depth_update":
		return p.parseDepthUpdate(msg.Params)
	case "bookTicker_update":
		return p.parseBookTickerUpdate(msg.Params)
	}

	return nil, nil
}

// parseLastPriceUpdate handles "lastprice_update" messages.
// Provides: symbol, price (basic last price only)
func (p *whitebitProtocol) parseLastPriceUpdate(params json.RawMessage) (*model.Response[model.CoinPrice], error) {
	var data []any
	if err := json.Unmarshal(params, &data); err != nil {
		return nil, nil
	}

	if len(data) < 2 {
		return nil, nil
	}

	symbol, ok := data[0].(string)
	if !ok {
		return nil, nil
	}

	priceStr, ok := data[1].(string)
	if !ok {
		return nil, nil
	}

	price, _ := strconv.ParseFloat(priceStr, 64)

	return &model.Response[model.CoinPrice]{
		Kind:     model.KindPrice,
		Provider: p.providerID,
		Data: model.CoinPrice{
			ID:       symbol,
			Symbol:   symbol,
			Price:    price,
			Currency: "USDT",
		},
	}, nil
}

// parseMarketUpdate handles "market_update" messages.
// Provides: symbol, price, volume, open, high, low, change24h
// This is used for the dual-subscription price stream to get full market stats.
func (p *whitebitProtocol) parseMarketUpdate(params json.RawMessage) (*model.Response[model.CoinPrice], error) {
	var data []any
	if err := json.Unmarshal(params, &data); err != nil {
		return nil, nil
	}

	if len(data) < 2 {
		return nil, nil
	}

	symbol, ok := data[0].(string)
	if !ok {
		return nil, nil
	}

	statsObj, ok := data[1].(map[string]any)
	if !ok {
		return nil, nil
	}

	price, _ := strconv.ParseFloat(getString(statsObj, "last"), 64)
	open, _ := strconv.ParseFloat(getString(statsObj, "open"), 64)
	high, _ := strconv.ParseFloat(getString(statsObj, "high"), 64)
	low, _ := strconv.ParseFloat(getString(statsObj, "low"), 64)
	volume, _ := strconv.ParseFloat(getString(statsObj, "volume"), 64)

	var changePct *float64
	if open > 0 {
		chg := (price - open) / open * 100
		changePct = &chg
	}

	return &model.Response[model.CoinPrice]{
		Kind:     model.KindPrice,
		Provider: p.providerID,
		Data: model.CoinPrice{
			ID:        symbol,
			Symbol:    symbol,
			Price:     price,
			Change24h: changePct,
			Open24h:   &open,
			High24h:   &high,
			Low24h:    &low,
			Volume24h: &volume,
			Currency:  "USDT",
			Time:      func() *time.Time { t := time.Now(); return &t }(),
		},
	}, nil
}

// parseDepthUpdate handles "depth_update" messages for order book streaming.
// Provides: symbol, bids, asks, timestamp, update_id (for delta detection)
func (p *whitebitProtocol) parseDepthUpdate(params json.RawMessage) (*model.Response[model.OrderBook], error) {
	var data []any
	if err := json.Unmarshal(params, &data); err != nil {
		return nil, nil
	}

	if len(data) < 3 {
		return nil, nil
	}

	symbol, ok := data[2].(string)
	if !ok {
		return nil, nil
	}

	depthObj, ok := data[1].(map[string]any)
	if !ok {
		return nil, nil
	}

	parseEntries := func(raw []any) []model.OrderBookEntry {
		entries := make([]model.OrderBookEntry, 0)
		for _, e := range raw {
			arr, ok := e.([]any)
			if !ok || len(arr) < 2 {
				continue
			}
			priceStr, _ := arr[0].(string)
			qtyStr, _ := arr[1].(string)
			price, _ := strconv.ParseFloat(priceStr, 64)
			qty, _ := strconv.ParseFloat(qtyStr, 64)
			if qty > 0 {
				entries = append(entries, model.OrderBookEntry{Price: price, Quantity: qty})
			}
		}
		return entries
	}

	bids := parseEntries(getArray(depthObj, "bids"))
	asks := parseEntries(getArray(depthObj, "asks"))

	var ts *time.Time
	tsFloat, ok := depthObj["timestamp"].(float64)
	if ok {
		t := time.Unix(int64(tsFloat), 0)
		ts = &t
	}

	// Determine if this is a snapshot or incremental update
	// Snapshot: first message (no past_update_id)
	// Incremental: subsequent messages (has past_update_id)
	var lastUpdateID *int64
	if updateID, ok := depthObj["update_id"].(float64); ok {
		id := int64(updateID)
		lastUpdateID = &id
	}

	return &model.Response[model.OrderBook]{
		Kind:     model.KindOrderBook,
		Provider: p.providerID,
		Data: model.OrderBook{
			Symbol:       symbol,
			Market:       model.MarketSpot,
			Bids:         bids,
			Asks:         asks,
			Time:         ts,
			LastUpdateID: lastUpdateID,
		},
	}, nil
}

// parseBookTickerUpdate handles "bookTicker_update" messages.
// Provides: symbol, bid_price, bid_size, ask_price, ask_size
// This is used for the dual-subscription price stream to get bid/ask data.
func (p *whitebitProtocol) parseBookTickerUpdate(params json.RawMessage) (*model.Response[model.CoinPrice], error) {
	var data []any
	if err := json.Unmarshal(params, &data); err != nil {
		return nil, nil
	}

	if len(data) == 0 {
		return nil, nil
	}

	// WhiteBit returns: data[0] is the ticker array directly (length 8)
	// Format: [transaction_time, message_time, market, update_id, bid_price, bid_size, ask_price, ask_size]
	var ticker []any
	var ok bool

	// Try: data[0] is the ticker array directly (length 8)
	if len(data) >= 1 {
		if arr, isArr := data[0].([]any); isArr && len(arr) >= 8 {
			ticker = arr
			ok = true
		}
	}

	if !ok {
		return nil, nil
	}

	symbol, ok := ticker[2].(string)
	if !ok {
		return nil, nil
	}

	bidPriceStr, ok := ticker[4].(string)
	if !ok {
		return nil, nil
	}
	bidSizeStr, ok := ticker[5].(string)
	if !ok {
		return nil, nil
	}
	askPriceStr, ok := ticker[6].(string)
	if !ok {
		return nil, nil
	}
	askSizeStr, ok := ticker[7].(string)
	if !ok {
		return nil, nil
	}

	bidPrice, _ := strconv.ParseFloat(bidPriceStr, 64)
	bidSize, _ := strconv.ParseFloat(bidSizeStr, 64)
	askPrice, _ := strconv.ParseFloat(askPriceStr, 64)
	askSize, _ := strconv.ParseFloat(askSizeStr, 64)

	resp := &model.Response[model.CoinPrice]{
		Kind:     model.KindPrice,
		Provider: p.providerID,
		Data: model.CoinPrice{
			ID:       symbol,
			Symbol:   symbol,
			BidPrice: &bidPrice,
			BidSize:  &bidSize,
			AskPrice: &askPrice,
			AskSize:  &askSize,
			Currency: "USDT",
		},
	}
	return resp, nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getArray(m map[string]any, key string) []any {
	if v, ok := m[key].([]any); ok {
		return v
	}
	return nil
}
