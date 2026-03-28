# Symbol Internalization Plan

## Problem Statement

Different exchanges use different symbol formats:
- Binance: `BTCUSDT` (concatenated)
- WhiteBit: `BTC_UST` (underscore separator)
- Users may input: `BTC-USDT`, `btc_usdt`, `BTCUSDT`, `btc.usdt`, etc.

Currently, symbols pass directly from CLI to providers without normalization, causing errors when users input the wrong format.

## Solution Overview

Create a symbol resolution layer that:
1. Normalizes user input to extract base/quote assets
2. Loads provider's exchange info (symbols) with caching
3. Matches base/quote to find the provider's native symbol format
4. Returns the correctly formatted symbol for the specific provider

## Architecture

```
User Input: "btc_usdt"
       ↓
Symbol Resolver
  - Parse: btc → BTC, usdt → USDT
  - Load cached ExchangeInfo (or fetch fresh)
  - Find symbol with BaseAsset=BTC, QuoteAsset=USDT
       ↓
Provider Native: "BTC_USDT" (WhiteBit) or "BTCUSDT" (Binance)
```

## Implementation Phases

---

## Phase 1: Core Symbol Resolver (Week 1)

### Goals
- Create symbol resolver package with parsing and matching logic
- Implement in-memory caching of ExchangeInfo
- Support basic normalization (case-insensitive, separator handling)

### Tasks

#### 1.1 Create symbol resolver package
**Files:**
- `internal/resolve/symbol/resolver.go` - SymbolResolver struct and methods
- `internal/resolve/symbol/cache.go` - Cache implementation

**Implementation:**
- `type SymbolResolver struct { provider provider.Provider, cache *symbolCache }`
- `func NewSymbolResolver(p provider.Provider) *SymbolResolver`
- `func (r *SymbolResolver) Resolve(ctx context.Context, input string) (string, error)` - main entry point
- `func normalize(input string) (base, quote string)` - parse user input
- `func (r *SymbolResolver) resolveFromCache(ctx context.Context) ([]model.Symbol, error)` - get cached or fetch

#### 1.2 Add ExchangeProvider dependency
- Extend SymbolResolver to call `provider.ExchangeProvider.ExchangeInfo()` when needed
- Store result in memory map keyed by provider+market

#### 1.3 Implement symbol matching
- `func findSymbol(symbols []model.Symbol, base, quote string) *model.Symbol` - linear search for match

#### 1.4 Wire into commands
- Modify `cmd/ticker.go`, `cmd/book.go`, `cmd/candles.go` to use symbol resolver
- Create helper in `cmd/factory.go`: `resolveSymbol(ctx, provider, input string) (string, error)`

### Completion Criteria
- [ ] `go build ./...` succeeds
- [ ] `go fmt ./...` clean
- [ ] `golangci-lint run` passes
- [ ] Manual test: `bits ticker btc_usdt -p whitebit` resolves to BTC_USDT
- [ ] Manual test: `bits ticker BTCUSDT -p binance` works as before

### Phase 1 Updates
*(Completed)*

**Files created:**
- `internal/resolve/symbol/normalize.go` - `Normalize(input) (base, quote, err)` parses user input (supports separators: `-`, `_`, `/`, `.`, concatenated)
- `internal/resolve/symbol/resolver.go` - `SymbolResolver` struct with `Resolve(ctx, input, market)`, loads ExchangeInfo, caches in memory

**Files modified:**
- `cmd/factory.go` - added `newSymbolResolver(p provider.Provider)` helper
- `cmd/ticker.go` - uses symbol resolver in FanOut
- `cmd/book.go` - uses symbol resolver before OrderBook call
- `cmd/candles.go` - uses symbol resolver before Candles call

**Tested:**
- `bits ticker BTCUSDT -p whitebit` → resolves to BTC_USDT
- `bits ticker BTC-USDT -p whitebit` → resolves to BTC_USDT
- `bits ticker btcusdt -p wb` → resolves to BTC_USDT (case-insensitive)
- `bits book BTCUSDT -p whitebit` → resolves to BTC_USDT
- `bits candles BTCUSDT -p whitebit` → resolves to BTC_USDT

---

## Phase 2: File-based Cache with TTL (Week 2)

