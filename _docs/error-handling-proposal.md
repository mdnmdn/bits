# Error Handling Proposal

## Overview

This document proposes a unified, ergonomic error handling strategy for `bits`. The current state uses `fmt.Errorf` strings throughout, which loses semantic information, makes programmatic handling impossible, and leaks provider internals inconsistently.

Two goals drive this proposal:
1. **Callers can act on errors** — distinguish auth failures from rate limits from network blips without string parsing.
2. **Provider specificity is preserved** — the raw provider code/message is always accessible, never silently dropped.

> **Status (2026-04-01):** Phase 1 is partially complete — `ErrorKind` and `ProviderError` exist in `model/errors.go` but `Unwrap`, `Is`, `WrapError`, and the sentinel migration are still pending. Phases 2–6 have not started.

---

## Part 1 — Error Handling Structures

### 1.1 Error Categories

`ErrorKind` is already defined in `model/errors.go`. No changes needed here.

```go
// model/errors.go  (already exists)

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

> **Note on 5xx and retryability:** All 5xx are marked `ErrKindServerError` and `Retryable() == true`. Callers should apply exponential backoff with a maximum attempt count; retry logic should give up eventually regardless.

> **Note on context errors:** `ErrKindCanceled` is intentionally excluded from `Retryable()`. A canceled context means the caller no longer wants the result; retrying would be wrong.

### 1.2 The `ProviderError` Type

`ProviderError` is already defined in `model/errors.go`. Two methods are **missing** and must be added:

```go
// model/errors.go  (ProviderError struct already exists — add these two methods)

func (e *ProviderError) Unwrap() error { return e.Cause }

// Is implements errors.Is matching by Kind for sentinel comparisons.
// Ensures errors.Is(err, model.ErrUnsupportedMarket) works for any *ProviderError
// with the matching Kind, not just the exact sentinel pointer.
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
        log.Printf("provider=%s code=%s http=%d", pe.ProviderID, pe.ProviderCode, pe.HTTPStatus)
    }
}
```

### 1.3 Sentinel Errors (backward-compat)

Currently the two sentinels are plain `errors.New(...)` values. They must be promoted to `*ProviderError` so that `errors.Is` matching against a freshly constructed error with the same `Kind` works:

```go
// model/errors.go  (replace the existing errors.New lines)

var (
    ErrUnsupportedMarket  = &ProviderError{Kind: ErrKindUnsupportedMarket,  ProviderMessage: "unsupported market type"}
    ErrUnsupportedFeature = &ProviderError{Kind: ErrKindUnsupportedFeature, ProviderMessage: "unsupported feature"}
)
```

Existing `errors.Is` call sites require no change after `Is()` is added (§1.2).

### 1.4 `ItemError` — partial failure in batch calls

`ItemError.Err` is currently typed `error` (an interface), which marshals to `{}` in JSON — effectively invisible. Changing to `*ProviderError` fixes serialisation as a side effect.

```go
// model/response.go  (change Err field type)

