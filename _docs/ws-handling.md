# WebSocket Infrastructure and Patterns

## Overview

This document specifies the architecture for WebSocket-based data streaming across all providers in `bits`. The design prioritises **simplicity at the call site, ergonomics for provider authors, and robustness at runtime** — reconnection, subscription restore, multi-connection pooling, and per-protocol specialisation are all handled by the infrastructure so individual providers stay thin.

Related documents:
- `_docs/error-handling-proposal.md` — typed `*model.ProviderError` used by this package
- `_docs/symbol-engine.md` — symbol normalisation; **Watch* methods receive provider-native symbols**, already resolved by the symbol engine before the call

---

## Provider Requirements Matrix

| Provider | Auth Required | Ping Format | Ping Interval | Max Subs/Conn | Compression | CRC32 | Special |
|---|---|---|---|---|---|---|---|
| Binance | No | WebSocket native (server) | Server-driven (20 s) | 1024 | None | No | URL-based stream list |
| Bitget | No (public) | `"ping"` string | 30 s | 1000 (rec. <50) | None | Yes (books) | instType per sub |
| MEXC | No | `{"method":"PING"}` | 60 s | **30** | **Protobuf** | No | Binary frames |
| Crypto.com | No | `public/respond-heartbeat` | 30 s (server-initiated) | 400 | None | No | JSON-RPC 2.0 |
| WhiteBit | No | `{"method":"ping"}` JSON-RPC | 50 s | Unlimited | None | No | JSON-RPC 2.0 |
| CoinGecko | Yes (API key) | WebSocket native (server) | 10 s (server) | 100 | ActionCable | No | Coin IDs, not symbols |

> **Symbol note:** provider-native symbols (e.g. `BTCUSDT` for Binance, `BTC_USDT` for WhiteBit) are the input to all Watch* methods. The symbol engine (`_docs/symbol-engine.md`) handles user-input normalisation before providers are called; the ws layer operates on already-resolved native symbols.

---

## Architecture

### Component Hierarchy

```
ws.Pool                         ← manages N Sessions for providers with sub limits
  └─ ws.Session                 ← one live connection with full lifecycle
       ├─ ws.Conn               ← thin gorilla/websocket wrapper + write lock
       ├─ ws.Protocol           ← provider-specific wire format (required)
       ├─ ws.Heartbeater        ← optional: server-initiated ping/pong (Crypto.com)
       ├─ ws.SubscriptionRestorer ← optional: batch restore override (CoinGecko)
       ├─ ws.SubscriptionStore  ← mutex-protected map of live subscriptions
       └─ ws.Pipeline           ← ordered middleware (decode, validate, reconstruct)
```

**For single-connection providers** (Binance, Bitget, Crypto.com, WhiteBit, CoinGecko), Pool with `MaxSubsPerConn=0` degenerates to a single Session — no overhead.

**For MEXC**, Pool automatically shards subscriptions across multiple Sessions, each capped at 30.

---

## Core Types

### `ws.ConnConfig`

All per-provider knobs in one place.

```go
// internal/ws/config.go

type SlowConsumerPolicy int

const (
    // SlowConsumerBlock applies back-pressure: read loop stalls until consumer drains.
    // Safe but risks server-side disconnect on very slow consumers.
    SlowConsumerBlock SlowConsumerPolicy = iota
    // SlowConsumerDrop discards the oldest buffered message and logs a warning.
    // Keeps the connection alive under bursty conditions.
    SlowConsumerDrop
)

type ConnConfig struct {
    URL              string
    UserAgent        string

    HandshakeTimeout time.Duration      // default: 10s
    BackoffMin       time.Duration      // default: 1s
    BackoffMax       time.Duration      // default: 30s

    PingInterval     time.Duration      // 0 = no outbound ping (server-driven)
    PingTimeout      time.Duration      // max wait for pong reply; 0 = fire-and-forget

    MaxSubsPerConn   int                // 0 = unlimited (single session)
    WriteBufferSize  int                // 0 = gorilla default (4096)
    ReadBufferSize   int                // 0 = gorilla default (4096)
    WriteTimeout     time.Duration      // per-write deadline; 0 = no deadline (risky on network partition)
    // Recommended output channel buffer: 100 messages for market data,
    // 10 for low-frequency feeds (CoinGecko). Larger = more memory;
    // smaller = faster back-pressure detection.
    OutChanBuffer    int                // default: 100

    SlowConsumer     SlowConsumerPolicy // default: SlowConsumerBlock

    // Observability hooks — all optional.
    OnConnect    func(url string)
    OnDisconnect func(url string, err error)
    OnDrop       func(dropped int64) // called when SlowConsumerDrop discards a message
}
```

