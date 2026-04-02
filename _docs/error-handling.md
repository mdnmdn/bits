# Error Handling

## Overview

All errors that cross a provider boundary are `*model.ProviderError`. This normalises heterogeneous provider errors (HTTP status codes, API envelope codes, network failures) into a single inspectable value without losing raw provider detail.

Two invariants hold throughout the codebase:
1. **Callers can act on errors** — distinguish auth failures from rate limits from network blips without string parsing.
2. **Provider specificity is preserved** — the raw provider code/message is always accessible, never silently dropped.

---

## Error Structures (`model/`)

### `ErrorKind` — normalised category

```go
// model/errors.go

type ErrorKind int

const (
    ErrKindUnknown          ErrorKind = iota // catch-all
    ErrKindAuth                              // 401 / 403 / invalid API key
    ErrKindRateLimit                         // 429 / quota exceeded
    ErrKindNotFound                          // 404 / symbol not found
    ErrKindInvalidRequest                    // 400 / bad parameters
    ErrKindServerError                       // 5xx / provider-side transient failure
    ErrKindNetwork                           // connection refused, DNS failure
    ErrKindCanceled                          // context.Canceled or context.DeadlineExceeded
    ErrKindParse                             // unexpected response shape
    ErrKindUnsupportedMarket                 // market type not supported by this provider
    ErrKindUnsupportedFeature                // feature not supported by this provider
)

func (k ErrorKind) Retryable() bool {
    return k == ErrKindRateLimit || k == ErrKindServerError || k == ErrKindNetwork
}
```

`Retryable()` returns `true` for transient failures where exponential backoff makes sense.
`ErrKindCanceled` is intentionally excluded — a cancelled context means the caller no longer
wants the result; retrying would be wrong.