type ItemError struct {
    Symbol string         `json:"sym" yaml:"sym" toon:"sym"`
    Err    *ProviderError `json:"err" yaml:"err" toon:"err"` // was: error
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

Each provider package gets a small unexported `errors.go` with construction helpers. The `httpStatusToKind` mapping lives here — **not** in `model/`, which must remain HTTP-agnostic.

```go
// provider/binance/errors.go  (pattern repeated per provider)

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

// httpStatusToKind is local to each provider package (not in model/).
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

> **CoinGecko note:** CoinGecko's `client.go` already checks the HTTP status in its `get()` helper (unlike the other four providers that skip status checks). Its `errors.go` helpers still apply; the integration point is inside `get()` rather than a separate `doRequest` method.

> **MEXC note:** MEXC's `doRequest` also already checks the HTTP status code. Its migration is limited to wrapping the existing check with `httpErr(...)` instead of `fmt.Errorf(...)`.

### 1.6 Migration safety valve — `WrapError`

During the transition, some construction sites (e.g. registry failures in the facade, third-party library errors from `go-binance/v2`) may not yet produce a `*ProviderError`. A single escape-hatch function prevents compile errors and leaves an audit trail:

```go
// model/errors.go  (add this function)

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

> **File path note:** The project has no `pkg/` prefix. All paths below are relative to the module root (e.g. `model/errors.go`, `provider/binance/`, `resolve/`, `client.go`).

### Phase 1 — Foundation (`model/`)

**Files:** `model/errors.go`, `model/response.go`

**Already done:**
- `ErrorKind` constants and `Retryable()` method
- `ProviderError` struct and `Error()` method

**Still needed:**
1. Add `Unwrap()` and `Is()` methods to `ProviderError`.
2. Add `WrapError` function.
3. Re-declare the two sentinel vars as `*ProviderError` values (replacing the current `errors.New` lines).
4. **Do not** change `ItemError.Err` yet — leave typed as `error` to avoid a cascade of compile errors.

Result: remaining model infrastructure exists, nothing broken, no providers use it yet.

---

### Phase 2 — Provider HTTP clients (parallel, one PR per provider)

**Files:** `provider/*/client.go`, `provider/*/market.go`, `provider/*/exchange.go`

> Phases 2a–2f are fully independent and can land in any order.

For each provider:

1. Add `provider/<name>/errors.go` with `providerErr`, `httpErr`, `apiErr`, `httpStatusToKind`, and `apiCodeToKind`.
2. Update the HTTP execution path to check `resp.StatusCode` on every call and return `httpErr` on non-2xx.
   - **Bug fix:** Bitget, WhiteBit, and Crypto.com currently do not check the HTTP status code at all — a 500 response body is silently passed to `json.Unmarshal`. This change fixes that behavior; add regression tests.
   - CoinGecko and MEXC already check status; migrate their existing check to use `httpErr`.
3. Wrap `net.Error` and connection errors with `providerErr(ErrKindNetwork, …, err)`.
4. Wrap `context.Canceled` / `context.DeadlineExceeded` with `providerErr(ErrKindCanceled, …, err)`.
5. Replace API-envelope error checks (`code != "00000"`, `!resp.Success`, `code != 0`, etc.) with `apiErr(code, msg)`.
6. Use `WrapError` for any site not yet convertible (e.g. third-party library errors from `go-binance/v2`).

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

**Files:** `model/response.go`, all `ItemError` construction sites

Change `ItemError.Err` from `error` to `*model.ProviderError`. Update every construction site to pass a `*ProviderError`; use `WrapError` for any remaining raw errors. Key construction sites:

- `resolve/fanout.go` — wraps per-symbol errors into `ItemError`
- provider `market.go` files that append to `Response.Errors`
- `client.go` (root) — `ComparePrices` stores errors into `ItemError{Err: err}` where `err` comes from provider calls; must be `WrapError`-guarded until Phase 5

---

### Phase 4 — Resolver

**Files:** `resolve/resolver.go`

Replace `fmt.Errorf("provider %q does not support feature %q …")` and similar with `*model.ProviderError`. Use `ErrKindUnsupportedFeature` / `ErrKindUnsupportedMarket` with an empty `ProviderID` (the resolution layer has no single provider to blame).

`resolve/fanout.go` requires no structural change after Phase 3.

---

### Phase 5 — Public `client.go` facade

**Files:** `client.go` (root-level bits library)

1. No signature changes — methods still return `(model.Response[T], error)`.
2. The "no data returned" path wraps with `model.WrapError("", …)` or a direct `*ProviderError{Kind: ErrKindNotFound}`.
3. `ComparePrices` / `ComparePricesWithResolution`: replace ad-hoc `ItemError{Err: err}` with `WrapError(pid, err)` to satisfy the typed field.
4. Add godoc on all public methods stating: *returned errors are always `*model.ProviderError`; use `errors.As` to inspect them.*

---

### Phase 6 — CLI layer (optional)

**Files:** `command/*.go`

CLI commands can optionally inspect `*model.ProviderError.Kind` for friendlier messages (e.g. "authentication failed — check your API key in config") without breaking anything. Defer until after Phase 5 stabilises.

---

### Migration checklist

```
[~] Phase 1:  model/ — ErrorKind ✅, ProviderError struct ✅, Error() ✅
              PENDING: Unwrap(), Is(), WrapError(), sentinel *ProviderError promotion
[ ] Phase 2a: provider/binance   — errors.go + HTTP/API error wrapping + tests
[ ] Phase 2b: provider/bitget    — errors.go + HTTP/API error wrapping + tests (bug fix: add HTTP status check)
[ ] Phase 2c: provider/coingecko — errors.go + HTTP/API error wrapping + tests (status already checked in get())
[ ] Phase 2d: provider/cryptocom — errors.go + HTTP/API error wrapping + tests (bug fix: add HTTP status check)
[ ] Phase 2e: provider/mexc      — errors.go + HTTP/API error wrapping + tests (status already checked in doRequest())
[ ] Phase 2f: provider/whitebit  — errors.go + HTTP/API error wrapping + tests (bug fix: add HTTP status check)
[ ] Phase 3:  ItemError.Err → *ProviderError  (gate: all Phase 2 complete)
[ ] Phase 4:  resolve/ — ProviderError for resolution failures
[ ] Phase 5:  client.go facade — WrapError at remaining sites + godoc
[ ] Phase 6:  (optional) command/ human-readable error messages
```

Phases 2a–2f are fully parallel. All other phases are sequential.
