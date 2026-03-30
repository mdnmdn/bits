package cryptocom

import (
	"encoding/json"
	"strconv"
)

// apiEnvelope is the common response envelope for all Crypto.com v2 REST API responses.
// code 0 = success; non-zero = error.
// Code can be either int or string depending on the endpoint.
// Error responses include "msg" field.
type apiEnvelope struct {
	ID     int64           `json:"id"`
	Method string          `json:"method"`
	Code   json.RawMessage `json:"code"`
	Msg    string          `json:"msg"`
	Result json.RawMessage `json:"result"`
}

func (e *apiEnvelope) GetCode() (int, error) {
	if len(e.Code) == 0 {
		return 0, nil
	}
	var codeInt int
	if err := json.Unmarshal(e.Code, &codeInt); err != nil {
		var codeStr string
		if err := json.Unmarshal(e.Code, &codeStr); err != nil {
			return 0, err
		}
		return strconv.Atoi(codeStr)
	}
	return codeInt, nil
}

// apiInstrumentsResult is the result payload for public/get-instruments.
type apiInstrumentsResult struct {
	Instruments []apiInstrument `json:"instruments"`
}

// apiInstrumentsV1Response is the response for v1 public/get-instruments.
type apiInstrumentsV1Response struct {
	ID     int64                `json:"id"`
	Method string               `json:"method"`
	Code   int                  `json:"code"`
	Result apiInstrumentsV1Data `json:"result"`
}

// apiInstrumentsV1Data contains the instruments data.
type apiInstrumentsV1Data struct {
	Data []apiInstrumentV1 `json:"data"`
}

// apiInstrumentV1 represents a single trading instrument in v1 API.
type apiInstrumentV1 struct {
	Symbol            string `json:"symbol"`
	InstType          string `json:"inst_type"`
	BaseCcy           string `json:"base_ccy"`
	QuoteCcy          string `json:"quote_ccy"`
	QuoteDecimals     int    `json:"quote_decimals"`
	QuantityDecimals  int    `json:"quantity_decimals"`
	PriceTickSize     string `json:"price_tick_size"`
	QtyTickSize       string `json:"qty_tick_size"`
	Tradable          bool   `json:"tradable"`
	MarginBuyEnabled  bool   `json:"margin_buy_enabled"`
	MarginSellEnabled bool   `json:"margin_sell_enabled"`
}

// apiInstrument represents a single trading instrument.
type apiInstrument struct {
	InstrumentName    string `json:"instrument_name"`
	ProductType       string `json:"product_type"`
	QuoteCurrency     string `json:"quote_currency"`
	BaseCurrency      string `json:"base_currency"`
	MinOrderSize      string `json:"min_order_size"`
	MaxOrderSize      string `json:"max_order_size"`
	PricePrecision    int    `json:"price_precision"`
	QuantityPrecision int    `json:"quantity_precision"`
}

// apiTickerResult is the result payload for public/get-ticker.
type apiTickerResult struct {
	Data []apiTickerData `json:"data"`
}

// apiTickerData holds the 24h rolling statistics for a single instrument.
// Field names follow the Crypto.com v2 API single-letter convention.
type apiTickerData struct {
	I  string `json:"i"`  // instrument name
	A  string `json:"a"`  // last price (ask)
	B  string `json:"b"`  // best bid price
	V  string `json:"v"`  // volume 24h (base currency)
	VV string `json:"vv"` // volume value 24h (quote currency)
	L  string `json:"l"`  // lowest price 24h
	H  string `json:"h"`  // highest price 24h
	O  string `json:"o"`  // open price
	C  string `json:"c"`  // price change (ratio, e.g., 0.0163 = 1.63%)
	P  string `json:"p"`  // price change (absolute: a - o)
	T  int64  `json:"t"`  // timestamp (ms)
}

// apiBookResult is the result payload for public/get-book.
// The API wraps the actual depth data in a one-element "data" slice.
type apiBookResult struct {
	InstrumentName string       `json:"instrument_name"`
	Depth          int          `json:"depth"`
	Data           []apiBookRow `json:"data"`
}

// apiBookRow holds one snapshot of bids and asks.
// Each entry is [price, quantity, num_orders] where num_orders may be absent.
// Price and quantity are returned as strings.
type apiBookRow struct {
	Bids [][]string `json:"bids"`
	Asks [][]string `json:"asks"`
	T    int64      `json:"t"` // snapshot timestamp (ms); may be absent
}

// ── WebSocket types ───────────────────────────────────────────────────────────

// wsEnvelope is the common envelope for all Crypto.com v2 WebSocket messages.
type wsEnvelope struct {
	ID     int64           `json:"id"`
	Method string          `json:"method"`
	Code   int             `json:"code"`
	Result json.RawMessage `json:"result"`
}

// wsResult is the result payload for subscription data messages.
// The same envelope is used for both subscription confirmations and data pushes.
type wsResult struct {
	Subscription   string          `json:"subscription"`
	Channel        string          `json:"channel"`
	InstrumentName string          `json:"instrument_name"`
	Data           json.RawMessage `json:"data"`
}

// wsTickerData is a single ticker update from the WebSocket "ticker" channel.
// IMPORTANT: unlike the REST ticker where t = trade count, here t = timestamp (ms).
type wsTickerData struct {
	I string  `json:"i"` // instrument name
	B float64 `json:"b"` // best bid price
	K float64 `json:"k"` // best ask price
	A float64 `json:"a"` // latest trade price
	T int64   `json:"t"` // timestamp (ms)
	V float64 `json:"v"` // volume 24h (base)
	H float64 `json:"h"` // high 24h
	L float64 `json:"l"` // low 24h
	C float64 `json:"c"` // last / close price
}

// wsBookData is a single order book snapshot from the WebSocket "book" channel.
// Each bid/ask entry is [price, qty, num_orders]; num_orders is optional.
type wsBookData struct {
	Bids [][]float64 `json:"bids"`
	Asks [][]float64 `json:"asks"`
	T    int64       `json:"t"` // snapshot timestamp (ms)
}

type apiCandlestickResult struct {
	Interval string            `json:"interval"`
	Data     []candlestickData `json:"data"`
}

type candlestickData struct {
	O string `json:"o"` // open
	H string `json:"h"` // high
	L string `json:"l"` // low
	C string `json:"c"` // close
	V string `json:"v"` // volume
	T int64  `json:"t"` // timestamp (ms)
}