### `ws.Subscription`

A named, idempotent unit of subscription that the infrastructure can restore after reconnect.

```go
// internal/ws/subscription.go

type Subscription struct {
    Key    string // stable unique ID: "<channel>:<nativeSymbol>:<marketType>"
    Params any    // opaque params passed verbatim to Protocol.Subscribe/Unsubscribe
                  // Protocol.Subscribe MUST validate the type assertion and return
                  // model.ErrKindInvalidRequest if the type does not match expectations.
}
```

`Key` is provider-defined and must be stable across reconnects. Suggested format: `"ticker:BTCUSDT:SPOT"`.

### `ws.SubscriptionStore`

Mutex-protected store of live subscriptions. Used by Session on Subscribe/Unsubscribe and replayed in full on every reconnect.

```go
// internal/ws/subscription.go

type SubscriptionStore struct {
    mu   sync.RWMutex
    subs map[string]Subscription
}

func (s *SubscriptionStore) Add(sub Subscription)
func (s *SubscriptionStore) Remove(key string)
func (s *SubscriptionStore) All() []Subscription   // returns snapshot; safe to iterate
func (s *SubscriptionStore) Len() int
```

All methods are safe for concurrent use. `All()` returns a copy — the caller does not need to hold any lock.

### `ws.Protocol`

The core interface every provider author implements. Methods are called by Session at the appropriate lifecycle point.

```go
// internal/ws/protocol.go

type Protocol interface {
    // Dial is called once after every successful connection (initial + reconnect).
    // Use it for authentication handshakes or connection-level setup.
    // Subscriptions are replayed AFTER Dial returns.
    Dial(ctx context.Context, conn *Conn) error

    // Ping is called on every PingInterval tick.
    // For server-driven providers (Binance, CoinGecko), set PingInterval=0
    // and implement Ping as `return nil`.
    Ping(ctx context.Context, conn *Conn) error

    // Subscribe sends the wire request for one Subscription.
    // Called for new subscriptions AND for every subscription in the store on reconnect.
    // MUST validate sub.Params type and return *model.ProviderError on mismatch.
    Subscribe(ctx context.Context, conn *Conn, sub Subscription) error

    // Unsubscribe sends the unsubscribe wire request.
    // Called only on explicit live removes, never during reconnect.
    Unsubscribe(ctx context.Context, conn *Conn, sub Subscription) error

    // Parse decodes a raw WebSocket frame.
    // Return (nil, nil) to silently ignore (pong frames, acks, system messages).
    // Any non-nil first return is fed into the middleware Pipeline.
    Parse(ctx context.Context, raw []byte) (any, error)
}
```

### Optional Protocol Extensions

These interfaces are checked at runtime by Session via type assertion. Providers implement only what they need.

```go
// internal/ws/protocol.go

// Heartbeater is implemented by protocols where the server initiates keep-alive
// messages that require a client response (e.g. Crypto.com public/heartbeat).
// Session calls Heartbeat BEFORE Parse for every incoming message.
// If handled==true, Parse is skipped for that message.
type Heartbeater interface {
    Heartbeat(ctx context.Context, raw []byte, conn *Conn) (handled bool, err error)
}

// SubscriptionRestorer is implemented by protocols that cannot restore subscriptions
// one-by-one (e.g. CoinGecko, which requires a single batched set_tokens call).
// When present, Session calls RestoreSubscriptions instead of looping Protocol.Subscribe.
type SubscriptionRestorer interface {
    RestoreSubscriptions(ctx context.Context, conn *Conn, subs []Subscription) error
}
```