### Goals
- Add persistent caching to `/tmp/bits/`
- Implement TTL-based invalidation
- Reduce API calls for ExchangeInfo

### Tasks

#### 2.1 Create cache file format
**File:** `internal/resolve/symbol/cache.go`

- Cache location: configurable via `config.Symbol.CacheDir` (default `/tmp/bits/<provider>-<market>.json`)
- TTL: configurable via `config.Symbol.CacheTTL` (default 5 minutes)
- Format: JSON with symbols array + timestamp

```go
type cacheEntry struct {
    Market     model.MarketType `json:"market"`
    Symbols    []model.Symbol   `json:"symbols"`
    CachedAt   time.Time        `json:"cached_at"`
    ExpiresAt time.Time        `json:"expires_at"`
}
```

#### 2.2 Implement cache read/write
- `func (c *symbolCache) get(ctx context.Context, providerID string, market model.MarketType) ([]model.Symbol, bool)` - returns symbols and `true` if valid cache
- `func (c *symbolCache) set(providerID string, market model.MarketType, symbols []model.Symbol)` - writes to file
- Create `/tmp/bits/` directory if not exists

#### 2.3 Add cache invalidation
- Check file mtime or stored ExpiresAt
- Force refresh: `resolveSymbol(..., forceRefresh bool)`

#### 2.4 Error handling
- If cache read fails: log warning, fallback to API
- If API fails and cache exists but expired: use stale cache with warning

### Completion Criteria
- [ ] `go build ./...` succeeds
- [ ] Cache file created at `/tmp/bits/whitebit-spot.json`
- [ ] Subsequent calls use cache (verify with timing)
- [ ] Cache expires after TTL
- [ ] `golangci-lint run` passes

### Phase 2 Updates
*(Completed: N/A - not yet started)*

---

## Phase 3: Provider Interface & Implementation (Week 3)

### Goals
- Formalize symbol resolution as a provider interface
- Support providers that need custom normalization
- Handle edge cases (e.g., stablecoins, different quote assets)

### Tasks

#### 3.1 Create SymbolResolverProvider interface
**File:** `internal/provider/symbol.go` (new)

```go
type SymbolResolverProvider interface {
    Provider
    // ResolveSymbol converts a user-input symbol to provider's native format
    ResolveSymbol(ctx context.Context, input string, market model.MarketType) (string, error)
    // GetSymbolInfo returns native symbol for base+quote pair
    GetSymbolInfo(ctx context.Context, base, quote string, market model.MarketType) (*model.Symbol, error)
}
```

#### 3.2 Implement for existing providers
**Files to modify:**
- `internal/provider/binance/market.go` - add ResolveSymbol method
- `internal/provider/whitebit/market.go` - add ResolveSymbol method  
- `internal/provider/bitget/market.go` - add ResolveSymbol method
- `internal/provider/coingecko/market.go` - add ResolveSymbol method (no-op or special handling)

Each provider implements:
- Parse input to base/quote using common normalizer
- Use cached/fetched ExchangeInfo to find exact match
- Return native symbol format

#### 3.3 Update capability matrix
**File:** `internal/capability/capability.go`

- Add `FeatureSymbolResolution` to Feature enum
- Providers declare symbol resolution capability

### Completion Criteria
- [ ] All providers implement SymbolResolverProvider
- [ ] `bits capabilities` shows SymbolResolution feature
- [ ] Cross-provider test: `bits ticker BTC-USDT -p whitebit` → works
- [ ] Error handling: unknown symbol returns clear error

### Phase 3 Updates
*(Completed: N/A - not yet started)*

---

## Phase 4: Advanced Normalization & Fallback (Week 4)

### Goals
- Handle more input formats (BTC-USDT, btc.usdt, etc.)
- Smart matching (fuzzy match for typos)
- Quote asset alternatives (USDT → USDC → USDT)

### Tasks

#### 4.1 Enhanced normalization
**File:** `internal/resolve/symbol/normalize.go`

- Support separators: `-`, `_`, `/`, `.`, `` (concatenated)
- Case insensitive: `btc` → `BTC`
- Common quote assets mapping: `USDT`, `USDC`, `USD`, `EUR`, `BTC`, `ETH`

