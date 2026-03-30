package cryptocom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/config"
	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/provider"
)

// Compile-time interface assertions.
var _ provider.Provider = (*Client)(nil)
var _ provider.ExchangeProvider = (*Client)(nil)
var _ provider.PriceProvider = (*Client)(nil)
var _ provider.CandleProvider = (*Client)(nil)
var _ provider.TickerProvider = (*Client)(nil)
var _ provider.OrderBookProvider = (*Client)(nil)
var _ provider.PriceStreamProvider = (*Client)(nil)
var _ provider.OrderBookStreamProvider = (*Client)(nil)

// ── helpers ──────────────────────────────────────────────────────────────────

// newTestClient creates a Client pointed at the given test server URL.
func newTestClient(serverURL string) *Client {
	return NewClient(config.CryptoComConfig{BaseURL: serverURL})
}

// envelope wraps a result value in the Crypto.com v2 response envelope.
func envelope(result any) []byte {
	raw, _ := json.Marshal(result)
	out, _ := json.Marshal(map[string]any{
		"id":     -1,
		"method": "public/test",
		"code":   0,
		"result": json.RawMessage(raw),
	})
	return out
}

// errorEnvelope returns a Crypto.com error response with the given code.
func errorEnvelope(code int) []byte {
	out, _ := json.Marshal(map[string]any{
		"id":     -1,
		"method": "public/test",
		"code":   code,
		"result": map[string]any{},
	})
	return out
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if want != "" {
		if s := err.Error(); len(s) == 0 {
			t.Fatalf("error message is empty, wanted to contain %q", want)
		}
	}
}

func assertEqual(t *testing.T, label string, got, want any) {
	t.Helper()
	if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
		t.Errorf("%s: got %v, want %v", label, got, want)
	}
}

func assertInDelta(t *testing.T, label string, got, want, delta float64) {
	t.Helper()
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > delta {
		t.Errorf("%s: got %v, want %v ± %v", label, got, want, delta)
	}
}

// ── fixtures ─────────────────────────────────────────────────────────────────

var sampleTicker = apiTickerResult{
	Data: []apiTickerData{
		{I: "BTC_USDT", A: "68000.50", B: "68000.00", C: "0.04615", O: "65000.00", P: "3000.50", H: "69000.00", L: "64500.00", V: "12345.678", VV: "850000000", T: 1700000000000},
	},
}

var sampleBook = apiBookResult{
	InstrumentName: "BTC_USDT",
	Depth:          5,
	Data: []apiBookRow{
		{
			Bids: [][]string{{"68000.0", "1.5", "1"}, {"67900.0", "2.0", "2"}},
			Asks: [][]string{{"68100.0", "0.5", "1"}, {"68200.0", "1.0", "1"}},
			T:    1700000000000,
		},
	},
}

var sampleInstruments = apiInstrumentsV1Response{
	Code: 0,
	Result: apiInstrumentsV1Data{
		Data: []apiInstrumentV1{
			{
				Symbol:           "BTC_USDT",
				InstType:         "CCY_PAIR",
				BaseCcy:          "BTC",
				QuoteCcy:         "USDT",
				QuoteDecimals:    2,
				QuantityDecimals: 6,
				PriceTickSize:    "0.01",
				QtyTickSize:      "0.00001",
				Tradable:         true,
			},
			{
				Symbol:   "ETH_USDT",
				InstType: "CCY_PAIR",
				BaseCcy:  "ETH",
				QuoteCcy: "USDT",
			},
			{
				Symbol:   "BTC_PERP",
				InstType: "PERPETUAL_SWAP", // non-CCY_PAIR; must be filtered
				BaseCcy:  "BTC",
				QuoteCcy: "USD",
			},
		},
	},
}

// ── Ticker24h ────────────────────────────────────────────────────────────────

