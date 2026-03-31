# Error Handling Proposal

## Overview

This document proposes a unified, ergonomic error handling strategy for `bits`. The current state uses `fmt.Errorf` strings throughout, which loses semantic information, makes programmatic handling impossible, and leaks provider internals inconsistently.

Two goals drive this proposal:
1. **Callers can act on errors** — distinguish auth failures from rate limits from network blips without string parsing.
2. **Provider specificity is preserved** — the raw provider code/message is always accessible, never silently dropped.

---

## Part 1 — Error Handling Structures

### 1.1 Error Categories

A single `ErrorKind` type covers all recoverable and unrecoverable conditions across every provider:

```go
// pkg/model/errors.go

type ErrorKind int

const (
    ErrKindUnknown          ErrorKind = iota // catch-all
    ErrKindAuth                              // 401 / 403 / invalid API key
    ErrKindRateLimit                         // 429 / quota exceeded
    ErrKindNotFound                          // 404 / symbol not found
    ErrKindInvalidRequest                    // 400 / bad parameters
    ErrKindServerError                       // 5xx / provider-side failure
    ErrKindNetwork                           // connection refused, timeout, DNS
    ErrKindParse                             // unexpected response shape
    ErrKindUnsupportedMarket                 // already exists, kept for compat
    ErrKindUnsupportedFeature                // already exists, kept for compat
)

func (k ErrorKind) Retryable() bool {
    return k == ErrKindRateLimit || k == ErrKindServerError || k == ErrKindNetwork
}
```