**HTTP status → kind mapping** (applied by every provider's HTTP client):

| Condition | Kind |
|---|---|
| HTTP 401, 403 | `ErrKindAuth` |
| HTTP 429 | `ErrKindRateLimit` |
| HTTP 404 | `ErrKindNotFound` |
| HTTP 400 | `ErrKindInvalidRequest` |
| HTTP 5xx | `ErrKindServerError` |
| `net.Error` (refused, DNS) | `ErrKindNetwork` |
| `context.Canceled` / `context.DeadlineExceeded` | `ErrKindCanceled` |
| JSON unmarshal failure | `ErrKindParse` |
| API-body code ≠ success | kind inferred from code; else `ErrKindUnknown` |

### `ProviderError` — the error type

```go
// model/errors.go

type ProviderError struct {
    Kind            ErrorKind // normalised category
    ProviderID      string    // "binance", "bitget", … empty for resolution-layer errors
    ProviderCode    string    // provider's native error code ("00000", "40001", …)
    ProviderMessage string    // provider's native message text
    HTTPStatus      int       // 0 when not an HTTP error
    Cause           error     // underlying cause (net error, json error, …)
}

func (e *ProviderError) Error() string  // formats ProviderID + ProviderMessage + ProviderCode
func (e *ProviderError) Unwrap() error  // returns Cause; enables errors.As chaining
func (e *ProviderError) Is(target error) bool  // matches sentinels by Kind
```

`Is()` enables `errors.Is(err, model.ErrUnsupportedMarket)` to match any `*ProviderError`
with `Kind == ErrKindUnsupportedMarket`, not just the exact sentinel pointer.

### Sentinel errors

```go
var (
    ErrUnsupportedMarket  = &ProviderError{Kind: ErrKindUnsupportedMarket,  ProviderMessage: "unsupported market type"}
    ErrUnsupportedFeature = &ProviderError{Kind: ErrKindUnsupportedFeature, ProviderMessage: "unsupported feature"}
)
```

### `ItemError` — partial failure in batch calls

```go
// model/response.go

type ItemError struct {
    Symbol string         `json:"sym" yaml:"sym" toon:"sym"`
    Err    *ProviderError `json:"err" yaml:"err" toon:"err"`
}
```

Batch methods (e.g. `Price` with multiple symbols) never return a top-level error for
per-symbol failures. Instead they collect them in `Response.Errors` and return whatever
succeeded in `Response.Data`.

### `WrapError` — adapter for plain errors

```go
// model/errors.go

func WrapError(providerID string, err error) *ProviderError
```

Wraps any `error` as `*ProviderError{Kind: ErrKindUnknown}`. If the error is already a
`*ProviderError` it is returned unchanged. Use this when a typed error is required but only
a plain `error` is available — for example, wrapping third-party library errors from
`go-binance/v2` where the library manages its own HTTP internals.

---

## Provider Error Helpers (`provider/*/errors.go`)

Every provider package contains an unexported `errors.go` with construction helpers.
`httpStatusToKind` and `apiCodeToKind` live here — not in `model/`, which must remain
HTTP-agnostic.

```go
// provider/<name>/errors.go

func providerErr(kind model.ErrorKind, msg string, cause error) *model.ProviderError
func httpErr(status int, body string) *model.ProviderError   // uses httpStatusToKind
func apiErr(code, msg string) *model.ProviderError           // uses apiCodeToKind; signature varies per provider

func httpStatusToKind(status int) model.ErrorKind
func apiCodeToKind(code string) model.ErrorKind              // provider-specific mapping
```

`apiErr` / `apiCodeToKind` are only defined for providers that return structured error codes
in the response body (Bitget, Crypto.com, WhiteBit). HTTP-only providers (Binance, CoinGecko,
MEXC) use `httpErr` and `WrapError` only.

Provider-specific code → kind mappings:

| Provider | Success signal | Error codes |
|---|---|---|
| Bitget | `code == "00000"` | `40001`, `40003` → Auth; `429xx` → RateLimit |
| Crypto.com | `code == 0` | `40001` → Auth; `10006` → RateLimit |
| WhiteBit | `success == true` | HTTP status codes only |
| Binance | HTTP 200 | HTTP status codes; library errors via `WrapError` |
| CoinGecko | HTTP 200 | HTTP status codes only |
| MEXC | HTTP 200 | HTTP status codes only |

---

## How Errors Flow

### Total failure (method-level)

When the entire call fails, the method returns a non-nil `error`:

```go
body, err := c.doRequest(ctx, path, query)
if err != nil {
    return model.Response[model.Ticker24h]{}, err  // already *ProviderError from doRequest
}
var envelope struct { Code string; Msg string; Data myTicker }
if err := json.Unmarshal(body, &envelope); err != nil {
    return model.Response[model.Ticker24h]{}, providerErr(model.ErrKindParse, err.Error(), err)
}
if envelope.Code != "0" {
    return model.Response[model.Ticker24h]{}, apiErr(envelope.Code, envelope.Msg)
}
```

### Partial failure (per-symbol)

When one symbol in a batch fails, the error goes into `Response.Errors`:

```go
for _, id := range ids {
    data, err := c.fetchTicker(id)
    if err != nil {
        errs = append(errs, model.ItemError{Symbol: id, Err: model.WrapError(providerID, err)})
        continue
    }
    prices = append(prices, convertPrice(data))
}
return model.Response[[]model.CoinPrice]{..., Errors: errs}, nil
```

### Resolution-layer errors

`resolve/resolver.go` emits `*ProviderError` values with an empty `ProviderID` when no
provider can satisfy the request:

```go
return nil, "", false, &model.ProviderError{
    Kind:            model.ErrKindUnsupportedFeature,
    ProviderMessage: "no provider supports feature ...",
}
```

---

## Caller Usage

```go
res, err := client.Price(ctx, []string{"BTCUSDT", "ETHUSDT"}, "")
if err != nil {
    var pe *model.ProviderError
    if errors.As(err, &pe) {
        switch pe.Kind {
        case model.ErrKindRateLimit:
            // back off and retry
        case model.ErrKindAuth:
            // surface config problem to user
        case model.ErrKindCanceled:
            return // do not retry
        }
        log.Printf("provider=%s code=%s http=%d msg=%s",
            pe.ProviderID, pe.ProviderCode, pe.HTTPStatus, pe.ProviderMessage)
    }
    return
}

// Partial failures in batch calls:
for _, ie := range res.Errors {
    if ie.Err.Kind == model.ErrKindNotFound {
        // skip missing symbols gracefully
    }
}
```

Sentinel matching:

```go
if errors.Is(err, model.ErrUnsupportedFeature) {
    // provider does not support this capability
}
```

---

## What is NOT in scope

- **Retry logic** — callers use `Retryable()` to decide; `bits` does not retry internally.
- **Circuit breaking** — out of scope.
- **Logging inside providers** — providers return errors, they do not log them.
- **Deep wrapping of third-party errors** — e.g. `go-binance/v2` library errors become `Cause`; `ProviderMessage` is set from `.Error()`. The library's internal HTTP handling is not re-implemented.
