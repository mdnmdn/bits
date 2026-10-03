package kraken

import (
	"context"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/mdnmdn/bits/model"
)

// isFuturesSymbol reports whether s is a native futures symbol (PF_, PI_, FF_,
// FI_). Price has no market argument and routes on this prefix.
func isFuturesSymbol(s string) bool {
	u := strings.ToUpper(strings.TrimSpace(s))
	for _, p := range []string{"PF_", "PI_", "FF_", "FI_"} {
		if strings.HasPrefix(u, p) {
			return true
		}
	}
	return false
}

// Price returns the last trade price of each symbol. Price has no market
// argument: a native futures symbol (PF_XBTUSD) is read from the futures
// tickers, any other symbol (BTC-USD, XBTUSD) from the spot Ticker. Mixing both
// kinds in one call is an error. Per-symbol failures go to Response.Errors.
// currency is ignored: the quote asset is part of the symbol.
func (c *Client) Price(ctx context.Context, symbols []string, currency string) (model.Response[[]model.CoinPrice], error) {
	resp := model.Response[[]model.CoinPrice]{Provider: providerID, Kind: model.KindPrice, Market: model.MarketSpot}
	fut, spot := 0, 0
	for _, s := range symbols {
		if isFuturesSymbol(s) {
			fut++
		} else {
			spot++
		}
	}
	if fut > 0 && spot > 0 {
		return resp, providerErr(model.ErrKindInvalidRequest, "kraken Price does not mix futures and spot symbols in one call", nil)
	}
	if fut > 0 {
		resp.Market = model.MarketFutures
	}
	out := make([]model.CoinPrice, 0, len(symbols))
	for _, s := range symbols {
		var p model.CoinPrice
		var err error
		if fut > 0 {
			p, err = c.futuresPrice(ctx, s)
		} else {
			p, err = c.spotPrice(ctx, s)
		}
		if err != nil {
			resp.Errors = append(resp.Errors, model.ItemError{Symbol: s, Err: model.WrapError(providerID, err)})
			continue
		}
		out = append(out, p)
	}
	resp.Data = out
	return resp, nil
}

func (c *Client) futuresPrice(ctx context.Context, symbol string) (model.CoinPrice, error) {
	t, err := c.futuresTicker(ctx, toNative(symbol))
	if err != nil {
		return model.CoinPrice{}, err
	}
	native := toNative(symbol)
	p := model.CoinPrice{
		ID: native, Symbol: normalize(native), OriginalSymbol: native, Currency: quoteOf(normalize(native)),
		Price: t.Last, Change24h: &t.Change24h, Volume24h: &t.Vol24h,
		High24h: &t.High24h, Low24h: &t.Low24h, Open24h: &t.Open24h,
		BidPrice: &t.Bid, AskPrice: &t.Ask,
	}
	if t.BidSize > 0 {
		p.BidSize = &t.BidSize
	}
	if t.AskSize > 0 {
		p.AskSize = &t.AskSize
	}
	if ts, err := time.Parse(time.RFC3339Nano, t.Time); err == nil {
		ts = ts.UTC()
		p.Time = &ts
	}
	return p, nil
}

type futuresTicker struct {
	Time      string  `json:"time"`
	Last      float64 `json:"last"`
	Bid       float64 `json:"bid"`
	BidSize   float64 `json:"bidSize"`
	Ask       float64 `json:"ask"`
	AskSize   float64 `json:"askSize"`
	Open24h   float64 `json:"open24h"`
	High24h   float64 `json:"high24h"`
	Low24h    float64 `json:"low24h"`
	Vol24h    float64 `json:"vol24h"`
	VolQuote  float64 `json:"volumeQuote"`
	Vwap24h   float64 `json:"vwap24h"`
	Change24h float64 `json:"change24h"`
}

func (c *Client) futuresTicker(ctx context.Context, native string) (*futuresTicker, error) {
	var body struct {
		Ticker *futuresTicker `json:"ticker"`
	}
	if err := c.get(ctx, "/derivatives/api/v3/tickers/"+url.PathEscape(native), &body); err != nil {
		return nil, err
	}
	if body.Ticker == nil {
		return nil, providerErr(model.ErrKindNotFound, "no ticker for "+native, nil)
	}
	return body.Ticker, nil
}

func quoteOf(normalized string) string {
	_, q, _ := strings.Cut(normalized, "-")
	return q
}