func TestTicker24h(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/public/get-tickers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if sym := r.URL.Query().Get("instrument_name"); sym != "BTC_USDT" {
			t.Errorf("unexpected instrument_name: %s", sym)
		}
		_, _ = w.Write(envelope(sampleTicker))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	res, err := c.Ticker24h(context.Background(), "BTC_USDT", model.MarketSpot)
	assertNoError(t, err)

	assertEqual(t, "provider", res.Provider, "cryptocom")
	assertEqual(t, "market", string(res.Market), "spot")
	assertEqual(t, "kind", res.Kind, model.KindTicker)
	assertEqual(t, "symbol", res.Data.Symbol, "BTC_USDT")
	assertInDelta(t, "last price", res.Data.LastPrice, 68000.50, 0.01)

	if res.Data.OpenPrice == nil {
		t.Fatal("OpenPrice must not be nil")
	}
	assertInDelta(t, "open price", *res.Data.OpenPrice, 65000.00, 0.01)

	if res.Data.HighPrice == nil {
		t.Fatal("HighPrice must not be nil")
	}
	assertInDelta(t, "high price", *res.Data.HighPrice, 69000.00, 0.01)

	if res.Data.LowPrice == nil {
		t.Fatal("LowPrice must not be nil")
	}
	assertInDelta(t, "low price", *res.Data.LowPrice, 64500.00, 0.01)

	if res.Data.Volume == nil {
		t.Fatal("Volume must not be nil")
	}
	assertInDelta(t, "volume", *res.Data.Volume, 12345.678, 0.001)

	if res.Data.PriceChange == nil {
		t.Fatal("PriceChange must not be nil")
	}
	assertInDelta(t, "price change", *res.Data.PriceChange, 3000.50, 0.01)

	if res.Data.PriceChangePercent == nil {
		t.Fatal("PriceChangePercent must not be nil")
	}
	// 3000.50 / 65000 * 100 ≈ 4.616%
	assertInDelta(t, "change pct", *res.Data.PriceChangePercent, 4.616, 0.01)
}

func TestTicker24h_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(envelope(apiTickerResult{Data: []apiTickerData{}}))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.Ticker24h(context.Background(), "UNKNOWN_USDT", model.MarketSpot)
	assertError(t, err, "no ticker data")
}

func TestTicker24h_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(errorEnvelope(10003))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.Ticker24h(context.Background(), "INVALID", model.MarketSpot)
	assertError(t, err, "API error")
}

// ── Price ────────────────────────────────────────────────────────────────────

func TestPrice_SingleSymbol(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(envelope(sampleTicker))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	res, err := c.Price(context.Background(), []string{"BTC_USDT"}, "USD")
	assertNoError(t, err)

	assertEqual(t, "kind", res.Kind, model.KindPrice)
	assertEqual(t, "provider", res.Provider, "cryptocom")
	if len(res.Data) != 1 {
		t.Fatalf("expected 1 price entry, got %d", len(res.Data))
	}
	assertInDelta(t, "price", res.Data[0].Price, 68000.50, 0.01)
	assertEqual(t, "symbol", res.Data[0].Symbol, "BTC_USDT")
	if res.Data[0].Change24h == nil {
		t.Fatal("Change24h must not be nil")
	}
	assertInDelta(t, "change24h", *res.Data[0].Change24h, 4.616, 0.01)
}

func TestPrice_MultiSymbol_PartialError(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("instrument_name") == "BTC_USDT" {
			_, _ = w.Write(envelope(sampleTicker))
		} else {
			_, _ = w.Write(envelope(apiTickerResult{Data: []apiTickerData{}}))
		}
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	res, err := c.Price(context.Background(), []string{"BTC_USDT", "UNKNOWN_XYZ"}, "USD")
	assertNoError(t, err)

	if len(res.Data) != 1 {
		t.Errorf("expected 1 successful price, got %d", len(res.Data))
	}
	if len(res.Errors) != 1 {
		t.Errorf("expected 1 item error, got %d", len(res.Errors))
	}
	assertEqual(t, "error symbol", res.Errors[0].Symbol, "UNKNOWN_XYZ")
	assertEqual(t, "call count", calls, 2)
}

// ── OrderBook ────────────────────────────────────────────────────────────────

func TestOrderBook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/public/get-book" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if sym := r.URL.Query().Get("instrument_name"); sym != "BTC_USDT" {
			t.Errorf("unexpected instrument_name: %s", sym)
		}
		if d := r.URL.Query().Get("depth"); d != "5" {
			t.Errorf("expected depth=5, got %q", d)
		}
		_, _ = w.Write(envelope(sampleBook))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	res, err := c.OrderBook(context.Background(), "BTC_USDT", model.MarketSpot, 5)
	assertNoError(t, err)

	assertEqual(t, "kind", res.Kind, model.KindOrderBook)
	assertEqual(t, "provider", res.Provider, "cryptocom")
	assertEqual(t, "symbol", res.Data.Symbol, "BTC_USDT")
	if len(res.Data.Bids) != 2 {
		t.Fatalf("expected 2 bids, got %d", len(res.Data.Bids))
	}
	if len(res.Data.Asks) != 2 {
		t.Fatalf("expected 2 asks, got %d", len(res.Data.Asks))
	}
	assertInDelta(t, "bid[0].price", res.Data.Bids[0].Price, 68000.0, 0.01)
	assertInDelta(t, "bid[0].qty", res.Data.Bids[0].Quantity, 1.5, 0.001)
	assertInDelta(t, "ask[0].price", res.Data.Asks[0].Price, 68100.0, 0.01)
	if res.Data.Time == nil {
		t.Error("Time must not be nil when T is provided")
	}
}

