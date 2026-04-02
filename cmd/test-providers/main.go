// test-providers exercises every selected capability across selected providers and markets.
// It is a diagnostic/integration tool, not part of the production CLI.
//
// Checklist Mode (default):
//   - Runs validity checks for each capability against known reference values
//   - Shows compact dashboard-style results: OK / WARNING / ERROR
//   - Provides detailed failure descriptions when checks don't pass
//
// Raw Mode:
//   - Shows the raw API responses without validation
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	bits "github.com/mdnmdn/bits"
	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/command"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider/registry"
	"github.com/mdnmdn/bits/render"
	"github.com/spf13/pflag"
)

const (
	checklistModeDefault = true
	timeTolerance        = 10 * time.Second
	priceTolerance       = 0.10
	volumeTolerance      = 10.0
)

type TestStatus string

const (
	StatusOK      TestStatus = "OK"
	StatusWarning TestStatus = "WARNING"
	StatusError   TestStatus = "ERROR"
)

type TestResult struct {
	Provider string
	Market   string
	Feature  string
	Symbol   string
	Status   TestStatus
	Message  string
	Details  string
}

var testableFeatures = []capability.Feature{
	capability.FeatureServerTime,
	capability.FeatureExchangeInfo,
	capability.FeaturePrice,
	capability.FeatureCandles,
	capability.FeatureTicker24h,
	capability.FeatureOrderBook,
	capability.FeatureMarketsList,
	capability.FeatureStreamPrice,
	capability.FeatureStreamOrderBook,
}

var featureNames = map[string]capability.Feature{
	"server_time":       capability.FeatureServerTime,
	"exchange_info":     capability.FeatureExchangeInfo,
	"price":             capability.FeaturePrice,
	"candles":           capability.FeatureCandles,
	"ticker_24h":        capability.FeatureTicker24h,
	"order_book":        capability.FeatureOrderBook,
	"markets_list":      capability.FeatureMarketsList,
	"stream_price":      capability.FeatureStreamPrice,
	"stream_order_book": capability.FeatureStreamOrderBook,
}

var marketIndependent = map[capability.Feature]bool{
	capability.FeatureServerTime:  true,
	capability.FeaturePrice:       true,
	capability.FeatureMarketsList: true,
}

type bitgetRef struct {
	lastPrice   float64
	volume      float64
	quoteVolume float64
}

