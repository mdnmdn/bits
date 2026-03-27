# CLAUDE.md

## Project Overview

`bits` is a multi-provider crypto CLI tool written in Go. It supports CoinGecko, Binance, and Bitget as data providers through a unified capability-based interface. All provider responses are wrapped in a typed `Response[T]` envelope with provenance tracking and automatic fallback.

## Build & Test

```sh
make build       # Build binary → ./bits
make test        # Run tests with -race
make lint        # Run golangci-lint
make clean       # Remove binary
```

Or directly:

```sh
go build -o bits .
go test -race ./...
```

## Project Structure

```
bits/
├── main.go
├── cmd/
│   ├── root.go              # RootCmd, Execute(), global flags (-p, -m, -o, -l)
│   ├── factory.go           # loadConfig(), newResolver(), flag helpers
│   ├── providers.go         # bits providers
│   ├── capabilities.go      # bits capabilities [--provider id]
│   ├── time.go              # bits time
│   ├── price.go             # bits price <id>...
│   ├── ticker.go            # bits ticker <symbol>...
│   ├── book.go              # bits book <symbol>
│   ├── candles.go           # bits candles <symbol>
│   ├── info.go              # bits info [--symbol S]
│   ├── markets.go           # bits markets
│   ├── stream.go            # bits stream (group)
│   ├── stream_price.go      # bits stream price <id>...
│   └── stream_book.go       # bits stream book <symbol>
├── internal/
│   ├── auth/
│   │   └── signature.go     # HMAC-SHA256 helpers
│   ├── capability/
│   │   └── capability.go    # MarketType, Feature, CapabilityKey, CapabilityMatrix
│   ├── config/
│   │   └── config.go        # Multi-provider config (YAML + Env + .env)
│   ├── model/               # Provider-agnostic data types
│   │   ├── market.go        # MarketType alias + constants
│   │   ├── response.go      # Response[T], ItemError
│   │   ├── exchange.go      # ServerTime, ExchangeInfo, Symbol
│   │   ├── candle.go        # Candle, CandleOpts
│   │   ├── ticker.go        # Ticker24h
│   │   ├── orderbook.go     # OrderBook, OrderBookEntry
│   │   ├── price.go         # CoinPrice
│   │   ├── coin.go          # CoinMarket, MarketOpts
│   │   └── errors.go        # ErrUnsupportedMarket, ErrUnsupportedFeature
│   ├── provider/            # Provider interfaces
│   │   ├── provider.go      # Provider base interface
│   │   ├── exchange.go      # ExchangeProvider
│   │   ├── aggregator.go    # AggregatorProvider
│   │   ├── capability.go    # PriceProvider, CandleProvider, TickerProvider, OrderBookProvider
│   │   ├── stream.go        # PriceStreamProvider, OrderBookStreamProvider
│   │   ├── binance/         # Binance implementation
│   │   ├── bitget/          # Bitget implementation
│   │   └── coingecko/       # CoinGecko implementation
│   ├── registry/
│   │   └── registry.go      # NewProvider factory (separate pkg avoids import cycle)
│   ├── resolve/
│   │   ├── resolver.go      # Resolver, ResolutionOpts, Resolve() with fallback
│   │   ├── require.go       # Require[T] type assertion helper
│   │   └── fanout.go        # FanOut[T] parallel multi-symbol helper
│   ├── render/
│   │   ├── renderer.go      # Format type + ParseFormat
│   │   ├── provenance.go    # FallbackFootnote, ProviderLabel
│   │   ├── json/            # Generic JSON renderer
│   │   └── table/           # Table renderers per data type
│   ├── process/
│   │   ├── process.go       # Processor[T] type + Apply combinator
│   │   ├── time.go          # TimeEnricher (latency + clock skew)
│   │   ├── orderbook.go     # SpreadCalculator
│   │   └── candles.go       # CandleStats
│   └── ws/
│       ├── base_client.go   # Generic WebSocket client (reconnect, backoff)
│       └── client.go        # CoinGecko WebSocket client (ActionCable)
```

## Conventions

- **Go version**: 1.26 (pinned in go.mod and CI)
- **Binary name**: `bits`
- **Module path**: `github.com/mdnmdn/bits`
- **Test framework**: `testify/assert` + `net/http/httptest`
- **Provider pattern**: implement `provider.Provider` + capability interfaces; return `model.Response[T]`
- **Response pattern**: every provider call returns `model.Response[T]` with `Provider` and `Market` populated
- **Config location**: platform-specific (`~/Library/Application Support/bits-cli/` on macOS, `~/.config/bits/` on Linux)
- **Auth tiers**: `demo`, `paid` (CoinGecko only)
- **Output modes**: global `-o table` / `-o json` flag
- **Stream contract**: stdout = data only, stderr = diagnostics/warnings
- **Formatting**: all code must be `gofmt`-clean
- **Commits**: conventional commit style (`feat:`, `fix:`, `chore:`)

## Command Pattern

Every command follows this pattern:

```go
func runXxx(cmd *cobra.Command, args []string) error {
    cfg, err := loadConfig()                        // load config
    opts := resolveOpts(cmd)                        // read -p, -m, -l flags
    format := resolveFormat(cmd)                    // read -o flag
    resolver := newResolver(cfg)

    p, market, fallback, err := resolver.Resolve(ctx, capability.FeatureXxx, opts)
    tp, err := resolve.Require[provider.XxxProvider](p, "xxx")

    res, err := tp.XxxMethod(ctx, args[0], market)  // or FanOut for multi-symbol

    if fallback { res.Fallback = true; ... }         // annotate if fallback
    res = process.Apply(res, process.XxxEnricher)    // optional enrichment

    switch format {
    case "json": return renderjson.Render(os.Stdout, res)
    default:     return rendertable.RenderXxx(os.Stdout, res)
    }
}
```

## Adding a New Provider

1. Create `internal/provider/<name>/client.go` implementing `Provider` base interface
2. Add capability interfaces as supported; all methods return `model.Response[T]` with `Provider` and `Market` populated
3. Register capabilities in `Capabilities()` using `capability.NewCapabilityMatrix(...)`
4. Add config struct to `internal/config/config.go`; add env var overrides in `applyEnvOverrides`
5. Register in `internal/registry/registry.go` — one new case in `NewProvider` and `AllProviderIDs`

## API Reference

- **CoinGecko API docs**: https://docs.coingecko.com
- Use `bits capabilities` to inspect the live capability matrix
