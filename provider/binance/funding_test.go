package binance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/model"
)

func TestFundingRates(t *testing.T) {
	var got *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(r.Context())
		_, _ = w.Write([]byte(`[
{"symbol":"BTCUSDT","fundingRate":"-0.00003755","fundingTime":1745654400000,"markPrice":"94000.10"},
{"symbol":"BTCUSDT","fundingRate":"0.00010000","fundingTime":1745683200000,"markPrice":""}]`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(config.BinanceConfig{Futures: config.MarketConfig{Enabled: true}})
	c.futuresClient.BaseURL = srv.URL

	from := time.UnixMilli(1745600000000)
	to := time.UnixMilli(1745700000000)
	limit := 500
	resp, err := c.FundingRates(context.Background(), "btcusdt", model.FundingRateOpts{From: &from, To: &to, Limit: &limit})
	require.NoError(t, err)

	assert.Equal(t, "/fapi/v1/fundingRate", got.URL.Path)
	q := got.URL.Query()
	assert.Equal(t, "BTCUSDT", q.Get("symbol"))
	assert.Equal(t, "1745600000000", q.Get("startTime"))
	assert.Equal(t, "1745700000000", q.Get("endTime"))
	assert.Equal(t, "500", q.Get("limit"))

	assert.Equal(t, model.KindFundingRate, resp.Kind)
	assert.Equal(t, model.MarketFutures, resp.Market)
	require.Len(t, resp.Data, 2)
	assert.Equal(t, -0.00003755, resp.Data[0].Rate)
	assert.Equal(t, int64(1745654400000), resp.Data[0].Time.UnixMilli())
	require.NotNil(t, resp.Data[0].MarkPrice)
	assert.Equal(t, 94000.10, *resp.Data[0].MarkPrice)
	assert.Nil(t, resp.Data[1].MarkPrice)
	assert.True(t, resp.Data[1].Time.After(resp.Data[0].Time))
}

func TestFundingRatesNeedsFutures(t *testing.T) {
	c := NewClient(config.BinanceConfig{Spot: config.MarketConfig{Enabled: true}})
	_, err := c.FundingRates(context.Background(), "BTCUSDT", model.FundingRateOpts{})
	var pe *model.ProviderError
	require.ErrorAs(t, err, &pe)
	assert.Equal(t, model.ErrKindUnsupportedMarket, pe.Kind)
	assert.False(t, c.Capabilities()[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureFundingRates}])
}
