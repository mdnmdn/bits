# Proposal: Generic Crypto CLI Architecture

This document outlines the strategy for transforming the current CoinGecko-specific CLI into a multi-provider, multi-capability crypto tool named `bits`.

## 1. Vision & Goals

- **Provider Agnostic**: Support multiple data sources (CoinGecko, CoinMarketCap) and exchanges (Binance, Bitget).
- **Capability Based**: Providers declare what they can do (Market Data, Trading, Real-time, Historical).
- **Uniform Experience**: A single command structure that works across different backends.
- **Pluggable Output**: Support for Table, JSON, CSV, MD, YAML, and TOON.

---

## 2. Hypothetical CLI Commands

The new CLI will use a `--provider` (or `-p`) flag to specify the backend. If omitted, it will use a default provider from the config.

### Market Data
```bash
# Fetch prices from different sources
bits price btc,eth -p coingecko
bits price btc,eth -p binance

# List markets
bits markets -p coingecko --category layer-1
bits markets -p binance --vs usdt

# Search for assets
bits search solana -p coingecko
```

### Trading (CEX/DEX)
```bash
# Check balance
bits balance -p binance

# Place orders
bits buy btc 0.1 --price 60000 -p binance
bits sell eth 1.5 --type market -p bitget

# View order book
bits orderbook btc/usdt -p binance
```

### Historical & Real-time
```bash
# Historical OHLC
bits history bitcoin --days 30 --ohlc -p coingecko
bits history btc/usdt --from 2024-01-01 --to 2024-01-31 -p binance

# Live monitoring
bits watch btc,eth -p binance
```

### Output Formatting
```bash
bits price btc -o json
bits price btc -o csv
bits price btc -o md
bits price btc -o yaml
bits price btc -o toon
```

---

## 3. Proposed Architecture

### Core Interfaces (`internal/provider`)

Instead of a single `Client`, we define interfaces based on **Capabilities**.

```go
type Provider interface {
    ID() string
    Name() string
    Capabilities() []Capability
}

// Capability represents a specific feature a provider supports
type Capability string
const (
    CapMarketData Capability = "market_data"
    CapTrading    Capability = "trading"
    CapRealtime   Capability = "realtime"
    CapHistory    Capability = "history"
)

type MarketDataProvider interface {
    GetPrice(ctx context.Context, ids []string, vs string) (api.PriceResponse, error)
    GetMarkets(ctx context.Context, params MarketParams) ([]api.MarketCoin, error)
}

type TradeProvider interface {
    GetBalance(ctx context.Context) (api.BalanceResponse, error)
    PlaceOrder(ctx context.Context, order api.OrderRequest) (api.OrderResponse, error)
}

type RealtimeProvider interface {
    StreamPrices(ctx context.Context, symbols []string) (<-chan api.PriceUpdate, error)
}
```

### Project Structure Evolution

```text
bits/
├── cmd/                           # Cobra commands (refactored to use provider registry)
├── internal/
│   ├── provider/                  # Core provider logic
│   │   ├── registry.go            # Factory to get providers by ID
│   │   ├── coingecko/             # Current logic moved here
│   │   ├── binance/               # New Binance implementation
│   │   └── bitget/                # New Bitget implementation
│   ├── output/                    # NEW: Universal formatter
│   │   ├── formatter.go           # Interface for table, json, md, etc.
│   │   ├── table.go
│   │   ├── json.go
│   │   ├── csv.go
│   │   ├── markdown.go
│   │   ├── yaml.go
│   │   └── toon.go
│   ├── model/                     # Generic data models (Price, Order, Market)
│   └── ...
```

---

## 4. Reusing Current Code

The existing code is highly modular and can be repurposed as the `coingecko` provider implementation:

1.  **`internal/api/`**: Becomes the foundation of `internal/provider/coingecko/`.
2.  **`internal/display/`**: Its logic (formatting prices, numbers, tables) moves into the new `internal/output/` package.
3.  **`internal/config/`**: Updated to handle multiple provider API keys and default provider selection.
4.  **`internal/tui/`**: Can remain generic if we feed it data from the `MarketDataProvider` interface.

