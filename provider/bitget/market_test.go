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

func newCandleServer(t *testing.T, body string, got *http.Request) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got != nil {
			*got = *r.Clone(r.Context())
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return NewClient(config.BitgetConfig{BaseURL: srv.URL})
}

func TestGranularityDayIsUTC(t *testing.T) {
	assert.Equal(t, "1Dutc", convertGranularitySpot("1d"))
	assert.Equal(t, "1Dutc", convertGranularityFutures("1d"))
}

func TestCandlesDayGranularityQuery(t *testing.T) {
	for _, mk := range []model.MarketType{model.MarketSpot, model.MarketFutures} {
		var req http.Request
		c := newCandleServer(t, `{"code":"00000","msg":"success","data":[["1756684800000","1","2","0.5","1.5","10","15"]]}`, &req)
		resp, err := c.Candles(context.Background(), "BTCUSDT", mk, "1d", model.CandleOpts{})
		require.NoError(t, err)
		assert.Equal(t, "1Dutc", req.URL.Query().Get("granularity"))
		require.Len(t, resp.Data, 1)
		assert.Equal(t, 0, resp.Data[0].OpenTime.UTC().Hour())
	}
}

func TestCandlesMalformedNumberReturnsError(t *testing.T) {
	rows := map[string]string{
		"timestamp": `["abc","1","2","0.5","1.5","10"]`,
		"open":      `["1756684800000","x","2","0.5","1.5","10"]`,
		"volume":    `["1756684800000","1","2","0.5","1.5","NaNx"]`,
	}
	for name, row := range rows {
		t.Run(name, func(t *testing.T) {
			c := newCandleServer(t, `{"code":"00000","msg":"success","data":[`+row+`]}`, nil)
			_, err := c.Candles(context.Background(), "BTCUSDT", model.MarketSpot, "1h", model.CandleOpts{})
			require.Error(t, err)
			var pe *model.ProviderError
			require.ErrorAs(t, err, &pe)
			assert.Equal(t, model.ErrKindParse, pe.Kind)
		})
	}
}
