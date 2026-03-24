# CLAUDE.md

## Project Overview

`bits` is a multi-provider crypto CLI tool written in Go. It provides both standard CLI commands and an interactive TUI for browsing cryptocurrency market data, starting with support for the CoinGecko API.

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
├── main.go                        # Entry point
├── cmd/                           # Cobra commands (auth, status, price, markets, search, trending, history, top_gainers_losers, watch, tui, version)
├── internal/
│   ├── provider/                  # Provider interfaces and registry
│   │   ├── types.go               # Provider interface definition (Capabilities)
│   │   ├── registry.go            # Provider factory
│   │   ├── coingecko/             # CoinGecko implementation (All capabilities)
│   │   ├── binance/               # Binance implementation (Prices, Tickers, OrderBook)
│   │   └── bitget/                # Bitget implementation (Prices, Tickers)
│   ├── config/
│   │   ├── config.go              # Multi-provider config (YAML + Env)
│   ├── auth/
│   │   └── signature.go           # Shared HMAC-SHA256 signature helpers
│   ├── ws/
│   │   ├── base_client.go         # Generic WebSocket client (reconnect, backoff)
│   │   ├── client.go              # CoinGecko WebSocket client
│   └── tui/
│       ├── styles.go              # Shared lipgloss styles
│       ├── markets.go             # Markets TUI model
│       ├── trending.go            # Trending TUI model
│       ├── detail.go              # Coin detail view
│       └── chart.go               # Braille price chart
├── Makefile
├── .goreleaser.yml
├── install.sh
└── .github/workflows/             # CI and Release
```

## Conventions

- **Go version**: 1.26 (pinned in go.mod and CI)
- **Binary name**: `bits`
- **Module path**: `github.com/mdnmdn/bits`
- **Test framework**: `testify/assert` + `net/http/httptest`
- **Provider Pattern**: All commands use `provider.Provider` interface via `newAPIClient` factory.
- **Config location**: `~/.config/coingecko-cli/config.yaml`
- **Auth tiers**: demo, paid
- **Error model**: `ErrInvalidAPIKey`, `ErrPlanRestricted`, `ErrRateLimited`
- **Output modes**: global `-o json` / `--output json` flag
- **Stream contract**: stdout = data only, stderr = diagnostics/warnings
- **Color output**: respects `NO_COLOR` env var and TTY detection
- **Formatting**: all code must be `gofmt`-clean
- **Commits**: conventional commit style (`feat:`, `fix:`, `chore:`)
- **TUI framework**: Bubble Tea (bubbletea) + Lip Gloss (lipgloss) + Bubbles + ntcharts
- **Interactive prompts**: `huh` (Charm ecosystem)

## API Reference (CoinGecko)

- **API docs**: https://docs.coingecko.com
- **Command catalog**: `bits commands` (hidden) outputs machine-readable JSON

### CLI → OAS Endpoint Mapping

| CLI Command | API Endpoint | OAS Operation ID |
|-------------|-------------|------------------|
| `bits price` | `/simple/price` | `simple-price` |
| `bits markets` | `/coins/markets` | `coins-markets` |
| `bits search` | `/search` | `search-data` |
| `bits trending` | `/search/trending` | `trending-search` |
| `bits history --date` | `/coins/{id}/history` | `coins-id-history` |
| `bits history --days` | `/coins/{id}/market_chart` | `coins-id-market-chart` |
| `bits history --days --interval hourly` | `/coins/{id}/market_chart/range` | `coins-id-market-chart-range` |
| `bits history --days --ohlc` | `/coins/{id}/ohlc` | `coins-id-ohlc` |
| `bits history --from/--to` | `/coins/{id}/market_chart/range` | `coins-id-market-chart-range` |
| `bits history --from/--to --interval hourly` | `/coins/{id}/market_chart/range` | `coins-id-market-chart-range` |
| `bits history --from/--to --ohlc` | `/coins/{id}/ohlc/range` | `coins-id-ohlc-range` |
| `bits top-gainers-losers` | `/coins/top_gainers_losers` | `coins-top-gainers-losers` |
| `bits watch` | `wss://stream.coingecko.com/v1` | — |

## Distribution

- **Homebrew**: tap repo at `coingecko/homebrew-coingecko-cli`
- **Goreleaser**: `.goreleaser.yml` generates Homebrew formula
- **Install script**: `install.sh` downloads the latest release binary
- **Go install**: `go install github.com/mdnmdn/bits@latest`
