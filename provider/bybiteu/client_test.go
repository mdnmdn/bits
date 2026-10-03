package bybiteu

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mdnmdn/bits/model"
)

func TestDefaultBaseURL(t *testing.T) {
	if got := NewClient(Config{}).cfg.BaseURL; got != "https://api.bybit.eu" {
		t.Errorf("default base url %q", got)
	}
	if got := NewClient(Config{BaseURL: "http://x"}).cfg.BaseURL; got != "http://x" {
		t.Errorf("override lost: %q", got)
	}
}

func TestToNative(t *testing.T) {
	for in, want := range map[string]string{"BTCUSDT": "BTCUSDT", "btc-usdt": "BTCUSDT", "BTC_USDT": "BTCUSDT", "eth/usdc": "ETHUSDC"} {
		if got := toNative(in); got != want {
			t.Errorf("toNative(%q)=%q want %q", in, got, want)
		}
	}
}

// fakeKline serves 1m candles in [first, last] (ms) like the venue: start and
// end inclusive, newest first, at most limit rows.
func fakeKline(t *testing.T, first, last int64, queries *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		*queries = append(*queries, r.URL.RawQuery)
		start, _ := strconv.ParseInt(q.Get("start"), 10, 64)
		end, _ := strconv.ParseInt(q.Get("end"), 10, 64)
		limit, _ := strconv.Atoi(q.Get("limit"))
		var rows []string
		for ts := last; ts >= first && len(rows) < limit; ts -= 60000 {
			if ts < start || ts > end {
				continue
			}
			rows = append(rows, fmt.Sprintf(`["%d","1","2","0.5","1.5","3","4"]`, ts))
		}
		fmt.Fprintf(w, `{"retCode":0,"retMsg":"OK","result":{"category":"spot","symbol":"BTCUSDT","list":[%s]},"time":1}`, strings.Join(rows, ","))
	}))
}

func TestCandles_RangePaging(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	const total, page = 2500, 1000
	var qs []string
	srv := fakeKline(t, base.UnixMilli(), base.Add((total-1)*time.Minute).UnixMilli(), &qs)
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL})

	to := base.Add(total * time.Minute)
	limit := 5000 // above the venue cap: clamped to 1000
	var got []model.Candle
	for cur := base; cur.Before(to); cur = cur.Add(page * time.Minute) {
		end := cur.Add(page * time.Minute)
		if end.After(to) {
			end = to
		}
		resp, err := c.Candles(context.Background(), "BTC-USDT", model.MarketSpot, "1m", model.CandleOpts{From: &cur, To: &end, Limit: &limit})
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, resp.Data...)
	}
	if len(got) != total {
		t.Fatalf("got %d candles, want %d", len(got), total)
	}
	for i, cd := range got {
		if want := base.Add(time.Duration(i) * time.Minute); !cd.OpenTime.Equal(want) || cd.OpenTime.Location() != time.UTC {
			t.Fatalf("candle %d at %s, want %s UTC", i, cd.OpenTime, want)
		}
		if cd.Volume == nil || *cd.Volume != 3 {
			t.Fatalf("candle %d volume %v", i, cd.Volume)
		}
	}
	if len(qs) != 3 {
		t.Fatalf("got %d requests, want 3", len(qs))
	}
	for _, q := range qs {
		if !strings.Contains(q, "symbol=BTCUSDT") || !strings.Contains(q, "interval=1&") || !strings.Contains(q, "limit=1000") || !strings.Contains(q, "start=") || !strings.Contains(q, "end=") {
			t.Errorf("bad query: %s", q)
		}
	}
}

func TestCandles_NoOlderHistoryAndBadInterval(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	var qs []string
	srv := fakeKline(t, base.UnixMilli(), base.Add(10*time.Minute).UnixMilli(), &qs)
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL})
	from, to := base.Add(-time.Hour), base.Add(-30*time.Minute)
	resp, err := c.Candles(context.Background(), "BTCUSDT", model.MarketSpot, "1m", model.CandleOpts{From: &from, To: &to})
	if err != nil || len(resp.Data) != 0 {
		t.Fatalf("got %d candles, err %v; want none", len(resp.Data), err)
	}
	if _, err := c.Candles(context.Background(), "BTCUSDT", model.MarketSpot, "7m", model.CandleOpts{}); err == nil {
		t.Error("want error for unsupported interval")
	}
	if _, err := c.Candles(context.Background(), "BTCUSDT", model.MarketFutures, "1m", model.CandleOpts{}); err == nil {
		t.Error("want error for futures")
	}
}

