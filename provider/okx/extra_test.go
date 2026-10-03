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

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider"
)

var (
	_ provider.PriceProvider       = (*Client)(nil)
	_ provider.OrderBookProvider   = (*Client)(nil)
	_ provider.FundingRateProvider = (*Client)(nil)
)

func TestCapabilitiesAllMarkets(t *testing.T) {
	cfg := config.OKXConfig{Spot: config.MarketConfig{Enabled: true}, Futures: config.MarketConfig{Enabled: true}}
	m := NewClient(cfg).Capabilities()
	for _, mk := range []capability.MarketType{capability.MarketSpot, capability.MarketMargin, capability.MarketFutures} {
		for _, f := range []capability.Feature{capability.FeatureExchangeInfo, capability.FeaturePrice, capability.FeatureCandles, capability.FeatureTicker24h, capability.FeatureOrderBook} {
			assert.True(t, m[capability.CapabilityKey{Market: mk, Feature: f}], "%s %s", mk, f)
		}
	}
	assert.True(t, m[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureFundingRates}])
	assert.False(t, m[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureFundingRates}])
	assert.False(t, m[capability.CapabilityKey{Market: capability.MarketMargin, Feature: capability.FeatureFundingRates}])
}

const tickerBody = `{"code":"0","msg":"","data":[{"instId":"%s","last":"110","open24h":"100","high24h":"120","low24h":"90","bidPx":"109","askPx":"111","vol24h":"500","volCcy24h":"5","ts":"1790000000000"}]}`

func TestPriceSpotAndSwapAndItemError(t *testing.T) {
	var queries []string
	c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("instId")
		queries = append(queries, id)
		if id == "NOPE-USDT" {
			_, _ = w.Write([]byte(`{"code":"51001","msg":"Instrument ID does not exist","data":[]}`))
			return
		}
		_, _ = fmt.Fprintf(w, tickerBody, id)
	})
	resp, err := c.Price(context.Background(), []string{"BTCUSDT", "eth-usdt-swap", "NOPE-USDT"}, "usdt")
	require.NoError(t, err)
	assert.Equal(t, []string{"BTC-USDT", "ETH-USDT-SWAP", "NOPE-USDT"}, queries)
	require.Len(t, resp.Data, 2)
	require.Len(t, resp.Errors, 1)
	assert.Equal(t, "NOPE-USDT", resp.Errors[0].Symbol)

	spot := resp.Data[0]
	assert.Equal(t, "BTC-USDT", spot.Symbol)
	assert.Equal(t, "usdt", spot.Currency)
	assert.Equal(t, 110.0, spot.Price)
	assert.InDelta(t, 10.0, *spot.Change24h, 1e-9)
	assert.Equal(t, 500.0, *spot.Volume24h)

	swap := resp.Data[1]
	assert.Equal(t, "ETH-USDT", swap.Symbol)
	assert.Equal(t, "ETH-USDT-SWAP", swap.OriginalSymbol)
	assert.Equal(t, 5.0, *swap.Volume24h) // base coin, not contracts
}