### `ws.Conn`

Thin wrapper over `gorilla/websocket` with write serialisation. Provides no reconnect logic — that belongs to Session.

```go
// internal/ws/conn.go

type Conn struct{ /* internal */ }

func NewConn(cfg ConnConfig) *Conn

// All Write* methods serialize concurrent callers with an internal mutex.
// If ConnConfig.WriteTimeout > 0, each write sets a per-call deadline via
// gorilla's SetWriteDeadline before sending. A timed-out write is treated as
// a connection failure and triggers the Session reconnect path.
func (c *Conn) WriteJSON(v any) error
func (c *Conn) WriteMessage(messageType int, data []byte) error
func (c *Conn) WritePing(data []byte) error

// IsConnected reports whether the underlying socket is currently open.
// Session uses this inside Subscribe to decide wire-send vs store-only.
// Callers outside Session should not rely on this for correctness — the state
// can change between the check and the subsequent write.
func (c *Conn) IsConnected() bool
```

### `ws.Pipeline`

Ordered chain of Middleware functions applied to every parsed message before it reaches the output channel. Each middleware is a pure function — composable and testable in isolation.

```go
// internal/ws/pipeline.go

// Middleware transforms or filters a message.
// Return (nil, nil) to drop the message. Return an error to surface it downstream.
type Middleware func(ctx context.Context, msg any, next func(any) (any, error)) (any, error)

type Pipeline []Middleware

// Apply runs msg through all middleware in order.
func (p Pipeline) Apply(ctx context.Context, msg any) (any, error)
```

Always terminate a pipeline with `TypeFilter[T]()` to catch upstream type errors early.

**Provided middleware** (`internal/ws/middleware/`):

| Middleware | Purpose | Used by |
|---|---|---|
| `CRC32Validator(extractFn)` | Validate first 25 bid/ask levels on `books` updates | Bitget |
| `OrderBookReconstructor()` | Merge snapshot + incremental delta into full book | Binance depth, WhiteBit depth |
| `ProtobufDecoder(newMsg func() proto.Message)` | Decode binary Protobuf frames | MEXC |
| `SymbolFilter(symbols)` | Drop messages for un-subscribed symbols | Any |
| `TypeFilter[T]()` | Drop messages not of type `*model.Response[T]`; log a warning | Any |

### `ws.Session`

One live WebSocket connection. Orchestrates the full lifecycle.

```go
// internal/ws/session.go

type SessionConfig struct {
    ConnConfig
    Protocol Protocol
    Pipeline Pipeline
}

func NewSession(cfg SessionConfig) *Session

// Start connects and returns the output channel.
// The channel is closed when ctx is cancelled or Stop() is called.
func (s *Session) Start(ctx context.Context) (<-chan StreamResponse[any], error)

// Subscribe adds a subscription.
// Thread-safe: if the connection is live, sends on the wire immediately.
// If reconnecting, the store update is picked up by the next restore pass.
func (s *Session) Subscribe(ctx context.Context, sub Subscription) error

// Unsubscribe removes a subscription.
// Thread-safe. Sends Unsubscribe on the wire only if currently connected.
func (s *Session) Unsubscribe(ctx context.Context, key string) error

// Stop initiates graceful shutdown (see Shutdown Sequence below).
func (s *Session) Stop()

// SubCount returns the number of live subscriptions.
func (s *Session) SubCount() int
```

**Subscribe concurrency contract:**

Session maintains a `connMu sync.RWMutex` protecting both the `Conn` state and the `SubscriptionStore`. Write lock is held by: `Subscribe`, `Unsubscribe`, and the reconnect path (dial + restore). Read lock is held by: `SubCount()` and `IsConnected()` when called externally. The reconnect path uses a write lock so that a concurrent `Subscribe` either completes before reconnect starts (and its subscription is already in the store for replay) or waits until reconnect finishes and then sends on the freshly connected wire — never lost, never sent twice.

**Session lifecycle:**

