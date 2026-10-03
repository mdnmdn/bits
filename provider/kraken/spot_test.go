package kraken

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/model"
)

func TestSpotSymbolRules(t *testing.T) {
	native := map[string]string{
		"BTC-USD": "XBTUSD", "btc/usd": "XBTUSD", "XBT/USD": "XBTUSD", "XBTUSD": "XBTUSD", "BTCUSD": "XBTUSD",
		"ETH-EUR": "ETHEUR", "ETH-BTC": "ETHXBT", "DOGE-USD": "XDGUSD", "XXBTZUSD": "XXBTZUSD",
	}
	for in, want := range native {
		if got := toSpot(in); got != want {
			t.Errorf("toSpot(%q)=%q want %q", in, got, want)
		}
	}
	norm := map[string]string{"XBTUSD": "BTC-USD", "BTC-USD": "BTC-USD", "ETHEUR": "ETH-EUR", "XDGUSD": "DOGE-USD", "ETH/XBT": "ETH-BTC", "SOLUSDT": "SOL-USDT"}
	for in, want := range norm {
		if got := spotNormalized(in); got != want {
			t.Errorf("spotNormalized(%q)=%q want %q", in, got, want)
		}
	}
	for in, want := range map[string]string{"XXBT": "BTC", "ZUSD": "USD", "XBT": "BTC", "XETH": "ETH", "SOL": "SOL", "ZEUR": "EUR"} {
		if got := normAsset(in); got != want {
			t.Errorf("normAsset(%q)=%q want %q", in, got, want)
		}
	}
}

func TestCapabilities(t *testing.T) {
	all := NewClient(Config{}).Capabilities()
	for _, m := range []capability.MarketType{capability.MarketSpot, capability.MarketMargin, capability.MarketFutures} {
		for _, f := range []capability.Feature{capability.FeatureServerTime, capability.FeatureExchangeInfo, capability.FeaturePrice,
			capability.FeatureCandles, capability.FeatureTicker24h, capability.FeatureOrderBook} {
			if !all[capability.CapabilityKey{Market: m, Feature: f}] {
				t.Errorf("missing %s %s", m, f)
			}
		}
	}
	if !all[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureFundingRates}] ||
		all[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureFundingRates}] {
		t.Error("funding rates must be futures only")
	}
	only := NewClient(Config{FuturesEnabled: true}).Capabilities()
	if only[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeaturePrice}] {
		t.Error("spot declared with only futures enabled")
	}
}

