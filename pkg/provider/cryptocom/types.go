package cryptocom

import "encoding/json"

// apiEnvelope is the common response envelope for all Crypto.com v2 REST API responses.
// code 0 = success; non-zero = error.
type apiEnvelope struct {
	ID     int64           `json:"id"`
	Method string          `json:"method"`
	Code   int             `json:"code"`
	Result json.RawMessage `json:"result"`
}

// apiInstrumentsResult is the result payload for public/get-instruments.
type apiInstrumentsResult struct {
	Instruments []apiInstrument `json:"instruments"`
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
	I  string  `json:"i"`  // instrument name
	V  float64 `json:"v"`  // volume 24h (base currency)
	VV float64 `json:"vv"` // volume value 24h (quote currency)
	L  float64 `json:"l"`  // lowest price 24h
	H  float64 `json:"h"`  // highest price 24h
	O  float64 `json:"o"`  // open price
	C  float64 `json:"c"`  // last / close price
	P  float64 `json:"p"`  // price change (absolute: c - o)
	T  int64   `json:"t"`  // trade count 24h
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
type apiBookRow struct {
	Bids [][]float64 `json:"bids"`
	Asks [][]float64 `json:"asks"`
	T    int64       `json:"t"` // snapshot timestamp (ms); may be absent
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
