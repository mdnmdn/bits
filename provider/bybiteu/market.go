package bybiteu

import (
	"cmp"
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

// spotMarket validates a market. Bybit EU has spot and spot margin; both read
// the same category=spot data, so margin only changes the Market label of the
// result. Futures (linear) are not offered by the EU entity.
func spotMarket(market model.MarketType) (model.MarketType, error) {
	switch market {
	case model.MarketSpot, "":
		return model.MarketSpot, nil
	case model.MarketMargin:
		return model.MarketMargin, nil
	}
	return "", model.ErrUnsupportedMarket
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
	mkt, err := spotMarket(market)
	if err != nil {
		return resp, err
	}
	resp.Market = mkt
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
	mkt, err := spotMarket(market)
	if err != nil {
		return resp, err
	}
	resp.Market = mkt
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
		Market:             mkt,
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

// Price returns the last price of each symbol in ids (BTCUSDT, BTC-USDT,
// BTC_USDT) from /v5/market/tickers (category=spot), one call per symbol. The
// currency argument is ignored: the quote asset of the symbol is not split
// from the native name, so CoinPrice.Currency stays empty. A failing symbol is
// an item error; the others still return.
func (c *Client) Price(ctx context.Context, ids []string, _ string) (model.Response[[]model.CoinPrice], error) {
	resp := model.Response[[]model.CoinPrice]{Kind: model.KindPrice, Provider: providerID, Market: model.MarketSpot}
	for _, id := range ids {
		cp, err := c.fetchPrice(ctx, id)
		if err != nil {
			resp.Errors = append(resp.Errors, model.ItemError{Symbol: id, Err: model.WrapError(providerID, err)})
			continue
		}
		resp.Data = append(resp.Data, *cp)
	}
	return resp, nil
}

func (c *Client) fetchPrice(ctx context.Context, id string) (*model.CoinPrice, error) {
	native := toNative(id)
	env, err := c.get(ctx, "/v5/market/tickers", url.Values{"category": {"spot"}, "symbol": {native}})
	if err != nil {
		return nil, err
	}
	var res struct {
		List []struct {
			Bid     string `json:"bid1Price"`
			BidSize string `json:"bid1Size"`
			Ask     string `json:"ask1Price"`
			AskSize string `json:"ask1Size"`
			Last    string `json:"lastPrice"`
			Prev24h string `json:"prevPrice24h"`
			Pcnt    string `json:"price24hPcnt"`
			High    string `json:"highPrice24h"`
			Low     string `json:"lowPrice24h"`
			Volume  string `json:"volume24h"`
		} `json:"list"`
	}
	if err := json.Unmarshal(env.Result, &res); err != nil {
		return nil, providerErr(model.ErrKindParse, "parse tickers result: "+err.Error(), err)
	}
	if len(res.List) == 0 {
		return nil, providerErr(model.ErrKindNotFound, "no ticker for "+native, nil)
	}
	t := res.List[0]
	pct := num(t.Pcnt) * 100 // the venue gives a fraction
	bid, bidSz, ask, askSz := num(t.Bid), num(t.BidSize), num(t.Ask), num(t.AskSize)
	hi, lo, open, vol := num(t.High), num(t.Low), num(t.Prev24h), num(t.Volume)
	ts := time.UnixMilli(env.Time).UTC()
	return &model.CoinPrice{
		ID: native, Symbol: model.NormalizeSymbol(native), OriginalSymbol: id,
		Price: num(t.Last), Change24h: &pct, Volume24h: &vol, High24h: &hi, Low24h: &lo, Open24h: &open,
		BidPrice: &bid, BidSize: &bidSz, AskPrice: &ask, AskSize: &askSz, Time: &ts,
	}, nil
}

// maxBookDepth is the largest spot depth /v5/market/orderbook accepts (1-200).
const maxBookDepth = 200

// OrderBook returns a snapshot from /v5/market/orderbook (category=spot).
// depth <= 0 means 20; depth is clamped to 200. LastUpdateID is the venue
// update id "u"; Time is the matching-engine timestamp "cts" (the system "ts"
// when absent).
func (c *Client) OrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error) {
	resp := model.Response[model.OrderBook]{Kind: model.KindOrderBook, Provider: providerID, Market: model.MarketSpot}
	mkt, err := spotMarket(market)
	if err != nil {
		return resp, err
	}
	resp.Market = mkt
	if depth <= 0 {
		depth = 20
	}
	depth = min(depth, maxBookDepth)
	native := toNative(symbol)
	env, err := c.get(ctx, "/v5/market/orderbook", url.Values{"category": {"spot"}, "symbol": {native}, "limit": {strconv.Itoa(depth)}})
	if err != nil {
		return resp, err
	}
	var res struct {
		Symbol string     `json:"s"`
		Bids   [][]string `json:"b"`
		Asks   [][]string `json:"a"`
		TS     int64      `json:"ts"`
		CTS    int64      `json:"cts"`
		U      int64      `json:"u"`
	}
	if err := json.Unmarshal(env.Result, &res); err != nil {
		return resp, providerErr(model.ErrKindParse, "parse orderbook result: "+err.Error(), err)
	}
	bids, err := parseLevels(res.Bids)
	if err != nil {
		return resp, err
	}
	asks, err := parseLevels(res.Asks)
	if err != nil {
		return resp, err
	}
	ob := model.OrderBook{Symbol: model.NormalizeSymbol(native), OriginalSymbol: native, Market: mkt, Bids: bids, Asks: asks}
	if res.U != 0 {
		u := res.U
		ob.LastUpdateID = &u
	}
	if ms := cmp.Or(res.CTS, res.TS); ms > 0 {
		t := time.UnixMilli(ms).UTC()
		ob.Time = &t
	}
	resp.Data = ob
	return resp, nil
}

func parseLevels(rows [][]string) ([]model.OrderBookEntry, error) {
	out := make([]model.OrderBookEntry, 0, len(rows))
	for _, r := range rows {
		if len(r) < 2 {
			return nil, providerErr(model.ErrKindParse, "short book level", nil)
		}
		out = append(out, model.OrderBookEntry{Price: num(r[0]), Quantity: num(r[1])})
	}
	return out, nil
}