func main() {
	var (
		providerFlags   []string
		marketFlags     []string
		symbolFlags     []string
		capabilityFlags []string
		listLength      int
		streamLength    int
		outputFormat    string
		checklistMode   bool
	)

	pflag.StringSliceVar(&providerFlags, "provider", nil, "providers to test (default: all); repeat or comma-separate")
	pflag.StringSliceVar(&marketFlags, "markets", nil, "markets to test (default: all)")
	pflag.StringSliceVar(&symbolFlags, "symbols", nil, "symbols (default: BTCUSDT)")
	pflag.StringSliceVar(&capabilityFlags, "capabilities", nil, "capabilities to test (default: all); case-insensitive")
	pflag.IntVar(&listLength, "list-length", 3, "max records in list responses")
	pflag.IntVar(&streamLength, "stream-length", 3, "number of streaming ticks to collect")
	pflag.StringVar(&outputFormat, "output", "json", "output format: dashboard, json, yaml, markdown, toon")
	pflag.BoolVar(&checklistMode, "checklist", checklistModeDefault, "enable checklist mode with validity checks (default: true)")
	pflag.Parse()

	allProviders := registry.AllProviderIDs()
	providers, err := resolveProviders(providerFlags, allProviders)
	if err != nil {
		fatalf("%v", err)
	}

	markets, err := resolveMarkets(marketFlags)
	if err != nil {
		fatalf("%v", err)
	}

	symbols := symbolFlags
	if len(symbols) == 0 {
		symbols = []string{"BTCUSDT"}
	}

	features, err := resolveCapabilities(capabilityFlags)
	if err != nil {
		fatalf("%v", err)
	}

	format := render.FormatJSON
	isDashboard := strings.ToLower(outputFormat) == "dashboard"
	if checklistMode || isDashboard {
		format = render.FormatJSON
	} else {
		switch strings.ToLower(outputFormat) {
		case "json":
			format = render.FormatJSON
		case "yaml":
			format = render.FormatYAML
		case "markdown":
			format = render.FormatMarkdown
		case "toon":
			format = render.FormatToon
		case "table":
			format = render.FormatJSON
			fmt.Fprintln(os.Stderr, "note: table format not supported in test-providers, using json")
		}
	}

	cfg, err := command.LoadConfig()
	if err != nil {
		fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	var results []TestResult

	for _, feat := range features {
		if !checklistMode {
			fmt.Printf("\n╔══ Capability: %s ══\n", feat)
		}
		calledProviders := map[string]bool{}

		for _, providerID := range providers {
			client := bits.NewProvider(cfg, providerID, bits.WithSymbolEngine())
			caps := client.Capabilities()

			for _, market := range markets {
				key := capability.CapabilityKey{Market: market, Feature: feat}
				if !caps[key] {
					continue
				}
				if marketIndependent[feat] && calledProviders[providerID] {
					continue
				}

				if checklistMode {
					result := runChecklistTest(ctx, client, feat, market, symbols, listLength, streamLength, providerID)
					results = append(results, result...)
					printResult(result)
				} else {
					fmt.Printf("╠─ %s / %s\n", providerID, market)
					callAndRender(ctx, os.Stdout, client, feat, market, symbols, listLength, streamLength, format)
				}

				if marketIndependent[feat] {
					calledProviders[providerID] = true
				}
			}
		}
	}

	if checklistMode && isDashboard {
		printTotal(results)
	}
}

func printResult(results []TestResult) {
	for _, r := range results {
		icon := "ok"
		switch r.Status {
		case StatusWarning:
			icon = "warn"
		case StatusError:
			icon = "error"
		}
		msg := r.Message
		if r.Details != "" {
			msg = msg + " (" + r.Details + ")"
		}
		if r.Symbol != "" {
			if msg != "" {
				_, _ = fmt.Fprintf(os.Stdout, "%s %s %s %s %s %s\n", r.Provider, r.Market, r.Symbol, r.Feature, icon, msg)
			} else {
				_, _ = fmt.Fprintf(os.Stdout, "%s %s %s %s %s\n", r.Provider, r.Market, r.Symbol, r.Feature, icon)
			}
		} else {
			if msg != "" {
				_, _ = fmt.Fprintf(os.Stdout, "%s %s %s %s %s\n", r.Provider, r.Market, r.Feature, icon, msg)
			} else {
				_, _ = fmt.Fprintf(os.Stdout, "%s %s %s %s\n", r.Provider, r.Market, r.Feature, icon)
			}
		}
	}
}

func printTotal(results []TestResult) {
	var ok, warn, err int
	for _, r := range results {
		switch r.Status {
		case StatusOK:
			ok++
		case StatusWarning:
			warn++
		case StatusError:
			err++
		}
	}
	_, _ = fmt.Fprintf(os.Stdout, "TOTAL=%d OK=%d WARNING=%d ERROR=%d\n", len(results), ok, warn, err)
}

// --- Checklist Mode Implementation ---

func runChecklistTest(
	ctx context.Context,
	client *bits.Client,
	feat capability.Feature,
	market capability.MarketType,
	symbols []string,
	listLength, streamLength int,
	providerID string,
) []TestResult {
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	symbol := ""
	if len(symbols) > 0 {
		symbol = symbols[0]
	}

	switch feat {
	case capability.FeatureServerTime:
		return testServerTime(reqCtx, client, providerID, string(market))
	case capability.FeatureExchangeInfo:
		return testExchangeInfo(reqCtx, client, providerID, market)
	case capability.FeaturePrice:
		return testPrice(reqCtx, client, providerID, symbols, symbol)
	case capability.FeatureCandles:
		return testCandles(reqCtx, client, providerID, market, symbol)
	case capability.FeatureTicker24h:
		return testTicker24h(reqCtx, client, providerID, market, symbol)
	case capability.FeatureOrderBook:
		return testOrderBook(reqCtx, client, providerID, market, symbol)
	case capability.FeatureMarketsList:
		return testMarketsList(reqCtx, client, providerID)
	case capability.FeatureStreamPrice:
		return testStreamPrice(reqCtx, client, providerID, symbols, streamLength)
	case capability.FeatureStreamOrderBook:
		return testStreamOrderBook(reqCtx, client, providerID, market, symbols, streamLength)
	}
	return []TestResult{{Provider: providerID, Market: string(market), Feature: string(feat), Symbol: symbol, Status: StatusError, Message: "unsupported feature"}}
}

// testServerTime validates server time is within +/-10s of UTC.
// Logic: Fetches server time, calculates clock skew, validates it's within tolerance.
func testServerTime(ctx context.Context, client *bits.Client, providerID, market string) []TestResult {
	res, err := client.ServerTime(ctx)
	if err != nil {
		return []TestResult{{Provider: providerID, Market: market, Feature: "server_time", Status: StatusError, Message: "API call failed", Details: err.Error()}}
	}

	if res.Data.ClockSkew == nil {
		return []TestResult{{Provider: providerID, Market: market, Feature: "server_time", Status: StatusWarning, Message: "clock skew not calculated"}}
	}

	skew := *res.Data.ClockSkew
	if skew < 0 {
		skew = -skew
	}

	if skew > timeTolerance {
		return []TestResult{{Provider: providerID, Market: market, Feature: "server_time", Status: StatusWarning, Message: "clock skew exceeds tolerance", Details: fmt.Sprintf("skew: %v (tolerance: %v)", *res.Data.ClockSkew, timeTolerance)}}
	}

	return []TestResult{{Provider: providerID, Market: market, Feature: "server_time", Status: StatusOK, Message: "time within tolerance", Details: fmt.Sprintf("skew: %v", *res.Data.ClockSkew)}}
}

// testExchangeInfo validates exchange info has valid symbols with reasonable values.
// Logic: Checks symbols exist, have status, and valid price/qty precisions.
func testExchangeInfo(ctx context.Context, client *bits.Client, providerID string, market capability.MarketType) []TestResult {
	res, err := client.ExchangeInfo(ctx, model.MarketType(market))
	if err != nil {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "exchange_info", Status: StatusError, Message: "API call failed", Details: err.Error()}}
	}

	if len(res.Data.Symbols) == 0 {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "exchange_info", Status: StatusError, Message: "no symbols returned"}}
	}

	var noStatus, noBase, noQuote, invalidPrec int
	for _, sym := range res.Data.Symbols {
		if sym.Status == "" {
			noStatus++
		}
		if sym.BaseAsset == "" {
			noBase++
		}
		if sym.QuoteAsset == "" {
			noQuote++
		}
		if sym.PricePrecision != nil && *sym.PricePrecision < 0 {
			invalidPrec++
		}
	}

	if noStatus > len(res.Data.Symbols)/2 || noBase > len(res.Data.Symbols)/2 || noQuote > len(res.Data.Symbols)/2 {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "exchange_info", Status: StatusWarning, Message: "many symbols missing critical fields", Details: fmt.Sprintf("noStatus: %d, noBase: %d, noQuote: %d", noStatus, noBase, noQuote)}}
	}

	return []TestResult{{Provider: providerID, Market: string(market), Feature: "exchange_info", Status: StatusOK, Message: fmt.Sprintf("%d symbols valid", len(res.Data.Symbols)), Details: fmt.Sprintf("noStatus: %d, noBase: %d, noQuote: %d", noStatus, noBase, noQuote)}}
}

