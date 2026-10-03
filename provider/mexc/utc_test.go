package mexc

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/model"
)

func TestFuturesCandles_UTCAndLimitClamp(t *testing.T) {
	var gotLimit string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLimit = r.URL.Query().Get("limit")
		gotPath = r.URL.Path
		fmt.Fprint(w, `{"success":true,"code":0,"data":{"time":[1761876000],"open":[1],"close":[2],"high":[3],"low":[0.5],"vol":[7]}}`)
	}))
	defer srv.Close()
	c := NewClient(config.MEXCConfig{BaseURL: srv.URL})

	for _, tc := range []struct {
		in   int
		want string
	}{{5000, "2000"}, {100, "100"}} {
		limit := tc.in
		resp, err := c.Candles(context.Background(), "BTC_USDT", model.MarketFutures, "1m", model.CandleOpts{Limit: &limit})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(gotPath, "/kline/BTC_USDT") {
			t.Fatalf("unexpected path %q", gotPath)
		}
		if gotLimit != tc.want {
			t.Fatalf("limit %d sent as %q, want %q", tc.in, gotLimit, tc.want)
		}
		if len(resp.Data) != 1 || resp.Data[0].OpenTime.Location() != time.UTC {
			t.Fatalf("futures candle not in UTC: %+v", resp.Data)
		}
	}
}

func TestSpotCandles_UTC(t *testing.T) {
	var reqs []string
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	srv := fakeKlines(base.UnixMilli(), base.UnixMilli(), &reqs)
	defer srv.Close()
	c := NewClient(config.MEXCConfig{BaseURL: srv.URL})
	to := base.Add(time.Minute)
	resp, err := c.Candles(context.Background(), "BTCUSDT", model.MarketSpot, "1m", model.CandleOpts{From: &base, To: &to})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 1 || resp.Data[0].OpenTime.Location() != time.UTC {
		t.Fatalf("spot candle not in UTC: %+v", resp.Data)
	}
}
