# Provider API Documentation Index

Master index for the formalized exchange provider API documentation.

## Target Structure

```
_docs/providers/apis/
├── provider-index.md          # This file
├── binance/
│   ├── binance-general.md        # Base URL, rate limits, auth, exchange info APIs
│   ├── binance-market.md         # Prices, orderbooks, candles, tickers, streams
│   ├── binance-market-ws.md      # WebSocket spot market streams
│   ├── binance-market-futures-ws.md  # WebSocket futures market streams (USDT-M + COIN-M)
│   ├── binance-accounts.md       # Account REST APIs (balance, subaccounts, transfers)
│   └── binance-spot-order.md     # Spot order REST APIs (place, cancel, query, history)
├── bitget/
│   ├── bitget-general.md         # Base URL, rate limits, auth, exchange info APIs
│   ├── bitget-market.md          # Prices, orderbooks, candles, tickers, streams
│   ├── bitget-market-ws.md       # WebSocket spot market streams
│   ├── bitget-market-futures-ws.md   # WebSocket futures market streams (USDT-M + Coin-M)
│   ├── bitget-accounts.md        # Account REST APIs (balance, subaccounts, transfers)
│   └── bitget-spot-order.md      # Spot order REST APIs (place, cancel, query, history)
├── coingecko/
│   ├── coingecko-general.md      # Base URL, rate limits, auth, exchange info APIs
│   ├── coingecko-market.md       # Prices, orderbooks, candles, tickers, streams
│   ├── coingecko-market-ws.md    # WebSocket spot market streams
│   └── coingecko-spot-order.md   # N/A — aggregator only, no order APIs
├── whitebit/
│   ├── whitebit-general.md       # Base URL, rate limits, auth, exchange info APIs
│   ├── whitebit-market.md        # Prices, orderbooks, candles, tickers, streams
│   ├── whitebit-market-ws.md     # WebSocket spot market streams
│   ├── whitebit-accounts.md      # Account REST APIs (balance, subaccounts, transfers)
│   └── whitebit-spot-order.md    # Spot order REST APIs (place, cancel, query, history)
├── cryptocom/
│   ├── cryptocom-general.md      # Base URL, rate limits, auth, exchange info APIs
│   ├── cryptocom-market.md       # Prices, orderbooks, candles, tickers, streams
│   ├── cryptocom-market-ws.md    # WebSocket spot market streams
│   ├── cryptocom-accounts.md     # Account REST APIs (balance, subaccounts, transfers)
│   └── cryptocom-spot-order.md   # Spot order REST APIs (place, cancel, query, history)
└── mexc/
    ├── mexc-general.md           # Base URL, rate limits, auth, exchange info APIs
    ├── mexc-market.md            # Prices, orderbooks, candles, tickers, streams
    ├── mexc-market-ws.md         # WebSocket spot market streams
    ├── mexc-market-futures-ws.md # WebSocket futures market streams
    ├── mexc-accounts.md          # Account REST APIs (balance, subaccounts, transfers)
    └── mexc-spot-order.md        # Spot order REST APIs (place, cancel, query, history)
```

## Document Template

### `{provider}-general.md`
- **Overview**: Provider description, official site, API docs link
- **Base URLs**: REST and WebSocket endpoints (prod + testnet if applicable)
- **Authentication**: Headers, signature algorithm, required fields
- **Rate Limits**: IP-based, UID-based, per-endpoint limits
- **Products**: Spot, Margin, Futures support
- **Exchange Info APIs**: Server time, symbol info, market configuration
- **WebSocket**: Connection details, keepalive, channels

### `{provider}-market.md`
For each market data API:
- **Description** and any quirks/notes
- **URL** (complete path with method)
- **Parameters** (table: name, type, required, description)
- **Return Values** (table: field, type, description)
- **Sample Request/Response** (JSON)
- **Rate Limit Weight** (if applicable)

APIs covered:
- Price / Ticker
- 24h Ticker
- Order Book / Depth
- Candles / Klines
- Recent Trades
- Exchange Info (market-specific)

### `{provider}-spot-order.md`
For each order API:
- **Description** and any quirks/notes
- **URL** (complete path with HTTP method)
- **Parameters** (table: name, type, required, description)
- **Response Fields** (table: field, type, description)
- **Sample Request/Response** (JSON)
- **Rate Limit Weight** (if applicable)

APIs covered:
- Place Order (limit, market, etc.)
- Test New Order (if available)
- Query Order
- Cancel Order
- Cancel All Open Orders
- Open Orders
- Order History / All Orders
- Batch Orders (if available)
- Order Types, Time In Force, Order Status enums
- Self Trade Prevention modes

## Progress

### REST API Documentation

| Provider | General Doc | Market Doc | Status |
|----------|-------------|------------|--------|
| Binance  | [x] | [x] | done |
| Bitget   | [x] | [x] | done |
| CoinGecko | [x] | [x] | done |
| WhiteBit | [x] | [x] | done |
| Crypto.com | [x] | [x] | done |
| MEXC     | [x] | [x] | done |

### WebSocket Documentation (Spot Markets)

| Provider | WS Market Doc | Status | Notes |
|----------|---------------|--------|-------|
| Binance  | [x] | done | Raw WS, combined streams, 8 stream types |
| Bitget   | [x] | done | V2 endpoint, 6 channels, CRC32 checksum for books |
| CoinGecko | [x] | done | ActionCable protocol, Beta, paid-only (Analyst+) |
| WhiteBit | [x] | done | JSON-RPC, 6 stream types |
| Crypto.com | [x] | done | JSON-RPC 2.0, 4 stream types |
| MEXC     | [x] | done | Protobuf format, 7 stream types, 30 subs/connection |

