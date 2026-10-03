package bitvavo

import (
	"context"
	"fmt"
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

var _ provider.Provider = (*Client)(nil)
var _ provider.ExchangeProvider = (*Client)(nil)
var _ provider.CandleProvider = (*Client)(nil)
var _ provider.TickerProvider = (*Client)(nil)

func newTestClient(url string) *Client { return NewClient(Config{BaseURL: url}) }

const marketsJSON = `[
 {"market":"BTC-EUR","status":"trading","base":"BTC","quote":"EUR","pricePrecision":null,"minOrderInBaseAsset":"0.00006607","minOrderInQuoteAsset":"5.00","maxOrderInBaseAsset":"13213.19","maxOrderInQuoteAsset":"1000000000.00","quantityDecimals":8,"notionalDecimals":2,"tickSize":"1.00","feeCategory":"A","orderTypes":["market","limit"]},
 {"market":"SHIB-EUR","status":"halted","base":"SHIB","quote":"EUR","pricePrecision":null,"minOrderInBaseAsset":"988149.08","minOrderInQuoteAsset":"5.00","maxOrderInBaseAsset":"1","maxOrderInQuoteAsset":"1","quantityDecimals":2,"notionalDecimals":2,"tickSize":"0.0000000100","feeCategory":"A","orderTypes":["limit"]}
]`

func TestCapabilities(t *testing.T) {
	m := newTestClient("").Capabilities()
	for _, f := range []capability.Feature{capability.FeatureServerTime, capability.FeatureExchangeInfo, capability.FeatureTicker24h, capability.FeatureCandles} {
		if !m[capability.CapabilityKey{Market: capability.MarketSpot, Feature: f}] {
			t.Errorf("spot %v missing", f)
		}
	}
	if m[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureCandles}] {
		t.Error("futures must not be supported")
	}
}

func TestExchangeInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Errorf("path %s", r.URL.Path)
		}
		fmt.Fprint(w, marketsJSON)
	}))
	defer srv.Close()
	resp, err := newTestClient(srv.URL).ExchangeInfo(context.Background(), model.MarketSpot)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data.Symbols) != 2 {
		t.Fatalf("got %d symbols", len(resp.Data.Symbols))
	}
	b := resp.Data.Symbols[0]
	if b.Symbol != "BTC-EUR" || b.NormalizedSymbol != "BTC-EUR" || b.BaseAsset != "BTC" || b.QuoteAsset != "EUR" || b.Status != model.SymbolStatusTrading {
		t.Errorf("bad symbol %+v", b)
	}
	if *b.PricePrecision != 0 || *b.QtyPrecision != 8 || *b.MinPrice != 1 || *b.StepSize != 1e-8 {
		t.Errorf("bad precision: pp=%d qp=%d min=%v step=%v", *b.PricePrecision, *b.QtyPrecision, *b.MinPrice, *b.StepSize)
	}
	if *b.MinQty != 0.00006607 || *b.MinNotional != 5 {
		t.Errorf("bad mins: qty=%v notional=%v", *b.MinQty, *b.MinNotional)
	}
	if b.MakerFee != nil || b.TakerFee != nil {
		t.Error("fees must be unset")
	}
	s := resp.Data.Symbols[1]
	if s.Status != model.SymbolStatusHalt || *s.PricePrecision != 8 {
		t.Errorf("bad halted symbol: %v pp=%d", s.Status, *s.PricePrecision)
	}
	if _, err := newTestClient(srv.URL).ExchangeInfo(context.Background(), model.MarketFutures); err == nil {
		t.Error("futures must fail")
	}
}

func TestTicker24h(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ticker/24h" || r.URL.Query().Get("market") != "BTC-EUR" {
			t.Errorf("bad request %s", r.URL)
		}
		fmt.Fprint(w, `{"market":"BTC-EUR","open":"7.5E+4","high":"7.754E+4","low":"74556","last":"75331","bid":"75330","ask":"75331","volume":"761.6","volumeQuote":"57824319.2","openTimestamp":1790942043742,"closeTimestamp":1791028423711}`)
	}))
	defer srv.Close()
	resp, err := newTestClient(srv.URL).Ticker24h(context.Background(), "BTCEUR", model.MarketSpot)
	if err != nil {
		t.Fatal(err)
	}
	d := resp.Data
	if d.Symbol != "BTC-EUR" || d.LastPrice != 75331 || *d.OpenPrice != 75000 || *d.QuoteVolume != 57824319.2 {
		t.Errorf("bad ticker %+v", d)
	}
	if *d.PriceChange != 331 || d.CloseTime == nil || d.CloseTime.Location() != time.UTC {
		t.Errorf("bad change/close time")
	}
}

