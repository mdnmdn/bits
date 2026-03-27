# Binance Go Package Documentation

The `github.com/adshao/go-binance/v2` package provides a Go client for the Binance API, supporting Spot, Futures, Delivery (Coin-M Futures), and Options trading.

## Package Structure

```
github.com/adshao/go-binance/v2/
├── client.go           # Main client with all services
├── futures/            # USDT-M futures client
├── delivery/           # Coin-M futures client
├── options/            # Options client
└── common/             # Shared types and utilities
```

## Main Client

### Creation

```go
func NewClient(apiKey, secretKey string) *Client
func NewProxiedClient(apiKey, secretKey, proxyUrl string) *Client
func NewFuturesClient(apiKey, secretKey string) *futures.Client
func NewDeliveryClient(apiKey, secretKey string) *delivery.Client
func NewOptionsClient(apiKey, secretKey string) *options.Client
```

### Client Struct

```go
type Client struct {
    APIKey     string
    SecretKey  string
    KeyType    string
    BaseURL    string
    UserAgent  string
    HTTPClient *http.Client
    Debug      bool
    Logger     *log.Logger
    TimeOffset int64
}
```

### Endpoints

```go
var BaseAPIMainURL    = "https://api.binance.com"
var BaseAPITestnetURL = "https://testnet.binance.vision"
var BaseAPIDemoURL    = "https://demo-api.binance.com"

var UseTestnet = false
var UseDemo    = false
```

## Core Types

### Enums

```go
type SideType string          // "BUY" or "SELL"
type OrderType string         // "LIMIT", "MARKET", "STOP_LOSS", etc.
type TimeInForceType string   // "GTC", "IOC", "FOK"
type OrderStatusType string   // "NEW", "FILLED", "CANCELED", etc.
type SymbolType string        // "SPOT"
type SymbolStatusType string   // "TRADING", "HALT", etc.
type SymbolFilterType string  // "LOT_SIZE", "PRICE_FILTER", etc.
type AccountType string       // "SPOT", "MARGIN", "USDT_FUTURE", etc.
```

### Constants

```go
const (
    SideTypeBuy  SideType = "BUY"
    SideTypeSell SideType = "SELL"

    OrderTypeLimit           OrderType = "LIMIT"
    OrderTypeMarket          OrderType = "MARKET"
    OrderTypeLimitMaker      OrderType = "LIMIT_MAKER"
    OrderTypeStopLoss        OrderType = "STOP_LOSS"
    OrderTypeStopLossLimit   OrderType = "STOP_LOSS_LIMIT"
    OrderTypeTakeProfit      OrderType = "TAKE_PROFIT"
    OrderTypeTakeProfitLimit OrderType = "TAKE_PROFIT_LIMIT"

    TimeInForceTypeGTC TimeInForceType = "GTC"
    TimeInForceTypeIOC TimeInForceType = "IOC"
    TimeInForceTypeFOK TimeInForceType = "FOK"

    OrderStatusTypeNew             OrderStatusType = "NEW"
    OrderStatusTypePartiallyFilled OrderStatusType = "PARTIALLY_FILLED"
    OrderStatusTypeFilled          OrderStatusType = "FILLED"
    OrderStatusTypeCanceled        OrderStatusType = "CANCELED"
    OrderStatusTypeRejected        OrderStatusType = "REJECTED"
    OrderStatusTypeExpired         OrderStatusType = "EXPIRED"
)
```

## Market Data Services

### Depth Service

```go
type DepthService struct {
    c      *Client
    symbol string
    limit  *int
}

func (s *DepthService) Symbol(symbol string) *DepthService
func (s *DepthService) Limit(limit int) *DepthService
func (s *DepthService) Do(ctx context.Context, opts ...RequestOption) (*DepthResponse, error)

type DepthResponse struct {
    LastUpdateID int64  `json:"lastUpdateId"`
    Bids         []Bid  `json:"bids"`
    Asks         []Ask  `json:"asks"`
}

type Bid = common.PriceLevel  // PriceLevel{Price, Quantity}
type Ask = common.PriceLevel
```

### Ticker Services