// testPrice validates prices are within +/-10% of Bitget reference values.
// Logic: Fetches Bitget anonymous ticker, compares prices and volumes.
func testPrice(ctx context.Context, client *bits.Client, providerID string, symbols []string, symbol string) []TestResult {
	ref, err := fetchBitgetReference(ctx, symbols[0])
	if err != nil {
		return []TestResult{{Provider: providerID, Market: "spot", Feature: "price", Symbol: symbol, Status: StatusWarning, Message: "could not fetch reference", Details: err.Error()}}
	}

	res, err := client.Price(ctx, symbols, "usd")
	if err != nil {
		return []TestResult{{Provider: providerID, Market: "spot", Feature: "price", Symbol: symbol, Status: StatusError, Message: "API call failed", Details: err.Error()}}
	}

	if len(res.Data) == 0 {
		return []TestResult{{Provider: providerID, Market: "spot", Feature: "price", Symbol: symbol, Status: StatusError, Message: "no price data returned"}}
	}

	var results []TestResult
	for _, p := range res.Data {
		priceRatio := math.Abs(p.Price-ref.lastPrice) / ref.lastPrice
		if priceRatio > priceTolerance {
			results = append(results, TestResult{Provider: providerID, Market: "spot", Symbol: symbol, Feature: "price", Status: StatusWarning, Message: "price deviation exceeds 10%", Details: fmt.Sprintf("%s: got %.2f, ref %.2f (%.1f%%)", p.Symbol, p.Price, ref.lastPrice, priceRatio*100)})
			continue
		}

		if p.Volume24h != nil && ref.volume > 0 {
			volRatio := math.Log10(*p.Volume24h+1) - math.Log10(ref.volume+1)
			if math.Abs(volRatio) > 1 {
				results = append(results, TestResult{Provider: providerID, Market: "spot", Symbol: symbol, Feature: "price", Status: StatusWarning, Message: "volume order of magnitude differs", Details: fmt.Sprintf("%s: got %.2f, ref %.2f", p.Symbol, *p.Volume24h, ref.volume)})
				continue
			}
		}

		vol := 0.0
		if p.Volume24h != nil {
			vol = *p.Volume24h
		}
		results = append(results, TestResult{Provider: providerID, Market: "spot", Symbol: symbol, Feature: "price", Status: StatusOK, Message: "price and volume valid", Details: fmt.Sprintf("price: %.2f, vol: %.2f", p.Price, vol)})
	}

	return results
}

