# CoinGecko API Documentation

## Overview

- **Official site**: https://www.coingecko.com
- **API Docs**: https://docs.coingecko.com/reference/introduction
- **Provider ID**: `coingecko`
- **Implementation**: Manual HTTP client (no external SDK)
- **Type**: Price aggregator (uses coin IDs, not trading symbols)

## Base URLs

| Environment | URL |
|-------------|-----|
| Pro API | `https://pro-api.coingecko.com/api/v3` |
| Demo API | `https://api.coingecko.com/api/v3` |

## Authentication

### API Key Header

```
x-cg-pro-api-key: <your-api-key>    # Pro API
x-cg-demo-api-key: <your-api-key>   # Demo API
```

### Configuration

Set via `coingecko.api_key` in config or `BITS_COINGECKO_API_KEY` env var.
Free demo keys available at coingecko.com/api.

## Rate Limits

| Tier | Requests/minute |
|------|-----------------|
| Free (Demo) | ~10-30 |
| Pro | Higher (varies by plan) |

## Products

| Product | Market Type | Supported |
|---------|-------------|-----------|
| Aggregator (Spot) | `spot` | Yes |
| Futures | N/A | No |
| Margin | N/A | No |

## Notes

- CoinGecko uses **coin IDs** (e.g. `bitcoin`, `ethereum`) not trading symbols
- Default currency is `usd`
- No order book or depth data (aggregator only)
- No dedicated server time endpoint

---

## Exchange Info APIs

### Ping (Health Check)

**Endpoint**: `GET /ping`

**Parameters**: None

**Response**:
```json
{
  "gecko_says": "(V3) To the Moon!"
}
```

### Coin List

Returns all supported coin IDs, symbols, and names.

**Endpoint**: `GET /coins/list`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| include_platform | boolean | No | Include platform contract addresses |

**Response**:
```json
[
  {"id": "bitcoin", "symbol": "btc", "name": "Bitcoin"},
  {"id": "ethereum", "symbol": "eth", "name": "Ethereum"}
]
```

### Supported VS Currencies

**Endpoint**: `GET /simple/supported_vs_currencies`

**Response**:
```json
["usd", "eur", "gbp", "jpy", "btc", ...]
```

### Asset Platforms

Returns blockchain platforms with contract address support.

**Endpoint**: `GET /asset_platforms`

### Search

**Endpoint**: `GET /search`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| query | string | Yes | Search term |

### Trending

**Endpoint**: `GET /search/trending`

---

## Price APIs

### Simple Price

Returns current prices for coin IDs.

**Endpoint**: `GET /simple/price`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| ids | string | Yes | Comma-separated coin IDs |
| vs_currencies | string | Yes | Comma-separated currencies |
| include_24hr_change | boolean | No | Include 24h change |
| include_24hr_vol | boolean | No | Include 24h volume |
| include_market_cap | boolean | No | Include market cap |
| include_last_updated_at | boolean | No | Include timestamp |

**Response**:
```json
{
  "bitcoin": {
    "usd": 50000,
    "usd_24h_change": 2.5
  },
  "ethereum": {
    "usd": 3000,
    "usd_24h_change": 1.8
  }
}
```

### Token Price (Multi-chain)

**Endpoint**: `GET /simple/token_price/{platform_id}`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| platform_id | string | Yes | Platform ID (e.g. `ethereum`) |
| contract_addresses | string | Yes | Comma-separated contract addresses |
| vs_currencies | string | Yes | Target currencies |

---

## Candles / OHLC

### Coin OHLC

Returns OHLCV data for a coin.

**Endpoint**: `GET /coins/{id}/ohlc`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Coin ID (e.g. `bitcoin`) |
| vs_currency | string | Yes | Target currency |
| days | string | Yes | `1`, `7`, `14`, `30`, `90`, `180`, `365`, `max` |
| interval | string | No | `daily` or `hourly` |

**Response** (array of arrays):
```json
[
  [1690196400000, 50000, 51000, 49000, 50500]
]
```

**Column order**: `[timestamp_ms, open, high, low, close]`

**Notes**: CoinGecko OHLC does not include volume data.

### Market Chart

**Endpoint**: `GET /coins/{id}/market_chart`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Coin ID |
| vs_currency | string | Yes | Target currency |
| days | string | Yes | Number of days |
| interval | string | No | Data interval |

**Response**:
```json
{
  "prices": [[timestamp, price], ...],
  "market_caps": [[timestamp, market_cap], ...],
  "total_volumes": [[timestamp, volume], ...]
}
```

---

## Market Data

### Coin Markets (Paginated)

Returns coins with market data, sorted by market cap.

**Endpoint**: `GET /coins/markets`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| vs_currency | string | Yes | Target currency |
| ids | string | No | Comma-separated coin IDs |
| order | string | No | `market_cap_asc`, `market_cap_desc`, `volume_asc`, `volume_desc` |
| per_page | int | No | 1-250, default 100 |
| page | int | No | Page number |
| sparkline | boolean | No | Include 7-day sparkline |
| price_change_percentage | string | No | `1h,24h,7d,14d,30d,200d,1y` |
| category | string | No | Filter by category |

**Response**:
```json
[
  {
    "id": "bitcoin",
    "symbol": "btc",
    "name": "Bitcoin",
    "current_price": 50000,
    "market_cap": 1000000000000,
    "market_cap_rank": 1,
    "total_volume": 50000000000,
    "high_24h": 51000,
    "low_24h": 49000,
    "price_change_percentage_24h": 2.0,
    "circulating_supply": 19000000,
    "total_supply": 21000000,
    "max_supply": 21000000,
    "last_updated": "2023-10-24T12:00:00.000Z"
  }
]
```

### Coin Details

**Endpoint**: `GET /coins/{id}`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Coin ID |
| localization | boolean | No | Include localized data |
| tickers | boolean | No | Include exchange tickers |
| market_data | boolean | No | Include market data |
| community_data | boolean | No | Include community data |
| developer_data | boolean | No | Include developer data |
| sparkline | boolean | No | Include 7-day sparkline |

---

## WebSocket Streaming

*Streaming endpoints exist but are not documented here.*