func TestMarginUsesSpotEndpoints(t *testing.T) {
	var got []string
	c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.URL.Path+"?"+r.URL.Query().Get("instId")+r.URL.Query().Get("instType"))
		if strings.HasSuffix(r.URL.Path, "/instruments") {
			_, _ = w.Write([]byte(`{"code":"0","msg":"","data":[{"instId":"BTC-USDT","baseCcy":"BTC","quoteCcy":"USDT","tickSz":"0.1","lotSz":"0.001","minSz":"0.001","state":"live"}]}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/books") {
			_, _ = w.Write([]byte(`{"code":"0","msg":"","data":[{"asks":[],"bids":[],"ts":"1"}]}`))
			return
		}
		_, _ = fmt.Fprintf(w, tickerBody, r.URL.Query().Get("instId"))
	})
	ctx := context.Background()

	info, err := c.ExchangeInfo(ctx, model.MarketMargin)
	require.NoError(t, err)
	assert.Equal(t, model.MarketMargin, info.Data.Symbols[0].Market)
	tk, err := c.Ticker24h(ctx, "BTCUSDT", model.MarketMargin)
	require.NoError(t, err)
	assert.Equal(t, model.MarketMargin, tk.Data.Market)
	ob, err := c.OrderBook(ctx, "BTCUSDT", model.MarketMargin, 5)
	require.NoError(t, err)
	assert.Equal(t, model.MarketMargin, ob.Market)

	assert.Equal(t, []string{
		"/api/v5/public/instruments?MARGIN",
		"/api/v5/market/ticker?BTC-USDT",
		"/api/v5/market/books?BTC-USDT",
	}, got)
}

func TestOrderBook(t *testing.T) {
	var req http.Request
	c := staticServer(t, `{"code":"0","msg":"","data":[{"asks":[["101","2","0","3"],["102","1","0","1"]],"bids":[["100","5","0","4"]],"ts":"1790000000000","seqId":1}]}`, &req)
	resp, err := c.OrderBook(context.Background(), "BTCUSDT", model.MarketFutures, 1000)
	require.NoError(t, err)
	assert.Equal(t, "BTC-USDT-SWAP", req.URL.Query().Get("instId"))
	assert.Equal(t, "400", req.URL.Query().Get("sz")) // capped

	ob := resp.Data
	assert.Equal(t, "BTC-USDT", ob.Symbol)
	assert.Equal(t, []model.OrderBookEntry{{Price: 100, Quantity: 5}}, ob.Bids)
	require.Len(t, ob.Asks, 2)
	assert.Equal(t, 101.0, ob.Asks[0].Price)
	require.NotNil(t, ob.Time)
	assert.Equal(t, time.UTC, ob.Time.Location())

	_, err = c.OrderBook(context.Background(), "BTC-USDT", model.MarketSpot, 0)
	require.NoError(t, err)
	assert.Equal(t, "20", req.URL.Query().Get("sz"))
}

func TestOrderBookErrors(t *testing.T) {
	c := staticServer(t, `{"code":"0","msg":"","data":[]}`, nil)
	_, err := c.OrderBook(context.Background(), "BTC-USDT", model.MarketSpot, 5)
	require.Error(t, err)

	c = staticServer(t, `{"code":"0","msg":"","data":[{"asks":[["x","1"]],"bids":[],"ts":"1"}]}`, nil)
	_, err = c.OrderBook(context.Background(), "BTC-USDT", model.MarketSpot, 5)
	require.Error(t, err)
}

const fundingStep = int64(8 * 3600 * 1000)

// fundingServer serves settlements at first + k*fundingStep for k in [0, n),
// newest first, strictly older than "after", at most "limit" rows.
func fundingServer(t *testing.T, first int64, n int, queries *[]string) *Client {
	return newServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		*queries = append(*queries, r.URL.RawQuery)
		limit, _ := strconv.Atoi(q.Get("limit"))
		after := int64(1<<62 - 1)
		if a := q.Get("after"); a != "" {
			after, _ = strconv.ParseInt(a, 10, 64)
		}
		var rows []string
		for k := n - 1; k >= 0 && len(rows) < limit; k-- {
			ts := first + int64(k)*fundingStep
			if ts >= after {
				continue
			}
			rows = append(rows, fmt.Sprintf(`{"instId":"BTC-USDT-SWAP","fundingRate":"0.0002","realizedRate":"0.%04d","fundingTime":"%d"}`, k+1, ts))
		}
		_, _ = w.Write([]byte(`{"code":"0","msg":"","data":[` + strings.Join(rows, ",") + `]}`))
	})
}

func TestFundingRatesDefaultNewest(t *testing.T) {
	first := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	var queries []string
	c := fundingServer(t, first, 250, &queries)
	resp, err := c.FundingRates(context.Background(), "BTCUSDT", model.FundingRateOpts{})
	require.NoError(t, err)
	assert.Equal(t, model.MarketFutures, resp.Market)
	require.Len(t, resp.Data, 100)
	assert.Contains(t, queries[0], "instId=BTC-USDT-SWAP")
	assert.Equal(t, "BTC-USDT", resp.Data[0].Symbol)
	assert.True(t, resp.Data[0].Time.Before(resp.Data[99].Time))
	assert.Equal(t, time.UTC, resp.Data[0].Time.Location())
	assert.Equal(t, first+249*fundingStep, resp.Data[99].Time.UnixMilli())
	assert.InDelta(t, 0.0250, resp.Data[99].Rate, 1e-12) // realizedRate
	assert.Len(t, queries, 1)
}

func TestFundingRatesWindowPagesAndBounds(t *testing.T) {
	first := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	var queries []string
	c := fundingServer(t, first, 250, &queries)
	from := time.UnixMilli(first + 10*fundingStep)
	to := time.UnixMilli(first + 200*fundingStep) // inclusive
	resp, err := c.FundingRates(context.Background(), "BTC-USDT-SWAP", model.FundingRateOpts{From: &from, To: &to})
	require.NoError(t, err)
	require.Len(t, resp.Data, 191)
	assert.Equal(t, from.UnixMilli(), resp.Data[0].Time.UnixMilli())
	assert.Equal(t, to.UnixMilli(), resp.Data[190].Time.UnixMilli())
	assert.Len(t, queries, 2)

	// From with Limit: earliest entries at or after From.
	limit := 5
	resp, err = c.FundingRates(context.Background(), "BTC-USDT-SWAP", model.FundingRateOpts{From: &from, To: &to, Limit: &limit})
	require.NoError(t, err)
	require.Len(t, resp.Data, 5)
	assert.Equal(t, from.UnixMilli(), resp.Data[0].Time.UnixMilli())
}

func TestFundingRatesLimitWithoutFromAndEmpty(t *testing.T) {
	first := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	var queries []string
	c := fundingServer(t, first, 30, &queries)
	limit := 3
	resp, err := c.FundingRates(context.Background(), "BTC-USDT", model.FundingRateOpts{Limit: &limit})
	require.NoError(t, err)
	require.Len(t, resp.Data, 3)
	assert.Equal(t, first+29*fundingStep, resp.Data[2].Time.UnixMilli())

	to := time.UnixMilli(first - fundingStep)
	resp, err = c.FundingRates(context.Background(), "BTC-USDT", model.FundingRateOpts{To: &to})
	require.NoError(t, err)
	assert.NotNil(t, resp.Data)
	assert.Empty(t, resp.Data)
}

func TestFundingRatesAPIError(t *testing.T) {
	c := staticServer(t, `{"code":"51001","msg":"Instrument ID does not exist","data":[]}`, nil)
	_, err := c.FundingRates(context.Background(), "NOPE", model.FundingRateOpts{})
	require.Error(t, err)
}
