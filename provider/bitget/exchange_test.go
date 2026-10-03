package bitget

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

func exchangeInfoClient(t *testing.T, body string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return NewClient(config.BitgetConfig{BaseURL: srv.URL})
}

func TestSpotExchangeInfoMinNotional(t *testing.T) {
	c := exchangeInfoClient(t, `{"code":"00000","msg":"success","data":[{"symbol":"BTCUSDT","baseCoin":"BTC","quoteCoin":"USDT","minTradeUSDT":"5","minTradeAmount":"0.0001","status":"online","pricePrecision":"2","quantityPrecision":"6"}]}`)
	resp, err := c.ExchangeInfo(context.Background(), model.MarketSpot)
	require.NoError(t, err)
	require.Len(t, resp.Data.Symbols, 1)
	s := resp.Data.Symbols[0]
	require.NotNil(t, s.MinNotional)
	assert.Equal(t, 5.0, *s.MinNotional)
	assert.Nil(t, s.MinPrice, "minTradeUSDT is a notional, not a price")
	require.NotNil(t, s.MinQty)
	assert.Equal(t, 0.0001, *s.MinQty)
}

func TestFuturesExchangeInfoMinNotional(t *testing.T) {
	c := exchangeInfoClient(t, `{"code":"00000","msg":"success","data":[{"symbol":"BTCUSDT","baseCoin":"BTC","quoteCoin":"USDT","minTradeUSDT":"5","minTradeNum":"0.01","symbolStatus":"normal","pricePlace":"1","volumePlace":"2"}]}`)
	resp, err := c.ExchangeInfo(context.Background(), model.MarketFutures)
	require.NoError(t, err)
	require.Len(t, resp.Data.Symbols, 1)
	s := resp.Data.Symbols[0]
	require.NotNil(t, s.MinNotional)
	assert.Equal(t, 5.0, *s.MinNotional)
	assert.Nil(t, s.MinPrice)
}
