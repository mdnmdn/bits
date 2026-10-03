package kraken

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/model"
)

// maxCandleCount is the most candles the charts API returns per request.
const maxCandleCount = 2000

// toNative converts a symbol to the Kraken Futures form. A native symbol
// (PF_XBTUSD, PI_ETHUSD, FF_SOLUSD_261225) is only upper-cased. Otherwise
// BASE-QUOTE, BASE_QUOTE, BASE/QUOTE or BASEQUOTE means the linear perpetual
// PF_<BASE><QUOTE>, with BTC written XBT (BTC-USD gives PF_XBTUSD).
func toNative(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	for _, p := range []string{"PF_", "PI_", "FF_", "FI_"} {
		if strings.HasPrefix(s, p) {
			return s
		}
	}
	s = strings.NewReplacer("-", "", "_", "", "/", "", ":", "").Replace(s)
	if strings.HasPrefix(s, "BTC") {
		s = "XBT" + s[3:]
	}
	return "PF_" + s
}

// normalize converts a native symbol to BASE-QUOTE where it can
// (PF_XBTUSD -> BTC-USD); instrument lists carry the exact base and quote.
func normalize(native string) string {
	_, rest, ok := strings.Cut(native, "_")
	if !ok {
		return native
	}
	rest, _, _ = strings.Cut(rest, "_") // drop a dated-contract suffix
	for _, q := range []string{"USDT", "USDC", "USD", "EUR", "GBP"} {
		if base, found := strings.CutSuffix(rest, q); found && base != "" {
			if base == "XBT" {
				base = "BTC"
			}
			return base + "-" + q
		}
	}
	return rest
}

var intervals = map[string]string{
	"1m": "1m", "5m": "5m", "15m": "15m", "30m": "30m",
	"1h": "1h", "4h": "4h", "12h": "12h", "1d": "1d", "1w": "1w",
}

func mapInterval(interval string) (string, error) {
	if v, ok := intervals[interval]; ok {
		return v, nil
	}
	return "", providerErr(model.ErrKindInvalidRequest, "unsupported interval "+interval, nil)
}

func requireFutures(market model.MarketType) error {
	if market != model.MarketFutures {
		return model.ErrUnsupportedMarket
	}
	return nil
}

func num(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// Candles fetches trade candles from /api/charts/v1/trade/{symbol}/{interval}.
// from and to are seconds. The venue returns candles oldest first starting at
// from, at most count (cap 2000) rows, and sets more_candles when the window
// holds more; a range wider than count candles yields the oldest ones. Callers
// page with windows of count candles. To is treated as exclusive.
func (c *Client) Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	resp := model.Response[[]model.Candle]{Provider: providerID, Market: model.MarketFutures, Kind: model.KindCandle}
	if err := requireFutures(market); err != nil {
		return resp, err
	}
	iv, err := mapInterval(interval)
	if err != nil {
		return resp, err
	}
	q := url.Values{}
	hasFrom := opts.From != nil && !opts.From.IsZero()
	hasTo := opts.To != nil && !opts.To.IsZero()
	if hasFrom {
		q.Set("from", strconv.FormatInt(opts.From.Unix(), 10))
	}
	if hasTo {
		// The venue's to is inclusive; ask for the last second before To.
		q.Set("to", strconv.FormatInt((opts.To.UnixMilli()-1)/1000, 10))
	}
	count := maxCandleCount
	if opts.Limit != nil && *opts.Limit > 0 {
		count = min(*opts.Limit, maxCandleCount)
	}
	q.Set("count", strconv.Itoa(count))

	var body struct {
		Candles []struct {
			Time   int64  `json:"time"`
			Open   string `json:"open"`
			High   string `json:"high"`
			Low    string `json:"low"`
			Close  string `json:"close"`
			Volume string `json:"volume"`
		} `json:"candles"`
	}
	path := fmt.Sprintf("/api/charts/v1/trade/%s/%s?%s", url.PathEscape(toNative(symbol)), iv, q.Encode())
	if err := c.get(ctx, path, &body); err != nil {
		return resp, err
	}
	out := make([]model.Candle, 0, len(body.Candles))
	for _, k := range body.Candles {
		t := time.UnixMilli(k.Time).UTC()
		if hasFrom && t.Before(*opts.From) {
			continue
		}
		if hasTo && !t.Before(*opts.To) {
			continue
		}
		vol := num(k.Volume)
		out = append(out, model.Candle{OpenTime: t, Open: num(k.Open), High: num(k.High), Low: num(k.Low), Close: num(k.Close), Volume: &vol})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OpenTime.Before(out[j].OpenTime) })
	resp.Data = out
	return resp, nil
}

// Ticker24h fetches the 24h statistics of one futures symbol. Volume is in
// the contract unit (base asset for linear PF_ contracts); QuoteVolume is
// volumeQuote.
func (c *Client) Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	resp := model.Response[model.Ticker24h]{Provider: providerID, Market: model.MarketFutures, Kind: model.KindTicker}
	if err := requireFutures(market); err != nil {
		return resp, err
	}
	native := toNative(symbol)
	var body struct {
		Ticker *struct {
			Last      float64 `json:"last"`
			Bid       float64 `json:"bid"`
			Ask       float64 `json:"ask"`
			Open24h   float64 `json:"open24h"`
			High24h   float64 `json:"high24h"`
			Low24h    float64 `json:"low24h"`
			Vol24h    float64 `json:"vol24h"`
			VolQuote  float64 `json:"volumeQuote"`
			Vwap24h   float64 `json:"vwap24h"`
			Change24h float64 `json:"change24h"`
		} `json:"ticker"`
	}
	if err := c.get(ctx, "/derivatives/api/v3/tickers/"+url.PathEscape(native), &body); err != nil {
		return resp, err
	}
	if body.Ticker == nil {
		return resp, providerErr(model.ErrKindNotFound, "no ticker for "+native, nil)
	}
	t := body.Ticker
	chg := t.Last - t.Open24h
	d := model.Ticker24h{
		Symbol:             normalize(native),
		OriginalSymbol:     native,
		Market:             model.MarketFutures,
		LastPrice:          t.Last,
		PriceChange:        &chg,
		PriceChangePercent: &t.Change24h, // already a percent
		HighPrice:          &t.High24h,
		LowPrice:           &t.Low24h,
		Volume:             &t.Vol24h,
		QuoteVolume:        &t.VolQuote,
		OpenPrice:          &t.Open24h,
		BidPrice:           &t.Bid,
		AskPrice:           &t.Ask,
	}
	if t.Vwap24h > 0 {
		d.WeightedAvgPrice = &t.Vwap24h
	}
	resp.Data = d
	return resp, nil
}
