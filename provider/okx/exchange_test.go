package okx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/model"
)

func newServer(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := NewClient(config.OKXConfig{BaseURL: srv.URL})
	c.pageDelay = 0
	return c
}

func staticServer(t *testing.T, body string, got *http.Request) *Client {
	t.Helper()
	return newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got != nil {
			*got = *r.Clone(r.Context())
		}
		_, _ = w.Write([]byte(body))
	})
}

func TestInstID(t *testing.T) {
	assert.Equal(t, "BTC-USDT", instID("BTCUSDT", model.MarketSpot))
	assert.Equal(t, "BTC-USDT", instID("btc-usdt", model.MarketSpot))
	assert.Equal(t, "BTC-USDT-SWAP", instID("BTCUSDT", model.MarketFutures))
	assert.Equal(t, "BTC-USDT-SWAP", instID("BTC-USDT-SWAP", model.MarketFutures))
	assert.Equal(t, "BTC-USDT", instID("BTC-USDT-SWAP", model.MarketSpot))
}

func TestServerTime(t *testing.T) {
	c := staticServer(t, `{"code":"0","msg":"","data":[{"ts":"1791028266281"}]}`, nil)
	resp, err := c.ServerTime(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1791028266281), resp.Data.Time.UnixMilli())
	assert.Equal(t, "UTC", resp.Data.Time.Location().String())
}

func TestSpotExchangeInfo(t *testing.T) {
	var req http.Request
	c := staticServer(t, `{"code":"0","msg":"","data":[
{"instId":"BTC-EUR","instType":"SPOT","baseCcy":"BTC","quoteCcy":"EUR","tickSz":"0.1","lotSz":"0.00000001","minSz":"0.0001","maxLmtSz":"9999999999","state":"live"},
{"instId":"XYZ-USDT","instType":"SPOT","baseCcy":"XYZ","quoteCcy":"USDT","tickSz":"0.01","lotSz":"1","minSz":"1","state":"suspend"}]}`, &req)
	resp, err := c.ExchangeInfo(context.Background(), model.MarketSpot)
	require.NoError(t, err)
	assert.Equal(t, "SPOT", req.URL.Query().Get("instType"))
	require.Len(t, resp.Data.Symbols, 2)

	s := resp.Data.Symbols[0]
	assert.Equal(t, "BTC-EUR", s.Symbol)
	assert.Equal(t, "BTC-EUR", s.NormalizedSymbol)
	assert.Equal(t, "BTC", s.BaseAsset)
	assert.Equal(t, "EUR", s.QuoteAsset)
	assert.Equal(t, model.SymbolStatusTrading, s.Status)
	require.NotNil(t, s.PricePrecision)
	assert.Equal(t, 1, *s.PricePrecision)
	require.NotNil(t, s.QtyPrecision)
	assert.Equal(t, 8, *s.QtyPrecision)
	require.NotNil(t, s.MinQty)
	assert.Equal(t, 0.0001, *s.MinQty)
	require.NotNil(t, s.StepSize)
	assert.Equal(t, 1e-8, *s.StepSize)
	assert.Equal(t, 0.1, s.Extra["tick_size"])
	assert.Nil(t, s.MinNotional, "OKX publishes no min notional")
	assert.Nil(t, s.MakerFee)
	assert.Nil(t, s.TakerFee)

	assert.Equal(t, model.SymbolStatusHalt, resp.Data.Symbols[1].Status)
	require.NotNil(t, resp.Data.Symbols[1].QtyPrecision)
	assert.Equal(t, 0, *resp.Data.Symbols[1].QtyPrecision)
}

func TestSwapExchangeInfo(t *testing.T) {
	var req http.Request
	c := staticServer(t, `{"code":"0","msg":"","data":[
{"instId":"BTC-USDT-SWAP","instType":"SWAP","uly":"BTC-USDT","ctType":"linear","ctVal":"0.01","ctValCcy":"BTC","settleCcy":"USDT","tickSz":"0.1","lotSz":"0.01","minSz":"0.01","maxLmtSz":"100000000","state":"live"},
{"instId":"BTC-USD-SWAP","instType":"SWAP","uly":"BTC-USD","ctType":"inverse","ctVal":"100","ctValCcy":"USD","settleCcy":"BTC","tickSz":"0.1","lotSz":"0.1","minSz":"0.1","state":"live"}]}`, &req)
	resp, err := c.ExchangeInfo(context.Background(), model.MarketFutures)
	require.NoError(t, err)
	assert.Equal(t, "SWAP", req.URL.Query().Get("instType"))
	require.Len(t, resp.Data.Symbols, 2)

	s := resp.Data.Symbols[0]
	assert.Equal(t, "BTC-USDT-SWAP", s.Symbol)
	assert.Equal(t, "BTC-USDT", s.NormalizedSymbol)
	assert.Equal(t, "BTC", s.BaseAsset)
	assert.Equal(t, "USDT", s.QuoteAsset)
	require.NotNil(t, s.MinQty)
	assert.Equal(t, 0.01, *s.MinQty)
	assert.Equal(t, 0.01, s.Extra["ct_val"])
	assert.Equal(t, "BTC", s.Extra["ct_val_ccy"])
	assert.Equal(t, "linear", s.Extra["ct_type"])
	assert.Equal(t, "USD", resp.Data.Symbols[1].QuoteAsset)
}

func TestAPIErrorKinds(t *testing.T) {
	c := staticServer(t, `{"code":"50011","msg":"Too Many Requests","data":[]}`, nil)
	_, err := c.ExchangeInfo(context.Background(), model.MarketSpot)
	var pe *model.ProviderError
	require.ErrorAs(t, err, &pe)
	assert.Equal(t, model.ErrKindRateLimit, pe.Kind)
	assert.Equal(t, "50011", pe.ProviderCode)

	c = newServer(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTooManyRequests) })
	_, err = c.ExchangeInfo(context.Background(), model.MarketSpot)
	require.ErrorAs(t, err, &pe)
	assert.Equal(t, model.ErrKindRateLimit, pe.Kind)
}
