# WIP: WebSocket Order Book Support & Multi-Provider Streaming

## Objective
Add support for real-time order book streaming across multiple providers (Binance, Bitget, CoinGecko) using a unified interface and a common WebSocket management core.

---

## 1. Architecture Overview

### Generic Streaming Core (`internal/ws/`)
Refactor the current CoinGecko-specific WebSocket logic into a reusable base client that handles common concerns:
- **Connection Management**: Dials endpoints, manages the `*websocket.Conn` lifecycle.
- **Robust Reconnection**: Exponential backoff with jitter (already implemented in CG client, to be generalized).
- **Read/Write Loops**: Standardized goroutines for handling inbound/outbound traffic.
- **Health Checks**: Generalized liveness deadlines and ping/pong handling.
- **Concurrency Safety**: Thread-safe connection access and state management.

### Capability Interfaces (`internal/provider/types.go`)
Define new interfaces for real-time data:

```go
// StreamProvider is the base interface for all streaming capabilities.
type StreamProvider interface {
    ID() string
}

// OrderBookStreamProvider supports real-time order book updates.
type OrderBookStreamProvider interface {
    StreamProvider
    // WatchOrderBook returns a channel for real-time OrderBook updates for multiple symbols.
    WatchOrderBook(ctx context.Context, symbols []string, limit int) (<-chan *model.OrderBook, error)
}

// PriceStreamProvider (formerly internal/ws/client.go logic)
type PriceStreamProvider interface {
    StreamProvider
    WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinUpdate, error)
}
```

---

## 2. Implementation Plan

### Step 1: Shared Libraries & Utilities
- **`internal/ws/base_client.go`**: Extract the reconnection, backoff, and read loop logic from the current CoinGecko client into a generic `BaseClient` struct.
- **`internal/auth/signature.go`**: Create a shared utility for HMAC-SHA256 signatures (useful for Bitget and future authenticated providers).
- **`internal/model/types.go`**: Ensure `OrderBook` and `CoinUpdate` models are sufficient for streaming (e.g., adding `Sequence` or `LastUpdateID` if needed for order book deltas).

### Step 2: Refactor CoinGecko Streaming
- Move existing CoinGecko WS logic from `internal/ws/` to `internal/provider/coingecko/ws.go`.
- Make it implement `PriceStreamProvider`.
- Use the new `BaseClient` to reduce boilerplate.

### Step 3: Binance Order Book Stream
- **Location**: `internal/provider/binance/ws.go`
- **Implementation**: Leverage `go-binance/v2`'s `WsDepthServe` or implement a raw client using the `BaseClient`.
- **Logic**: Map Binance `@depth` stream payloads to `model.OrderBook`.

### Step 4: Bitget Order Book Stream
- **Location**: `internal/provider/bitget/ws.go`
- **Implementation**: Implement Bitget's WS protocol (Login -> Subscribe `books`).
- **Logic**: Handle Bitget's specific JSON-RPC style messages and map them to `model.OrderBook`.

### Step 5: CLI Integration
- **`cmd/watch.go`**: Extend the `watch` command to support order books.
  ```bash
  bits watch orderbook BTCUSDT -p binance
  bits watch orderbook BTCUSDT ETHUSDT -p binance
  ```
- **Capability Checks**: Ensure the selected provider implements `OrderBookStreamProvider`.

---

## 3. Detailed Skeleton Design (`internal/ws/base_client.go`)

```go
type BaseClient struct {
    URL        string
    UserAgent  string
    Backoff    BackoffConfig
    
    onConnect  func(ctx context.Context, conn *websocket.Conn) error
    onMessage  func(raw []byte) error
    pingTicker *time.Ticker
    
    mu         sync.Mutex
    conn       *websocket.Conn
    // ... (reconnect logic)
}
```

---

## 4. Verification Plan
- **Unit Tests**: Mock WebSocket servers to test the `BaseClient` reconnection logic and provider-specific parsers.
- **Integration Tests**: Real-world connectivity tests (limited) to Binance/Bitget/CoinGecko public endpoints.
- **TUI Integration**: Future plan to integrate real-time order books into the `bits tui` detail view.
