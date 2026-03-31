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
    ErrKindServerError                       // 502/503/504 / provider-side transient failure
    ErrKindNetwork                           // connection refused, DNS failure
    ErrKindCanceled                          // context.Canceled or context.DeadlineExceeded
    ErrKindParse                             // unexpected response shape
    ErrKindUnsupportedMarket                 // already exists, kept for compat
    ErrKindUnsupportedFeature                // already exists, kept for compat
)

// Retryable reports whether errors of this kind are worth retrying with backoff.
// Callers are responsible for applying exponential backoff; this method only
// signals intent.
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
| HTTP 502, 503, 504 | `ErrKindServerError` |
| HTTP 500 (and other 5xx) | `ErrKindServerError` |
| `net.Error` (refused, DNS) | `ErrKindNetwork` |
| `context.Canceled` / `context.DeadlineExceeded` | `ErrKindCanceled` |
| JSON unmarshal failure | `ErrKindParse` |
| API-body code ≠ success | kind inferred from code; else `ErrKindUnknown` |

> **Note on 5xx and retryability:** All 5xx are marked `ErrKindServerError` and `Retryable() == true`. In practice, callers should apply exponential backoff and a maximum attempt count. A persistent 500 that never clears is a provider bug; retry logic should give up eventually regardless of this flag.

> **Note on context errors:** `ErrKindCanceled` is intentionally excluded from `Retryable()`. A canceled context means the caller no longer wants the result; retrying would be wrong.

### 1.2 The `ProviderError` Type

One concrete type is used everywhere an error crosses a provider boundary:

```go
// pkg/model/errors.go

type ProviderError struct {
    // Normalized — callers use these without knowing which provider.
    Kind       ErrorKind
    ProviderID string // "binance", "bitget", … empty for resolution-layer errors

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
    if e.ProviderID != "" {
        return fmt.Sprintf("[%s] %s", e.ProviderID, e.ProviderMessage)
    }
    return e.ProviderMessage
}

func (e *ProviderError) Unwrap() error { return e.Cause }

// Is implements errors.Is matching by Kind for sentinel comparisons.
// This ensures errors.Is(err, model.ErrUnsupportedMarket) keeps working even
// when the error is a freshly constructed *ProviderError rather than the exact
// sentinel pointer.
func (e *ProviderError) Is(target error) bool {
    t, ok := target.(*ProviderError)
    if !ok {
        return false
    }
    // Match sentinels by Kind when the target has no ProviderID (i.e. is a sentinel).
    if t.ProviderID == "" {
        return e.Kind == t.Kind
    }
    return e == t
}
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
        case model.ErrKindCanceled:
            return // caller canceled, do not retry
        }
        // always available for logging / telemetry:
        log.Printf("provider=%s code=%s http=%d", pe.ProviderID, pe.ProviderCode, pe.HTTPStatus)
    }
}
```

**Provider-specific code access** is explicit and zero-surprise: `pe.ProviderCode` is the raw string the exchange returned (`"00000"`, `"10001"`, `"429"`, …), and `pe.ProviderMessage` is its description verbatim.

### 1.3 Sentinel Errors (backward-compat)

The two existing sentinels are promoted into `*ProviderError` values. The `Is` method on §1.2 ensures `errors.Is(err, model.ErrUnsupportedMarket)` keeps working for any `*ProviderError` with `Kind == ErrKindUnsupportedMarket`, not just the exact pointer — so wrapped errors and freshly constructed errors both match.

```go
var (
    ErrUnsupportedMarket  = &ProviderError{Kind: ErrKindUnsupportedMarket,  ProviderMessage: "unsupported market type"}
    ErrUnsupportedFeature = &ProviderError{Kind: ErrKindUnsupportedFeature, ProviderMessage: "unsupported feature"}
)
```

Existing `errors.Is` call sites require no change. `errors.As` call sites that then inspect `.Kind` also work without change.

### 1.4 `ItemError` — partial failure in batch calls

`ItemError` gains a typed `Err` so JSON serialisation is meaningful and callers can programmatically inspect per-symbol failures.

> **Pre-existing issue fixed by this change:** `ItemError.Err` is currently typed `error` (an interface), which marshals to `{}` in JSON — effectively invisible. Changing to `*ProviderError` fixes serialisation as a side effect.

```go
// pkg/model/response.go

type ItemError struct {
    Symbol string         `json:"sym"`
    Err    *ProviderError `json:"err"` // typed — always a *ProviderError
}
```

