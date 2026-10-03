package bitvavo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/model"
)

// maxCandleLimit is the largest limit the candles endpoint accepts (default 1440).
const maxCandleLimit = 1440

// intervals maps bits intervals to Bitvavo candle intervals. Others are rejected.
var intervals = map[string]string{
	"1m": "1m", "5m": "5m", "15m": "15m", "30m": "30m",
	"1h": "1h", "2h": "2h", "4h": "4h", "6h": "6h", "8h": "8h", "12h": "12h",
	"1d": "1d",
}

type apiTime struct {
	Time int64 `json:"time"`
}

type apiMarket struct {
	Market               string   `json:"market"`
	Status               string   `json:"status"`
	Base                 string   `json:"base"`
	Quote                string   `json:"quote"`
	MinOrderInBaseAsset  string   `json:"minOrderInBaseAsset"`
	MinOrderInQuoteAsset string   `json:"minOrderInQuoteAsset"`
	MaxOrderInBaseAsset  string   `json:"maxOrderInBaseAsset"`
	MaxOrderInQuoteAsset string   `json:"maxOrderInQuoteAsset"`
	QuantityDecimals     int      `json:"quantityDecimals"`
	NotionalDecimals     int      `json:"notionalDecimals"`
	TickSize             string   `json:"tickSize"`
	FeeCategory          string   `json:"feeCategory"`
	OrderTypes           []string `json:"orderTypes"`
}

type apiTicker struct {
	Market         string `json:"market"`
	Open           string `json:"open"`
	High           string `json:"high"`
	Low            string `json:"low"`
	Last           string `json:"last"`
	Bid            string `json:"bid"`
	Ask            string `json:"ask"`
	Volume         string `json:"volume"`
	VolumeQuote    string `json:"volumeQuote"`
	OpenTimestamp  int64  `json:"openTimestamp"`
	CloseTimestamp int64  `json:"closeTimestamp"`
}

// ServerTime returns the exchange time from /time.
func (c *Client) ServerTime(ctx context.Context) (model.Response[model.ServerTime], error) {
	before := time.Now()
	body, err := c.get(ctx, "/time", "")
	latency := time.Since(before)
	if err != nil {
		return model.Response[model.ServerTime]{}, err
	}
	var t apiTime
	if err := json.Unmarshal(body, &t); err != nil {
		return model.Response[model.ServerTime]{}, providerErr(model.ErrKindParse, "failed to parse time", err)
	}
	return model.Response[model.ServerTime]{
		Kind:     model.KindServerTime,
		Provider: providerID,
		Data:     model.ServerTime{Time: time.UnixMilli(t.Time).UTC(), Latency: &latency},
	}, nil
}