```
Start()
  │
  ├─ [write lock] Conn.Dial()
  ├─ Protocol.Dial()               ← auth handshake
  ├─ restore subscriptions          ← SubscriptionRestorer or loop Protocol.Subscribe
  ├─ [write lock released]
  │
  ├─ read loop (goroutine)
  │   ├─ Heartbeater.Heartbeat()   ← if implemented (Crypto.com); skip Parse if handled
  │   ├─ Protocol.Parse(raw)
  │   ├─ Pipeline.Apply(msg)
  │   ├─ → outChan (non-blocking if SlowConsumerDrop; blocking if SlowConsumerBlock)
  │   └─ on read error:
  │       ├─ [write lock] mark disconnected, reset backoff counter
  │       ├─ backoff + jitter
  │       ├─ Conn.Dial()
  │       ├─ Protocol.Dial()
  │       ├─ restore subscriptions
  │       ├─ [write lock released]
  │       └─ reset backoff to BackoffMin on first successful message after reconnect
  │
  └─ ping loop (goroutine, only if PingInterval > 0)
      └─ Protocol.Ping()
```

### `ws.Pool`

For providers with a per-connection subscription cap (MEXC: 30). Transparent to callers — the API is identical to Session.

```go
// internal/ws/pool.go

func NewPool(cfg SessionConfig, maxPerConn int) *Pool

func (p *Pool) Start(ctx context.Context) (<-chan StreamResponse[any], error)
func (p *Pool) Subscribe(ctx context.Context, sub Subscription) error
func (p *Pool) Unsubscribe(ctx context.Context, key string) error
func (p *Pool) Stop()
```

**Fan-in mechanism:**

```
Pool.Start()
  ├─ create Session[0], Start() → outChan[0]
  ├─ merged = make(chan StreamResponse[any], bufSize)
  ├─ wg.Add(1); go fanIn(outChan[0], merged, wg)
  └─ go func() { wg.Wait(); close(merged) }()

Pool.Subscribe() when Session[N] is full:
  ├─ create Session[N+1], Start() → outChan[N+1]
  ├─ wg.Add(1); go fanIn(outChan[N+1], merged, wg)
  └─ add sub to Session[N+1]
```

Each `fanIn` goroutine forwards all messages from one session's channel to `merged` and calls `wg.Done()` when the session closes. `merged` is closed only after all sessions are done — no data is lost.

**Session failure handling:**

If a Session exhausts its reconnect budget (BackoffMax reached repeatedly), it sends `StreamResponse{Error: &model.ProviderError{Kind: model.ErrKindNetwork, ProviderMessage: "reconnect budget exhausted"}}` on its output channel, then closes it. The fan-in goroutine forwards this error to `merged` before exiting. The caller receives the typed error and can decide whether to stop or restart the Pool.

Pool does **not** automatically redistribute subscriptions from a dead Session to live ones — doing so would silently re-subscribe without the caller's knowledge. Instead, surface the error and let the caller decide.

**For providers without a cap** (`maxPerConn=0`), Pool holds exactly one Session. No fan-in overhead.

### `ws.StreamResponse`

```go
type StreamResponse[T any] struct {
    Response T
    Error    error // *model.ProviderError post-migration; plain error during transition
}
```

> **Migration note:** During the error-handling migration (Phase 2 of `_docs/error-handling-proposal.md`), `Error` may be a plain `error`. Use `model.WrapError(providerID, err)` at the ws package boundary until providers are fully migrated. After Phase 2, the field type can be narrowed to `*model.ProviderError`.

---

## Typed Output Helper

Eliminates the per-provider filter goroutine. Returns both a data channel and an error channel so callers can act on errors.