This is the only breaking change in the public model. It allows:

```go
for _, ie := range res.Errors {
    if ie.Err.Kind == model.ErrKindNotFound {
        // skip missing symbols gracefully
    }
}
```

### 1.5 Provider constructor helpers

Each provider package has a small unexported `errors.go` with construction helpers. The `httpStatusToKind` mapping lives here too — **not** in `pkg/model`, which must remain HTTP-agnostic.

```go
// pkg/provider/binance/errors.go  (pattern repeated per provider)

func providerErr(kind model.ErrorKind, msg string, cause error) *model.ProviderError {
    return &model.ProviderError{
        Kind:            kind,
        ProviderID:      ProviderID, // "binance"
        ProviderMessage: msg,
        Cause:           cause,
    }
}

func httpErr(status int, body string) *model.ProviderError {
    return &model.ProviderError{
        Kind:            httpStatusToKind(status),
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

// httpStatusToKind is local to each provider package (not in pkg/model).
func httpStatusToKind(status int) model.ErrorKind {
    switch {
    case status == 401 || status == 403:
        return model.ErrKindAuth
    case status == 404:
        return model.ErrKindNotFound
    case status == 429:
        return model.ErrKindRateLimit
    case status == 400:
        return model.ErrKindInvalidRequest
    case status >= 500:
        return model.ErrKindServerError
    default:
        return model.ErrKindUnknown
    }
}
```

> **CoinGecko note:** CoinGecko's `client.go` uses a `get(ctx, path, &result)` helper that decodes JSON internally and returns a plain `error`. It does not follow the two-step fetch-then-unmarshal pattern of the other five providers. Its `errors.go` helpers still apply, but the integration point is inside `get()` rather than a separate `doRequest` method. No structural change to the client is needed — only the error return sites change.

### 1.6 Migration safety valve — `WrapError`

During the transition, some construction sites (e.g. registry failures in the facade, third-party library errors from `go-binance/v2`) may not yet produce a `*ProviderError`. A single escape-hatch function prevents compile errors and leaves an audit trail:

```go
// pkg/model/errors.go

// WrapError wraps an arbitrary error as ErrKindUnknown. Use only as a
// temporary shim during migration; replace with a typed providerErr call.
func WrapError(providerID string, err error) *ProviderError {
    if err == nil {
        return nil
    }
    // If already a ProviderError, preserve it.
    var pe *ProviderError
    if errors.As(err, &pe) {
        return pe
    }
    return &ProviderError{
        Kind:            ErrKindUnknown,
        ProviderID:      providerID,
        ProviderMessage: err.Error(),
        Cause:           err,
    }
}
```

### 1.7 What is NOT in scope

- Retry logic — callers use `Retryable()` to decide; `bits` does not retry internally.
- Circuit breaking — out of scope.
- Logging inside providers — providers return errors, they do not log them.
- Wrapping third-party library errors (e.g. `go-binance/v2`) beyond the `Cause` field — the library error becomes `Cause`; `ProviderMessage` is set from its `.Error()` string.

---

## Part 2 — Transition

The transition is additive first, then breaking. Each phase compiles and passes tests independently.

### Phase 1 — Foundation (`pkg/model`)

**Files:** `pkg/model/errors.go`, `pkg/model/response.go`

1. Add `ErrorKind`, `ProviderError` (with `Is` and `Unwrap`), and `WrapError` to `errors.go`.
2. Re-declare the two sentinel vars as `*ProviderError` values.
3. **Do not** change `ItemError` yet — leave `Err` typed as `error` to avoid a cascade of compile errors.

Result: new types exist, nothing broken, no providers use them yet.

---

### Phase 2 — Provider HTTP clients (parallel, one PR per provider)

**Files:** `pkg/provider/*/client.go`, `pkg/provider/*/market.go`, `pkg/provider/*/exchange.go`

> Phases 2a–2f are fully independent and can land in any order.

For each provider:

1. Add `pkg/provider/<name>/errors.go` with `providerErr`, `httpErr`, `apiErr`, `httpStatusToKind`, and `apiCodeToKind`.
2. Update the HTTP execution path to check `resp.StatusCode` on every call and return `httpErr` on non-2xx.
   - **Bug fix:** Bitget, WhiteBit, and Crypto.com currently do not check the HTTP status code at all — a 500 response body is silently passed to `json.Unmarshal`. This change fixes that behavior; add regression tests.
