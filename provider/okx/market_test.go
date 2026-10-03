package okx

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mdnmdn/bits/model"
)

const minute = int64(60_000)

// historyServer serves 1m candles for open times in [first, last] (ms) the way
// OKX does: newest first, records strictly older than "after", at most "limit".
func historyServer(t *testing.T, first, last int64, queries *[]string) *Client {
	t.Helper()
	return newServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		*queries = append(*queries, r.URL.RawQuery)
		limit, _ := strconv.Atoi(q.Get("limit"))
		if limit > maxCandlePage {
			limit = maxCandlePage
		}
		start := last
		if a := q.Get("after"); a != "" {
			after, _ := strconv.ParseInt(a, 10, 64)
			if after-minute < start {
				start = (after - 1) / minute * minute
			}
		}
		var rows []string
		for ts := start; ts >= first && len(rows) < limit; ts -= minute {
			rows = append(rows, fmt.Sprintf(`["%d","1","2","0.5","1.5","10","%d","15","1"]`, ts, ts/minute))
		}
		_, _ = w.Write([]byte(`{"code":"0","msg":"","data":[` + strings.Join(rows, ",") + `]}`))
	})
}

func TestCandlesPagesBackAcrossRange(t *testing.T) {
	base := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	n := int64(700) // three pages of 300/300/100
	var queries []string
	c := historyServer(t, base-5*minute, base+n*minute, &queries)

	from := time.UnixMilli(base).UTC()
	to := time.UnixMilli(base + (n-1)*minute).UTC()
	resp, err := c.Candles(context.Background(), "BTCUSDT", model.MarketSpot, "1m", model.CandleOpts{From: &from, To: &to})
	require.NoError(t, err)

	require.Len(t, resp.Data, int(n))
	assert.Equal(t, from, resp.Data[0].OpenTime)
	assert.Equal(t, to, resp.Data[len(resp.Data)-1].OpenTime)
	for i := 1; i < len(resp.Data); i++ {
		require.Equal(t, minute, resp.Data[i].OpenTime.Sub(resp.Data[i-1].OpenTime).Milliseconds(), "gap at %d", i)
	}
	assert.Equal(t, "UTC", resp.Data[0].OpenTime.Location().String())
	require.NotNil(t, resp.Data[0].Volume)
	assert.Equal(t, 10.0, *resp.Data[0].Volume, "spot volume is the base-asset column")

	require.GreaterOrEqual(t, len(queries), 3)
	assert.Contains(t, queries[0], "instId=BTC-USDT")
	assert.Contains(t, queries[0], "bar=1m")
	assert.Contains(t, queries[0], "limit=300")
	assert.Contains(t, queries[0], fmt.Sprintf("after=%d", to.UnixMilli()+1), "To is inclusive")
	assert.Contains(t, queries[1], fmt.Sprintf("after=%d", to.UnixMilli()-299*minute), "cursor is the oldest open time of the previous page")
}

func TestCandlesNoFromReturnsNewestLimit(t *testing.T) {
	var queries []string
	c := historyServer(t, 0, 1000*minute, &queries)
	limit := 5
	resp, err := c.Candles(context.Background(), "BTC-USDT", model.MarketSpot, "1m", model.CandleOpts{Limit: &limit})
	require.NoError(t, err)
	require.Len(t, resp.Data, 5)
	assert.Equal(t, int64(1000*minute), resp.Data[4].OpenTime.UnixMilli())
	assert.Equal(t, int64(996*minute), resp.Data[0].OpenTime.UnixMilli())
	require.Len(t, queries, 1)
	assert.Contains(t, queries[0], "limit=5")
	assert.NotContains(t, queries[0], "after=")
}

func TestCandlesFromAndLimitReturnsEarliest(t *testing.T) {
	var queries []string
	c := historyServer(t, 0, 99*minute, &queries)
	from := time.UnixMilli(10 * minute).UTC()
	limit := 3
	resp, err := c.Candles(context.Background(), "BTC-USDT", model.MarketSpot, "1m", model.CandleOpts{From: &from, Limit: &limit})
	require.NoError(t, err)
	require.Len(t, resp.Data, 3)
	assert.Equal(t, from, resp.Data[0].OpenTime)
}

func TestCandlesStopsWhenHistoryEnds(t *testing.T) {
	var queries []string
	c := historyServer(t, 100*minute, 150*minute, &queries)
	from := time.UnixMilli(0).UTC()
	to := time.UnixMilli(150 * minute).UTC()
	resp, err := c.Candles(context.Background(), "BTC-USDT", model.MarketSpot, "1m", model.CandleOpts{From: &from, To: &to})
	require.NoError(t, err)
	assert.Len(t, resp.Data, 51)
	assert.Len(t, queries, 2, "second page is empty and ends the loop")
}

