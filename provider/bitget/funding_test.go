package bitget

import (
	"context"
	"fmt"
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

// fundingServer serves total synthetic rows newest-first, 8h apart, paged like Bitget.
func fundingServer(t *testing.T, total int, newest time.Time, pages *int) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "/api/v2/mix/market/history-fund-rate", r.URL.Path)
		assert.Equal(t, "USDT-FUTURES", q.Get("productType"))
		*pages++
		var page, size int
		_, _ = fmt.Sscan(q.Get("pageNo"), &page)
		_, _ = fmt.Sscan(q.Get("pageSize"), &size)
		body := `{"code":"00000","msg":"success","data":[`
		first := true
		for i := (page - 1) * size; i < page*size && i < total; i++ {
			ts := newest.Add(-time.Duration(i) * 8 * time.Hour).UnixMilli()
			if !first {
				body += ","
			}
			first = false
			body += fmt.Sprintf(`{"symbol":"BTCUSDT","fundingRate":"0.0001%02d","fundingTime":"%d"}`, i%100, ts)
		}
		_, _ = w.Write([]byte(body + `]}`))
	}))
	t.Cleanup(srv.Close)
	return NewClient(config.BitgetConfig{BaseURL: srv.URL})
}

func TestFundingRatesAscendingAndWindow(t *testing.T) {
	newest := time.Date(2025, 4, 10, 16, 0, 0, 0, time.UTC)
	var pages int
	c := fundingServer(t, 250, newest, &pages)

	from := newest.Add(-30 * 8 * time.Hour)
	to := newest.Add(-10 * 8 * time.Hour)
	resp, err := c.FundingRates(context.Background(), "BTCUSDT", model.FundingRateOpts{From: &from, To: &to})
	require.NoError(t, err)
	assert.Equal(t, model.KindFundingRate, resp.Kind)
	assert.Equal(t, model.MarketFutures, resp.Market)
	require.Len(t, resp.Data, 21)
	assert.True(t, resp.Data[0].Time.Equal(from))
	assert.True(t, resp.Data[20].Time.Equal(to))
	for i := 1; i < len(resp.Data); i++ {
		assert.True(t, resp.Data[i].Time.After(resp.Data[i-1].Time))
	}
	assert.Equal(t, 1, pages, "window inside first page needs one request")
}

func TestFundingRatesPaginatesAndLimit(t *testing.T) {
	newest := time.Date(2025, 4, 10, 16, 0, 0, 0, time.UTC)
	var pages int
	c := fundingServer(t, 250, newest, &pages)

	from := newest.Add(-150 * 8 * time.Hour)
	resp, err := c.FundingRates(context.Background(), "BTCUSDT", model.FundingRateOpts{From: &from})
	require.NoError(t, err)
	require.Len(t, resp.Data, 151)
	assert.Equal(t, 2, pages)
	assert.Equal(t, 0.000100, resp.Data[150].Rate)

	// No From: most recent Limit entries, ascending.
	pages = 0
	limit := 3
	resp, err = c.FundingRates(context.Background(), "BTCUSDT", model.FundingRateOpts{Limit: &limit})
	require.NoError(t, err)
	require.Len(t, resp.Data, 3)
	assert.True(t, resp.Data[2].Time.Equal(newest))
	assert.True(t, resp.Data[0].Time.Before(resp.Data[2].Time))

	// From + Limit: earliest Limit entries at or after From.
	resp, err = c.FundingRates(context.Background(), "BTCUSDT", model.FundingRateOpts{From: &from, Limit: &limit})
	require.NoError(t, err)
	require.Len(t, resp.Data, 3)
	assert.True(t, resp.Data[0].Time.Equal(from))
}

func TestFundingRatesErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":"00000","msg":"ok","data":[{"symbol":"X","fundingRate":"abc","fundingTime":"1"}]}`))
	}))
	t.Cleanup(srv.Close)
	c := NewClient(config.BitgetConfig{BaseURL: srv.URL})
	_, err := c.FundingRates(context.Background(), "X", model.FundingRateOpts{})
	var pe *model.ProviderError
	require.ErrorAs(t, err, &pe)
	assert.Equal(t, model.ErrKindParse, pe.Kind)
}

func TestFundingRatesCapability(t *testing.T) {
	c := NewClient(config.BitgetConfig{Futures: config.MarketConfig{Enabled: true}})
	assert.True(t, c.Capabilities()[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureFundingRates}])
	assert.False(t, c.Capabilities()[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureFundingRates}])
}
