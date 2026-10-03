package bybiteu

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider"
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

var _ provider.PriceProvider = (*Client)(nil)
var _ provider.OrderBookProvider = (*Client)(nil)
var _ provider.TickerProvider = (*Client)(nil)
var _ provider.CandleProvider = (*Client)(nil)
var _ provider.ExchangeProvider = (*Client)(nil)

func TestCapabilities(t *testing.T) {
	m := NewClient(Config{}).Capabilities()
	for _, mk := range []capability.MarketType{capability.MarketSpot, capability.MarketMargin} {
		for _, f := range []capability.Feature{capability.FeatureServerTime, capability.FeatureExchangeInfo, capability.FeaturePrice, capability.FeatureTicker24h, capability.FeatureOrderBook, capability.FeatureCandles} {
			if !m[capability.CapabilityKey{Market: mk, Feature: f}] {
				t.Errorf("%v %v missing", mk, f)
			}
		}
	}
	if m[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureCandles}] || m[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureFundingRates}] {
		t.Error("futures must not be supported")
	}
}

const tickerJSON = `{"retCode":0,"retMsg":"OK","result":{"category":"spot","list":[{"symbol":"BTCUSDT","bid1Price":"84760.6","bid1Size":"0.6","ask1Price":"84760.7","ask1Size":"0.06","lastPrice":"84760.7","prevPrice24h":"84000","price24hPcnt":"0.0090","highPrice24h":"85000","lowPrice24h":"83000","turnover24h":"1000","volume24h":"12.5"}]},"time":1791029915078}`

func TestPrice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/v5/market/tickers" || q.Get("category") != "spot" {
			t.Errorf("bad request %s", r.URL)
		}
		if q.Get("symbol") != "BTCUSDT" {
			fmt.Fprint(w, `{"retCode":0,"result":{"list":[]},"time":1}`)
			return
		}
		fmt.Fprint(w, tickerJSON)
	}))
	defer srv.Close()
	resp, err := NewClient(Config{BaseURL: srv.URL}).Price(context.Background(), []string{"btc-usdt", "NOPEUSDT"}, "usd")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 1 || len(resp.Errors) != 1 {
		t.Fatalf("data=%d errors=%d", len(resp.Data), len(resp.Errors))
	}
	p := resp.Data[0]
	if p.ID != "BTCUSDT" || p.Symbol != "BTC-USDT" || p.OriginalSymbol != "btc-usdt" || p.Price != 84760.7 || math.Abs(*p.Change24h-0.9) > 1e-9 || *p.BidSize != 0.6 || p.Time.Location() != time.UTC {
		t.Errorf("bad price %+v", p)
	}
}

func TestOrderBook(t *testing.T) {
	var query string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		if r.URL.Path != "/v5/market/orderbook" {
			t.Errorf("path %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"retCode":0,"retMsg":"OK","result":{"s":"BTCUSDT","a":[["84760.7","0.06"]],"b":[["84760.6","0.6"],["84760.4","0.1"]],"ts":1791029914954,"u":55987382,"seq":1,"cts":1791029914950},"time":1}`)
	}))
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL})
	for _, mk := range []model.MarketType{model.MarketSpot, model.MarketMargin} {
		resp, err := c.OrderBook(context.Background(), "BTC-USDT", mk, 1000)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(query, "category=spot") || !strings.Contains(query, "limit=200") || !strings.Contains(query, "symbol=BTCUSDT") {
			t.Errorf("bad query %s", query)
		}
		d := resp.Data
		if resp.Market != mk || d.Market != mk || len(d.Bids) != 2 || len(d.Asks) != 1 || d.Bids[0].Price != 84760.6 || *d.LastUpdateID != 55987382 {
			t.Errorf("bad book %+v", d)
		}
		if d.Time == nil || d.Time.UnixMilli() != 1791029914950 || d.Time.Location() != time.UTC {
			t.Errorf("bad time %v", d.Time)
		}
	}
	if _, err := c.OrderBook(context.Background(), "BTCUSDT", "", 0); err != nil || !strings.Contains(query, "limit=20") {
		t.Errorf("default depth: %s %v", query, err)
	}
	if _, err := c.OrderBook(context.Background(), "BTCUSDT", model.MarketFutures, 5); err == nil {
		t.Error("futures must fail")
	}
}

func TestMarginExchangeInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"retCode":0,"result":{"list":[{"symbol":"BTCUSDT","baseCoin":"BTC","quoteCoin":"USDT","status":"Trading","marginTrading":"utaOnly","lotSizeFilter":{"basePrecision":"0.000001","minOrderQty":"0.000001","maxOrderQty":"230","minOrderAmt":"5"},"priceFilter":{"tickSize":"0.1"}},{"symbol":"XYZUSDT","baseCoin":"XYZ","quoteCoin":"USDT","status":"Trading","marginTrading":"none","lotSizeFilter":{"basePrecision":"1"},"priceFilter":{"tickSize":"0.01"}}]},"time":1}`)
	}))
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL})
	m, err := c.ExchangeInfo(context.Background(), model.MarketMargin)
	if err != nil || len(m.Data.Symbols) != 1 || m.Data.Symbols[0].Market != model.MarketMargin || m.Data.Market != model.MarketMargin {
		t.Fatalf("margin: %+v %v", m.Data, err)
	}
	s, err := c.ExchangeInfo(context.Background(), model.MarketSpot)
	if err != nil || len(s.Data.Symbols) != 2 {
		t.Fatalf("spot: %v", err)
	}
	if _, err := c.ExchangeInfo(context.Background(), model.MarketFutures); err == nil {
		t.Error("futures must fail")
	}
}
