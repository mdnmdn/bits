package bitget

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/model"
)

const (
	fundingPageSize = 100 // endpoint maximum
	fundingMaxPages = 100
)

type bitgetFundingRateResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Symbol      string `json:"symbol"`
		FundingRate string `json:"fundingRate"`
		FundingTime string `json:"fundingTime"`
	} `json:"data"`
}

// FundingRates fetches USDT-FUTURES funding-rate history
// (/api/v2/mix/market/history-fund-rate), ascending by time. The endpoint has
// no time filter and pages newest-first, so pages are walked until the window
// start is covered, then filtered by From/To.
func (c *Client) FundingRates(_ context.Context, symbol string, opts model.FundingRateOpts) (model.Response[[]model.FundingRate], error) {
	var out []model.FundingRate

	for page := 1; page <= fundingMaxPages; page++ {
		query := fmt.Sprintf("symbol=%s&productType=USDT-FUTURES&pageSize=%d&pageNo=%d", symbol, fundingPageSize, page)
		body, err := c.doRequest("GET", "/api/v2/mix/market/history-fund-rate", query)
		if err != nil {
			return model.Response[[]model.FundingRate]{}, err
		}
		var resp bitgetFundingRateResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return model.Response[[]model.FundingRate]{}, fmt.Errorf("failed to parse funding rate response: %w", err)
		}
		if resp.Code != "00000" {
			return model.Response[[]model.FundingRate]{}, apiErr(resp.Code, resp.Msg)
		}

		covered := false
		for i, row := range resp.Data {
			ms, err := strconv.ParseInt(row.FundingTime, 10, 64)
			if err != nil {
				return model.Response[[]model.FundingRate]{}, providerErr(model.ErrKindParse, fmt.Sprintf("invalid funding time at index %d: %q", i, row.FundingTime), err)
			}
			rate, err := strconv.ParseFloat(row.FundingRate, 64)
			if err != nil {
				return model.Response[[]model.FundingRate]{}, providerErr(model.ErrKindParse, fmt.Sprintf("invalid funding rate at index %d: %q", i, row.FundingRate), err)
			}
			t := time.UnixMilli(ms)
			if opts.From != nil && t.Before(*opts.From) {
				covered = true
				continue
			}
			if opts.To != nil && t.After(*opts.To) {
				continue
			}
			out = append(out, model.FundingRate{Symbol: row.Symbol, Time: t, Rate: rate})
		}

		if covered || len(resp.Data) < fundingPageSize {
			break
		}
		// Without From, stop once Limit entries are collected (newest first).
		if opts.From == nil && opts.Limit != nil && len(out) >= *opts.Limit {
			break
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Time.Before(out[j].Time) })
	if opts.Limit != nil && *opts.Limit > 0 && len(out) > *opts.Limit {
		if opts.From != nil {
			out = out[:*opts.Limit]
		} else {
			out = out[len(out)-*opts.Limit:]
		}
	}
	if out == nil {
		out = []model.FundingRate{}
	}

	return model.Response[[]model.FundingRate]{
		Kind:     model.KindFundingRate,
		Provider: providerID,
		Market:   model.MarketFutures,
		Data:     out,
	}, nil
}
