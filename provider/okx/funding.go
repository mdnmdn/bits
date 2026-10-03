package okx

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
	// fundingPageSize is the page size requested from funding-rate-history.
	fundingPageSize = 100
	fundingMaxPages = 100
	// defaultFundingLimit applies when neither From nor Limit is set.
	defaultFundingLimit = 100
)

type okxFunding struct {
	InstID       string `json:"instId"`
	FundingRate  string `json:"fundingRate"`
	RealizedRate string `json:"realizedRate"`
	FundingTime  string `json:"fundingTime"`
}

// FundingRates fetches SWAP funding-rate history, ascending by time, via
// /api/v5/public/funding-rate-history. The endpoint pages newest-first with the
// "after" cursor (records older than the given ms timestamp); OKX only keeps
// about the last three months. The realized rate is used (it equals the
// published rate for the settled period). opts.From and opts.To are inclusive.
// With no From and no Limit, the newest 100 entries are returned.
func (c *Client) FundingRates(ctx context.Context, symbol string, opts model.FundingRateOpts) (model.Response[[]model.FundingRate], error) {
	id := instID(symbol, model.MarketFutures)

	limit := 0
	if opts.Limit != nil && *opts.Limit > 0 {
		limit = *opts.Limit
	}
	if limit == 0 && opts.From == nil {
		limit = defaultFundingLimit
	}

	var cursor int64
	if opts.To != nil {
		cursor = opts.To.UnixMilli() + 1
	}

	out := []model.FundingRate{}
	for page := 0; page < fundingMaxPages; page++ {
		if page > 0 && c.pageDelay > 0 {
			select {
			case <-ctx.Done():
				return model.Response[[]model.FundingRate]{}, wrapTransportError(ctx.Err())
			case <-time.After(c.pageDelay):
			}
		}
		query := fmt.Sprintf("instId=%s&limit=%d", id, fundingPageSize)
		if cursor > 0 {
			query += fmt.Sprintf("&after=%d", cursor)
		}
		data, err := c.get(ctx, "/api/v5/public/funding-rate-history", query)
		if err != nil {
			return model.Response[[]model.FundingRate]{}, err
		}
		var rows []okxFunding
		if err := json.Unmarshal(data, &rows); err != nil {
			return model.Response[[]model.FundingRate]{}, providerErr(model.ErrKindParse, "failed to parse funding rates", err)
		}

		oldest := int64(0)
		covered := false
		for i, r := range rows {
			ms, err := strconv.ParseInt(r.FundingTime, 10, 64)
			if err != nil {
				return model.Response[[]model.FundingRate]{}, providerErr(model.ErrKindParse, fmt.Sprintf("invalid funding time at index %d: %q", i, r.FundingTime), err)
			}
			if cursor != 0 && ms >= cursor {
				continue // not older than the cursor: never loop on it
			}
			if oldest == 0 || ms < oldest {
				oldest = ms
			}
			if opts.From != nil && ms < opts.From.UnixMilli() {
				covered = true
				continue
			}
			rateStr := r.RealizedRate
			if rateStr == "" {
				rateStr = r.FundingRate
			}
			rate, err := strconv.ParseFloat(rateStr, 64)
			if err != nil {
				return model.Response[[]model.FundingRate]{}, providerErr(model.ErrKindParse, fmt.Sprintf("invalid funding rate at index %d: %q", i, rateStr), err)
			}
			out = append(out, model.FundingRate{
				Symbol: normalizedFromInstID(r.InstID),
				Time:   time.UnixMilli(ms).UTC(),
				Rate:   rate,
			})
		}

		if oldest == 0 || covered || len(rows) < fundingPageSize {
			break
		}
		cursor = oldest
		// Without From, stop once Limit entries are collected (newest first).
		if opts.From == nil && len(out) >= limit {
			break
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Time.Before(out[j].Time) })
	if limit > 0 && len(out) > limit {
		if opts.From != nil {
			out = out[:limit]
		} else {
			out = out[len(out)-limit:]
		}
	}

	return model.Response[[]model.FundingRate]{
		Kind:     model.KindFundingRate,
		Provider: providerID,
		Market:   model.MarketFutures,
		Data:     out,
	}, nil
}
