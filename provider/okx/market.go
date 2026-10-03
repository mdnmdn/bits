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

// maxCandlePage is the largest limit accepted by /market/history-candles.
const maxCandlePage = 300

// convertBar maps a bits interval ("1m", "1h", "1d") to an OKX bar. Day and
// longer bars use the UTC variants so they open at 00:00 UTC.
func convertBar(interval string) (string, error) {
	switch interval {
	case "1m", "3m", "5m", "15m", "30m":
		return interval, nil
	case "1h":
		return "1H", nil
	case "2h":
		return "2H", nil
	case "4h":
		return "4H", nil
	case "6h":
		return "6Hutc", nil
	case "12h":
		return "12Hutc", nil
	case "1d":
		return "1Dutc", nil
	case "2d":
		return "2Dutc", nil
	case "3d":
		return "3Dutc", nil
	case "1w":
		return "1Wutc", nil
	case "1M":
		return "1Mutc", nil
	}
	return "", providerErr(model.ErrKindInvalidRequest, fmt.Sprintf("unsupported interval %q", interval), nil)
}

// Candles fetches OHLCV candles, oldest first. It pages backwards through
// /market/history-candles (limit 300, cursor "after" = records older than ts).
//
// opts.To is inclusive on the candle open time; opts.From is inclusive. With no
// From, the newest Limit candles at or before To are returned (default one
// page). With From and Limit, the first Limit candles from From are returned.
// Candles OKX has not yet confirmed (still forming) are dropped. Volume is in
// base asset (for SWAP, base-coin volume, not contracts).
func (c *Client) Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	bar, err := convertBar(interval)
	if err != nil {
		return model.Response[[]model.Candle]{}, err
	}
	id := instID(symbol, market)

	var fromMs, toMs int64
	if opts.From != nil {
		fromMs = opts.From.UnixMilli()
	}
	hasTo := opts.To != nil
	if hasTo {
		toMs = opts.To.UnixMilli()
	}
	limit := 0
	if opts.Limit != nil && *opts.Limit > 0 {
		limit = *opts.Limit
	}

	var cursor int64
	if hasTo {
		cursor = toMs + 1
	}

	var out []model.Candle
	for page := 0; ; page++ {
		pageSize := maxCandlePage
		if opts.From == nil {
			want := 100
			if limit > 0 {
				want = limit
			}
			if rem := want - len(out); rem < pageSize {
				pageSize = rem
			}
		}
		if pageSize <= 0 {
			break
		}

		if page > 0 && c.pageDelay > 0 {
			select {
			case <-ctx.Done():
				return model.Response[[]model.Candle]{}, wrapTransportError(ctx.Err())
			case <-time.After(c.pageDelay):
			}
		}

		query := fmt.Sprintf("instId=%s&bar=%s&limit=%d", id, bar, pageSize)
		if cursor > 0 {
			query += fmt.Sprintf("&after=%d", cursor)
		}
		data, err := c.get(ctx, "/api/v5/market/history-candles", query)
		if err != nil {
			return model.Response[[]model.Candle]{}, err
		}
		var rows [][]string
		if err := json.Unmarshal(data, &rows); err != nil {
			return model.Response[[]model.Candle]{}, providerErr(model.ErrKindParse, "failed to parse candles", err)
		}
		if len(rows) == 0 {
			break
		}

		oldest := int64(0)
		for _, row := range rows {
			cd, ts, confirmed, err := parseCandleRow(row, market)
			if err != nil {
				return model.Response[[]model.Candle]{}, err
			}
			if cursor != 0 && ts >= cursor {
				continue // not older than the cursor: ignore, never loop on it
			}
			if oldest == 0 || ts < oldest {
				oldest = ts
			}
			if !confirmed || (opts.From != nil && ts < fromMs) || (hasTo && ts > toMs) {
				continue
			}
			out = append(out, cd)
		}

		if oldest == 0 {
			break // nothing older than the cursor
		}
		cursor = oldest
		if opts.From != nil && oldest <= fromMs {
			break
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].OpenTime.Before(out[j].OpenTime) })
	if limit > 0 && len(out) > limit {
		if opts.From != nil {
			out = out[:limit]
		} else {
			out = out[len(out)-limit:]
		}
	}

	return model.Response[[]model.Candle]{
		Kind:     model.KindCandle,
		Provider: providerID,
		Market:   market,
		Data:     out,
	}, nil
}

