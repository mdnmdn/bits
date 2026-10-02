package binance

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/model"
)

// FundingRates fetches perpetual funding-rate history (/fapi/v1/fundingRate),
// ascending by time. Binance returns at most 1000 entries per call, starting
// at opts.From when set.
func (c *Client) FundingRates(ctx context.Context, symbol string, opts model.FundingRateOpts) (model.Response[[]model.FundingRate], error) {
	if c.futuresClient == nil {
		return model.Response[[]model.FundingRate]{}, providerErr(model.ErrKindUnsupportedMarket, "futures client not configured", nil)
	}
	sym := strings.ToUpper(symbol)

	svc := c.futuresClient.NewFundingRateService().Symbol(sym)
	if opts.From != nil {
		svc = svc.StartTime(opts.From.UnixMilli())
	}
	if opts.To != nil {
		svc = svc.EndTime(opts.To.UnixMilli())
	}
	if opts.Limit != nil {
		svc = svc.Limit(*opts.Limit)
	}

	rows, err := svc.Do(ctx)
	if err != nil {
		return model.Response[[]model.FundingRate]{}, model.WrapError(providerID, err)
	}

	out := make([]model.FundingRate, 0, len(rows))
	for i, r := range rows {
		rate, err := strconv.ParseFloat(r.FundingRate, 64)
		if err != nil {
			return model.Response[[]model.FundingRate]{}, providerErr(model.ErrKindParse, "invalid funding rate at index "+strconv.Itoa(i)+": "+strconv.Quote(r.FundingRate), err)
		}
		fr := model.FundingRate{Symbol: r.Symbol, Time: time.UnixMilli(r.FundingTime), Rate: rate}
		// markPrice is empty for old entries.
		if mp, err := strconv.ParseFloat(r.MarkPrice, 64); err == nil {
			fr.MarkPrice = &mp
		}
		out = append(out, fr)
	}

	return model.Response[[]model.FundingRate]{
		Kind:     model.KindFundingRate,
		Provider: providerID,
		Market:   model.MarketFutures,
		Data:     out,
	}, nil
}