#### 4.2 Fuzzy matching
- If exact match fails: try case-insensitive match
- If still fails: try prefix/suffix matching
- Log suggestions if close match found

#### 4.3 Quote asset fallback
- If BTC-USDT not found: try BTC-USDC, then BTC-USD
- Configurable fallback order per provider

#### 4.4 Self-healing cache
- Background refresh: update cache 1 minute before expiry
- Prefetch popular symbols on startup

### Completion Criteria
- [ ] All these inputs resolve correctly for WhiteBit:
  - `btc_usdt` → BTC_USDT
  - `BTC-USDT` → BTC_USDT  
  - `btc.usdt` → BTC_USDT
  - `BTCUSDT` → BTC_USDT (fuzzy match)
- [ ] Error shows suggestions for typo: `BTC_ETH` → "Did you mean BTC-USDT?"
- [ ] Performance: <100ms for symbol resolution

### Phase 4 Updates
*(Completed: N/A - not yet started)*

---

## Phase 5: Integration & Polish (Week 5)

### Goals
- Full integration with all symbol-using commands
- Configuration options
- Documentation

### Tasks

#### 5.1 Update all commands
**Files:**
- `cmd/ticker.go` - use symbol resolver
- `cmd/book.go` - use symbol resolver
- `cmd/candles.go` - use symbol resolver
- `cmd/price.go` - use symbol resolver (for exchange prices)
- `cmd/info.go` - use symbol resolver for --symbol flag

#### 5.2 Add configuration
**File:** `internal/config/config.go` (COMPLETED)

Added `SymbolConfig` struct and integrated into `Config`:
- `config.Symbol.CacheTTL` - TTL for symbol cache (default: 5m)
- `config.Symbol.CacheDir` - Cache directory (default: /tmp/bits)
- Config key: `[symbol]` in config file
- Environment: `BITS_SYMBOL_CACHE_TTL`, `BITS_SYMBOL_CACHE_DIR`
- .env: `symbol.cache_ttl`, `symbol.cache_dir`
- Methods: `GetCacheTTL()`, `GetCacheDir()` provide defaults if not set

Updated files:
- `internal/config/config.go` - added SymbolConfig struct, integrated into Config, added to ConfigTemplate, added defaults in Load(), added env var overrides in applyEnvOverrides() and applyEnvMap(), added to Redacted()

#### 5.3 Commands for cache management
- `bits cache clear` - clear symbol cache
- `bits cache status` - show cache stats

#### 5.4 Documentation
- Updated `_docs/config.md` with Symbol config section and environment variables

### Completion Criteria
- [ ] All commands use symbol resolution
- [ ] Configuration works via config file and env vars
- [ ] Cache management commands functional
- [ ] Documentation complete
- [ ] Full test suite passes

### Phase 5 Updates
*(Completed: N/A - not yet started)*

---

## File Structure Summary

```
internal/
├── resolve/
│   └── symbol/
│       ├── resolver.go      # SymbolResolver, Resolve()
│       ├── cache.go         # file-based cache with TTL
│       └── normalize.go     # input parsing, base/quote extraction
└── provider/
    └── symbol.go            # SymbolResolverProvider interface (Phase 3)

cmd/
├── ticker.go                # use symbol resolver (Phase 1)
├── book.go                  # use symbol resolver (Phase 1)
├── candles.go               # use symbol resolver (Phase 1)
├── price.go                 # use symbol resolver (Phase 5)
├── info.go                  # use symbol resolver (Phase 5)
└── cache.go                 # new cache management command (Phase 5)
```

## Key Interfaces

```go
// Phase 3: Provider interface
type SymbolResolverProvider interface {
    Provider
    ResolveSymbol(ctx context.Context, input string, market model.MarketType) (string, error)
}

// Phase 1-2: Internal resolver (used by commands)
type SymbolResolver interface {
    Resolve(ctx context.Context, input string, market model.MarketType) (string, error)
    Refresh(ctx context.Context, market model.MarketType) error
}
```

## Testing Strategy

1. **Unit tests**: normalize.go - test various input formats
2. **Integration tests**: resolver - mock ExchangeInfo, test matching
3. **E2E tests**: CLI commands with real providers
4. **Performance tests**: benchmark symbol resolution with large symbol lists