```go
// ListBookTickersService - Best price/qty on order book
type ListBookTickersService struct {
    c       *Client
    symbol  *string
    symbols []string
}
func (s *ListBookTickersService) Symbol(symbol string) *ListBookTickersService
func (s *ListBookTickersService) Symbols(symbols ...string) *ListBookTickersService
func (s *ListBookTickersService) Do(ctx context.Context, opts ...RequestOption) ([]*BookTicker, error)

type BookTicker struct {
    Symbol      string `json:"symbol"`
    BidPrice    string `json:"bidPrice"`
    BidQuantity string `json:"bidQty"`
    AskPrice    string `json:"askPrice"`
    AskQuantity string `json:"askQty"`
}

// ListPricesService - Latest price for symbols
type ListPricesService struct {
    c       *Client
    symbol  *string
    symbols []string
}

type SymbolPrice struct {
    Symbol string `json:"symbol"`
    Price  string `json:"price"`
}

// ListPriceChangeStatsService - 24h price change stats
type ListPriceChangeStatsService struct {
    c       *Client
    symbol  *string
    symbols []string
}

type PriceChangeStats struct {
    Symbol             string `json:"symbol"`
    PriceChange        string `json:"priceChange"`
    PriceChangePercent string `json:"priceChangePercent"`
    WeightedAvgPrice   string `json:"weightedAvgPrice"`
    PrevClosePrice     string `json:"prevClosePrice"`
    LastPrice          string `json:"lastPrice"`
    LastQty            string `json:"lastQty"`
    BidPrice           string `json:"bidPrice"`
    BidQty             string `json:"bidQty"`
    AskPrice           string `json:"askPrice"`
    AskQty             string `json:"askQty"`
    OpenPrice          string `json:"openPrice"`
    HighPrice          string `json:"highPrice"`
    LowPrice           string `json:"lowPrice"`
    Volume             string `json:"volume"`
    QuoteVolume        string `json:"quoteVolume"`
    OpenTime           int64  `json:"openTime"`
    CloseTime          int64  `json:"closeTime"`
    FirstID            int64  `json:"firstId"`
    LastID             int64  `json:"lastId"`
    Count              int64  `json:"count"`
}

// AveragePriceService - Current average price
type AveragePriceService struct {
    c      *Client
    symbol string
}

type AvgPrice struct {
    Mins  int64  `json:"mins"`
    Price string `json:"price"`
}

// ListSymbolTickerService - Rolling window price change stats
type ListSymbolTickerService struct {
    c          *Client
    symbol     *string
    symbols    []string
    windowSize *string  // "1m", "1h", "1d", etc.
}

type SymbolTicker struct {
    Symbol             string `json:"symbol"`
    PriceChange        string `json:"priceChange"`
    PriceChangePercent string `json:"priceChangePercent"`
    WeightedAvgPrice   string `json:"weightedAvgPrice"`
    OpenPrice          string `json:"openPrice"`
    HighPrice          string `json:"highPrice"`
    LowPrice           string `json:"lowPrice"`
    LastPrice          string `json:"lastPrice"`
    Volume             string `json:"volume"`
    QuoteVolume        string `json:"quoteVolume"`
    OpenTime           int64  `json:"openTime"`
    CloseTime          int64  `json:"closeTime"`
    FirstId            int64  `json:"firstId"`
    LastId             int64  `json:"lastId"`
    Count              int64  `json:"count"`
}
```

### Kline Service

```go
type KlineService struct {
    c        *Client
    symbol   string
    interval string    // "1m", "5m", "1h", "1d", etc.
    startTime *int64
    endTime   *int64
    limit     *int
}
```

## WebSocket Services

### WebSocket Configuration

```go
type WsConfig struct {
    Endpoint string
    Header   http.Header
    Proxy    *string
}

type WsHandler func(message []byte)
type ErrHandler func(err error)
type ConnHandler func(context.Context, *websocket.Conn)
```

### Functions

```go
func newWsConfig(endpoint string) *WsConfig
func wsServe(cfg *WsConfig, handler WsHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, error)
func keepAliveWithPing(interval time.Duration, pongTimeout time.Duration) ConnHandler
func keepAliveWithPong(ctx context.Context, c *websocket.Conn, timeout time.Duration)
func WsGetReadWriteConnection(cfg *WsConfig) (*websocket.Conn, error)
```

### Variables

```go
var WebsocketKeepalive bool
var WebsocketTimeout time.Duration
var WebsocketPingTimeout time.Duration
var WebsocketPongTimeout time.Duration
var wsServeWithConnHandler func(cfg *WsConfig, handler WsHandler, errHandler ErrHandler, connHandler ConnHandler) (doneC, stopC chan struct{}, error)
```

## Utility Functions

```go
func currentTimestamp() int64
func FormatTimestamp(t time.Time) int64  // Returns Unix ms
func NewClient(apiKey, secretKey string) *Client
func getAPIEndpoint() string
```

## Service Creation Pattern

All services follow a fluent builder pattern:

```go
// Create service
service := client.NewDepthService()

// Configure with chainable methods
service.Symbol("BTCUSDT").Limit(10)

// Execute
response, err := service.Do(context.Background())
```

## Key Sub-packages

### Futures Client (`futures/`)

```go
var BaseApiMainUrl    = "https://fapi.binance.com"
var BaseApiTestnetUrl = "https://testnet.binancefuture.com"

type Client struct {
    APIKey     string
    SecretKey  string
    KeyType    string
    BaseURL    string
    // ...
}

func NewClient(apiKey, secretKey string) *Client
```

### Delivery Client (`delivery/`)

Coin-M futures API client.

### Options Client (`options/`)

Binance Options API client.

### Common Package (`common/`)

Shared types including:
- `PriceLevel` - Bid/Ask level with Price and Quantity
- `SignFunc` - Signature generation function
- `UsedWeight`, `OrderCount` - Rate limiting tracking

## Integration with bits

This package can be used to implement the Binance provider for the `bits` CLI tool, supporting:
- Spot market data (prices, tickers, order book)
- Futures market data
- WebSocket streams for real-time data
