package bybiteu

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/model"
)

// maxCandleLimit is the largest limit /v5/market/kline accepts.
const maxCandleLimit = 1000

// toNative converts a symbol to the Bybit form: BTC-USDT, btc_usdt and
// BTC/USDT all become BTCUSDT.
func toNative(symbol string) string {
	return strings.ToUpper(strings.NewReplacer("-", "", "_", "", "/", "").Replace(symbol))
}

// intervals maps bits interval strings to V5 kline intervals.
var intervals = map[string]string{
	"1m": "1", "3m": "3", "5m": "5", "15m": "15", "30m": "30",
	"1h": "60", "2h": "120", "4h": "240", "6h": "360", "12h": "720",
	"1d": "D", "1w": "W", "1M": "M",
}

func mapInterval(interval string) (string, error) {
	if v, ok := intervals[interval]; ok {
		return v, nil
	}
	return "", providerErr(model.ErrKindInvalidRequest, "unsupported interval "+interval, nil)
}

func requireSpot(market model.MarketType) error {
	if market != model.MarketSpot && market != "" {
		return model.ErrUnsupportedMarket
	}
	return nil
}

func num(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// Candles fetches klines. The venue returns candles newest first, in [start,
// end] (both inclusive), at most limit rows: a window wider than limit
// candles yields the newest ones. Callers page with windows of limit
// candles. To is treated as exclusive. Results are ascending.
func (c *Client) Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	resp := model.Response[[]model.Candle]{Provider: providerID, Market: model.MarketSpot, Kind: model.KindCandle}
	if err := requireSpot(market); err != nil {
		return resp, err
	}
	iv, err := mapInterval(interval)
	if err != nil {
		return resp, err
	}
	q := url.Values{"category": {"spot"}, "symbol": {toNative(symbol)}, "interval": {iv}}
	hasFrom := opts.From != nil && !opts.From.IsZero()
	hasTo := opts.To != nil && !opts.To.IsZero()
	if hasFrom {
		q.Set("start", strconv.FormatInt(opts.From.UnixMilli(), 10))
	}
	if hasTo {
		q.Set("end", strconv.FormatInt(opts.To.UnixMilli()-1, 10))
	}
	limit := maxCandleLimit
	if opts.Limit != nil && *opts.Limit > 0 {
		limit = min(*opts.Limit, maxCandleLimit)
	}
	q.Set("limit", strconv.Itoa(limit))

	env, err := c.get(ctx, "/v5/market/kline", q)
	if err != nil {
		return resp, err
	}
	var res struct {
		List [][]string `json:"list"`
	}
	if err := json.Unmarshal(env.Result, &res); err != nil {
		return resp, providerErr(model.ErrKindParse, "parse kline result: "+err.Error(), err)
	}
	out := make([]model.Candle, 0, len(res.List))
	for _, row := range res.List {
		if len(row) < 6 {
			return resp, providerErr(model.ErrKindParse, "short kline row", nil)
		}
		ms, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			return resp, providerErr(model.ErrKindParse, "bad kline timestamp "+row[0], err)
		}
		t := time.UnixMilli(ms).UTC()
		if hasFrom && t.Before(*opts.From) {
			continue
		}
		if hasTo && !t.Before(*opts.To) {
			continue
		}
		vol := num(row[5])
		out = append(out, model.Candle{OpenTime: t, Open: num(row[1]), High: num(row[2]), Low: num(row[3]), Close: num(row[4]), Volume: &vol})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OpenTime.Before(out[j].OpenTime) })
	resp.Data = out
	return resp, nil
}

// Ticker24h fetches the 24h statistics of one spot symbol.
func (c *Client) Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	resp := model.Response[model.Ticker24h]{Provider: providerID, Market: model.MarketSpot, Kind: model.KindTicker}
	if err := requireSpot(market); err != nil {
		return resp, err
	}
	native := toNative(symbol)
	env, err := c.get(ctx, "/v5/market/tickers", url.Values{"category": {"spot"}, "symbol": {native}})
	if err != nil {
		return resp, err
	}
	var res struct {
		List []struct {
			Symbol   string `json:"symbol"`
			Bid      string `json:"bid1Price"`
			Ask      string `json:"ask1Price"`
			Last     string `json:"lastPrice"`
			Prev24h  string `json:"prevPrice24h"`
			Pcnt     string `json:"price24hPcnt"`
			High     string `json:"highPrice24h"`
			Low      string `json:"lowPrice24h"`
			Turnover string `json:"turnover24h"`
			Volume   string `json:"volume24h"`
		} `json:"list"`
	}
	if err := json.Unmarshal(env.Result, &res); err != nil {
		return resp, providerErr(model.ErrKindParse, "parse tickers result: "+err.Error(), err)
	}
	if len(res.List) == 0 {
		return resp, providerErr(model.ErrKindNotFound, "no ticker for "+native, nil)
	}
	t := res.List[0]
	last, prev := num(t.Last), num(t.Prev24h)
	chg := last - prev
	pct := num(t.Pcnt) * 100 // the venue gives a fraction
	bid, ask, hi, lo := num(t.Bid), num(t.Ask), num(t.High), num(t.Low)
	vol, qvol := num(t.Volume), num(t.Turnover)
	resp.Data = model.Ticker24h{
		Symbol:             model.NormalizeSymbol(native),
		OriginalSymbol:     native,
		Market:             model.MarketSpot,
		LastPrice:          last,
		PriceChange:        &chg,
		PriceChangePercent: &pct,
		HighPrice:          &hi,
		LowPrice:           &lo,
		Volume:             &vol,
		QuoteVolume:        &qvol,
		OpenPrice:          &prev,
		BidPrice:           &bid,
		AskPrice:           &ask,
	}
	return resp, nil
}