func TestOrderBook_NoDepthParam(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if d := r.URL.Query().Get("depth"); d != "" {
			t.Errorf("expected no depth param, got %q", d)
		}
		_, _ = w.Write(envelope(sampleBook))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	res, err := c.OrderBook(context.Background(), "BTC_USDT", model.MarketSpot, 0)
	assertNoError(t, err)
	if len(res.Data.Bids) == 0 {
		t.Error("expected bids, got none")
	}
}

// ── ExchangeInfo ─────────────────────────────────────────────────────────────

func TestExchangeInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, "path", r.URL.Path, "/public/get-instruments")
		data, _ := json.Marshal(sampleInstruments.Result.Data)
		resp := map[string]any{
			"id":     -1,
			"method": "public/get-instruments",
			"code":   0,
			"result": map[string]any{"data": json.RawMessage(data)},
		}
		out, _ := json.Marshal(resp)
		_, _ = w.Write(out)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	res, err := c.ExchangeInfo(context.Background(), model.MarketSpot)
	assertNoError(t, err)

	assertEqual(t, "kind", res.Kind, model.KindExchangeInfo)
	assertEqual(t, "provider", res.Provider, "cryptocom")
	assertEqual(t, "market", string(res.Data.Market), "spot")
	// We return only CCY_PAIR symbols
	if len(res.Data.Symbols) != 2 {
		t.Fatalf("expected 2 CCY_PAIR symbols, got %d", len(res.Data.Symbols))
	}

	btc := res.Data.Symbols[0]
	assertEqual(t, "symbol", btc.Symbol, "BTC_USDT")
	assertEqual(t, "base", btc.BaseAsset, "BTC")
	assertEqual(t, "quote", btc.QuoteAsset, "USDT")
	assertEqual(t, "status", string(btc.Status), "trading")
	if btc.PricePrecision == nil || *btc.PricePrecision != 2 {
		t.Errorf("PricePrecision: got %v, want 2", btc.PricePrecision)
	}
}

// ── ServerTime ───────────────────────────────────────────────────────────────

func TestServerTime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, "path", r.URL.Path, "/public/get-ticker")
		_, _ = w.Write(envelope(sampleTicker))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	res, err := c.ServerTime(context.Background())
	assertNoError(t, err)

	assertEqual(t, "kind", res.Kind, model.KindServerTime)
	assertEqual(t, "provider", res.Provider, "cryptocom")
	if res.Data.Time.IsZero() {
		t.Error("Time must not be zero")
	}
	if res.Data.Latency == nil {
		t.Fatal("Latency must not be nil")
	}
	if res.Data.Latency.Nanoseconds() <= 0 {
		t.Error("Latency must be positive")
	}
}

// ── Capabilities ─────────────────────────────────────────────────────────────

func TestCapabilities(t *testing.T) {
	c := NewClient(config.CryptoComConfig{})
	caps := c.Capabilities()

	if len(caps) < 7 {
		t.Errorf("expected at least 7 capabilities, got %d", len(caps))
	}

	expected := []capability.CapabilityKey{
		{Market: capability.MarketSpot, Feature: capability.FeatureServerTime},
		{Market: capability.MarketSpot, Feature: capability.FeatureExchangeInfo},
		{Market: capability.MarketSpot, Feature: capability.FeaturePrice},
		{Market: capability.MarketSpot, Feature: capability.FeatureTicker24h},
		{Market: capability.MarketSpot, Feature: capability.FeatureOrderBook},
		{Market: capability.MarketSpot, Feature: capability.FeatureStreamPrice},
		{Market: capability.MarketSpot, Feature: capability.FeatureStreamOrderBook},
	}
	for _, key := range expected {
		if !caps[key] {
			t.Errorf("capability missing: market=%s feature=%s", key.Market, key.Feature)
		}
	}
}

func TestID(t *testing.T) {
	c := NewClient(config.CryptoComConfig{})
	if id := c.ID(); id != "cryptocom" {
		t.Errorf("ID: got %q, want %q", id, "cryptocom")
	}
}

func TestSetUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ua := r.Header.Get("User-Agent"); ua != "bits-test/1.0" {
			t.Errorf("User-Agent: got %q, want %q", ua, "bits-test/1.0")
		}
		_, _ = w.Write(envelope(sampleTicker))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	c.SetUserAgent("bits-test/1.0")
	_, err := c.Ticker24h(context.Background(), "BTC_USDT", model.MarketSpot)
	assertNoError(t, err)
}