func TestTicker24h(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v5/market/tickers" || r.URL.Query().Get("category") != "spot" || r.URL.Query().Get("symbol") != "BTCUSDT" {
			t.Errorf("unexpected request %s", r.URL)
		}
		fmt.Fprint(w, `{"retCode":0,"retMsg":"OK","result":{"category":"spot","list":[{"symbol":"BTCUSDT","bid1Price":"99","ask1Price":"101","lastPrice":"100","prevPrice24h":"125","price24hPcnt":"-0.2","highPrice24h":"130","lowPrice24h":"90","turnover24h":"5000","volume24h":"50"}]}}`)
	}))
	defer srv.Close()
	resp, err := NewClient(Config{BaseURL: srv.URL}).Ticker24h(context.Background(), "BTC-USDT", model.MarketSpot)
	if err != nil {
		t.Fatal(err)
	}
	d := resp.Data
	if d.Symbol != "BTC-USDT" || d.OriginalSymbol != "BTCUSDT" || d.LastPrice != 100 {
		t.Errorf("bad ticker %+v", d)
	}
	if *d.PriceChange != -25 || *d.PriceChangePercent != -20 || *d.HighPrice != 130 || *d.LowPrice != 90 || *d.Volume != 50 || *d.QuoteVolume != 5000 || *d.OpenPrice != 125 {
		t.Errorf("bad 24h stats %+v", d)
	}
}

func TestExchangeInfo(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("cursor") == "" {
			fmt.Fprint(w, `{"retCode":0,"result":{"category":"spot","list":[{"symbol":"BTCUSDT","baseCoin":"BTC","quoteCoin":"USDT","status":"Trading","lotSizeFilter":{"basePrecision":"0.000001","minOrderQty":"0.000001","maxOrderQty":"230","minOrderAmt":"5"},"priceFilter":{"tickSize":"0.1"}}],"nextPageCursor":"c2"}}`)
			return
		}
		fmt.Fprint(w, `{"retCode":0,"result":{"category":"spot","list":[{"symbol":"ETHEUR","baseCoin":"ETH","quoteCoin":"EUR","status":"PreLaunch","lotSizeFilter":{"basePrecision":"0.00001","minOrderQty":"0.00001","minOrderAmt":"5"},"priceFilter":{"tickSize":"0.01"}}],"nextPageCursor":""}}`)
	}))
	defer srv.Close()
	resp, err := NewClient(Config{BaseURL: srv.URL}).ExchangeInfo(context.Background(), model.MarketSpot)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(resp.Data.Symbols) != 2 {
		t.Fatalf("calls %d symbols %d", calls, len(resp.Data.Symbols))
	}
	s := resp.Data.Symbols[0]
	if s.NormalizedSymbol != "BTC-USDT" || s.Status != model.SymbolStatusTrading || *s.StepSize != 0.000001 || *s.MinQty != 0.000001 ||
		*s.MinNotional != 5 || *s.PricePrecision != 1 || *s.QtyPrecision != 6 || s.Extra["tick_size"] != 0.1 || s.MakerFee != nil || s.TakerFee != nil {
		t.Errorf("bad symbol %+v", s)
	}
	if resp.Data.Symbols[1].Status != model.SymbolStatusHalt {
		t.Error("non-trading status must be halt")
	}
}

func TestAPIErrorKinds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"retCode":10006,"retMsg":"Too many visits"}`)
	}))
	defer srv.Close()
	_, err := NewClient(Config{BaseURL: srv.URL}).ServerTime(context.Background())
	var pe *model.ProviderError
	if !errors.As(err, &pe) || pe.Kind != model.ErrKindRateLimit {
		t.Fatalf("want rate limit ProviderError, got %v", err)
	}
}

func TestServerTime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"retCode":0,"result":{"timeSecond":"1791028509","timeNano":"1791028509381756764"},"time":1791028509381}`)
	}))
	defer srv.Close()
	resp, err := NewClient(Config{BaseURL: srv.URL}).ServerTime(context.Background())
	if err != nil || resp.Data.Time.Unix() != 1791028509 || resp.Data.Time.Location() != time.UTC {
		t.Fatalf("bad time %v %v", resp.Data.Time, err)
	}
}
