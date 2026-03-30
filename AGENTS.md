# CLAUDE.md

## Project Overview

`bits` is a multi-provider crypto CLI tool written in Go. It supports CoinGecko, Binance, Bitget, WhiteBit, Crypto.com, and MEXC as data providers through a unified capability-based interface. All provider responses are wrapped in a typed `Response[T]` envelope with provenance tracking and automatic fallback.

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
│   ├── render.go            # shared render dispatch helper
│   ├── tui.go               # bits tui
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
├── pkg/                     # Public library packages (importable by external tools)
│   ├── bits/                # High-level facade: Client, GetPrice, ComparePrices
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
│   ├── provider/            # Provider interfaces + implementations
│   │   ├── provider.go      # Provider base interface
│   │   ├── exchange.go      # ExchangeProvider
│   │   ├── aggregator.go    # AggregatorProvider
│   │   ├── capability.go    # PriceProvider, CandleProvider, TickerProvider, OrderBookProvider
│   │   ├── stream.go        # PriceStreamProvider, OrderBookStreamProvider
│   │   ├── binance/         # Binance implementation (spot + futures)
│   │   ├── bitget/          # Bitget implementation (spot + futures)
│   │   ├── coingecko/       # CoinGecko implementation
│   │   ├── whitebit/        # WhiteBit implementation
│   │   ├── cryptocom/       # Crypto.com implementation
│   │   ├── mexc/            # MEXC implementation
│   │   └── registry/
│   │       └── registry.go  # NewProvider factory (separate pkg to avoid import cycle)
│   └── resolve/
│       ├── resolver.go      # Resolver, ResolutionOpts, Resolve() with fallback
│       ├── require.go       # Require[T] type assertion helper
│       ├── fanout.go        # FanOut[T] parallel multi-symbol helper
│       └── symbol/          # Symbol resolution + normalization
├── internal/                # CLI-only internals (not importable by external tools)
│   ├── auth/
│   │   └── signature.go     # HMAC-SHA256 helpers (used by Bitget)
│   ├── logger/              # Structured logging helpers
│   ├── render/
│   │   ├── renderer.go      # Format type + ParseFormat
│   │   ├── provenance.go    # FallbackFootnote, ProviderLabel
│   │   ├── json/            # Generic JSON renderer
│   │   ├── yaml/            # Generic YAML renderer
│   │   ├── toon/            # Generic TOON renderer
│   │   ├── markdown/        # Generic Markdown renderer
│   │   └── table/           # Table renderers per data type
│   ├── process/
│   │   ├── process.go       # Processor[T] type + Apply combinator
│   │   ├── time.go          # TimeEnricher (latency + clock skew)
│   │   ├── orderbook.go     # SpreadCalculator
│   │   └── candles.go       # CandleStats
│   ├── tui/                 # Terminal UI components
│   │   └── section/
│   └── ws/
│       ├── base_client.go   # Generic WebSocket client (reconnect, backoff)
│       └── client.go        # CoinGecko WebSocket client (ActionCable)
├── examples/
│   ├── basic_usage/         # Fetch price from a single provider
│   └── price_comparison/    # Compare prices across multiple exchanges
```

## Conventions

- **Go version**: 1.24 (pinned in go.mod and CI)
- **Binary name**: `bits`
- **Module path**: `github.com/mdnmdn/bits`
- **Test framework**: `testify/assert` + `net/http/httptest`
- **Provider pattern**: implement `provider.Provider` + capability interfaces; return `model.Response[T]`
- **Response pattern**: every provider call returns `model.Response[T]` with `Provider` and `Market` populated
- **Config location**: platform-specific (`~/Library/Application Support/bits-cli/` on macOS, `~/.config/bits/` on Linux)
- **Output modes**: global `-o table` / `-o json` / `-o yaml` / `-o toon` / `-o markdown` flag
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



See `_docs/provider-structure.md` for the full step-by-step guide 

## References

The _docs folder contains all the documentation:
- `_docs` general documents
- `_docs/wip` work in progress
- `_docs/backlog` future tasks
- `_docs/_old` old docs kept for historical reason
- `_docs/providers` dirty documentationg
 
- Use `bits capabilities` to inspect the live capability matrix
- See `_docs/renderers.md` for output format specifications (table, json, yaml, toon, markdown)
- See `_docs/architecture.md` for full architecture and provider capability matrix
