# Crypto.com Provider — WIP

## Status: In Progress

## Summary

Implementing `cryptocom` as an exchange provider backed by the Crypto.com Exchange v2 REST API (`https://api.crypto.com/v2`). Only the **spot** market is supported via this API tier.

## Progress

### ✅ Completed

- [x] `internal/config/config.go` — added `CryptoComConfig` struct, added to `Config`, wired env overrides (`BITS_CRYPTOCOM_*`), `.env` map parsing, `Redacted()` masking, and `ConfigTemplate`
- [x] `internal/provider/cryptocom/client.go` — `Client` struct, `NewClient()`, `ID()`, `SetUserAgent()`, `Capabilities()`, `doRequest()`
- [x] `internal/provider/cryptocom/types.go` — internal JSON structs: `apiEnvelope`, `apiInstrumentsResult`, `apiInstrument`, `apiTickerResult`, `apiTickerData`, `apiBookResult`, `apiBookRow`
- [x] `internal/provider/cryptocom/market.go` — implements `provider.PriceProvider`, `provider.TickerProvider`, `provider.OrderBookProvider`
- [x] `internal/provider/cryptocom/exchange.go` — implements `provider.ExchangeProvider` (`ServerTime`, `ExchangeInfo`)
- [x] `internal/registry/registry.go` — registered `cryptocom` with aliases `cdc` and `cro`

### 🔲 Remaining / Future Work

- [ ] Add candle support (`public/get-candlestick`) once endpoint behavior is verified
- [ ] Add streaming support via Crypto.com WebSocket API
- [ ] Write unit tests (`client_test.go`) with `httptest` mocks
- [ ] Update `_docs/architecture.md` capability table

## API Reference

- **Base URL**: `https://api.crypto.com/v2`
- **Docs**: `_docs/providers/cryptocom-apis.md`
- **Auth**: HMAC-SHA512 (only needed for private endpoints; public endpoints used here require no auth)

## Capabilities

| Market | Feature        | Endpoint                   | Status |
|--------|----------------|----------------------------|--------|
| Spot   | server_time    | via `public/get-instruments` timing | ✅ |
| Spot   | exchange_info  | `public/get-instruments`   | ✅ |
| Spot   | price          | `public/get-ticker`        | ✅ |
| Spot   | ticker_24h     | `public/get-ticker`        | ✅ |
| Spot   | order_book     | `public/get-book`          | ✅ |
| Spot   | candles        | `public/get-candlestick`   | 🔲 future |
| Spot   | stream_price   | WebSocket API              | 🔲 future |
| Spot   | stream_orderbook | WebSocket API            | 🔲 future |

## Notes

### ServerTime Approximation

The Crypto.com v2 REST API does not expose a dedicated `public/get-time` endpoint. `ServerTime()` estimates the server time by measuring the round-trip latency of a `public/get-instruments` call and using the midpoint (`before + latency/2`). The `Latency` field is populated so callers can judge accuracy.

### OrderBook Format

The `public/get-book` response wraps depth data in a one-element `data` array:
```json
{
  "result": {
    "instrument_name": "BTC_USDT",
    "depth": 10,
    "data": [
      {
        "bids": [[price, qty, num_orders], ...],
        "asks": [[price, qty, num_orders], ...],
        "t": 1700000000000
      }
    ]
  }
}
```
The third element (`num_orders`) is ignored; only price and quantity are used.

### Ticker Fields

The ticker response uses single-letter field names:
- `i` = instrument name
- `c` = last price (used as current price)
- `o` = open price
- `p` = price change (absolute: `c - o`)
- `h` / `l` = high / low 24h
- `v` / `vv` = base / quote volume 24h
- `t` = trade count 24h (not a timestamp)

Price change percent is derived as `p / o * 100`.

## Configuration

```toml
[cryptocom]
api_key = ""       # optional; only needed for private endpoints
api_secret = ""    # optional; only needed for private endpoints
# base_url = "https://api.crypto.com/v2"

[cryptocom.spot]
enabled = true
```

Environment variable overrides:
- `BITS_CRYPTOCOM_API_KEY`
- `BITS_CRYPTOCOM_API_SECRET`
- `BITS_CRYPTOCOM_BASE_URL`
- `BITS_CRYPTOCOM_SPOT_ENABLED`

## Usage

```bash
bits price BTC_USDT -p cryptocom
bits ticker BTC_USDT -p cryptocom
bits book BTC_USDT -p cryptocom
bits info -p cryptocom
bits time -p cryptocom
bits providers
bits capabilities -p cryptocom
```
