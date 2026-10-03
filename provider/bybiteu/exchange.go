package bybiteu

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/mdnmdn/bits/model"
)

// ServerTime returns the venue clock from /v5/market/time.
func (c *Client) ServerTime(ctx context.Context) (model.Response[model.ServerTime], error) {
	before := time.Now()
	env, err := c.get(ctx, "/v5/market/time", nil)
	latency := time.Since(before)
	if err != nil {
		return model.Response[model.ServerTime]{}, err
	}
	var res struct {
		TimeNano string `json:"timeNano"`
	}
	if err := json.Unmarshal(env.Result, &res); err != nil {
		return model.Response[model.ServerTime]{}, providerErr(model.ErrKindParse, "parse time result: "+err.Error(), err)
	}
	t := time.UnixMilli(env.Time).UTC()
	if ns := int64(num(res.TimeNano)); ns > 0 {
		t = time.Unix(0, ns).UTC()
	}
	return model.Response[model.ServerTime]{
		Kind: model.KindServerTime, Provider: providerID, Market: model.MarketSpot,
		Data: model.ServerTime{Time: t, Latency: &latency},
	}, nil
}

// decimals counts the digits after the decimal point of a step such as "0.001".
func decimals(s string) int {
	_, frac, ok := strings.Cut(s, ".")
	if !ok {
		return 0
	}
	return len(strings.TrimRight(frac, "0"))
}

// ExchangeInfo lists the spot instruments. Fees are not published by a public
// endpoint (/v5/account/fee-rate needs credentials), so MakerFee and TakerFee
// stay nil. The tick size goes to Extra["tick_size"] (model.Symbol has no
// tick field).
func (c *Client) ExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	if err := requireSpot(market); err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}
	var symbols []model.Symbol
	cursor := ""
	for {
		q := url.Values{"category": {"spot"}, "limit": {"1000"}}
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		env, err := c.get(ctx, "/v5/market/instruments-info", q)
		if err != nil {
			return model.Response[model.ExchangeInfo]{}, err
		}
		var res struct {
			List []struct {
				Symbol        string `json:"symbol"`
				BaseCoin      string `json:"baseCoin"`
				QuoteCoin     string `json:"quoteCoin"`
				Status        string `json:"status"`
				LotSizeFilter struct {
					BasePrecision string `json:"basePrecision"`
					MinOrderQty   string `json:"minOrderQty"`
					MaxOrderQty   string `json:"maxOrderQty"`
					MinOrderAmt   string `json:"minOrderAmt"`
				} `json:"lotSizeFilter"`
				PriceFilter struct {
					TickSize string `json:"tickSize"`
				} `json:"priceFilter"`
			} `json:"list"`
			NextPageCursor string `json:"nextPageCursor"`
		}
		if err := json.Unmarshal(env.Result, &res); err != nil {
			return model.Response[model.ExchangeInfo]{}, providerErr(model.ErrKindParse, "parse instruments result: "+err.Error(), err)
		}
		for _, in := range res.List {
			status := model.SymbolStatusHalt
			if in.Status == "Trading" {
				status = model.SymbolStatusTrading
			}
			s := model.Symbol{
				Symbol:           in.Symbol,
				NormalizedSymbol: in.BaseCoin + "-" + in.QuoteCoin,
				BaseAsset:        in.BaseCoin,
				QuoteAsset:       in.QuoteCoin,
				Status:           status,
				Market:           model.MarketSpot,
				Extra:            map[string]any{"tick_size": num(in.PriceFilter.TickSize)},
			}
			pp, qp := decimals(in.PriceFilter.TickSize), decimals(in.LotSizeFilter.BasePrecision)
			s.PricePrecision, s.QtyPrecision = &pp, &qp
			if v := num(in.LotSizeFilter.BasePrecision); v > 0 {
				s.StepSize = &v
			}
			if v := num(in.LotSizeFilter.MinOrderQty); v > 0 {
				s.MinQty = &v
			}
			if v := num(in.LotSizeFilter.MaxOrderQty); v > 0 {
				s.MaxQty = &v
			}
			if v := num(in.LotSizeFilter.MinOrderAmt); v > 0 {
				s.MinNotional = &v
			}
			symbols = append(symbols, s)
		}
		if res.NextPageCursor == "" || len(res.List) == 0 {
			break
		}
		cursor = res.NextPageCursor
	}
	return model.Response[model.ExchangeInfo]{
		Kind: model.KindExchangeInfo, Provider: providerID, Market: model.MarketSpot,
		Data: model.ExchangeInfo{ExchangeID: providerID, Market: model.MarketSpot, Symbols: symbols},
	}, nil
}