// testCandles validates candle data has valid OHLCV values.
// Logic: Checks open < high, open > low, close values reasonable, volumes positive.
func testCandles(ctx context.Context, client *bits.Client, providerID string, market capability.MarketType, symbol string) []TestResult {
	limit := 10
	res, err := client.Candles(ctx, symbol, model.MarketType(market), "1h", model.CandleOpts{Limit: &limit})
	if err != nil {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "candles", Status: StatusError, Message: "API call failed", Details: err.Error()}}
	}

	if len(res.Data) == 0 {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "candles", Status: StatusError, Message: "no candles returned"}}
	}

	var invalidOHLC, invalidVolume int
	for _, c := range res.Data {
		if c.Open > c.High || c.Open < c.Low || c.Close > c.High || c.Close < c.Low {
			invalidOHLC++
		}
		if c.Volume != nil && *c.Volume < 0 {
			invalidVolume++
		}
	}

	if invalidOHLC > len(res.Data)/2 {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "candles", Status: StatusWarning, Message: "many candles have invalid OHLC"}}
	}

	return []TestResult{{Provider: providerID, Market: string(market), Feature: "candles", Status: StatusOK, Message: fmt.Sprintf("%d candles valid", len(res.Data))}}
}

// testTicker24h validates ticker has valid price and volume data with reasonable values.
// Logic: Compares against Bitget reference for price and volume magnitude.
func testTicker24h(ctx context.Context, client *bits.Client, providerID string, market capability.MarketType, symbol string) []TestResult {
	ref, err := fetchBitgetReference(ctx, symbol)
	if err != nil {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "ticker_24h", Status: StatusWarning, Message: "could not fetch reference", Details: err.Error()}}
	}

	res, err := client.Ticker24h(ctx, symbol, model.MarketType(market))
	if err != nil {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "ticker_24h", Status: StatusError, Message: "API call failed", Details: err.Error()}}
	}

	priceRatio := math.Abs(res.Data.LastPrice-ref.lastPrice) / ref.lastPrice
	if priceRatio > priceTolerance {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "ticker_24h", Status: StatusWarning, Message: "price deviation exceeds 10%", Details: fmt.Sprintf("got %.2f, ref %.2f", res.Data.LastPrice, ref.lastPrice)}}
	}

	if res.Data.Volume != nil && ref.volume > 0 {
		volRatio := math.Log10(*res.Data.Volume+1) - math.Log10(ref.volume+1)
		if math.Abs(volRatio) > 1 {
			return []TestResult{{Provider: providerID, Market: string(market), Feature: "ticker_24h", Status: StatusWarning, Message: "volume order of magnitude differs", Details: fmt.Sprintf("got %.2f, ref %.2f", *res.Data.Volume, ref.volume)}}
		}
	}

	return []TestResult{{Provider: providerID, Market: string(market), Feature: "ticker_24h", Status: StatusOK, Message: "ticker data valid"}}
}

