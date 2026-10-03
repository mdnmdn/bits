package kraken

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
	if got := NewClient(Config{}).cfg.BaseURL; got != "https://futures.kraken.com" {
		t.Errorf("default base url %q", got)
	}
}

func TestSymbolRules(t *testing.T) {
	native := map[string]string{
		"PF_XBTUSD": "PF_XBTUSD", "pf_ethusd": "PF_ETHUSD", "PI_XBTUSD": "PI_XBTUSD",
		"BTC-USD": "PF_XBTUSD", "BTCUSD": "PF_XBTUSD", "XBT/USD": "PF_XBTUSD",
		"ETH_USD": "PF_ETHUSD", "sol-usd": "PF_SOLUSD", "FF_SOLUSD_261225": "FF_SOLUSD_261225",
	}
	for in, want := range native {
		if got := toNative(in); got != want {
			t.Errorf("toNative(%q)=%q want %q", in, got, want)
		}
	}
	norm := map[string]string{"PF_XBTUSD": "BTC-USD", "PF_ETHUSD": "ETH-USD", "FF_SOLUSD_261225": "SOL-USD"}
	for in, want := range norm {
		if got := normalize(in); got != want {
			t.Errorf("normalize(%q)=%q want %q", in, got, want)
		}
	}
}

// fakeCharts serves 1m candles in [first, last] (ms) like the venue: from and
// to in seconds and inclusive, oldest first, at most count rows.
func fakeCharts(t *testing.T, first, last int64, requests *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/charts/v1/trade/PF_XBTUSD/1m" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		q := r.URL.Query()
		*requests = append(*requests, r.URL.RawQuery)
		from, _ := strconv.ParseInt(q.Get("from"), 10, 64)
		to, _ := strconv.ParseInt(q.Get("to"), 10, 64)
		count, _ := strconv.Atoi(q.Get("count"))
		var rows []string
		for ts := first; ts <= last && len(rows) < count; ts += 60000 {
			if ts < from*1000 || ts > to*1000 {
				continue
			}
			rows = append(rows, fmt.Sprintf(`{"time":%d,"open":"1","high":"2","low":"0.5","close":"1.5","volume":"3"}`, ts))
		}
		fmt.Fprintf(w, `{"candles":[%s],"more_candles":false}`, strings.Join(rows, ","))
	}))
}

func TestCandles_RangePaging(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	const total, page = 5000, 2000
	var reqs []string
	srv := fakeCharts(t, base.UnixMilli(), base.Add((total-1)*time.Minute).UnixMilli(), &reqs)
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL})

	to := base.Add(total * time.Minute)
	limit := 9999 // above the venue cap: clamped to 2000
	var got []model.Candle
	for cur := base; cur.Before(to); cur = cur.Add(page * time.Minute) {
		end := cur.Add(page * time.Minute)
		if end.After(to) {
			end = to
		}
		resp, err := c.Candles(context.Background(), "BTC-USD", model.MarketFutures, "1m", model.CandleOpts{From: &cur, To: &end, Limit: &limit})
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
		if cd.Volume == nil || *cd.Volume != 3 || cd.Close != 1.5 {
			t.Fatalf("candle %d bad values %+v", i, cd)
		}
	}
	if len(reqs) != 3 {
		t.Fatalf("got %d requests, want 3", len(reqs))
	}
	for _, r := range reqs {
		if !strings.Contains(r, "from=") || !strings.Contains(r, "to=") || !strings.Contains(r, "count=2000") {
			t.Errorf("bad query: %s", r)
		}
	}
}

func TestCandles_EmptyAndErrors(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	var reqs []string
	srv := fakeCharts(t, base.UnixMilli(), base.Add(10*time.Minute).UnixMilli(), &reqs)
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL})
	from, to := base.Add(-time.Hour), base.Add(-30*time.Minute)
	resp, err := c.Candles(context.Background(), "PF_XBTUSD", model.MarketFutures, "1m", model.CandleOpts{From: &from, To: &to})
	if err != nil || len(resp.Data) != 0 {
		t.Fatalf("got %d candles, err %v; want none", len(resp.Data), err)
	}
	if _, err := c.Candles(context.Background(), "PF_XBTUSD", model.MarketFutures, "7m", model.CandleOpts{}); err == nil {
		t.Error("want error for unsupported interval")
	}
	if _, err := c.Candles(context.Background(), "PF_XBTUSD", model.MarketSpot, "1m", model.CandleOpts{}); err == nil {
		t.Error("want error for spot")
	}
}

