# CLAUDE.md

## Project Overview

`cg` is a CoinGecko CLI tool written in Go. It provides both standard CLI commands and an interactive TUI for browsing cryptocurrency market data from the CoinGecko API.

## Build & Test

```sh
make build       # Build binary → ./cg
make test        # Run tests with -race
make lint        # Run golangci-lint
make clean       # Remove binary
```

Or directly:

```sh
go build -o cg .
go test -race ./...
```

## Project Structure

```
coingecko-cli/
├── main.go                        # Entry point
├── cmd/                           # Cobra commands (auth, status, price, markets, search, trending, history, top_gainers_losers, tui, version)
├── internal/
│   ├── api/
│   │   ├── client.go              # HTTP client, auth, error handling
│   │   ├── client_test.go         # httptest-based tests
│   │   ├── coins.go               # API endpoint methods
│   │   └── types.go               # JSON response structs
│   ├── config/
│   │   ├── config.go              # Viper-based config (API key, tier)
│   │   └── config_test.go
│   ├── display/
│   │   ├── banner.go              # ASCII logo, welcome box, brand banner
│   │   ├── table.go               # Table rendering
│   │   ├── format.go              # Price/number formatting
│   │   └── color.go               # ANSI color (NO_COLOR/TTY aware)
│   ├── export/
│   │   └── csv.go                 # CSV file export
│   └── tui/
│       ├── styles.go              # Shared lipgloss styles, brand colors, frame/layout helpers
│       ├── markets.go             # Markets TUI model
│       ├── trending.go            # Trending TUI model
│       ├── detail.go              # Coin detail view
│       └── chart.go               # Braille price chart (ntcharts)
├── Makefile
├── .goreleaser.yml
├── install.sh
└── .github/workflows/             # CI (lint, test, build) and Release (goreleaser)
```

## Conventions

- **Go version**: 1.26 (pinned in go.mod and CI)
- **Binary name**: `cg`
- **Module path**: `github.com/coingecko/coingecko-cli`
- **Test framework**: `testify/assert` + `net/http/httptest`
- **API client tests**: use `httptest.NewServer` with `SetBaseURL`, never hit real API
- **Config location**: `~/.config/coingecko-cli/config.yaml` (dir `0700`, file `0600`)
- **Auth tiers**: demo, paid — demo uses `x-cg-demo-api-key` header and `api.coingecko.com`, paid uses `x-cg-pro-api-key` and `pro-api.coingecko.com`
- **Error model**: `ErrInvalidAPIKey` (401), `ErrPlanRestricted` (403), `ErrRateLimited` (429) — all user-friendly messages
- **Paid-only endpoints**: use `requirePaid()` pre-flight check before API call
- **Output modes**: global `-o json` / `--output json` flag on all data commands; JSON mode emits raw API structs to stdout
- **Stream contract**: stdout = data only, stderr = diagnostics/warnings (use `warnf()` helper in `cmd/output.go`)
- **Color output**: respects `NO_COLOR` env var and TTY detection — never hardcode ANSI in output that may be piped
- **Formatting**: all code must be `gofmt`-clean
- **Commits**: conventional commit style (`feat:`, `fix:`, `chore:`)
- **TUI framework**: Bubble Tea (bubbletea) + Lip Gloss (lipgloss) + Bubbles + ntcharts
- **Interactive prompts**: `huh` (Charm ecosystem) — API key input uses password echo mode
- **Auth key input priority**: env var `CG_API_KEY` > `--key` flag > interactive prompt — env vars are preferred to avoid shell history/process listing exposure

## API Reference

- **OAS specs**: https://github.com/coingecko/coingecko-api-oas (`coingecko-demo.json`, `coingecko-pro.json`)
- **API docs**: https://docs.coingecko.com
- **Rate limiting**: 429 responses include `Retry-After: <seconds>` header (integer only, no HTTP-date)
- **Structured errors**: `-o json` mode emits `{"error":"rate_limited","message":"...","retry_after":54}` to stderr
- **Dry run**: `--dry-run` on any data command outputs the API request as JSON without executing
- **Command catalog**: `cg commands` (hidden) outputs machine-readable JSON with all commands, flags, OAS operation IDs, and enums

### CLI → OAS Endpoint Mapping

| CLI Command | API Endpoint | OAS Operation ID |
|-------------|-------------|------------------|
| `cg price` | `/simple/price` | `simple-price` |
| `cg markets` | `/coins/markets` | `coins-markets` |
| `cg search` | `/search` | `search-data` |
| `cg trending` | `/search/trending` | `trending-search` |
| `cg history --date` | `/coins/{id}/history` | `coins-id-history` |
| `cg history --days` | `/coins/{id}/market_chart` | `coins-id-market-chart` |
| `cg history --days --ohlc` | `/coins/{id}/ohlc` | `coins-id-ohlc` |
| `cg history --from/--to` | `/coins/{id}/market_chart/range` | `coins-id-market-chart-range` |
| `cg history --from/--to --ohlc` | `/coins/{id}/ohlc/range` | `coins-id-ohlc-range` |
| `cg top-gainers-losers` | `/coins/top_gainers_losers` | `coins-top-gainers-losers` |

## Key Design Decisions

- CoinGecko `/coins/{id}/market_chart/range` expects UNIX timestamps in seconds — CLI accepts `YYYY-MM-DD` and converts in the command layer
- CoinGecko `/coins/{id}/history` uses `DD-MM-YYYY` date format — CLI accepts `YYYY-MM-DD` and converts
- Symbol resolution (for `cg price --symbols`) uses `/search` endpoint, picks exact case-insensitive match with highest market_cap_rank
- TUI detail view fetches coin detail + OHLC concurrently via `tea.Batch`
- `RateLimitError` typed error carries `RetryAfter` seconds, satisfies `errors.Is(err, ErrRateLimited)` via custom `Is()` method
- API text (coin names, symbols) sanitized via `display.SanitizeCell` to strip terminal escape injection