```go
// internal/ws/typed.go

// TypedChan filters a raw output channel into typed responses and errors.
// Data messages that do not match *model.Response[T] are silently dropped.
//
// IMPORTANT: the caller MUST drain the returned error channel, typically in a
// background goroutine. If the error channel fills up, the forwarding goroutine
// blocks and data messages stop flowing — effectively a deadlock. Errors from
// a healthy stream are rare; a buffer of 8–16 is sufficient for bursts.
// The simplest safe pattern:
//
//	data, errs := ws.TypedChan[model.CoinPrice](out, 100)
//	go func() { for err := range errs { log.Warn(err) } }()
func TypedChan[T any](
    in <-chan StreamResponse[any],
    bufSize int,
) (<-chan *model.Response[T], <-chan *model.ProviderError) {
    out  := make(chan *model.Response[T], bufSize)
    errs := make(chan *model.ProviderError, 16) // small buffer; errors are rare
    go func() {
        defer close(out)
        defer close(errs)
        for sr := range in {
            if sr.Error != nil {
                if pe, ok := sr.Error.(*model.ProviderError); ok {
                    select {
                    case errs <- pe:
                    default:
                        // error channel full: drop error, never block data flow
                        // OnDrop hook on ConnConfig is called upstream for data drops;
                        // log this separately during implementation
                    }
                }
                continue
            }
            if resp, ok := sr.Response.(*model.Response[T]); ok {
                out <- resp
            }
        }
    }()
    return out, errs
}
```

---

## Error Handling

All errors from the ws package are wrapped as `*model.ProviderError`. Mapping:

| Event | ErrorKind | Retryable |
|---|---|---|
| Connection refused / DNS failure | `ErrKindNetwork` | Yes (Session retries internally) |
| Auth handshake rejected | `ErrKindAuth` | No — surfaced to caller |
| Provider error in message body | `ErrKindAuth / ErrKindRateLimit / …` | Per kind |
| JSON / Protobuf parse failure | `ErrKindParse` | No — surfaced to caller |
| `context.Canceled` / `DeadlineExceeded` | `ErrKindCanceled` | No |
| Server close 1008 (Policy Violation) | `ErrKindAuth` | No |
| Reconnect budget exhausted | `ErrKindNetwork` | No — terminal, surfaced to caller |

> **Distinction from `error-handling-proposal.md`:** `ErrKindNetwork` is normally `Retryable()==true`, meaning the *caller* should retry the operation. For WebSocket reconnect, Session retries internally and only surfaces `ErrKindNetwork` when it has given up. Callers receiving this error should treat it as terminal for the current stream and restart from scratch.

---

## Graceful Shutdown

`Stop()` follows this sequence:

1. Set `shuttingDown` flag; reject new `Subscribe` calls with an error.
2. For each live subscription: call `Protocol.Unsubscribe` (best-effort, errors logged).
3. Send a WebSocket close frame (`1000 Normal Closure`).
4. Wait for the read loop to exit (short drain timeout, default 2 s).
5. Force-close the underlying socket if the read loop has not exited.
6. Close `outChan`.

For Pool: run the above sequence for every Session concurrently, then close `merged`.

Callers that cancel the parent context instead of calling `Stop()` get steps 4–6 only — no unsubscribe messages are sent on the wire.

---

## Provider Implementation Guide

### Minimal Protocol

```go
// pkg/provider/myprovider/ws_protocol.go

type myProtocol struct {
    providerID string
    apiKey     string
}

func (p *myProtocol) Dial(ctx context.Context, conn *ws.Conn) error {
    if p.apiKey == "" {
        return nil
    }
    return conn.WriteJSON(map[string]any{"method": "auth", "key": p.apiKey})
}

func (p *myProtocol) Ping(ctx context.Context, conn *ws.Conn) error {
    return conn.WriteJSON("ping")
}

func (p *myProtocol) Subscribe(ctx context.Context, conn *ws.Conn, sub ws.Subscription) error {
    args, ok := sub.Params.([]map[string]string)
    if !ok {
        return &model.ProviderError{Kind: model.ErrKindInvalidRequest,
            ProviderID: p.providerID, ProviderMessage: "invalid subscribe params type"}
    }
    return conn.WriteJSON(map[string]any{"op": "subscribe", "args": args})
}

func (p *myProtocol) Unsubscribe(ctx context.Context, conn *ws.Conn, sub ws.Subscription) error {
    args, ok := sub.Params.([]map[string]string)
    if !ok {
        return nil // best-effort
    }
    return conn.WriteJSON(map[string]any{"op": "unsubscribe", "args": args})
}

func (p *myProtocol) Parse(ctx context.Context, raw []byte) (any, error) {
    // decode and return *model.Response[T], or (nil, nil) to ignore
    return nil, nil
}
```