func TestTicker24h(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/derivatives/api/v3/tickers/PF_XBTUSD" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"result":"success","ticker":{"symbol":"PF_XBTUSD","last":100,"bid":99,"ask":101,"open24h":125,"high24h":130,"low24h":90,"vol24h":50,"volumeQuote":5000,"vwap24h":110,"change24h":-20}}`)
	}))
	defer srv.Close()
	resp, err := NewClient(Config{BaseURL: srv.URL}).Ticker24h(context.Background(), "BTC-USD", model.MarketFutures)
	if err != nil {
		t.Fatal(err)
	}
	d := resp.Data
	if d.Symbol != "BTC-USD" || d.OriginalSymbol != "PF_XBTUSD" || d.LastPrice != 100 {
		t.Errorf("bad ticker %+v", d)
	}
	if *d.PriceChange != -25 || *d.PriceChangePercent != -20 || *d.HighPrice != 130 || *d.LowPrice != 90 || *d.Volume != 50 || *d.QuoteVolume != 5000 || *d.OpenPrice != 125 || *d.WeightedAvgPrice != 110 {
		t.Errorf("bad 24h stats %+v", d)
	}
}

func TestExchangeInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/derivatives/api/v3/instruments":
			fmt.Fprint(w, `{"result":"success","serverTime":"2026-10-03T11:00:00.000Z","instruments":[
{"symbol":"PF_XBTUSD","type":"flexible_futures","base":"BTC","quote":"USD","tickSize":1,"contractSize":1,"contractValueTradePrecision":4,"tradeable":true,"maxPositionSize":1200,"feeScheduleUid":"s1"},
{"symbol":"PF_PEPEUSD","type":"flexible_futures","base":"PEPE","quote":"USD","tickSize":0.0000001,"contractSize":1,"contractValueTradePrecision":-3,"tradeable":false,"feeScheduleUid":"s1"}]}`)
		case "/derivatives/api/v3/feeschedules":
			fmt.Fprint(w, `{"result":"success","serverTime":"2026-10-03T11:00:00.000Z","feeSchedules":[{"uid":"s1","name":"x","tiers":[{"makerFee":0.02,"takerFee":0.05,"usdVolume":0}]}]}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()
	resp, err := NewClient(Config{BaseURL: srv.URL}).ExchangeInfo(context.Background(), model.MarketFutures)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data.Symbols) != 2 || resp.Data.ServerTime == nil {
		t.Fatalf("bad info %+v", resp.Data)
	}
	s := resp.Data.Symbols[0]
	if s.NormalizedSymbol != "BTC-USD" || s.Status != model.SymbolStatusTrading || s.Extra["tick_size"] != 1.0 ||
		*s.QtyPrecision != 4 || *s.StepSize < 0.00009999 || *s.StepSize > 0.00010001 || *s.PricePrecision != 0 ||
		*s.MaxQty != 1200 || s.MinQty != nil || s.MinNotional != nil {
		t.Errorf("bad symbol %+v", s)
	}
	if *s.MakerFee != 0.0002 || *s.TakerFee != 0.0005 {
		t.Errorf("bad fees %v %v", *s.MakerFee, *s.TakerFee)
	}
	p := resp.Data.Symbols[1]
	if p.Status != model.SymbolStatusHalt || *p.StepSize != 1000 || *p.QtyPrecision != 0 || *p.PricePrecision != 7 {
		t.Errorf("bad pepe %+v step %v", p, *p.StepSize)
	}
}

func TestAPIErrorKinds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"result":"error","error":"apiLimitExceeded"}`)
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
		fmt.Fprint(w, `{"result":"success","serverTime":"2026-10-03T11:55:28.19Z","feeSchedules":[]}`)
	}))
	defer srv.Close()
	resp, err := NewClient(Config{BaseURL: srv.URL}).ServerTime(context.Background())
	want := time.Date(2026, 10, 3, 11, 55, 28, 190000000, time.UTC)
	if err != nil || !resp.Data.Time.Equal(want) {
		t.Fatalf("bad time %v %v", resp.Data.Time, err)
	}
}
