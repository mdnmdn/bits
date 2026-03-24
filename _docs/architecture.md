# bits Architecture

## Overview

`bits` is a multi-provider crypto CLI tool written in Go. It uses a capability-based provider architecture that allows different data sources (CoinGecko, Binance, Bitget) to be used interchangeably through a unified command interface.

## Provider Architecture

### Core Design

The provider system is built on Go interfaces. A base `Provider` interface defines the minimum contract (price and OHLC data), while optional capability interfaces extend functionality for providers that support additional features.

```
internal/provider/types.go
├── Provider (base)           → ID, SetUserAgent, SimplePrice, CoinOHLC
├── SymbolPricer              → SimplePriceBySymbols
├── MarketLister              → CoinMarkets, FetchAllMarkets
├── Searcher                  → Search
├── TrendingProvider          → SearchTrending
├── HistoricalProvider        → CoinHistory, CoinMarketChart, CoinMarketChartRange, CoinOHLCRange
├── GainersLosersProvider     → TopGainersLosers
├── DetailProvider            → CoinDetail
├── TickerProvider            → Ticker24h
└── OrderBookProvider         → OrderBook
```

### Provider Capabilities

| Capability | CoinGecko | Binance | Bitget |
|-----------|-----------|---------|--------|
| Price (SimplePrice) | Yes | Yes | Yes |
| OHLC/Candles | Yes | Yes | Yes |
| Symbol Price Lookup | Yes | — | — |
| Market Listings | Yes | — | — |
| Search | Yes | — | — |
| Trending | Yes | — | — |
| Historical Charts | Yes | — | — |
| Gainers/Losers | Yes | — | — |
| Coin Detail | Yes | — | — |
| 24h Ticker | — | Yes | Yes |
| Order Book | — | Yes | — |

### Generic Data Model

All providers return types from `internal/model/`, not provider-specific types. This decouples command handlers and TUI views from any particular API's response format.

Key types:
- `model.PriceResponse` — `map[string]map[string]float64`
- `model.OHLCData` — `[][]float64` (each: `[timestamp, open, high, low, close]`)
- `model.Ticker24h` — 24hr stats struct
- `model.OrderBook` — Bids/asks with float64 prices
- `model.MarketCoin`, `model.CoinDetail`, etc.

### Runtime Capability Checking

Commands check for capabilities at runtime using Go type assertions:

```go
tp, ok := client.(provider.TickerProvider)
if !ok {
    return fmt.Errorf("%s provider does not support ticker data", client.ID())
}
```

This allows commands to gracefully degrade when a provider doesn't support a feature.

## Provider Implementations

### CoinGecko (`internal/provider/coingecko/`)
- HTTP client with Viper-based config
- Supports demo and paid API tiers
- Internal types for JSON unmarshaling, conversion layer to model types
- All capability interfaces implemented

### Binance (`internal/provider/binance/`)
- Uses `go-binance/v2` library
- Auth: API key + secret (HMAC-SHA256 handled by library)
- Testnet support via config
- Trading methods available (unexposed via CLI)

### Bitget (`internal/provider/bitget/`)
- Raw HTTP client, no external API dependency
- Auth: 3-part (key + secret + passphrase) with HMAC-SHA256 signatures
- Trading pairs cache with 5-minute TTL
- Trading methods available (unexposed via CLI)

## Configuration

Config file: `~/.config/coingecko-cli/config.yaml`

```yaml
provider: coingecko        # default provider
api_key: "CG-xxx"          # CoinGecko API key
tier: demo                 # CoinGecko tier (demo/paid)

binance:
  api_key: ""
  api_secret: ""
  use_testnet: false

bitget:
  key: ""
  secret: ""
  passphrase: ""
```

Environment variable overrides (take priority over config file):
- `BITS_PROVIDER` — active provider
- `BITS_COINGECKO_API_KEY`, `BITS_COINGECKO_TIER`
- `BITS_BINANCE_API_KEY`, `BITS_BINANCE_API_SECRET`
- `BITS_BITGET_KEY`, `BITS_BITGET_SECRET`, `BITS_BITGET_PASSPHRASE`

## CLI Provider Selection

The `--provider` / `-p` flag selects the active provider per-command:

```bash
bits price --ids bitcoin                    # CoinGecko (default)
bits price --ids BTCUSDT -p binance         # Binance
bits ticker BTCUSDT -p binance              # Binance 24h stats
bits orderbook BTCUSDT -p binance           # Binance order book
```

Resolution order: `--provider` flag → `BITS_PROVIDER` env → config file → `"coingecko"` default.

## Adding a New Provider

1. Create `internal/provider/<name>/` with `client.go`, `types.go`, `market.go`
2. Implement at minimum the `Provider` interface (ID, SetUserAgent, SimplePrice, CoinOHLC)
3. Implement additional capability interfaces as supported
4. Add a config struct in `internal/config/config.go`
5. Register in `internal/provider/registry.go`
6. Add env var overrides in `config.applyEnvOverrides()`

## Project Structure

```
bits/
├── main.go
├── cmd/                              # Cobra commands
│   ├── root.go                       # Root command, --provider/-p, --output/-o flags
│   ├── client_factory.go             # Provider instantiation with flag resolution
│   ├── price.go, markets.go, ...     # Commands with capability checks
│   ├── ticker.go                     # 24h ticker (Binance/Bitget)
│   └── orderbook.go                  # Order book (Binance)
├── internal/
│   ├── model/                        # Provider-agnostic data types and errors
│   │   ├── types.go
│   │   └── errors.go
│   ├── provider/
│   │   ├── types.go                  # Capability interfaces
│   │   ├── registry.go               # Provider factory
│   │   ├── coingecko/                # CoinGecko implementation
│   │   ├── binance/                  # Binance implementation
│   │   └── bitget/                   # Bitget implementation
│   ├── config/                       # Multi-provider config
│   ├── display/                      # Terminal output formatting
│   ├── export/                       # CSV export
│   ├── tui/                          # Bubble Tea TUI
│   └── ws/                           # WebSocket streaming (CoinGecko)
```