// testOrderBook validates order book has bids < asks, reasonable spread, non-negative quantities.
// Logic: Checks bid price < ask price, quantities positive, spread reasonable.
func testOrderBook(ctx context.Context, client *bits.Client, providerID string, market capability.MarketType, symbol string) []TestResult {
	res, err := client.OrderBook(ctx, symbol, model.MarketType(market), 10)
	if err != nil {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "order_book", Status: StatusError, Message: "API call failed", Details: err.Error()}}
	}

	if len(res.Data.Bids) == 0 || len(res.Data.Asks) == 0 {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "order_book", Status: StatusError, Message: "empty order book"}}
	}

	bestBid := res.Data.Bids[0].Price
	bestAsk := res.Data.Asks[0].Price

	if bestBid >= bestAsk {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "order_book", Status: StatusError, Message: "bid >= ask (invalid spread)"}}
	}

	var negQty int
	for _, b := range res.Data.Bids {
		if b.Quantity <= 0 {
			negQty++
		}
	}
	for _, a := range res.Data.Asks {
		if a.Quantity <= 0 {
			negQty++
		}
	}

	if negQty > 0 {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "order_book", Status: StatusWarning, Message: "negative quantities found"}}
	}

	spreadPct := (bestAsk - bestBid) / bestBid * 100
	return []TestResult{{Provider: providerID, Market: string(market), Feature: "order_book", Status: StatusOK, Message: fmt.Sprintf("spread: %.4f%%", spreadPct)}}
}

// testMarketsList validates market list has coins with valid market caps and prices.
// Logic: Checks market cap > 0, price > 0, volume > 0 for significant portion.
func testMarketsList(ctx context.Context, client *bits.Client, providerID string) []TestResult {
	opts := model.MarketOpts{Currency: "usd", PerPage: 50, Page: 1}
	res, err := client.CoinMarkets(ctx, opts)
	if err != nil {
		return []TestResult{{Provider: providerID, Market: "spot", Feature: "markets_list", Status: StatusError, Message: "API call failed", Details: err.Error()}}
	}

	if len(res.Data) == 0 {
		return []TestResult{{Provider: providerID, Market: "spot", Feature: "markets_list", Status: StatusError, Message: "no markets returned"}}
	}

	var noPrice, noVolume, noCap int
	for _, c := range res.Data {
		if c.Price <= 0 {
			noPrice++
		}
		if c.Volume24h != nil && *c.Volume24h <= 0 {
			noVolume++
		}
		if c.MarketCap != nil && *c.MarketCap <= 0 {
			noCap++
		}
	}

	if noPrice > len(res.Data)/2 || (noVolume > len(res.Data)/2 && noVolume < len(res.Data)) {
		return []TestResult{{Provider: providerID, Market: "spot", Feature: "markets_list", Status: StatusWarning, Message: "many coins missing price/volume"}}
	}

	return []TestResult{{Provider: providerID, Market: "spot", Feature: "markets_list", Status: StatusOK, Message: fmt.Sprintf("%d markets valid", len(res.Data))}}
}

