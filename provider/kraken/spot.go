package kraken

import (
	"context"
	"encoding/json"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/model"
)

// Spot and margin share the spot REST API (https://api.kraken.com/0/public).
// Margin has no endpoints of its own: it is the spot pairs whose AssetPairs
// entry lists margin leverage (leverage_buy), so every margin call reads spot
// data and ExchangeInfo(margin) only filters the pair list.

const (
	// spotOHLCMax is how many candles OHLC returns: the most recent 720 of the
	// interval, whatever "since" says. There is no way to page further back.
	spotOHLCMax = 720
	// spotDepthMax is the largest Depth count.
	spotDepthMax = 500
)

func isSpotMarket(m model.MarketType) bool {
	return m == model.MarketSpot || m == model.MarketMargin
}

// assetNames maps legacy Kraken asset codes to the common ticker.
var assetNames = map[string]string{
	"XBT": "BTC", "XXBT": "BTC", "XDG": "DOGE", "XXDG": "DOGE",
	"XETH": "ETH", "XLTC": "LTC", "XXRP": "XRP", "XXLM": "XLM", "XETC": "ETC",
	"XMLN": "MLN", "XREP": "REP", "XZEC": "ZEC", "XXMR": "XMR", "XXTZ": "XTZ",
	"ZUSD": "USD", "ZEUR": "EUR", "ZGBP": "GBP", "ZJPY": "JPY", "ZCAD": "CAD", "ZAUD": "AUD",
}

// normAsset converts a Kraken asset code (XXBT, ZUSD, XBT) to BTC, USD, ...
func normAsset(a string) string {
	a = strings.ToUpper(a)
	if v, ok := assetNames[a]; ok {
		return v
	}
	return a
}

// nativeAsset converts a common ticker to the form used in pair altnames.
func nativeAsset(a string) string {
	switch a {
	case "BTC":
		return "XBT"
	case "DOGE":
		return "XDG"
	}
	return a
}

var spotQuotes = []string{"USDT", "USDC", "USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF", "XBT", "BTC", "ETH", "DAI"}

// splitSpot splits a symbol into normalised base and quote (BTC, USD). BASE-QUOTE,
// BASE_QUOTE and BASE/QUOTE split on the separator; XBTUSD or BTCUSD split on
// a known quote suffix. Unknown forms return the whole symbol as base.
func splitSpot(symbol string) (base, quote string) {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	if i := strings.IndexAny(s, "-_/:"); i > 0 {
		return normAsset(s[:i]), normAsset(s[i+1:])
	}
	for _, q := range spotQuotes {
		if b, ok := strings.CutSuffix(s, q); ok && len(b) >= 2 {
			return normAsset(b), normAsset(q)
		}
	}
	return normAsset(s), ""
}

// toSpot converts a symbol to the spot pair form accepted by the public API
// (the altname, XBTUSD). BTC-USD, BTC/USD, BTC_USD, XBT/USD, XBTUSD and BTCUSD
// all give XBTUSD; an unseparated BASEQUOTE only has a leading BTC or DOGE
// rewritten (BTC is XBT, DOGE is XDG at Kraken). Kraken's own long keys
// (XXBTZUSD) are passed through.
func toSpot(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	if i := strings.IndexAny(s, "-_/:"); i > 0 {
		return nativeAsset(s[:i]) + nativeAsset(s[i+1:])
	}
	for _, p := range [][2]string{{"BTC", "XBT"}, {"DOGE", "XDG"}} {
		if strings.HasPrefix(s, p[0]) {
			return p[1] + s[len(p[0]):]
		}
	}
	return s
}

func spotNormalized(symbol string) string {
	b, q := splitSpot(symbol)
	if q == "" {
		return b
	}
	return b + "-" + q
}

// spotIntervals maps interval strings to OHLC minutes. Kraken has no 12h.
var spotIntervals = map[string]int{
	"1m": 1, "5m": 5, "15m": 15, "30m": 30, "1h": 60, "4h": 240, "1d": 1440, "1w": 10080, "15d": 21600,
}

func spotErr(codes []string) *model.ProviderError {
	code := codes[0]
	kind := model.ErrKindUnknown
	switch {
	case strings.HasPrefix(code, "EAPI:Rate limit"), strings.HasPrefix(code, "EGeneral:Too many"):
		kind = model.ErrKindRateLimit
	case strings.HasPrefix(code, "EQuery:Unknown asset pair"):
		kind = model.ErrKindNotFound
	case strings.HasPrefix(code, "EQuery"), strings.HasPrefix(code, "EGeneral:Invalid arguments"):
		kind = model.ErrKindInvalidRequest
	case strings.HasPrefix(code, "EService"), strings.HasPrefix(code, "EGeneral:Internal"):
		kind = model.ErrKindServerError
	}
	return &model.ProviderError{Kind: kind, ProviderID: providerID, ProviderCode: code, ProviderMessage: strings.Join(codes, "; ")}
}