// ExchangeInfo lists spot markets with their trading rules.
//
// Mapping (Bitvavo has no maker/taker fields, see the package comment):
//   - PricePrecision: decimals of tickSize. The "pricePrecision" field is
//     always null; Bitvavo limits prices to 15 significant digits and a tick
//     that depends on the price magnitude, so tickSize is the usable rule. The
//     tick is read per market at call time and can change with the price.
//   - MinPrice: tickSize (smallest price increment, as the cryptocom provider).
//   - QtyPrecision: quantityDecimals. StepSize: 10^-quantityDecimals.
//   - MinQty: minOrderInBaseAsset. MinNotional: minOrderInQuoteAsset.
//   - MaxQty / Extra max_notional: maxOrderInBaseAsset / maxOrderInQuoteAsset.
func (c *Client) ExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	if market != model.MarketSpot {
		return model.Response[model.ExchangeInfo]{}, model.ErrUnsupportedMarket
	}
	body, err := c.get(ctx, "/markets", "")
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to fetch markets: %w", err)
	}
	var markets []apiMarket
	if err := json.Unmarshal(body, &markets); err != nil {
		return model.Response[model.ExchangeInfo]{}, providerErr(model.ErrKindParse, "failed to parse markets", err)
	}

	symbols := make([]model.Symbol, 0, len(markets))
	for _, m := range markets {
		status := model.SymbolStatusHalt
		if m.Status == "trading" {
			status = model.SymbolStatusTrading
		}
		s := model.Symbol{
			Symbol:           m.Market,
			NormalizedSymbol: m.Base + "-" + m.Quote,
			BaseAsset:        m.Base,
			QuoteAsset:       m.Quote,
			Status:           status,
			Market:           model.MarketSpot,
			Extra: map[string]any{
				"fee_category":      m.FeeCategory,
				"notional_decimals": m.NotionalDecimals,
				"order_types":       m.OrderTypes,
				"status_raw":        m.Status,
			},
		}
		qd := m.QuantityDecimals
		s.QtyPrecision = &qd
		step := math.Pow10(-qd)
		s.StepSize = &step
		if v, ok := parseFloat(m.TickSize); ok {
			s.MinPrice = &v
			pp := decimals(m.TickSize)
			s.PricePrecision = &pp
		}
		if v, ok := parseFloat(m.MinOrderInBaseAsset); ok {
			s.MinQty = &v
		}
		if v, ok := parseFloat(m.MaxOrderInBaseAsset); ok {
			s.MaxQty = &v
		}
		if v, ok := parseFloat(m.MinOrderInQuoteAsset); ok {
			s.MinNotional = &v
		}
		if v, ok := parseFloat(m.MaxOrderInQuoteAsset); ok {
			s.Extra["max_notional"] = v
		}
		symbols = append(symbols, s)
	}

	return model.Response[model.ExchangeInfo]{
		Kind:     model.KindExchangeInfo,
		Provider: providerID,
		Market:   market,
		Data:     model.ExchangeInfo{ExchangeID: providerID, Market: market, Symbols: symbols},
	}, nil
}

// Ticker24h returns the rolling 24h statistics of one market.
func (c *Client) Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	if market != model.MarketSpot {
		return model.Response[model.Ticker24h]{}, model.ErrUnsupportedMarket
	}
	sym, err := toMarket(symbol)
	if err != nil {
		return model.Response[model.Ticker24h]{}, err
	}
	body, err := c.get(ctx, "/ticker/24h", "market="+url.QueryEscape(sym))
	if err != nil {
		return model.Response[model.Ticker24h]{}, err
	}
	var t apiTicker
	if err := json.Unmarshal(body, &t); err != nil {
		return model.Response[model.Ticker24h]{}, providerErr(model.ErrKindParse, "failed to parse ticker", err)
	}
	last, ok := parseFloat(t.Last)
	if !ok {
		return model.Response[model.Ticker24h]{}, providerErr(model.ErrKindParse, "ticker has no last price", nil)
	}

	out := model.Ticker24h{
		Symbol:         sym,
		OriginalSymbol: symbol,
		Market:         market,
		LastPrice:      last,
		HighPrice:      optFloat(t.High),
		LowPrice:       optFloat(t.Low),
		OpenPrice:      optFloat(t.Open),
		BidPrice:       optFloat(t.Bid),
		AskPrice:       optFloat(t.Ask),
		Volume:         optFloat(t.Volume),
		QuoteVolume:    optFloat(t.VolumeQuote),
	}
	if out.OpenPrice != nil {
		chg := last - *out.OpenPrice
		out.PriceChange = &chg
		if *out.OpenPrice != 0 {
			pct := chg / *out.OpenPrice * 100
			out.PriceChangePercent = &pct
		}
	}
	if t.OpenTimestamp > 0 {
		ot := time.UnixMilli(t.OpenTimestamp).UTC()
		out.OpenTime = &ot
	}
	if t.CloseTimestamp > 0 {
		ct := time.UnixMilli(t.CloseTimestamp).UTC()
		out.CloseTime = &ct
	}
	return model.Response[model.Ticker24h]{Kind: model.KindTicker, Provider: providerID, Market: market, Data: out}, nil
}