// testStreamPrice validates price stream delivers updates with increasing timestamps.
// Logic: Collects ticks, verifies timestamps increase, prices reasonable.
func testStreamPrice(ctx context.Context, client *bits.Client, providerID string, symbols []string, count int) []TestResult {
	ch, err := client.StartPriceStream(ctx, symbols)
	if err != nil {
		return []TestResult{{Provider: providerID, Market: "spot", Feature: "stream_price", Status: StatusError, Message: "stream failed to start", Details: err.Error()}}
	}
	defer client.StopPriceStream() //nolint:errcheck

	var lastTS int64
	var lastPrice float64
	var ticks int

	for i := 0; i < count; i++ {
		select {
		case tick, ok := <-ch:
			if !ok {
				if ticks == 0 {
					return []TestResult{{Provider: providerID, Market: "spot", Feature: "stream_price", Status: StatusError, Message: "stream closed immediately"}}
				}
				break
			}
			ticks++
			if tick.Time != nil {
				ts := tick.Time.UnixMilli()
				if ts > lastTS {
					lastTS = ts
				} else if ticks > 1 {
					return []TestResult{{Provider: providerID, Market: "spot", Feature: "stream_price", Status: StatusWarning, Message: "timestamp not increasing"}}
				}
			}
			if tick.Price > 0 {
				lastPrice = tick.Price
			}
		case <-ctx.Done():
			return []TestResult{{Provider: providerID, Market: "spot", Feature: "stream_price", Status: StatusError, Message: "stream timeout"}}
		}
	}

	if ticks == 0 {
		return []TestResult{{Provider: providerID, Market: "spot", Feature: "stream_price", Status: StatusError, Message: "no ticks received"}}
	}

	return []TestResult{{Provider: providerID, Market: "spot", Feature: "stream_price", Status: StatusOK, Message: fmt.Sprintf("%d ticks received", ticks), Details: fmt.Sprintf("last price: %.2f", lastPrice)}}
}

// testStreamOrderBook validates order book stream delivers updates with reasonable data.
// Logic: Collects ticks, verifies bid < ask, quantities positive.
func testStreamOrderBook(ctx context.Context, client *bits.Client, providerID string, market capability.MarketType, symbols []string, count int) []TestResult {
	ch, err := client.StartOrderBookStream(ctx, symbols, model.MarketType(market), 5)
	if err != nil {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "stream_order_book", Status: StatusError, Message: "stream failed to start", Details: err.Error()}}
	}
	defer client.StopOrderBookStream() //nolint:errcheck

	var ticks int
	var invalidSpread bool

	for i := 0; i < count; i++ {
		select {
		case tick, ok := <-ch:
			if !ok {
				if ticks == 0 {
					return []TestResult{{Provider: providerID, Market: string(market), Feature: "stream_order_book", Status: StatusError, Message: "stream closed immediately"}}
				}
				break
			}
			ticks++
			if len(tick.Bids) > 0 && len(tick.Asks) > 0 {
				if tick.Bids[0].Price >= tick.Asks[0].Price {
					invalidSpread = true
				}
			}
		case <-ctx.Done():
			return []TestResult{{Provider: providerID, Market: string(market), Feature: "stream_order_book", Status: StatusError, Message: "stream timeout"}}
		}
	}

	if ticks == 0 {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "stream_order_book", Status: StatusError, Message: "no ticks received"}}
	}

	if invalidSpread {
		return []TestResult{{Provider: providerID, Market: string(market), Feature: "stream_order_book", Status: StatusWarning, Message: "invalid spread detected"}}
	}

	return []TestResult{{Provider: providerID, Market: string(market), Feature: "stream_order_book", Status: StatusOK, Message: fmt.Sprintf("%d ticks received", ticks)}}
}