// spotGet performs a GET on /0/public/<endpoint>?<query> and returns the
// result object of the {error, result} envelope.
func (c *Client) spotGet(ctx context.Context, endpoint string, q url.Values) (map[string]json.RawMessage, error) {
	path := "/0/public/" + endpoint
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	body, err := c.fetch(ctx, c.cfg.SpotBaseURL, path)
	if err != nil {
		return nil, err
	}
	var env struct {
		Error  []string                   `json:"error"`
		Result map[string]json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, providerErr(model.ErrKindParse, "parse response: "+err.Error(), err)
	}
	if len(env.Error) > 0 {
		return nil, spotErr(env.Error)
	}
	return env.Result, nil
}

// spotPair queries an endpoint for one pair and returns that pair's payload
// (the result is keyed by Kraken's long pair name, XXBTZUSD, not by the
// request). Result keys other than "last" count as pairs.
func (c *Client) spotPair(ctx context.Context, endpoint, symbol string, extra url.Values) (json.RawMessage, error) {
	q := url.Values{"pair": {toSpot(symbol)}}
	for k, v := range extra {
		q[k] = v
	}
	res, err := c.spotGet(ctx, endpoint, q)
	if err != nil {
		return nil, err
	}
	for k, v := range res {
		if k != "last" {
			return v, nil
		}
	}
	return nil, providerErr(model.ErrKindNotFound, "no "+endpoint+" data for "+symbol, nil)
}

func firstNum(v []string) float64 {
	if len(v) == 0 {
		return 0
	}
	return num(v[0])
}

func idx(v []string, i int) float64 {
	if i < len(v) {
		return num(v[i])
	}
	return 0
}