3. Wrap `net.Error` and connection errors with `providerErr(ErrKindNetwork, …, err)`.
4. Wrap `context.Canceled` / `context.DeadlineExceeded` with `providerErr(ErrKindCanceled, …, err)`.
5. Replace API-envelope error checks (`code != "00000"`, `!resp.Success`, etc.) with `apiErr(code, msg)`.
6. Use `WrapError` for any site not yet convertible (e.g. third-party library errors).

Provider-specific code→kind mappings:

| Provider | Success signal | Auth codes | Rate-limit codes |
|---|---|---|---|
| Bitget | `code == "00000"` | `40001`, `40003` | `429xx` series |
| Crypto.com | `code == 0` | `40001` | `10006` |
| Binance | HTTP 200 | HTTP 401/403 | HTTP 429 |
| CoinGecko | HTTP 200 | HTTP 401/403 | HTTP 429 |
| MEXC | HTTP 200 | HTTP 401/403 | HTTP 429 |
| WhiteBit | `success == true` | HTTP 401/403 | HTTP 429 |

Each provider's `errors.go` should include table-driven tests mapping HTTP statuses and API codes to expected `ErrorKind` values.

---

### Phase 3 — `ItemError` migration

> **Gate:** Phase 2 must be complete for all six providers before this phase lands. If partial, use `WrapError` to bridge un-migrated paths before opening this PR.

**Files:** `pkg/model/response.go`, all `ItemError` construction sites

Change `ItemError.Err` from `error` to `*model.ProviderError`. Update every construction site to pass a `*ProviderError`; use `WrapError` for any remaining raw errors. This includes:

- `pkg/resolve/fanout.go` — wraps per-symbol errors into `ItemError`
- all provider `market.go` files that append to `Response.Errors`
- `pkg/bits/client.go` — `ComparePrices` stores errors into `ItemError{Err: err}` where `err` comes from `GetPrice`; this must be `WrapError`-guarded until Phase 5

---

### Phase 4 — Resolver

**Files:** `pkg/resolve/resolver.go`

Replace `fmt.Errorf("no provider supports …")` and similar with `*model.ProviderError`. Use `ErrKindUnsupportedFeature` / `ErrKindUnsupportedMarket` with an empty `ProviderID` (the resolution layer has no single provider to blame).

`fanout.go` requires no structural change after Phase 3.

---

### Phase 5 — Public `pkg/bits` facade

**Files:** `pkg/bits/client.go`

1. No signature changes — methods still return `(model.Response[T], error)`.
2. The "no data returned" path wraps with `model.WrapError("", …)` or a direct `*ProviderError{Kind: ErrKindNotFound}`.
3. `ComparePrices` / `ComparePricesWithResolution`: replace ad-hoc `ItemError{Err: err}` with `WrapError(pid, err)` to satisfy the typed field.
4. Add godoc on all public methods stating: *returned errors are always `*model.ProviderError`; use `errors.As` to inspect them.*

---

### Phase 6 — CLI layer (optional)

**Files:** `cmd/*.go`

CLI commands can optionally inspect `*model.ProviderError.Kind` for friendlier messages (e.g. "authentication failed — check your API key in config") without breaking anything. Defer until after Phase 5 stabilises.

---

### Migration checklist

```
[ ] Phase 1:  pkg/model — ErrorKind, ProviderError (with Is/Unwrap), WrapError, sentinels
[ ] Phase 2a: pkg/provider/binance   — errors.go + HTTP/API error wrapping + tests
[ ] Phase 2b: pkg/provider/bitget    — errors.go + HTTP/API error wrapping + tests (bug fix: add HTTP status check)
[ ] Phase 2c: pkg/provider/coingecko — errors.go + HTTP/API error wrapping + tests (note: get() pattern differs)
[ ] Phase 2d: pkg/provider/cryptocom — errors.go + HTTP/API error wrapping + tests (bug fix: add HTTP status check)
[ ] Phase 2e: pkg/provider/mexc      — errors.go + HTTP/API error wrapping + tests
[ ] Phase 2f: pkg/provider/whitebit  — errors.go + HTTP/API error wrapping + tests (bug fix: add HTTP status check)
[ ] Phase 3:  ItemError.Err → *ProviderError  (gate: all Phase 2 complete)
[ ] Phase 4:  resolver — ProviderError for resolution failures
[ ] Phase 5:  pkg/bits facade — WrapError at remaining sites + godoc
[ ] Phase 6:  (optional) CLI human-readable error messages
```

Phases 2a–2f are fully parallel. All other phases are sequential.
