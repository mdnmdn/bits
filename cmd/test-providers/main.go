// test-providers exercises every selected capability across selected providers and markets.
// It is a diagnostic/integration tool, not part of the production CLI.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// testableFeatures are the capabilities that have a client-method implementation here.
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

// featureNames maps lowercase name → Feature constant.
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

// marketIndependent features should be called only once per provider regardless of market.
var marketIndependent = map[capability.Feature]bool{
	capability.FeatureServerTime:  true,
	capability.FeaturePrice:       true,
	capability.FeatureMarketsList: true,
}

func main() {
	var (
		providerFlags    []string
		marketFlags      []string
		symbolFlags      []string
		capabilityFlags  []string
		listLength       int
		streamLength     int
		outputFormat     string
	)

	pflag.StringSliceVar(&providerFlags, "provider", nil, "providers to test (default: all); repeat or comma-separate")
	pflag.StringSliceVar(&marketFlags, "markets", nil, "markets to test (default: all)")
	pflag.StringSliceVar(&symbolFlags, "symbols", nil, "symbols (default: BTCUSDT)")
	pflag.StringSliceVar(&capabilityFlags, "capabilities", nil, "capabilities to test (default: all); case-insensitive")
	pflag.IntVar(&listLength, "list-length", 3, "max records in list responses")
	pflag.IntVar(&streamLength, "stream-length", 3, "number of streaming ticks to collect")
	pflag.StringVar(&outputFormat, "output", "json", "output format: json, yaml, markdown, toon")
	pflag.Parse()

	// --- resolve providers ---
	allProviders := registry.AllProviderIDs()
	providers, err := resolveProviders(providerFlags, allProviders)
	if err != nil {
		fatalf("%v", err)
	}

	// --- resolve markets ---
	markets, err := resolveMarkets(marketFlags)
	if err != nil {
		fatalf("%v", err)
	}

	// --- resolve symbols ---
	symbols := symbolFlags
	if len(symbols) == 0 {
		symbols = []string{"BTCUSDT"}
	}

	// --- resolve capabilities ---
	features, err := resolveCapabilities(capabilityFlags)
	if err != nil {
		fatalf("%v", err)
	}

	// --- resolve output format ---
	format := render.ParseFormat(outputFormat)
	if format == render.FormatTable {
		fmt.Fprintln(os.Stderr, "note: table format not supported in test-providers, using json")
		format = render.FormatJSON
	}

	// --- load config ---
	cfg, err := command.LoadConfig()
	if err != nil {
		fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	// --- main loop: capability → provider → market ---
	for _, feat := range features {
		fmt.Printf("\n╔══ Capability: %s ══\n", feat)
		calledProviders := map[string]bool{}

		for _, providerID := range providers {
			client := bits.NewProvider(cfg, providerID)
			caps := client.Capabilities()

			for _, market := range markets {
				key := capability.CapabilityKey{Market: market, Feature: feat}
				if !caps[key] {
					continue
				}
				if marketIndependent[feat] && calledProviders[providerID] {
					continue
				}

				fmt.Printf("╠─ %s / %s\n", providerID, market)
				callAndRender(ctx, os.Stdout, client, feat, market, symbols, listLength, streamLength, format)

				if marketIndependent[feat] {
					calledProviders[providerID] = true
				}
			}
		}
	}
}

// callAndRender dispatches to the correct client method based on feature and renders the result.
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
		fmt.Fprintf(w, "  (no test implementation for %s)\n", feat)
	}
}

func renderOrError[T any](w io.Writer, format render.Format, res model.Response[T], err error) {
	if err != nil {
		printError(w, err)
		return
	}
	for _, ie := range res.Errors {
		fmt.Fprintf(w, "  partial error [%s]: ", ie.Symbol)
		printError(w, ie.Err)
	}
	if _, renderErr := command.RenderGeneric(w, format, res); renderErr != nil {
		fmt.Fprintf(os.Stderr, "render error: %v\n", renderErr)
	}
}

func printError(w io.Writer, err error) {
	var pe *model.ProviderError
	if errors.As(err, &pe) {
		fmt.Fprintf(w, "  ERROR kind=%v provider=%q code=%q http=%d msg=%q\n",
			pe.Kind, pe.ProviderID, pe.ProviderCode, pe.HTTPStatus, pe.ProviderMessage)
		if pe.Cause != nil {
			fmt.Fprintf(w, "    caused by: %v\n", pe.Cause)
		}
	} else {
		fmt.Fprintf(w, "  ERROR: %v\n", err)
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
				fmt.Fprintln(w, "  stream closed")
				return
			}
			printJSON(w, fmt.Sprintf("tick %d", i+1), tick)
		case <-ctx.Done():
			fmt.Fprintln(w, "  stream timeout")
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
				fmt.Fprintln(w, "  stream closed")
				return
			}
			printJSON(w, fmt.Sprintf("tick %d", i+1), tick)
		case <-ctx.Done():
			fmt.Fprintln(w, "  stream timeout")
			return
		}
	}
}

func printJSON(w io.Writer, label string, v any) {
	b, err := json.MarshalIndent(v, "  ", "  ")
	if err != nil {
		fmt.Fprintf(w, "  %s: (marshal error: %v)\n", label, err)
		return
	}
	fmt.Fprintf(w, "  %s: %s\n", label, b)
}

// --- flag resolution helpers ---

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