### Provider `WatchPrices` Pattern

```go
// pkg/provider/myprovider/stream.go

func (c *Client) pool(ctx context.Context) (*ws.Pool, <-chan ws.StreamResponse[any], error) {
    cfg := ws.SessionConfig{
        ConnConfig: ws.ConnConfig{
            URL:          wsPublicURL,
            PingInterval: 30 * time.Second,
            OutChanBuffer: 100,
        },
        Protocol: &myProtocol{providerID: providerID},
    }
    pool := ws.NewPool(cfg, 0) // 0 = unlimited
    out, err := pool.Start(ctx)
    return pool, out, err
}

func (c *Client) WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
    pool, out, err := c.pool(ctx)
    if err != nil {
        return nil, err
    }
    for _, id := range ids {
        if err := pool.Subscribe(ctx, ws.Subscription{
            Key:    "ticker:" + id + ":SPOT",
            Params: []map[string]string{{"instType": "SPOT", "channel": "ticker", "instId": id}},
        }); err != nil {
            pool.Stop()
            return nil, err
        }
    }
    // ids are already provider-native symbols (resolved by symbol engine before this call)
    data, errs := ws.TypedChan[model.CoinPrice](out, 100)
    go func() {
        for err := range errs {
            logger.Default.Warn("price stream error", "provider", providerID, "err", err)
        }
    }()
    // unwrap model.Response to raw CoinPrice for the streaming interface
    prices := make(chan *model.CoinPrice, 100)
    go func() {
        defer close(prices)
        for resp := range data {
            prices <- &resp.Data
        }
    }()
    return prices, nil
}
```

---

## Provider-Specific Notes

### Binance
- `PingInterval = 0` — gorilla handles WebSocket-level pong automatically.
- Spot uses URL-based combined streams. `Dial` may embed the initial stream list in the URL; dynamic subscribe/unsubscribe use JSON-RPC `{"method":"SUBSCRIBE","params":[...],"id":N}`.
- 24-hour connection limit — Session will reconnect automatically; subscriptions restore.
- Symbol format: concatenated uppercase (`BTCUSDT`). Already provided as-is from symbol engine.

### Bitget
- `PingInterval = 30s`; `Ping` sends plain string `"ping"` (not JSON).
- `books` channel: add `CRC32Validator(bitget.ExtractChecksum)` + `OrderBookReconstructor` middleware.
- Symbol format: `instType` field in the subscription arg (`"SPOT"`, `"USDT-FUTURES"`, `"COIN-FUTURES"`) differentiates market types; the `instId` value itself does **not** carry a `_PERP` suffix in the WS API (unlike the REST API). Verify this against `_docs/providers/apis/bitget/bitget-market-ws.md` during implementation.

### MEXC
- `MaxSubsPerConn = 30` → Pool shards automatically.
- All frames are binary Protobuf — `Parse` must use `ProtobufDecoder`.
- `PingInterval = 60s`; idle connections disconnect after 1 minute.
- Symbol format: `BTCUSDT` (spot), `BTC_USDT` (futures) — symbol engine handles this split.

### Crypto.com
- Server sends `public/heartbeat` every 30 s; client must respond within 5 s.
- Implement `Heartbeater`: detect `method == "public/heartbeat"`, write `public/respond-heartbeat` with same `id`, return `handled=true`.
- `Ping` returns `nil` (set `PingInterval=0`).
- Add 1 s delay at the start of `Dial` (API requirement).
- Symbol format: `BTC_USDT` (spot), `BTCUSD-PERP` (perpetuals).

### WhiteBit
- JSON-RPC 2.0 subscribe: `{"id":N, "method":"lastprice_subscribe", "params":["BTC_USDT"]}`.
- `PingInterval = 50s` (60 s inactivity timeout).
- Depth stream: snapshot then incremental updates — add `OrderBookReconstructor` middleware.
- Symbol format: `BTC_USDT` (spot), `BTC_PERP` (perpetuals).