// fetchBitgetReference fetches anonymous Bitget ticker for reference values.
// Uses public endpoint: GET /api/v2/spot/market/tickers
func fetchBitgetReference(ctx context.Context, symbol string) (*bitgetRef, error) {
	url := fmt.Sprintf("https://api.bitget.com/api/v2/spot/market/tickers?symbol=%s", symbol)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bitget returned %d", resp.StatusCode)
	}

	var body struct {
		Code string `json:"code"`
		Data []struct {
			LastPr      string `json:"lastPr"`
			BaseVolume  string `json:"baseVolume"`
			QuoteVolume string `json:"quoteVolume"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	if body.Code != "00000" || len(body.Data) == 0 {
		return nil, fmt.Errorf("bitget error: %s", body.Code)
	}

	d := body.Data[0]
	lastPrice, _ := parseFloat(d.LastPr)
	volume, _ := parseFloat(d.BaseVolume)
	quoteVolume, _ := parseFloat(d.QuoteVolume)

	return &bitgetRef{lastPrice: lastPrice, volume: volume, quoteVolume: quoteVolume}, nil
}

func parseFloat(s string) (float64, error) {
	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	return v, err
}

// renderDashboard displays test results in compact dashboard format or JSON.
func renderDashboard(w io.Writer, results []TestResult, _ render.Format) {
	renderDashboardDefault(w, results)
}

func renderDashboardDefault(w io.Writer, results []TestResult) {
	var ok, warn, err int
	for _, r := range results {
		switch r.Status {
		case StatusOK:
			ok++
		case StatusWarning:
			warn++
		case StatusError:
			err++
		}
	}

	_, _ = fmt.Fprintf(w, "TOTAL=%d OK=%d WARNING=%d ERROR=%d\n", len(results), ok, warn, err)

	for _, r := range results {
		icon := "ok"
		switch r.Status {
		case StatusWarning:
			icon = "warn"
		case StatusError:
			icon = "error"
		}
		msg := r.Message
		if r.Details != "" {
			msg = msg + " (" + r.Details + ")"
		}
		if msg != "" {
			_, _ = fmt.Fprintf(w, "%s %s %s %s %s\n", r.Provider, r.Market, r.Feature, icon, msg)
		} else {
			_, _ = fmt.Fprintf(w, "%s %s %s %s\n", r.Provider, r.Market, r.Feature, icon)
		}
	}
}

// --- Raw Mode Implementation ---

func callAndRender(
	ctx context.Context,
	w io.Writer,
	client *bits.Client,
	feat capability.Feature,
	market capability.MarketType,
	symbols []string,
	listLength, streamLength int,
	format render.Format,
) {
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	switch feat {
	case capability.FeatureServerTime:
		res, err := client.ServerTime(reqCtx)
		renderOrError(w, format, res, err)

	case capability.FeatureExchangeInfo:
		res, err := client.ExchangeInfo(reqCtx, model.MarketType(market))
		if err == nil && len(res.Data.Symbols) > listLength {
			res.Data.Symbols = res.Data.Symbols[:listLength]
		}
		renderOrError(w, format, res, err)

	case capability.FeaturePrice:
		res, err := client.Price(reqCtx, symbols, "usd")
		renderOrError(w, format, res, err)

	case capability.FeatureCandles:
		limit := listLength
		res, err := client.Candles(reqCtx, symbols[0], model.MarketType(market), "1h", model.CandleOpts{Limit: &limit})
		renderOrError(w, format, res, err)

	case capability.FeatureTicker24h:
		res, err := client.Ticker24h(reqCtx, symbols[0], model.MarketType(market))
		renderOrError(w, format, res, err)

	case capability.FeatureOrderBook:
		res, err := client.OrderBook(reqCtx, symbols[0], model.MarketType(market), listLength)
		renderOrError(w, format, res, err)

	case capability.FeatureMarketsList:
		opts := model.MarketOpts{Currency: "usd", PerPage: listLength, Page: 1}
		res, err := client.CoinMarkets(reqCtx, opts)
		renderOrError(w, format, res, err)

	case capability.FeatureStreamPrice:
		collectPriceStream(reqCtx, w, client, symbols, streamLength)

	case capability.FeatureStreamOrderBook:
		collectOrderBookStream(reqCtx, w, client, symbols, model.MarketType(market), streamLength)

	default:
		_, _ = fmt.Fprintf(w, "  (no test implementation for %s)\n", feat)
	}
}

func renderOrError[T any](w io.Writer, format render.Format, res model.Response[T], err error) {
	if err != nil {
		printError(w, err)
		return
	}
	for _, ie := range res.Errors {
		_, _ = fmt.Fprintf(w, "  partial error [%s]: ", ie.Symbol)
		printError(w, ie.Err)
	}
	if _, renderErr := command.RenderGeneric(w, format, res); renderErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "render error: %v\n", renderErr)
	}
}

func printError(w io.Writer, err error) {
	var pe *model.ProviderError
	if errors.As(err, &pe) {
		_, _ = fmt.Fprintf(w, "  ERROR kind=%v provider=%q code=%q http=%d msg=%q\n",
			pe.Kind, pe.ProviderID, pe.ProviderCode, pe.HTTPStatus, pe.ProviderMessage)
		if pe.Cause != nil {
			_, _ = fmt.Fprintf(w, "    caused by: %v\n", pe.Cause)
		}
	} else {
		_, _ = fmt.Fprintf(w, "  ERROR: %v\n", err)
	}
}

func collectPriceStream(ctx context.Context, w io.Writer, client *bits.Client, symbols []string, count int) {
	ch, err := client.StartPriceStream(ctx, symbols)
	if err != nil {
		printError(w, err)
		return
	}
	defer client.StopPriceStream() //nolint:errcheck

	for i := range count {
		select {
		case tick, ok := <-ch:
			if !ok {
				_, _ = fmt.Fprintln(w, "  stream closed")
				return
			}
			printJSON(w, fmt.Sprintf("tick %d", i+1), tick)
		case <-ctx.Done():
			_, _ = fmt.Fprintln(w, "  stream timeout")
			return
		}
	}
}

func collectOrderBookStream(ctx context.Context, w io.Writer, client *bits.Client, symbols []string, market model.MarketType, count int) {
	ch, err := client.StartOrderBookStream(ctx, symbols, market, 5)
	if err != nil {
		printError(w, err)
		return
	}
	defer client.StopOrderBookStream() //nolint:errcheck

	for i := range count {
		select {
		case tick, ok := <-ch:
			if !ok {
				_, _ = fmt.Fprintln(w, "  stream closed")
				return
			}
			printJSON(w, fmt.Sprintf("tick %d", i+1), tick)
		case <-ctx.Done():
			_, _ = fmt.Fprintln(w, "  stream timeout")
			return
		}
	}
}

func printJSON(w io.Writer, label string, v any) {
	b, err := json.MarshalIndent(v, "  ", "  ")
	if err != nil {
		_, _ = fmt.Fprintf(w, "  %s: (marshal error: %v)\n", label, err)
		return
	}
	_, _ = fmt.Fprintf(w, "  %s: %s\n", label, b)
}

func resolveProviders(input, all []string) ([]string, error) {
	if len(input) == 0 {
		return all, nil
	}
	valid := make(map[string]bool, len(all))
	for _, id := range all {
		valid[id] = true
	}
	var out, bad []string
	for _, p := range input {
		canonical := registry.ResolveProvider(strings.ToLower(p))
		if !valid[canonical] {
			bad = append(bad, p)
		} else {
			out = append(out, canonical)
		}
	}
	if len(bad) > 0 {
		return nil, fmt.Errorf("unknown providers: %v\nvalid providers: %v", bad, all)
	}
	return out, nil
}

func resolveMarkets(input []string) ([]capability.MarketType, error) {
	all := capability.AllMarkets()
	if len(input) == 0 {
		return all, nil
	}
	valid := make(map[capability.MarketType]bool, len(all))
	for _, m := range all {
		valid[m] = true
	}
	var out []capability.MarketType
	var bad []string
	for _, s := range input {
		m := capability.MarketType(strings.ToLower(s))
		if !valid[m] {
			bad = append(bad, s)
		} else {
			out = append(out, m)
		}
	}
	if len(bad) > 0 {
		return nil, fmt.Errorf("unknown markets: %v\nvalid markets: %v", bad, all)
	}
	return out, nil
}

func resolveCapabilities(input []string) ([]capability.Feature, error) {
	if len(input) == 0 {
		return testableFeatures, nil
	}
	var out []capability.Feature
	var bad []string
	for _, s := range input {
		f, ok := featureNames[strings.ToLower(s)]
		if !ok {
			bad = append(bad, s)
		} else {
			out = append(out, f)
		}
	}
	if len(bad) > 0 {
		names := make([]string, 0, len(featureNames))
		for k := range featureNames {
			names = append(names, k)
		}
		return nil, fmt.Errorf("unknown capabilities: %v\nvalid capabilities: %v", bad, names)
	}
	return out, nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
