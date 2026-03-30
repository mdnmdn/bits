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
Symbol Resolver (in-memory)
  - Parse: btc → BTC, usdt → USDT
  - Load cached ExchangeInfo (or fetch fresh)
  - Find symbol with BaseAsset=BTC, QuoteAsset=USDT
       ↓
Provider Native: "BTC_USDT" (WhiteBit) or "BTCUSDT" (Binance)
       ↓
Output: "BTC-USDT" (normalized)
```

## Files

| File | Purpose |
|------|---------|
| `pkg/resolve/symbol/normalize.go` | Input parsing (`BTC-USDT` → `base=BTC, quote=USDT`) |
| `pkg/resolve/symbol/resolver.go` | Symbol resolution with in-memory cache |
| `pkg/resolve/symbol/lookup.go` | Efficient symbol lookup by base/quote |
| `pkg/resolve/symbol/cache.go` | Disk cache with TTL support |
| `pkg/resolve/symbol/engine.go` | Advanced engine with provider-specific translators |
| `pkg/resolve/symbol/translators/` | Provider-specific symbol translation |

## Provider Patterns

| Provider | Spot Symbol | Futures Symbol | Strategy |
|----------|-------------|---------------|----------|
| Binance | BTCUSDT | BTCUSDT | Same symbol, different API |
| WhiteBit | BTC_USDT | BTC_PERP | Different symbols per market |
| Crypto.com | BTC_USDT | BTCUSD-PERP | Different symbols per market |
| MEXC | BTCUSDT | BTC_USDT | Underscore differentiates futures |
| Bitget | BTCUSDT | BTCUSDT_PERP | Suffix differentiates futures |

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

## API

```go
// Resolve input to provider-specific symbol
func (r *SymbolResolver) Resolve(ctx context.Context, input string, market model.MarketType) (string, error)

// Normalize symbol for output display
func NormalizeSymbol(symbol string) string
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