func spotServer(t *testing.T, handlers map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := handlers[r.URL.Path]
		if !ok {
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
}

const assetPairsBody = `{"error":[],"result":{
"XXBTZUSD":{"altname":"XBTUSD","wsname":"XBT/USD","base":"XXBT","quote":"ZUSD","pair_decimals":1,"lot_decimals":8,"leverage_buy":[2,3,5],"fees":[[0,0.4],[10000,0.35]],"fees_maker":[[0,0.25]],"ordermin":"0.00005","costmin":"0.5","tick_size":"0.1","status":"online"},
"SOLEUR":{"altname":"SOLEUR","wsname":"SOL/EUR","base":"SOL","quote":"ZEUR","pair_decimals":2,"lot_decimals":8,"leverage_buy":[],"fees":[],"fees_maker":[],"ordermin":"0.02","costmin":"0.45","tick_size":"0.01","status":"cancel_only"},
"XXBTZUSD.d":{"altname":"XBTUSD.d","base":"XXBT","quote":"ZUSD","pair_decimals":1,"lot_decimals":8,"leverage_buy":[],"status":"online"}}}`

func TestSpotExchangeInfo(t *testing.T) {
	srv := spotServer(t, map[string]string{"/0/public/AssetPairs": assetPairsBody})
	defer srv.Close()
	c := NewClient(Config{SpotBaseURL: srv.URL})
	resp, err := c.ExchangeInfo(context.Background(), model.MarketSpot)
	if err != nil {
		t.Fatal(err)
	}
	syms := resp.Data.Symbols
	if len(syms) != 2 || syms[0].Symbol != "SOLEUR" || syms[1].Symbol != "XBTUSD" {
		t.Fatalf("symbols = %+v", syms)
	}
	b := syms[1]
	if b.NormalizedSymbol != "BTC-USD" || b.BaseAsset != "BTC" || b.QuoteAsset != "USD" || b.Status != model.SymbolStatusTrading {
		t.Errorf("btc = %+v", b)
	}
	if *b.MinQty != 0.00005 || *b.MinNotional != 0.5 || *b.StepSize != 1e-8 || *b.PricePrecision != 1 || b.Extra["tick_size"] != 0.1 || b.Extra["margin_leverage"] != 5 {
		t.Errorf("btc rules = %+v", b)
	}
	if *b.TakerFee != 0.004 || *b.MakerFee != 0.0025 {
		t.Errorf("fees = %v %v", *b.TakerFee, *b.MakerFee)
	}
	if syms[0].Status != model.SymbolStatusHalt {
		t.Errorf("cancel_only status = %v", syms[0].Status)
	}
	m, err := c.ExchangeInfo(context.Background(), model.MarketMargin)
	if err != nil || len(m.Data.Symbols) != 1 || m.Data.Symbols[0].Symbol != "XBTUSD" || m.Data.Symbols[0].Market != model.MarketMargin {
		t.Fatalf("margin = %+v, %v", m.Data.Symbols, err)
	}
}

func TestSpotCandles(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"error":[],"result":{"XXBTZUSD":[
[1788436800,"1","2","0.5","1.5","1.2","3.5",10],
[1788440400,"1.5","2","1","1.8","1.4","4",11],
[1788444000,"1.8","2","1","1.9","1.4","5",12]],"last":1788440400}}`))
	}))
	defer srv.Close()
	c := NewClient(Config{SpotBaseURL: srv.URL})
	from := time.Unix(1788440400, 0).UTC()
	to := time.Unix(1788444000, 0).UTC()
	resp, err := c.Candles(context.Background(), "BTC-USD", model.MarketSpot, "1h", model.CandleOpts{From: &from, To: &to})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "pair=XBTUSD") || !strings.Contains(gotQuery, "interval=60") || !strings.Contains(gotQuery, "since=1788440399") {
		t.Errorf("query = %s", gotQuery)
	}
	if len(resp.Data) != 1 || !resp.Data[0].OpenTime.Equal(from) || resp.Data[0].OpenTime.Location() != time.UTC || resp.Data[0].Close != 1.8 || *resp.Data[0].Volume != 4 {
		t.Fatalf("candles = %+v", resp.Data)
	}
	// A range older than the 720 window returns nothing, never fake data.
	oldFrom, oldTo := from.AddDate(-1, 0, 0), from.AddDate(-1, 0, 7)
	resp, err = c.Candles(context.Background(), "BTC-USD", model.MarketMargin, "1h", model.CandleOpts{From: &oldFrom, To: &oldTo})
	if err != nil || len(resp.Data) != 0 {
		t.Fatalf("old range = %+v, %v", resp.Data, err)
	}
	limit := 2
	resp, _ = c.Candles(context.Background(), "BTC-USD", model.MarketSpot, "1h", model.CandleOpts{Limit: &limit})
	if len(resp.Data) != 2 || resp.Data[1].Close != 1.9 {
		t.Fatalf("limit without from = %+v", resp.Data)
	}
	if _, err := c.Candles(context.Background(), "BTC-USD", model.MarketSpot, "12h", model.CandleOpts{}); err == nil {
		t.Fatal("12h accepted on spot")
	}
}

func TestSpotTickerPriceBook(t *testing.T) {
	srv := spotServer(t, map[string]string{
		"/0/public/Ticker": `{"error":[],"result":{"XXBTZUSD":{"a":["101.0","1","1.000"],"b":["100.0","2","2.000"],"c":["100.5","0.1"],"v":["10","200"],"p":["99","98"],"t":[5,50],"l":["90","80"],"h":["110","120"],"o":"100"}}}`,
		"/0/public/Depth":  `{"error":[],"result":{"XXBTZUSD":{"asks":[["101.0","1.5",1788436800],["102.0","2",1788436801]],"bids":[["100.0","3",1788436802],["99.0","4",1788436799]]}}}`,
	})
	defer srv.Close()
	c := NewClient(Config{SpotBaseURL: srv.URL})
	tk, err := c.Ticker24h(context.Background(), "BTC-USD", model.MarketSpot)
	if err != nil {
		t.Fatal(err)
	}
	d := tk.Data
	if d.Symbol != "BTC-USD" || d.LastPrice != 100.5 || *d.Volume != 200 || *d.HighPrice != 120 || *d.LowPrice != 80 || *d.WeightedAvgPrice != 98 || *d.OpenPrice != 100 || *d.PriceChange != 0.5 || d.QuoteVolume != nil {
		t.Errorf("ticker = %+v", d)
	}
	pr, err := c.Price(context.Background(), []string{"XBTUSD"}, "")
	if err != nil || pr.Market != model.MarketSpot || len(pr.Data) != 1 || pr.Data[0].Price != 100.5 || pr.Data[0].Symbol != "BTC-USD" || pr.Data[0].Currency != "USD" || *pr.Data[0].BidSize != 2 {
		t.Fatalf("price = %+v, %v", pr, err)
	}
	ob, err := c.OrderBook(context.Background(), "BTC-USD", model.MarketMargin, 2)
	if err != nil {
		t.Fatal(err)
	}
	b := ob.Data
	if len(b.Bids) != 2 || b.Bids[0].Price != 100 || b.Bids[0].Quantity != 3 || b.Asks[0].Price != 101 || b.Time == nil || b.Time.Unix() != 1788436802 || b.Market != model.MarketMargin {
		t.Errorf("book = %+v", b)
	}
}

func TestSpotErrors(t *testing.T) {
	srv := spotServer(t, map[string]string{"/0/public/Ticker": `{"error":["EQuery:Unknown asset pair"]}`})
	defer srv.Close()
	c := NewClient(Config{SpotBaseURL: srv.URL})
	_, err := c.Ticker24h(context.Background(), "NOPE-USD", model.MarketSpot)
	pe, ok := err.(*model.ProviderError)
	if !ok || pe.Kind != model.ErrKindNotFound {
		t.Fatalf("err = %#v", err)
	}
	pr, err := c.Price(context.Background(), []string{"NOPE-USD"}, "")
	if err != nil || len(pr.Errors) != 1 {
		t.Fatalf("price errors = %+v, %v", pr, err)
	}
}

func TestFuturesBookPriceFunding(t *testing.T) {
	srv := spotServer(t, map[string]string{
		"/derivatives/api/v3/orderbook":              `{"result":"success","serverTime":"2026-10-03T12:00:00.000Z","orderBook":{"bids":[[1,5],[100,2],[99,3]],"asks":[[103,1],[101,4]]}}`,
		"/derivatives/api/v3/tickers/PF_XBTUSD":      `{"result":"success","ticker":{"time":"2026-10-03T12:00:00.000Z","last":100.5,"bid":100,"bidSize":2,"ask":101,"askSize":0,"open24h":99,"high24h":110,"low24h":90,"vol24h":7,"change24h":1.5}}`,
		"/derivatives/api/v4/historicalfundingrates": `{"result":"success","rates":[{"timestamp":"2026-10-01T08:00:00Z","fundingRate":1,"relativeFundingRate":0.0001},{"timestamp":"2026-10-01T09:00:00Z","fundingRate":2,"relativeFundingRate":0.0002},{"timestamp":"2026-10-01T10:00:00Z","fundingRate":3,"relativeFundingRate":0.0003}]}`,
	})
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL})
	ob, err := c.OrderBook(context.Background(), "BTC-USD", model.MarketFutures, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(ob.Data.Bids) != 2 || ob.Data.Bids[0].Price != 100 || ob.Data.Bids[1].Price != 99 || ob.Data.Asks[0].Price != 101 || ob.Data.Symbol != "BTC-USD" {
		t.Errorf("book = %+v", ob.Data)
	}
	pr, err := c.Price(context.Background(), []string{"PF_XBTUSD"}, "")
	if err != nil || pr.Market != model.MarketFutures || pr.Data[0].Price != 100.5 || pr.Data[0].AskSize != nil || *pr.Data[0].BidSize != 2 || pr.Data[0].Time == nil {
		t.Fatalf("price = %+v, %v", pr, err)
	}
	if _, err := c.Price(context.Background(), []string{"PF_XBTUSD", "BTC-USD"}, ""); err == nil {
		t.Fatal("mixed Price accepted")
	}
	fr, err := c.FundingRates(context.Background(), "BTC-USD", model.FundingRateOpts{})
	if err != nil || len(fr.Data) != 3 || fr.Data[0].Rate != 0.0001 || fr.Data[0].Symbol != "PF_XBTUSD" {
		t.Fatalf("funding = %+v, %v", fr.Data, err)
	}
	from := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	one := 1
	fr, _ = c.FundingRates(context.Background(), "PF_XBTUSD", model.FundingRateOpts{From: &from, Limit: &one})
	if len(fr.Data) != 1 || fr.Data[0].Rate != 0.0002 {
		t.Fatalf("from+limit = %+v", fr.Data)
	}
	fr, _ = c.FundingRates(context.Background(), "PF_XBTUSD", model.FundingRateOpts{Limit: &one})
	if len(fr.Data) != 1 || fr.Data[0].Rate != 0.0003 {
		t.Fatalf("limit = %+v", fr.Data)
	}
	to := time.Date(2026, 10, 1, 8, 30, 0, 0, time.UTC)
	fr, _ = c.FundingRates(context.Background(), "PF_XBTUSD", model.FundingRateOpts{To: &to})
	if len(fr.Data) != 1 {
		t.Fatalf("to = %+v", fr.Data)
	}
}
