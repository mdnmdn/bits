# Symbol Engine

The symbol engine provides unified symbol handling across all providers, converting between various symbol formats and normalizing output.

## Overview

Different exchanges use different symbol formats:
- Binance: `BTCUSDT` (concatenated)
- WhiteBit: `BTC_USDT` (underscore separator)
- Crypto.com: `BTC_USDT` (spot), `BTCUSD-PERP` (futures)

The symbol engine handles:
1. **Input normalization**: Converts user input (`BTC-USDT`, `btc_usdt`, `BTCUSDT`) to provider-specific format
2. **Output normalization**: Always displays symbols in `BASE-QUOTE` format (e.g., `BTC-USDT`)
3. **Disk caching**: Caches exchange symbol lists to reduce API calls

## Architecture

```
User Input: "btc_usdt"
       ↓
Symbol Engine (lookup tables per {provider, market})
  - Parse: btc → BTC, usdt → USDT
  - Load provider-specific, market-specific lookup table
  - Find symbol with BaseAsset=BTC, QuoteAsset=USDT
       ↓
Provider Native: "BTC_USDT" (WhiteBit) or "BTCUSDT" (Binance)
       ↓
Output: Symbol = "BTC-USDT" (normalized)
        OriginalSymbol = "BTC_USDT" or "BTCUSDT" (provider-native)
```

Input resolution is only active when `bits.WithSymbolEngine()` option is used when creating a client. Without it, input symbols are passed directly to the provider.

## Files

| File | Purpose |
|------|---------|
| `resolve/symbol/normalize.go` | Symbol normalization utility: `NormalizeSymbol(s)` → `BASE-QUOTE` form |
| `resolve/symbol/engine.go` | Symbol engine with provider-specific, market-specific lookup tables |
| `resolve/symbol/translators/` | Provider-specific symbol translation |
| `model/normalize.go` | Core normalization function used by providers in output structs |

## Provider Patterns

| Provider | Spot Symbol | Futures Symbol | Strategy |
|----------|-------------|---------------|----------|
| Binance | BTCUSDT | BTCUSDT | Same symbol, different API |
| WhiteBit | BTC_USDT | BTC_PERP | Different symbols per market |
| Crypto.com | BTC_USDT | BTCUSD-PERP | Different symbols per market |
| MEXC | BTCUSDT | BTC_USDT | Underscore differentiates futures |
| Bitget | BTCUSDT | BTCUSDT_PERP | Suffix differentiates futures |
| OKX | BTC-USDT | BTC-USDT-SWAP | `-SWAP` suffix; linear `BTC-USDT-SWAP`, inverse `BTC-USD-SWAP` |
| Bitvavo | BTC-EUR | — | Already canonical; spot only |
| Bybit EU | BTCUSDT | — | Spot only |
| Kraken | — | PF_XBTUSD | `PF_` perpetual prefix, BTC is `XBT`; `BTC-USD` maps to `PF_XBTUSD` |

OKX, Bitvavo, Bybit EU and Kraken have no translator in `resolve/symbol/translators/`: each provider converts any accepted spelling (`BTC-USDT`, `BTC_USDT`, `BTCUSDT`) to its native form itself, and returns `BASE-QUOTE` in `Symbol` with the native name in `OriginalSymbol`.

## Configuration

Symbol caching is configured via:

```toml
[symbol]
cache_ttl = "5m"      # Cache duration
cache_dir = "/tmp/bits"  # Cache directory
```

Environment variables:
- `BITS_SYMBOL_CACHE_TTL`
- `BITS_SYMBOL_CACHE_DIR`

## Usage

The symbol resolver is automatically used by commands:

```bash
# These all resolve to the same symbol internally
bits ticker btc_usdt -p whitebit
bits ticker BTCUSDT -p binance  
bits ticker btc-usdt -p whitebit

# Output is always normalized
# SYMBOL
# BTC-USDT
```

## Symbol Fields in Output

All provider output structs carrying a symbol (`Ticker24h`, `OrderBook`, `CoinPrice`) now include both:

- **`Symbol string`** — normalized form in `BASE-QUOTE` (e.g., `BTC-USDT`), set by calling `model.NormalizeSymbol(nativeSymbol)`
- **`OriginalSymbol string`** — provider-native form exactly as received from the exchange API (e.g., `BTCUSDT`, `BTC_USDT`)

The `model.Symbol` type in `ExchangeInfo` carries:

- **`NormalizedSymbol string`** — always `BaseAsset + "-" + QuoteAsset` (e.g., `BTC-USDT`)
- **`Symbol string`** — provider-native format (used by the symbol engine for API calls)

## API

```go
// Normalize symbol for output display
func NormalizeSymbol(symbol string) string  // in model/normalize.go
```

## Disk Cache

Cache files are stored at:
```
/tmp/bits/symbols/
├── binance_spot.json
├── binance_futures.json
├── whitebit_spot.json
├── whitebit_futures.json
└── ...
```

Each cache file contains:
```json
{
  "symbols": [...],
  "cached_at": "2024-01-01T00:00:00Z",
  "expires_at": "2024-01-01T00:05:00Z"
}
```