**Mapping rules** (applied by each provider's HTTP client):

| Condition | Kind |
|---|---|
| HTTP 401, 403 | `ErrKindAuth` |
| HTTP 429 | `ErrKindRateLimit` |
| HTTP 404 | `ErrKindNotFound` |
| HTTP 400 | `ErrKindInvalidRequest` |
| HTTP 5xx | `ErrKindServerError` |
| `net.Error` (timeout/refused) | `ErrKindNetwork` |
| JSON unmarshal failure | `ErrKindParse` |
| API-body code ≠ success | kind inferred from code; else `ErrKindUnknown` |

### 1.2 The `ProviderError` Type

One concrete type is used everywhere an error crosses a provider boundary:

```go
// pkg/model/errors.go

type ProviderError struct {
    // Normalized — callers use these without knowing which provider.
    Kind       ErrorKind
    ProviderID string // "binance", "bitget", …

    // Raw — always preserved, never dropped.
    ProviderCode    string // provider's native error code ("00000", "400", …)
    ProviderMessage string // provider's native message text

    // Optional HTTP layer information.
    HTTPStatus int // 0 when not an HTTP error

    // Underlying cause (network error, json error, …).
    Cause error
}

func (e *ProviderError) Error() string {
    if e.ProviderCode != "" {
        return fmt.Sprintf("[%s] %s (code %s)", e.ProviderID, e.ProviderMessage, e.ProviderCode)
    }
    if e.HTTPStatus != 0 {
        return fmt.Sprintf("[%s] HTTP %d: %s", e.ProviderID, e.HTTPStatus, e.ProviderMessage)
    }
    return fmt.Sprintf("[%s] %s", e.ProviderID, e.ProviderMessage)
}

func (e *ProviderError) Unwrap() error { return e.Cause }
```

**Usage at the call site:**

```go
res, err := client.GetPrice(ctx, "BTCUSDT", "binance")
if err != nil {
    var pe *model.ProviderError
    if errors.As(err, &pe) {
        switch pe.Kind {
        case model.ErrKindRateLimit:
            time.Sleep(backoff)
            // retry
        case model.ErrKindAuth:
            // surface config problem to user
        }
        // always available for logging / telemetry:
        log.Printf("provider=%s code=%s http=%d", pe.ProviderID, pe.ProviderCode, pe.HTTPStatus)
    }
}
```

**Provider-specific code access** is explicit and zero-surprise: `pe.ProviderCode` is the raw string the exchange returned (`"00000"`, `"10001"`, `"429"`, …), and `pe.ProviderMessage` is its description verbatim.

### 1.3 Sentinel Errors (keep backward-compat)

The two existing sentinels are promoted into `ProviderError` instances and keep their original `errors.Is` surface:

```go
var (
    ErrUnsupportedMarket  = &ProviderError{Kind: ErrKindUnsupportedMarket,  ProviderMessage: "unsupported market type"}
    ErrUnsupportedFeature = &ProviderError{Kind: ErrKindUnsupportedFeature, ProviderMessage: "unsupported feature"}
)
```

Existing `errors.Is(err, model.ErrUnsupportedMarket)` checks continue to work because `ProviderError` equality is by pointer identity for these package-level values.

### 1.4 `ItemError` — partial failure in batch calls

`ItemError` gains a typed `Err` so JSON serialisation is meaningful and callers can programmatically inspect per-symbol failures:

```go
// pkg/model/response.go

type ItemError struct {
    Symbol string         `json:"sym"`
    Err    *ProviderError `json:"err"`            // typed — always a *ProviderError
}
```

Changing from `error` to `*ProviderError` is the only breaking change in the public model. It allows:

```go
for _, ie := range res.Errors {
    if ie.Err.Kind == model.ErrKindNotFound {
        // skip missing symbols gracefully
    }
}
```

### 1.5 Provider constructor helpers

Each provider package exposes a small unexported helper so error construction is consistent and never duplicated:

```go
// pkg/provider/binance/errors.go  (one per provider)

func providerErr(kind model.ErrorKind, msg string, cause error) *model.ProviderError {
    return &model.ProviderError{
        Kind:            kind,
        ProviderID:      ProviderID, // "binance"
        ProviderMessage: msg,
        Cause:           cause,
    }
}

func httpErr(status int, body string) *model.ProviderError {
    kind := httpStatusToKind(status)
    return &model.ProviderError{
        Kind:            kind,
        ProviderID:      ProviderID,
        HTTPStatus:      status,
        ProviderMessage: body,
    }
}

func apiErr(code, msg string) *model.ProviderError {
    return &model.ProviderError{
        Kind:            apiCodeToKind(code), // provider-specific mapping
        ProviderID:      ProviderID,
        ProviderCode:    code,
        ProviderMessage: msg,
    }
}
```

A shared `httpStatusToKind(status int) ErrorKind` utility lives in `pkg/model/errors.go` to keep the HTTP→Kind mapping DRY across all providers.

### 1.6 What is NOT in scope

- Retry logic — out of scope; callers use `Retryable()` to decide.
- Circuit breaking — out of scope.
- Logging inside providers — providers return errors, they do not log them.
- Wrapping third-party library errors (e.g. `go-binance/v2`) beyond the single `Cause` field — the library's error becomes `Cause`; `ProviderMessage` is set from it.

---

## Part 2 — Transition

The transition is additive first, then breaking. It can be merged incrementally without breaking the CLI.

### Phase 1 — Foundation (`pkg/model`)

**Files:** `pkg/model/errors.go`, `pkg/model/response.go`

1. Add `ErrorKind`, `ProviderError`, and `httpStatusToKind` to `errors.go`.
2. Keep the two existing sentinel vars; re-declare them as `*ProviderError` values.
3. **Do not** change `ItemError` yet — leave it `error` to avoid a cascade of compile errors.

Result: the new types exist, nothing is broken, no providers use them yet.

---

### Phase 2 — Provider HTTP clients

**Files:** `pkg/provider/*/client.go` (all six providers)

For each provider:

1. Add a small `errors.go` file in the provider package with `providerErr`, `httpErr`, `apiErr` helpers.
2. Update `client.go` (or equivalent HTTP execution path) to:
   - Check `resp.StatusCode` after every HTTP call (currently missing in Bitget, WhiteBit, Crypto.com).
   - Return `httpErr(status, body)` on non-2xx instead of `fmt.Errorf`.
   - Wrap `net.Error` / `context` errors with `providerErr(ErrKindNetwork, …, err)`.
3. Update market/exchange method files (`market.go`, `exchange.go`) to:
   - Return `apiErr(code, msg)` when the response envelope signals failure (e.g. Bitget `code != "00000"`, Crypto.com `code != 0`).

No interface changes yet — all methods still return `(model.Response[T], error)`.

Provider-specific code→kind mappings go into each provider's `errors.go`. Examples:

| Provider | Success signal | Auth codes | Rate-limit codes |
|---|---|---|---|
| Bitget | `code == "00000"` | `40001`, `40003` | `429xx` series |
| Crypto.com | `code == 0` | `40001` | `10006` |
| Binance | HTTP 200 | HTTP 401/403 | HTTP 429 |
| CoinGecko | HTTP 200 | HTTP 401/403 | HTTP 429 |
| MEXC | HTTP 200 | HTTP 401/403 | HTTP 429 |
| WhiteBit | `success == true` | HTTP 401/403 | HTTP 429 |

---

### Phase 3 — `ItemError` migration

**Files:** `pkg/model/response.go`, all call sites that construct `ItemError`

Change `ItemError.Err` from `error` to `*model.ProviderError`.

Update every construction site (fanout, provider market files) to always pass a `*ProviderError`. Any site that currently creates a raw `fmt.Errorf` must be wrapped in the provider's `providerErr` helper first.

This is the only breaking change for library consumers. Callers that stored `ItemError.Err` as `error` continue to work via the `error` interface; callers that did type assertions need updating.

---

### Phase 4 — Resolver and FanOut

**Files:** `pkg/resolve/resolver.go`, `pkg/resolve/fanout.go`

- `resolver.go`: replace `fmt.Errorf("no provider supports …")` with `&model.ProviderError{Kind: ErrKindUnsupportedFeature, …}`. No `ProviderID` here since the error is from the resolution layer, not a specific provider.
- `fanout.go`: no structural change needed — `ItemError` already accepts the new type from Phase 3.

---

### Phase 5 — Public `pkg/bits` facade

**Files:** `pkg/bits/client.go`

The facade is what external consumers import, so this phase ensures the public surface is clean:

1. `GetPrice`, `ComparePrices`, etc. already return `(model.Response[T], error)` — no signature change.
2. Document in godoc that returned errors are always `*model.ProviderError` when they originate from a provider, so callers can use `errors.As`.
3. The "no data returned" fallback path also wraps with `*model.ProviderError{Kind: ErrKindNotFound, …}` instead of `fmt.Errorf`.

---

### Phase 6 — CLI layer (optional cleanup)

**Files:** `cmd/*.go`

The CLI commands currently print `err.Error()` on failure. After the transition, they can optionally surface the `Kind` for better UX (e.g. "authentication failed — check your API key in config") without breaking anything.

This phase is optional and can be deferred.

---

### Migration checklist

```
[ ] Phase 1: pkg/model — ErrorKind, ProviderError, httpStatusToKind
[ ] Phase 2a: pkg/provider/binance   — errors.go + client/market updates
[ ] Phase 2b: pkg/provider/bitget    — errors.go + client/market updates
[ ] Phase 2c: pkg/provider/coingecko — errors.go + client/market updates
[ ] Phase 2d: pkg/provider/cryptocom — errors.go + client/market updates
[ ] Phase 2e: pkg/provider/mexc      — errors.go + client/market updates
[ ] Phase 2f: pkg/provider/whitebit  — errors.go + client/market updates
[ ] Phase 3: ItemError.Err → *ProviderError
[ ] Phase 4: resolver + fanout
[ ] Phase 5: pkg/bits facade + godoc
[ ] Phase 6: (optional) CLI human-readable error messages
```

Each phase compiles and passes tests independently. Phases 2a–2f are fully parallel.
