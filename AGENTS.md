# AGENTS.md

## Project Overview

`bits` is a multi-provider crypto library and CLI tool written in Go. It supports CoinGecko, Binance, Bitget, WhiteBit, Crypto.com, and MEXC as data providers through a unified capability-based interface. All provider responses are wrapped in a typed `Response[T]` envelope with provenance tracking and automatic fallback.

**Architectural Principle**: The bits library (`github.com/mdnmdn/bits`) is the first citizen of this project. The CLI is a thin wrapper that uses the public library interface. External projects can import and extend the CLI commands.

## Build & Test

```sh
make build       # Build binary → ./bits
make test        # Run tests with -race
make lint        # Run golangci-lint
make clean       # Remove binary
```

Or directly:

```sh
go build -o bits ./cmd/bits
go test -race ./...
```

## Project Structure

```
bits/
├── bits.go                     # Main library: Client, NewProvider, NullProvider
├── client.go                  # Client implementation (alias of bits.go)
├── command/                   # CLI commands (exported for external use)
│   ├── commands.go            # Root command definition
│   ├── helpers.go             # Config loading, option resolution, render dispatch
│   ├── price.go              # price command
│   ├── ticker.go             # ticker command
│   ├── book.go               # book command
│   ├── candles.go            # candles command
│   ├── info.go               # info command
│   ├── time.go               # time command
│   ├── markets.go            # markets command
│   ├── stream.go             # stream commands
│   └── providers.go          # providers, capabilities commands
├── cmd/bits/main.go            # CLI entry point
├── capability/                # Capability system
├── config/                    # Multi-provider config (YAML + Env + .env)
├── model/                     # Provider-agnostic data types
├── provider/                  # Provider interfaces + implementations
│   ├── binance/               # Binance implementation (spot + futures)
│   ├── bitget/                # Bitget implementation (spot + futures)
│   ├── coingecko/             # CoinGecko implementation
│   ├── whitebit/              # WhiteBit implementation
│   ├── cryptocom/             # Crypto.com implementation
│   ├── mexc/                  # MEXC implementation
│   └── registry/              # NewProvider factory
├── resolve/                   # Resolver, symbol resolution
├── render/                    # Output renderers (exported)
│   ├── renderer.go            # Format type + ParseFormat
│   ├── provenance.go          # FallbackFootnote, ProviderLabel
│   ├── table/                 # Table renderers
│   ├── json/                  # JSON renderer
│   ├── yaml/                  # YAML renderer
│   ├── markdown/              # Markdown renderer
│   └── toon/                  # TOON renderer
├── process/                   # Data processors (exported)
│   ├── process.go             # Processor[T] type + Apply
│   ├── time.go                # TimeEnricher
│   ├── orderbook.go           # SpreadCalculator
│   └── candles.go             # CandleStats
├── internal/                  # CLI-only internals
│   ├── auth/                  # HMAC-SHA256 helpers
│   ├── logger/                # Structured logging
│   ├── tui/                   # Terminal UI components
│   └── ws/                    # WebSocket clients
└── examples/                   # Example applications
```

## Using the Library

```go
import "github.com/mdnmdn/bits"

// Create a provider-specific client
cfg := &config.Config{
    Binance: config.BinanceConfig{
        Spot: config.MarketConfig{Enabled: true},
    },
}
client := bits.NewProvider(cfg, "binance", bits.WithSymbolEngine())

// Use public methods
price, err := client.Price(ctx, []string{"btc", "eth"}, "usd")
```

## Using the CLI

External projects can import and extend the CLI:

```go
import "github.com/mdnmdn/bits/command"

func main() {
    command.Root.AddCommand(myCustomCmd)
    if err := command.Root.Execute(); err != nil {
        // handle error
    }
}
```

## Conventions

- **Go version**: 1.24 (pinned in go.mod and CI)
- **Binary name**: `bits`
- **Module path**: `github.com/mdnmdn/bits`
- **Test framework**: `testify/assert` + `net/http/httptest`
- **Provider pattern**: implement `provider.Provider` + capability interfaces; return `model.Response[T]`
- **Response pattern**: every provider call returns `model.Response[T]` with `Provider` and `Market` populated
- **Error pattern**: all errors cross provider boundaries as `*model.ProviderError`; use `errors.As` to inspect; never `fmt.Errorf` at provider boundaries
- **Config location**: platform-specific (`~/Library/Application Support/bits-cli/` on macOS, `~/.config/bits/` on Linux)
- **Output modes**: global `-o table` / `-o json` / `-o yaml` / `-o toon` / `-o markdown` flag
- **Stream contract**: stdout = data only, stderr = diagnostics/warnings
- **Formatting**: all code must be `gofmt`-clean
- **Commits**: conventional commit style (`feat:`, `fix:`, `chore:`)

## Command Pattern (CLI)

Every CLI command uses the public bits library:

```go
func runPrice(cmd *cobra.Command, args []string) error {
    cfg, _, err := command.LoadConfig()           // load config
    providerID, market, format := command.ResolveOptions(cmd)

    client := bits.NewProvider(cfg, providerID, bits.WithSymbolEngine())
    res, err := client.Price(ctx, args, currency)

    if ok, err := command.RenderGeneric(w, format, res); ok || err != nil {
        return err
    }
    return table.RenderPrices(w, res)
}
```

## Adding a New Provider

See `_docs/provider-structure.md` for the full step-by-step guide.

## References

The _docs folder contains all the documentation:
- `_docs` general documents
- `_docs/wip/` work in progress folder
- `_docs/backlog/` future tasks
- `_docs/_old/` old docs kept for historical reason
- `_docs/providers/` provider documentation
    - `provider-index.md`  index of the documentation
    - `_drafts/`  documentation drafts  
- Use `bits capabilities` to inspect the live capability matrix
- See `_docs/renderers.md` for output format specifications (table, json, yaml, toon, markdown)
- See `_docs/architecture.md` for full architecture and provider capability matrix
- See `_docs/error-handling.md` for error handling design (`ProviderError`, `ErrorKind`, provider helpers)