### WebSocket Documentation (Futures Markets)

| Provider | WS Futures Doc | Status | Notes |
|----------|----------------|--------|-------|
| Binance  | [x] | done | Separate USDT-M/COIN-M endpoints, 16+ streams, mark price, liquidations, composite index |
| Bitget   | [x] | done | Unified endpoint, 8 channels, mark price, funding rate, open interest, liquidations |
| MEXC     | [x] | done | Gzip compressed, 8 channels, fair price, index price, funding rate |
| CoinGecko | N/A | n/a | Spot-only aggregator, no futures support |
| WhiteBit | [x] | done | Unified WebSocket endpoint — same streams as spot, perpetual symbols use `_PERP` suffix (e.g., `BTC_PERP`) |
| Crypto.com | [x] | done | Unified WebSocket endpoint — same streams as spot, perpetual symbols use `-PERP` suffix (e.g., `BTCUSD-PERP`) |

### Account REST API Documentation

| Provider | Accounts Doc | Status | Notes |
|----------|--------------|--------|-------|
| Binance  | [x] | done | Spot account, subaccount CRUD, universal transfers, futures account/balance |
| Bitget   | [x] | done | Account assets, bills, subaccount CRUD, transfer V2, sub-transfer |
| WhiteBit | [x] | done | Main balance, subaccount CRUD, internal transfers, transfer history |
| Crypto.com | [x] | done | User balance, balance history, accounts, positions, subaccount balance/transfer |
| MEXC     | [x] | done | Account info, subaccount CRUD, universal transfers, internal transfers |
| CoinGecko | N/A | n/a | Aggregator only, no account APIs |

### Spot Order REST API Documentation

| Provider | Spot Order Doc | Status | Notes |
|----------|----------------|--------|-------|
| Binance  | [x] | done | Unified endpoint, 7 order types, ACK/RESULT/FULL response modes, STP, pegged orders |
| Bitget   | [x] | done | Unified endpoint, limit/market, batch place/cancel (50 orders), lowercase enums |
| WhiteBit | [x] | done | Separate endpoints per order type (limit/market/stop-limit/stop-market), bulk orders (20), HMAC-SHA512 |
| Crypto.com | [x] | done | JSON-RPC envelope in body, limit/market, async operations, `BASE_QUOTE` symbol format |
| MEXC     | [x] | done | Binance-compatible API, 5 order types, batch orders (20), HMAC-SHA256 |
| CoinGecko | N/A | n/a | Aggregator only, no order APIs |

## Provider Capability Summary

Based on `pkg/provider/` implementation:

| Provider | ID | Markets | Capabilities |
|----------|-----|---------|--------------|
| Binance | `binance` | spot, margin, futures | Price, Candles, Ticker24h, OrderBook, StreamPrice, StreamOrderBook (spot+futures) |
| Bitget | `bitget` | spot, futures | Price, Candles, Ticker24h, OrderBook, StreamPrice, StreamOrderBook |
| CoinGecko | `coingecko` | spot (aggregator) | Price, Ticker24h, StreamPrice |
| WhiteBit | `whitebit` | spot | Price, Candles, Ticker24h, OrderBook |
| Crypto.com | `cryptocom` | spot | Price, Candles, Ticker24h, OrderBook |
| MEXC | `mexc` | spot, futures | Price, Candles, Ticker24h, OrderBook, StreamPrice, StreamOrderBook |

## Source References

- **Binance**: Uses `adshao/go-binance/v2` SDK. API docs: https://binance-docs.github.io/apidocs/
- **Bitget**: Manual implementation. API docs: https://www.bitget.com/api-doc/common/intro
- **CoinGecko**: Manual implementation. API docs: https://docs.coingecko.com/reference/introduction
- **WhiteBit**: Manual implementation. API docs: https://docs.zondacrypto.exchange/reference/introduction
- **Crypto.com**: Manual implementation. API docs: https://exchange-docs.crypto.com
- **MEXC**: Manual implementation. API docs: https://www.mexc.com/api-docs/spot-v3/introduction

## Notes

- This documentation is formalized from both the bits provider implementation (`pkg/provider/`) and the official exchange API documentation.
- All market data endpoints are public (no auth required) except where noted.
- Symbol formats vary by provider (e.g., MEXC futures uses underscore separator `BTC_USDT`, spot uses `BTCUSDT`).
- WebSocket stream naming conventions differ per provider.
- **Unified WebSocket pattern**: WhiteBit and Crypto.com use a single WebSocket endpoint for both spot and perpetual futures — the same stream methods/channels work for both, differentiated only by symbol format:
  - WhiteBit: Spot `BTC_USDT` → Perpetual `BTC_PERP`
  - Crypto.com: Spot `BTC_USDT` → Perpetual `BTCUSD-PERP`
- **Order API patterns**:
  - **Binance/MEXC**: Unified `POST /order` endpoint with `type` parameter; uppercase enums (`BUY`, `SELL`, `LIMIT`, `MARKET`)
  - **Bitget**: Unified endpoint; lowercase enums (`buy`, `sell`, `limit`, `market`)
  - **WhiteBit**: Separate endpoints per order type (`/order/new`, `/order/market`, `/order/stop_limit`, `/order/stop_market`); underscore symbol format (`BTC_USDT`)
  - **Crypto.com**: JSON-RPC envelope in request body; `BASE_QUOTE` symbol format; async operations
  - **CoinGecko**: No order APIs (price aggregator only)