---

## 5. Output Formatter Design

We will implement an `OutputFormatter` interface that handles different data types:

```go
type OutputFormatter interface {
    Format(data any, writer io.Writer) error
}
```

- **MD (Markdown)**: Renders data as GitHub-flavored markdown tables.
- **CSV**: Standard comma-separated values.
- **JSON/YAML**: Standard serialization.
- **TOON**: A stylized, highly-visual terminal format (conceptually similar to the current "Pretty Table" but with more "cartoonish"/bold ASCII borders and vibrant colors).

---

## 7. Unified Data Model

To support advanced trading features and multi-provider data, we define a core `model` package with the following structures.

### Asset & Symbol Metadata
This model handles both pure data (CoinGecko) and exchange-specific pairs (Binance).

```go
type Asset struct {
    ID       string // e.g. "bitcoin" (Coingecko) or "BTC" (Binance)
    Symbol   string // e.g. "btc"
    Name     string // e.g. "Bitcoin"
}

type SymbolStatus string
const (
    StatusEnabled  SymbolStatus = "enabled"
    StatusDisabled SymbolStatus = "disabled"
    StatusBreak    SymbolStatus = "break"
)

type SymbolInfo struct {
    ID            string       // Provider-specific ID (e.g. "BTCUSDT")
    BaseAsset     Asset
    QuoteAsset    Asset
    Status        SymbolStatus
    
    // Precisions
    PricePrecision int // Decimal places for price
    QtyPrecision   int // Decimal places for quantity
    
    // Trading Filters/Constraints
    MinQty         float64 // Minimum order quantity
    MaxQty         float64 // Maximum order quantity
    StepSize       float64 // Quantity increment
    MinNotional    float64 // Minimum order value (Price * Qty)
    TickSize       float64 // Price increment
}
```

### Market Data
```go
type PriceTicker struct {
    Symbol      string
    Price       float64
    Change24h   float64 // Percentage
    High24h     float64
    Low24h      float64
    Volume24h   float64
    Timestamp   time.Time
}

type Candle struct {
    Timestamp time.Time
    Open      float64
    High      float64
    Low       float64
    Close     float64
    Volume    float64
}
```

### Trading & Account
```go
type Balance struct {
    Asset     string
    Free      float64
    Locked    float64 // Funds in open orders
    Total     float64
}

type OrderType string
const (
    OrderLimit  OrderType = "limit"
    OrderMarket OrderType = "market"
    OrderStop   OrderType = "stop_loss"
)

type OrderSide string
const (
    SideBuy  OrderSide = "buy"
    SideSell OrderSide = "sell"
)

type OrderStatus string
const (
    StatusNew             OrderStatus = "new"
    StatusPartiallyFilled OrderStatus = "partially_filled"
    StatusFilled          OrderStatus = "filled"
    StatusCanceled        OrderStatus = "canceled"
)

type Order struct {
    ID            string
    Symbol        string
    Side          OrderSide
    Type          OrderType
    Status        OrderStatus
    Price         float64
    Qty           float64
    ExecutedQty   float64
    CumulativeQuoteQty float64 // Total spend/receive in quote asset
    CreatedAt     time.Time
}
```

### Provider Implementation Note
- **CoinGecko**: Maps `PriceTicker` but returns empty/zero for `Trading Filters` since it's a data-only provider.
- **Binance/Bitget**: Fully populates `SymbolInfo` filters from their `exchangeInfo` endpoints to allow for client-side validation before sending orders.

1.  **Phase 1: Extraction**: Move existing CoinGecko client into a provider-specific package. Define the first generic interfaces for `Price` and `Market` data.
2.  **Phase 2: Registry**: Implement a provider registry and update CLI flags to support `-p`.
3.  **Phase 3: Output Abstraction**: Replace direct `display.PrintTable` calls with a formatter factory.
4.  **Phase 4: New Providers**: Add a second provider (e.g., Binance Read-Only) to validate the architecture.
5.  **Phase 5: Trading**: Introduce the `TradeProvider` interface and commands for order placement.