// Candles fetches one page of candles, oldest first.
//
// Observed on the live API (the docs are vague on this): the endpoint returns
// rows newest first; start is inclusive and end is exclusive; when the range
// holds more than limit rows it returns the newest limit rows, so a caller
// pages a range by moving To back (or by windows of at most limit candles as
// the kaboom datasync does). Minutes without a trade have no candle (gaps).
// limit is clamped to 1440. An interval Bitvavo lacks is an error.
func (c *Client) Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	resp := model.Response[[]model.Candle]{Provider: providerID, Market: market, Kind: model.KindCandle}
	if market != model.MarketSpot {
		return resp, model.ErrUnsupportedMarket
	}
	iv, ok := intervals[interval]
	if !ok {
		return resp, providerErr(model.ErrKindInvalidRequest, fmt.Sprintf("unsupported interval %q", interval), nil)
	}
	sym, err := toMarket(symbol)
	if err != nil {
		return resp, err
	}

	q := url.Values{}
	q.Set("interval", iv)
	if opts.From != nil && !opts.From.IsZero() {
		q.Set("start", strconv.FormatInt(opts.From.UnixMilli(), 10))
	}
	if opts.To != nil && !opts.To.IsZero() {
		q.Set("end", strconv.FormatInt(opts.To.UnixMilli(), 10))
	}
	if opts.Limit != nil && *opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(min(*opts.Limit, maxCandleLimit)))
	}
	body, err := c.get(ctx, "/"+url.PathEscape(sym)+"/candles", q.Encode())
	if err != nil {
		return resp, err
	}
	var rows [][]json.RawMessage
	if err := json.Unmarshal(body, &rows); err != nil {
		return resp, providerErr(model.ErrKindParse, "failed to parse candles", err)
	}

	candles := make([]model.Candle, 0, len(rows))
	for _, r := range rows {
		cd, err := parseCandle(r)
		if err != nil {
			return resp, err
		}
		if opts.From != nil && !opts.From.IsZero() && cd.OpenTime.Before(*opts.From) {
			continue
		}
		if opts.To != nil && !opts.To.IsZero() && !cd.OpenTime.Before(*opts.To) {
			continue
		}
		candles = append(candles, cd)
	}
	sort.Slice(candles, func(i, j int) bool { return candles[i].OpenTime.Before(candles[j].OpenTime) })
	resp.Data = candles
	return resp, nil
}

// parseCandle reads [timestamp, open, high, low, close, volume]; a row without
// a volume is rejected.
func parseCandle(r []json.RawMessage) (model.Candle, error) {
	bad := func(msg string) (model.Candle, error) {
		return model.Candle{}, providerErr(model.ErrKindParse, "bad candle row: "+msg, nil)
	}
	if len(r) < 6 {
		return bad("fewer than 6 fields")
	}
	var ts int64
	if err := json.Unmarshal(r[0], &ts); err != nil {
		return bad("timestamp")
	}
	var f [5]float64
	for i := range f {
		var s string
		if err := json.Unmarshal(r[i+1], &s); err != nil {
			return bad("value is not a string")
		}
		v, ok := parseFloat(s)
		if !ok {
			return bad("value " + s)
		}
		f[i] = v
	}
	vol := f[4]
	return model.Candle{OpenTime: time.UnixMilli(ts).UTC(), Open: f[0], High: f[1], Low: f[2], Close: f[3], Volume: &vol}, nil
}

// toMarket converts any accepted spelling (BTC-EUR, BTC_EUR, BTCEUR) to the
// Bitvavo market name BTC-EUR.
func toMarket(symbol string) (string, error) {
	s := model.NormalizeSymbol(symbol)
	if !strings.Contains(s, "-") {
		return "", providerErr(model.ErrKindInvalidRequest, fmt.Sprintf("cannot split symbol %q, use BASE-QUOTE", symbol), nil)
	}
	return s, nil
}

func parseFloat(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

func optFloat(s string) *float64 {
	if v, ok := parseFloat(s); ok {
		return &v
	}
	return nil
}

// decimals counts the decimal places of a plain decimal string, ignoring
// trailing zeros: "0.0000010" is 6, "1.00" is 0.
func decimals(s string) int {
	i := strings.IndexByte(s, '.')
	if i < 0 {
		return 0
	}
	return len(strings.TrimRight(s[i+1:], "0"))
}