// parseCandleRow parses [ts, o, h, l, c, vol, volCcy, volCcyQuote, confirm].
// Spot vol is in base asset; for SWAP vol is in contracts and volCcy is in base.
func parseCandleRow(row []string, market model.MarketType) (model.Candle, int64, bool, error) {
	volIdx := 5
	if market == model.MarketFutures {
		volIdx = 6
	}
	if len(row) <= volIdx {
		return model.Candle{}, 0, false, providerErr(model.ErrKindParse, fmt.Sprintf("expected at least %d fields, got %d", volIdx+1, len(row)), nil)
	}
	ts, err := strconv.ParseInt(row[0], 10, 64)
	if err != nil {
		return model.Candle{}, 0, false, providerErr(model.ErrKindParse, fmt.Sprintf("invalid timestamp %q", row[0]), err)
	}
	var f [4]float64
	for j := range f {
		f[j], err = strconv.ParseFloat(row[j+1], 64)
		if err != nil {
			return model.Candle{}, 0, false, providerErr(model.ErrKindParse, fmt.Sprintf("invalid value %q", row[j+1]), err)
		}
	}
	vol, err := strconv.ParseFloat(row[volIdx], 64)
	if err != nil {
		return model.Candle{}, 0, false, providerErr(model.ErrKindParse, fmt.Sprintf("invalid volume %q", row[volIdx]), err)
	}
	confirmed := len(row) < 9 || row[8] == "1"
	return model.Candle{
		OpenTime: time.UnixMilli(ts).UTC(),
		Open:     f[0],
		High:     f[1],
		Low:      f[2],
		Close:    f[3],
		Volume:   &vol,
	}, ts, confirmed, nil
}

type okxTicker struct {
	InstID    string `json:"instId"`
	Last      string `json:"last"`
	AskPx     string `json:"askPx"`
	BidPx     string `json:"bidPx"`
	Open24h   string `json:"open24h"`
	High24h   string `json:"high24h"`
	Low24h    string `json:"low24h"`
	VolCcy24h string `json:"volCcy24h"`
	Vol24h    string `json:"vol24h"`
	Ts        string `json:"ts"`
}

// Ticker24h fetches the 24-hour rolling statistics for a symbol. For SWAP the
// base volume is volCcy24h (vol24h is in contracts) and the quote volume is
// derived as base volume times last price (an approximation).
func (c *Client) Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	id := instID(symbol, market)
	data, err := c.get(ctx, "/api/v5/market/ticker", "instId="+id)
	if err != nil {
		return model.Response[model.Ticker24h]{}, err
	}
	var rows []okxTicker
	if err := json.Unmarshal(data, &rows); err != nil {
		return model.Response[model.Ticker24h]{}, providerErr(model.ErrKindParse, "failed to parse ticker", err)
	}
	if len(rows) == 0 {
		return model.Response[model.Ticker24h]{}, providerErr(model.ErrKindNotFound, "no ticker for "+id, nil)
	}
	r := rows[0]

	last, err := strconv.ParseFloat(r.Last, 64)
	if err != nil {
		return model.Response[model.Ticker24h]{}, providerErr(model.ErrKindParse, fmt.Sprintf("invalid last price %q", r.Last), err)
	}
	open := parseOpt(r.Open24h)
	t := model.Ticker24h{
		Symbol:         normalizedFromInstID(r.InstID),
		OriginalSymbol: r.InstID,
		Market:         market,
		LastPrice:      last,
		HighPrice:      parseOpt(r.High24h),
		LowPrice:       parseOpt(r.Low24h),
		OpenPrice:      open,
		BidPrice:       parseOpt(r.BidPx),
		AskPrice:       parseOpt(r.AskPx),
	}
	if open != nil && *open != 0 {
		chg := last - *open
		pct := chg / *open * 100
		t.PriceChange = &chg
		t.PriceChangePercent = &pct
	}
	if market == model.MarketFutures {
		t.Volume = parseOpt(r.VolCcy24h)
		if t.Volume != nil {
			q := *t.Volume * last
			t.QuoteVolume = &q
		}
	} else {
		t.Volume = parseOpt(r.Vol24h)
		t.QuoteVolume = parseOpt(r.VolCcy24h)
	}
	if ms, err := strconv.ParseInt(r.Ts, 10, 64); err == nil {
		ct := time.UnixMilli(ms).UTC()
		ot := ct.Add(-24 * time.Hour)
		t.CloseTime, t.OpenTime = &ct, &ot
	}

	return model.Response[model.Ticker24h]{
		Kind:     model.KindTicker,
		Provider: providerID,
		Market:   market,
		Data:     t,
	}, nil
}