### CoinGecko
- API key required in connection URL query param (`?x_cg_pro_api_key=...`).
- ActionCable protocol: subscription is a single channel for all coins; all coin IDs are sent together in `set_tokens`.
- Implement `SubscriptionRestorer`: collect all stored subscriptions, build a single `set_tokens` message with all coin IDs, send it. Without this, looped `Protocol.Subscribe` would issue N separate `set_tokens` messages, each overwriting the previous.
- `Ping` returns `nil` (server-driven; gorilla handles it).
- Symbol note: CoinGecko uses coin IDs (`bitcoin`, `ethereum`), not trading symbols. The symbol engine resolves user input to these IDs before calling Watch*. See `_docs/symbol-engine.md`.

---

## Migration from Current `Manager`

The current `Manager` type remains functional. Migration is **additive** — new providers use the new API; existing ones migrate one PR at a time.

### Mapping

| Current | New |
|---|---|
| `MessageHandler` | `Protocol` |
| `MessageHandler.Handle` | `Protocol.Parse` |
| `MessageHandler.OnCommand` | `Protocol.Subscribe` / `Protocol.Unsubscribe` |
| `MessageHandler.OnPing` | `Protocol.Ping` |
| `Manager` | `Session` |
| — | `Pool` (new, for MEXC) |
| — | `Heartbeater` (new, for Crypto.com) |
| — | `SubscriptionRestorer` (new, for CoinGecko) |
| per-provider filter goroutine | `ws.TypedChan[T]` |
| hardcoded `pingTimeout: 30s` | `ConnConfig.PingInterval` per provider |
| `Manager.subs` map (broken key collision) | `SubscriptionStore` with stable `Key` |

### Migration Checklist

```
[ ] Add ws.Conn, ws.Protocol, ws.Heartbeater, ws.SubscriptionRestorer (internal/ws/)
[ ] Add ws.SubscriptionStore
[ ] Add ws.Session, ws.Pool
[ ] Add ws.TypedChan[T]
[ ] Add ws.ConnConfig with SlowConsumerPolicy and observability hooks
[ ] Add middleware package (CRC32Validator, OrderBookReconstructor, ProtobufDecoder)
[ ] Migrate bitget:   bitgetHandler → bitgetProtocol, Manager → Pool
[ ] Migrate binance:  same pattern
[ ] Migrate mexc:     Pool(maxPerConn=30) + ProtobufDecoder
[ ] Migrate cryptocom: implement Heartbeater
[ ] Migrate whitebit: JSON-RPC subscribe, OrderBookReconstructor
[ ] Migrate coingecko: implement SubscriptionRestorer, ActionCable parse
[ ] Remove Manager after all providers migrated (or keep as deprecated alias)
```

---

## Testing Strategy

- **Protocol in isolation**: implement a `MockConn` that records `WriteJSON` calls; assert subscribe/unsubscribe payloads without a live connection.
- **Pipeline in isolation**: `Pipeline.Apply` is a pure function — unit-test each middleware with table-driven inputs.
- **Session integration**: use `net/http/httptest` WebSocket echo server; verify reconnect, subscription restore, and backoff behaviour.
- **Pool**: test with a server that accepts at most N subscriptions per connection; verify sharding.
- **CRC32 / OrderBookReconstructor**: test with recorded real provider payloads.

---

## Usage Example (call site)

```go
// No knowledge of ws internals at the call site.
// Symbols are already in provider-native format (resolved by symbol engine).
prices, err := provider.WatchPrices(ctx, []string{"BTCUSDT", "ETHUSDT"})
if err != nil {
    // err is *model.ProviderError
}

for price := range prices {
    fmt.Printf("%s: %.2f\n", price.Symbol, price.Price)
}
```

---

## What Is Out of Scope

- **Retry logic at call site** — callers use `model.ProviderError.Kind.Retryable()`.
- **Circuit breaking** — out of scope.
- **Rate-limiting outbound commands** — providers must respect per-provider message rate limits in their `Protocol` implementation.
- **Authenticated (private) streams** — the `Dial` hook is the extension point; credentials flow through `ConnConfig` or a closure; full private stream support is deferred.