func TestServerTime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"time":1791028462976,"timeNs":1791028462976364819}`)
	}))
	defer srv.Close()
	resp, err := newTestClient(srv.URL).ServerTime(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Data.Time.UnixMilli() != 1791028462976 || resp.Data.Time.Location() != time.UTC {
		t.Errorf("bad time %v", resp.Data.Time)
	}
}

func TestErrorMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		fmt.Fprint(w, `{"errorCode":205,"error":"market parameter is invalid."}`)
	}))
	defer srv.Close()
	_, err := newTestClient(srv.URL).Ticker24h(context.Background(), "NOPE-EUR", model.MarketSpot)
	pe, ok := err.(*model.ProviderError)
	if !ok || pe.Kind != model.ErrKindInvalidRequest || !strings.Contains(pe.ProviderMessage, "205") {
		t.Fatalf("got %#v", err)
	}
}

// fakeCandleServer serves 1m candles in [first, last] (ms) like the venue:
// start inclusive, end exclusive, newest first, newest `limit` rows win.
func fakeCandleServer(t *testing.T, first, last int64, requests *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		*requests = append(*requests, r.URL.RawQuery)
		if r.URL.Path != "/BTC-EUR/candles" {
			t.Errorf("path %s", r.URL.Path)
		}
		start, _ := strconv.ParseInt(q.Get("start"), 10, 64)
		end, _ := strconv.ParseInt(q.Get("end"), 10, 64)
		limit, err := strconv.Atoi(q.Get("limit"))
		if err != nil {
			limit = 1440
		}
		var rows []string
		for ts := last; ts >= first && len(rows) < limit; ts -= 60000 {
			if (start != 0 && ts < start) || (end != 0 && ts >= end) {
				continue
			}
			rows = append(rows, fmt.Sprintf(`[%d,"1","2","0.5","1.5","3"]`, ts))
		}
		fmt.Fprintf(w, "[%s]", strings.Join(rows, ","))
	}))
}

func TestCandles_RangePaging(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	const total = 3000
	var reqs []string
	srv := fakeCandleServer(t, base.UnixMilli(), base.Add((total-1)*time.Minute).UnixMilli(), &reqs)
	defer srv.Close()
	c := newTestClient(srv.URL)

	to := base.Add(total * time.Minute)
	limit := 5000 // above the venue cap: the provider clamps it to 1440
	var got []model.Candle
	for cursor := base; cursor.Before(to); cursor = cursor.Add(1440 * time.Minute) {
		end := cursor.Add(1440 * time.Minute)
		if end.After(to) {
			end = to
		}
		resp, err := c.Candles(context.Background(), "BTC-EUR", model.MarketSpot, "1m", model.CandleOpts{From: &cursor, To: &end, Limit: &limit})
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
			t.Fatalf("candle %d at %s, want %s", i, cd.OpenTime, want)
		}
	}
	if cd := got[0]; cd.Open != 1 || cd.High != 2 || cd.Low != 0.5 || cd.Close != 1.5 || cd.Volume == nil || *cd.Volume != 3 {
		t.Errorf("bad OHLCV %+v", cd)
	}
	if len(reqs) != 3 {
		t.Fatalf("got %d requests, want 3", len(reqs))
	}
	for _, r := range reqs {
		if !strings.Contains(r, "start=") || !strings.Contains(r, "end=") || !strings.Contains(r, "limit=1440") || !strings.Contains(r, "interval=1m") {
			t.Errorf("request lacks params: %s", r)
		}
	}
}

func TestCandles_UnsupportedInterval(t *testing.T) {
	var reqs []string
	srv := fakeCandleServer(t, 0, 0, &reqs)
	defer srv.Close()
	if _, err := newTestClient(srv.URL).Candles(context.Background(), "BTC-EUR", model.MarketSpot, "3m", model.CandleOpts{}); err == nil {
		t.Fatal("want error for 3m")
	}
	if len(reqs) != 0 {
		t.Error("no request expected")
	}
}

func TestCandles_BadRow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[[1759599900000,"1","2","0.5","1.5"]]`)
	}))
	defer srv.Close()
	if _, err := newTestClient(srv.URL).Candles(context.Background(), "BTC-EUR", model.MarketSpot, "1m", model.CandleOpts{}); err == nil {
		t.Fatal("want error for row without volume")
	}
}