// spotExchangeInfo lists the spot pairs from AssetPairs. Per pair: tick size
// (Extra["tick_size"]), step 10^-lot_decimals, MinQty = ordermin, MinNotional =
// costmin, first-tier fees as fractions (the venue gives percents; they are
// not account-specific). Symbol is the altname (XBTUSD), NormalizedSymbol
// BASE-QUOTE with Kraken codes normalised (XXBT, XBT give BTC; ZUSD gives USD).
// Dark-pool ".d" pairs are skipped. For margin only pairs with margin leverage
// are listed; Extra["margin_leverage"] holds the highest buy leverage.
func (c *Client) spotExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	resp := model.Response[model.ExchangeInfo]{Kind: model.KindExchangeInfo, Provider: providerID, Market: market}
	res, err := c.spotGet(ctx, "AssetPairs", nil)
	if err != nil {
		return resp, err
	}
	symbols := make([]model.Symbol, 0, len(res))
	for key, raw := range res {
		var p struct {
			Altname      string      `json:"altname"`
			Wsname       string      `json:"wsname"`
			Base         string      `json:"base"`
			Quote        string      `json:"quote"`
			PairDecimals int         `json:"pair_decimals"`
			LotDecimals  int         `json:"lot_decimals"`
			LeverageBuy  []int       `json:"leverage_buy"`
			Fees         [][]float64 `json:"fees"`
			FeesMaker    [][]float64 `json:"fees_maker"`
			OrderMin     string      `json:"ordermin"`
			CostMin      string      `json:"costmin"`
			TickSize     string      `json:"tick_size"`
			Status       string      `json:"status"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return resp, providerErr(model.ErrKindParse, "parse pair "+key+": "+err.Error(), err)
		}
		if p.Altname == "" || strings.HasSuffix(p.Altname, ".d") {
			continue
		}
		if market == model.MarketMargin && len(p.LeverageBuy) == 0 {
			continue
		}
		status := model.SymbolStatusHalt
		if p.Status == "online" {
			status = model.SymbolStatusTrading
		}
		base, quote := normAsset(p.Base), normAsset(p.Quote)
		step := math.Pow(10, -float64(p.LotDecimals))
		qp, pp := p.LotDecimals, p.PairDecimals
		s := model.Symbol{
			Symbol: p.Altname, NormalizedSymbol: base + "-" + quote,
			BaseAsset: base, QuoteAsset: quote, Status: status, Market: market,
			StepSize: &step, QtyPrecision: &qp, PricePrecision: &pp,
			Extra: map[string]any{"pair_key": key, "wsname": p.Wsname},
		}
		if t, err := strconv.ParseFloat(p.TickSize, 64); err == nil && t > 0 {
			s.Extra["tick_size"] = t
		}
		if v, err := strconv.ParseFloat(p.OrderMin, 64); err == nil && v > 0 {
			s.MinQty = &v
		}
		if v, err := strconv.ParseFloat(p.CostMin, 64); err == nil && v > 0 {
			s.MinNotional = &v
		}
		if n := len(p.LeverageBuy); n > 0 {
			s.Extra["margin_leverage"] = p.LeverageBuy[n-1]
		}
		if len(p.Fees) > 0 && len(p.Fees[0]) == 2 {
			t := p.Fees[0][1] / 100
			s.TakerFee = &t
		}
		if len(p.FeesMaker) > 0 && len(p.FeesMaker[0]) == 2 {
			m := p.FeesMaker[0][1] / 100
			s.MakerFee = &m
		}
		symbols = append(symbols, s)
	}
	sort.Slice(symbols, func(i, j int) bool { return symbols[i].Symbol < symbols[j].Symbol })
	resp.Data = model.ExchangeInfo{ExchangeID: providerID, Market: market, Symbols: symbols}
	return resp, nil
}

// spotCandles reads OHLC. The venue returns only the most recent 720 candles
// of the interval and "since" cannot reach further back: a range older than
// that window yields no candles (or only its overlap), never made-up ones.
// From, To (exclusive) and Limit are applied here: with From the earliest
// Limit candles of the range, without From the most recent Limit. The last
// row is the still-forming candle; it is returned as the venue sends it, so
// callers drop it (the kaboom adapter does).
func (c *Client) spotCandles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	resp := model.Response[[]model.Candle]{Provider: providerID, Market: market, Kind: model.KindCandle}
	mins, ok := spotIntervals[interval]
	if !ok {
		return resp, providerErr(model.ErrKindInvalidRequest, "unsupported spot interval "+interval, nil)
	}
	hasFrom := opts.From != nil && !opts.From.IsZero()
	hasTo := opts.To != nil && !opts.To.IsZero()
	extra := url.Values{"interval": {strconv.Itoa(mins)}}
	if hasFrom {
		// since is exclusive: ask from the second before From.
		extra.Set("since", strconv.FormatInt(opts.From.Unix()-1, 10))
	}
	raw, err := c.spotPair(ctx, "OHLC", symbol, extra)
	if err != nil {
		return resp, err
	}
	var rows [][]json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		return resp, providerErr(model.ErrKindParse, "parse OHLC: "+err.Error(), err)
	}
	out := make([]model.Candle, 0, len(rows))
	for _, r := range rows {
		if len(r) < 7 {
			return resp, providerErr(model.ErrKindParse, "short OHLC row", nil)
		}
		var ts int64
		var o, h, l, cl, v string
		for i, dst := range []any{&ts, &o, &h, &l, &cl, nil, &v} {
			if dst == nil {
				continue
			}
			if err := json.Unmarshal(r[i], dst); err != nil {
				return resp, providerErr(model.ErrKindParse, "parse OHLC row: "+err.Error(), err)
			}
		}
		t := time.Unix(ts, 0).UTC()
		if hasFrom && t.Before(*opts.From) || hasTo && !t.Before(*opts.To) {
			continue
		}
		vol := num(v)
		out = append(out, model.Candle{OpenTime: t, Open: num(o), High: num(h), Low: num(l), Close: num(cl), Volume: &vol})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OpenTime.Before(out[j].OpenTime) })
	if opts.Limit != nil && *opts.Limit > 0 && len(out) > *opts.Limit {
		if hasFrom {
			out = out[:*opts.Limit]
		} else {
			out = out[len(out)-*opts.Limit:]
		}
	}
	resp.Data = out
	return resp, nil
}

type spotTicker struct {
	A []string `json:"a"` // ask: price, whole lot volume, lot volume
	B []string `json:"b"` // bid
	C []string `json:"c"` // last trade: price, lot volume
	V []string `json:"v"` // volume: today, last 24h
	P []string `json:"p"` // vwap: today, last 24h
	T []int64  `json:"t"` // trade count: today, last 24h
	L []string `json:"l"` // low: today, last 24h
	H []string `json:"h"` // high: today, last 24h
	O string   `json:"o"` // today's opening price (UTC day)
}

func (c *Client) spotTickerRaw(ctx context.Context, symbol string) (*spotTicker, error) {
	raw, err := c.spotPair(ctx, "Ticker", symbol, nil)
	if err != nil {
		return nil, err
	}
	var t spotTicker
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, providerErr(model.ErrKindParse, "parse Ticker: "+err.Error(), err)
	}
	if len(t.C) == 0 {
		return nil, providerErr(model.ErrKindParse, "Ticker without last price", nil)
	}
	return &t, nil
}

// spotTicker24h reads Ticker. Volume, high, low and vwap are the rolling 24h
// values. The venue has no price 24h ago: OpenPrice is today's UTC-day open
// and PriceChange(Percent) is measured from it (Extra["open_basis"]="utc_day").
// QuoteVolume is not published and stays nil.
func (c *Client) spotTicker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	resp := model.Response[model.Ticker24h]{Provider: providerID, Market: market, Kind: model.KindTicker}
	t, err := c.spotTickerRaw(ctx, symbol)
	if err != nil {
		return resp, err
	}
	last := firstNum(t.C)
	hi, lo, vol, vwap := idx(t.H, 1), idx(t.L, 1), idx(t.V, 1), idx(t.P, 1)
	bid, ask := firstNum(t.B), firstNum(t.A)
	d := model.Ticker24h{
		Symbol: spotNormalized(symbol), OriginalSymbol: toSpot(symbol), Market: market,
		LastPrice: last, HighPrice: &hi, LowPrice: &lo, Volume: &vol,
		BidPrice: &bid, AskPrice: &ask,
		Extra: map[string]any{"open_basis": "utc_day"},
	}
	if vwap > 0 {
		d.WeightedAvgPrice = &vwap
	}
	if open := num(t.O); open > 0 {
		chg := last - open
		pct := chg / open * 100
		d.OpenPrice, d.PriceChange, d.PriceChangePercent = &open, &chg, &pct
	}
	resp.Data = d
	return resp, nil
}

// spotPrice builds a CoinPrice from Ticker.
func (c *Client) spotPrice(ctx context.Context, symbol string) (model.CoinPrice, error) {
	t, err := c.spotTickerRaw(ctx, symbol)
	if err != nil {
		return model.CoinPrice{}, err
	}
	norm := spotNormalized(symbol)
	hi, lo, vol := idx(t.H, 1), idx(t.L, 1), idx(t.V, 1)
	bid, ask, bidSz, askSz := firstNum(t.B), firstNum(t.A), idx(t.B, 2), idx(t.A, 2)
	p := model.CoinPrice{
		ID: toSpot(symbol), Symbol: norm, OriginalSymbol: toSpot(symbol), Currency: quoteOf(norm),
		Price: firstNum(t.C), Volume24h: &vol, High24h: &hi, Low24h: &lo,
		BidPrice: &bid, BidSize: &bidSz, AskPrice: &ask, AskSize: &askSz,
	}
	if open := num(t.O); open > 0 {
		p.Open24h = &open // UTC-day open, see spotTicker24h
	}
	return p, nil
}

// spotOrderBook reads Depth: count levels per side (venue cap 500, default
// 100 when depth <= 0). Entries are [price, volume, unix seconds]; Time is the
// newest entry time.
func (c *Client) spotOrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error) {
	resp := model.Response[model.OrderBook]{Provider: providerID, Market: market, Kind: model.KindOrderBook}
	count := 100
	if depth > 0 {
		count = min(depth, spotDepthMax)
	}
	raw, err := c.spotPair(ctx, "Depth", symbol, url.Values{"count": {strconv.Itoa(count)}})
	if err != nil {
		return resp, err
	}
	var d struct {
		Asks [][]json.RawMessage `json:"asks"`
		Bids [][]json.RawMessage `json:"bids"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return resp, providerErr(model.ErrKindParse, "parse Depth: "+err.Error(), err)
	}
	var newest int64
	conv := func(rows [][]json.RawMessage) ([]model.OrderBookEntry, error) {
		out := make([]model.OrderBookEntry, 0, len(rows))
		for _, r := range rows {
			if len(r) < 3 {
				return nil, providerErr(model.ErrKindParse, "short Depth row", nil)
			}
			var p, q string
			var ts int64
			if json.Unmarshal(r[0], &p) != nil || json.Unmarshal(r[1], &q) != nil || json.Unmarshal(r[2], &ts) != nil {
				return nil, providerErr(model.ErrKindParse, "bad Depth row", nil)
			}
			newest = max(newest, ts)
			out = append(out, model.OrderBookEntry{Price: num(p), Quantity: num(q)})
		}
		return out, nil
	}
	ob := model.OrderBook{Symbol: spotNormalized(symbol), OriginalSymbol: toSpot(symbol), Market: market}
	if ob.Bids, err = conv(d.Bids); err != nil {
		return resp, err
	}
	if ob.Asks, err = conv(d.Asks); err != nil {
		return resp, err
	}
	if newest > 0 {
		t := time.Unix(newest, 0).UTC()
		ob.Time = &t
	}
	resp.Data = ob
	return resp, nil
}