func TestCandlesSwapUsesBaseVolumeAndDropsUnconfirmed(t *testing.T) {
	var req http.Request
	c := staticServer(t, `{"code":"0","msg":"","data":[
["1791028260000","3","4","2","3.5","999","0.5","42","0"],
["1791028200000","1","2","0.5","1.5","2715.07","27.1507","2298630.5","1"]]}`, &req)
	resp, err := c.Candles(context.Background(), "BTCUSDT", model.MarketFutures, "1m", model.CandleOpts{})
	require.NoError(t, err)
	assert.Equal(t, "BTC-USDT-SWAP", req.URL.Query().Get("instId"))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, int64(1791028200000), resp.Data[0].OpenTime.UnixMilli())
	assert.Equal(t, 27.1507, *resp.Data[0].Volume)
}

func TestCandlesIntervals(t *testing.T) {
	for in, want := range map[string]string{"1m": "1m", "1h": "1H", "4h": "4H", "1d": "1Dutc", "1w": "1Wutc"} {
		got, err := convertBar(in)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
	_, err := convertBar("7m")
	var pe *model.ProviderError
	require.ErrorAs(t, err, &pe)
	assert.Equal(t, model.ErrKindInvalidRequest, pe.Kind)
}

func TestCandlesMalformedNumberReturnsError(t *testing.T) {
	rows := map[string]string{
		"timestamp": `["abc","1","2","0.5","1.5","10"]`,
		"open":      `["1756684800000","x","2","0.5","1.5","10"]`,
		"volume":    `["1756684800000","1","2","0.5","1.5","NaNx"]`,
	}
	for name, row := range rows {
		t.Run(name, func(t *testing.T) {
			c := staticServer(t, `{"code":"0","msg":"","data":[`+row+`]}`, nil)
			_, err := c.Candles(context.Background(), "BTCUSDT", model.MarketSpot, "1m", model.CandleOpts{})
			var pe *model.ProviderError
			require.ErrorAs(t, err, &pe)
			assert.Equal(t, model.ErrKindParse, pe.Kind)
		})
	}
}

func TestTicker24hSpot(t *testing.T) {
	var req http.Request
	c := staticServer(t, `{"code":"0","msg":"","data":[{"instType":"SPOT","instId":"BTC-USDT","last":"110","askPx":"111","bidPx":"109","open24h":"100","high24h":"120","low24h":"90","volCcy24h":"5000","vol24h":"50","ts":"1791028289971"}]}`, &req)
	resp, err := c.Ticker24h(context.Background(), "BTCUSDT", model.MarketSpot)
	require.NoError(t, err)
	assert.Equal(t, "BTC-USDT", req.URL.Query().Get("instId"))
	d := resp.Data
	assert.Equal(t, "BTC-USDT", d.Symbol)
	assert.Equal(t, 110.0, d.LastPrice)
	assert.InDelta(t, 10.0, *d.PriceChangePercent, 1e-9)
	assert.Equal(t, 50.0, *d.Volume)
	assert.Equal(t, 5000.0, *d.QuoteVolume)
	assert.Equal(t, int64(1791028289971), d.CloseTime.UnixMilli())
	assert.Equal(t, 24*time.Hour, d.CloseTime.Sub(*d.OpenTime))
}

func TestTicker24hSwapUsesBaseVolume(t *testing.T) {
	var req http.Request
	c := staticServer(t, `{"code":"0","msg":"","data":[{"instType":"SWAP","instId":"BTC-USDT-SWAP","last":"100","open24h":"100","high24h":"101","low24h":"99","volCcy24h":"7","vol24h":"700","ts":"1791028289971"}]}`, &req)
	resp, err := c.Ticker24h(context.Background(), "BTC-USDT", model.MarketFutures)
	require.NoError(t, err)
	assert.Equal(t, "BTC-USDT-SWAP", req.URL.Query().Get("instId"))
	assert.Equal(t, "BTC-USDT", resp.Data.Symbol)
	assert.Equal(t, 7.0, *resp.Data.Volume)
	assert.Equal(t, 700.0, *resp.Data.QuoteVolume)
}

func TestTicker24hEmptyIsNotFound(t *testing.T) {
	c := staticServer(t, `{"code":"0","msg":"","data":[]}`, nil)
	_, err := c.Ticker24h(context.Background(), "BTC-USDT", model.MarketSpot)
	var pe *model.ProviderError
	require.ErrorAs(t, err, &pe)
	assert.Equal(t, model.ErrKindNotFound, pe.Kind)
}