// OrderBook returns a depth snapshot. Futures: /derivatives/api/v3/orderbook
// returns the whole book, so it is sorted (bids high to low, asks low to high)
// and cut to depth. Spot and margin: Depth (see spotOrderBook). depth <= 0
// means the venue default (futures: whole book, spot: 100 levels).
func (c *Client) OrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error) {
	if isSpotMarket(market) {
		return c.spotOrderBook(ctx, symbol, market, depth)
	}
	resp := model.Response[model.OrderBook]{Provider: providerID, Market: model.MarketFutures, Kind: model.KindOrderBook}
	if err := requireFutures(market); err != nil {
		return resp, err
	}
	native := toNative(symbol)
	var body struct {
		ServerTime string `json:"serverTime"`
		OrderBook  *struct {
			Bids [][2]float64 `json:"bids"`
			Asks [][2]float64 `json:"asks"`
		} `json:"orderBook"`
	}
	if err := c.get(ctx, "/derivatives/api/v3/orderbook?symbol="+url.QueryEscape(native), &body); err != nil {
		return resp, err
	}
	if body.OrderBook == nil {
		return resp, providerErr(model.ErrKindNotFound, "no order book for "+native, nil)
	}
	conv := func(rows [][2]float64, desc bool) []model.OrderBookEntry {
		out := make([]model.OrderBookEntry, 0, len(rows))
		for _, r := range rows {
			out = append(out, model.OrderBookEntry{Price: r[0], Quantity: r[1]})
		}
		sort.SliceStable(out, func(i, j int) bool {
			if desc {
				return out[i].Price > out[j].Price
			}
			return out[i].Price < out[j].Price
		})
		if depth > 0 && len(out) > depth {
			out = out[:depth]
		}
		return out
	}
	ob := model.OrderBook{
		Symbol: normalize(native), OriginalSymbol: native, Market: model.MarketFutures,
		Bids: conv(body.OrderBook.Bids, true), Asks: conv(body.OrderBook.Asks, false),
	}
	if t, err := time.Parse(time.RFC3339Nano, body.ServerTime); err == nil {
		t = t.UTC()
		ob.Time = &t
	}
	resp.Data = ob
	return resp, nil
}

// FundingRates returns the perpetual funding history of one symbol from
// /derivatives/api/v4/historicalfundingrates, ascending by time. The venue
// publishes hourly settlements and has no range parameters, so the whole
// history is read and cut here: From and To are inclusive bounds; with From
// and a Limit the earliest Limit entries at or after From are kept, without
// From the most recent Limit. Rate is relativeFundingRate (the per-hour ratio
// of the mark price: 0.0001 = 0.01%); the absolute fundingRate (quote units per
// contract) is not used. No mark price is published.
func (c *Client) FundingRates(ctx context.Context, symbol string, opts model.FundingRateOpts) (model.Response[[]model.FundingRate], error) {
	resp := model.Response[[]model.FundingRate]{Provider: providerID, Market: model.MarketFutures, Kind: model.KindFundingRate}
	native := toNative(symbol)
	var body struct {
		Rates []struct {
			Timestamp    string  `json:"timestamp"`
			RelativeRate float64 `json:"relativeFundingRate"`
		} `json:"rates"`
	}
	if err := c.get(ctx, "/derivatives/api/v4/historicalfundingrates?symbol="+url.QueryEscape(native), &body); err != nil {
		return resp, err
	}
	out := make([]model.FundingRate, 0, len(body.Rates))
	for _, r := range body.Rates {
		t, err := time.Parse(time.RFC3339Nano, r.Timestamp)
		if err != nil {
			return resp, providerErr(model.ErrKindParse, "bad funding timestamp "+r.Timestamp, err)
		}
		t = t.UTC()
		if opts.From != nil && t.Before(*opts.From) || opts.To != nil && t.After(*opts.To) {
			continue
		}
		out = append(out, model.FundingRate{Symbol: native, Time: t, Rate: r.RelativeRate})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Time.Before(out[j].Time) })
	if opts.Limit != nil && *opts.Limit > 0 && len(out) > *opts.Limit {
		if opts.From != nil {
			out = out[:*opts.Limit]
		} else {
			out = out[len(out)-*opts.Limit:]
		}
	}
	resp.Data = out
	return resp, nil
}
