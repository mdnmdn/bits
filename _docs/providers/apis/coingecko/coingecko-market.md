# CoinGecko Market Data APIs

## Price

### Simple Price

Returns current prices for one or more coin IDs.

**Endpoint**: `GET /simple/price`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| ids | string | Yes | Comma-separated coin IDs (e.g. `bitcoin,ethereum`) |
| vs_currencies | string | Yes | Comma-separated target currencies (e.g. `usd`) |
| include_24hr_change | boolean | No | Include 24h price change percentage |
| include_24hr_vol | boolean | No | Include 24h trading volume |
| include_market_cap | boolean | No | Include market cap |
| include_last_updated_at | boolean | No | Include last update timestamp |

**Sample Request**:
```
GET /simple/price?ids=bitcoin,ethereum&vs_currencies=usd&include_24hr_change=true
```

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

**Notes**:
- Response is a nested map: `{coin_id: {currency: value, ...}, ...}`
- Change values are percentages (not ratios)
- CoinGecko does not support bid/ask prices (aggregator model)

---

## Candles / OHLC

### OHLC Data

Returns OHLC data for a coin over a specified time range.

**Endpoint**: `GET /coins/{id}/ohlc`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | CoinGecko coin ID |
| vs_currency | string | Yes | Target currency (e.g. `usd`) |
| days | string | Yes | Time range: `1`, `7`, `14`, `30`, `90`, `180`, `365`, `max` |
| interval | string | No | `daily` or `hourly` |

**Sample Request**:
```
GET /coins/bitcoin/ohlc?vs_currency=usd&days=7
```

**Response**:
```json
[
  [1690196400000, 50000, 51000, 49000, 50500],
  [1690100000000, 49500, 50500, 49000, 50000]
]
```

**Column order**: `[timestamp_ms, open, high, low, close]`

**Notes**:
- No volume data in OHLC endpoint
- `days` parameter determines granularity (more days = coarser candles)
- For custom time ranges, use `/coins/{id}/market_chart/range`

### OHLC Range

**Endpoint**: `GET /coins/{id}/ohlc/range`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Coin ID |
| vs_currency | string | Yes | Target currency |
| from | string | Yes | Start date (UNIX or ISO) |
| to | string | Yes | End date |
| interval | string | No | `daily` or `hourly` |

---

## Market Data

### Coin Markets

Paginated list of coins with market data.

**Endpoint**: `GET /coins/markets`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| vs_currency | string | Yes | Target currency |
| ids | string | No | Comma-separated coin IDs |
| order | string | No | Sort order |
| per_page | int | No | 1-250, default 100 |
| page | int | No | Page number |
| sparkline | boolean | No | Include 7-day sparkline |
| price_change_percentage | string | No | Timeframes |
| category | string | No | Filter by category |

**Sample Request**:
```
GET /coins/markets?vs_currency=usd&order=market_cap_desc&per_page=10&page=1
```

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
    "ath": 69000,
    "ath_date": "2021-11-10T00:00:00.000Z",
    "atl": 67.81,
    "last_updated": "2023-10-24T12:00:00.000Z"
  }
]
```

---

## Global Data

### Global Market Data

**Endpoint**: `GET /global`

**Response**:
```json
{
  "data": {
    "active_cryptocurrencies": 10000,
    "market_cap_change_percentage_24h_usd": 2.5,
    "total_market_cap": {"usd": 2000000000000},
    "total_volume": {"usd": 100000000000},
    "market_cap_percentage": {"btc": 50, "eth": 20}
  }
}
```

### Global DeFi Data

**Endpoint**: `GET /global/decentralized_finance_defi`

---

## Not Supported

CoinGecko as a price aggregator does **not** provide:

- **Order Book / Depth**: No bid/ask depth data
- **Live Order Book Streaming**: Not available
- **Futures Data**: No derivatives market data
- **Server Time**: No dedicated endpoint
- **Exchange Info**: No trading pair configuration (use `/coins/list` instead)
