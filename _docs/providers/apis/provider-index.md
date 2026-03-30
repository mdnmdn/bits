# Provider API Documentation Index

Master index for the formalized exchange provider API documentation.

## Target Structure

```
_docs/providers/apis/
├── provider-index.md          # This file
├── binance/
│   ├── binance-general.md     # Base URL, rate limits, auth, exchange info APIs
│   ├── binance-market.md      # Prices, orderbooks, candles, tickers, streams
│   └── binance-market-ws.md   # WebSocket spot market streams
├── bitget/
│   ├── bitget-general.md
│   ├── bitget-market.md
│   └── bitget-market-ws.md    # WebSocket spot market streams (pending)
├── coingecko/
│   ├── coingecko-general.md
│   ├── coingecko-market.md
│   └── coingecko-market-ws.md # WebSocket spot market streams
├── whitebit/
│   ├── whitebit-general.md
│   ├── whitebit-market.md
│   └── whitebit-market-ws.md  # WebSocket spot market streams
├── cryptocom/
│   ├── cryptocom-general.md
│   ├── cryptocom-market.md
│   └── cryptocom-market-ws.md # WebSocket spot market streams
└── mexc/
    ├── mexc-general.md
    ├── mexc-market.md
    └── mexc-market-ws.md      # WebSocket spot market streams
```

## Document Template

Each provider follows this structure:

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
