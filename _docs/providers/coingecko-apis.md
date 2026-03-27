# CoinGecko API Documentation

Based on the official [coingecko-typescript](https://github.com/coingecko/coingecko-typescript) SDK.

## Base URLs

| Environment | URL |
|-------------|-----|
| Pro API | `https://pro-api.coingecko.com/api/v3` |
| Demo API | `https://api.coingecko.com/api/v3` |

## Authentication

### API Keys

CoinGecko supports two API tiers:

- **Pro API Key** - Higher rate limits, requires paid subscription
- **Demo API Key** - Limited rate limits, free tier

### Headers

```
x-cg-pro-api-key: <your-pro-api-key>
x-cg-demo-api-key: <your-demo-api-key>
```

### Client Initialization (TypeScript Example)

```typescript
import { Coingecko } from 'coingecko-typescript';

const client = new Coingecko({
  proAPIKey: process.env.COINGECKO_PRO_API_KEY,
  // or
  demoAPIKey: process.env.COINGECKO_DEMO_API_KEY,
  environment: 'pro',  // or 'demo' (default: 'pro')
});
```

Environment variables:
- `COINGECKO_PRO_API_KEY`
- `COINGECKO_DEMO_API_KEY`
- `COINGECKO_BASE_URL`
- `COINGECKO_LOG`

---

## Resources Overview

| Resource | Description |
|----------|-------------|
| `coins` | Coin metadata, market data, OHLC, history |
| `simple` | Simple price, token price, supported currencies |
| `exchanges` | Exchange listings and data |
| `derivatives` | Derivatives market data |
| `global` | Global market data |
| `exchangeRates` | Exchange rates |
| `assetPlatforms` | Blockchain platforms |
| `search` | Search for coins, categories |
| `ping` | API health check |

---

## Simple API

### Get Simple Price

```
GET /simple/price
```

Query Parameters:
- `vs_currencies` (required) - Target currency (e.g., "usd", "eur", "btc")
- `ids` - Coin IDs (comma-separated)
- `include_24hr_change` - Include 24h change
- `include_24hr_vol` - Include 24h volume
- `include_market_cap` - Include market cap
- `include_last_updated_at` - Include last updated timestamp

Response:
```json
{
  "bitcoin": {
    "usd": 50000,
    "usd_24h_change": 2.5,
    "usd_24h_vol": 30000000000,
    "usd_market_cap": 1000000000000,
    "last_updated_at": 1690196141
  }
}
```

### Get Token Price (Multi-chain)

```
GET /simple/token_price/{id}
```

Query Parameters:
- `contract_addresses` - Token contract addresses
- `vs_currencies` - Target currencies
- `include_24hr_change`, `include_24hr_vol`, `include_market_cap`

### Get Supported VS Currencies

```
GET /simple/supported_vs_currencies
```

Returns list of supported fiat currencies (usd, eur, gbp, etc.)

---

## Coins API

### Get Coin by ID

```
GET /coins/{id}
```

Path: `id` - CoinGecko coin ID (e.g., "bitcoin", "ethereum")

Query Parameters:
- `localization` - Include localized languages (default: true)
- `tickers` - Include exchange tickers (default: true)
- `market_data` - Include market data (default: true)
- `community_data` - Include community data (default: true)
- `developer_data` - Include developer data (default: true)
- `sparkline` - Include 7-day sparkline (default: false)

Response includes:
```json
{
  "id": "bitcoin",
  "symbol": "btc",
  "name": "Bitcoin",
  "image": { "large": "...", "small": "...", "thumb": "..." },
  "market_data": {
    "current_price": { "usd": 50000 },
    "market_cap": { "usd": 1000000000000 },
    "total_volume": { "usd": 50000000000 },
    "high_24h": { "usd": 51000 },
    "low_24h": { "usd": 49000 },
    "price_change_24h": 1000,
    "price_change_percentage_24h": 2.0,
    "ath": { "usd": 69000 },
    "ath_date": { "usd": "2021-11-10T00:00:00.000Z" },
    "atl": { "usd": 67.81 },
    "circulating_supply": 19000000,
    "total_supply": 21000000,
    "max_supply": 21000000
  },
  "tickers": [...],
  "description": { "en": "..." },
  "links": { "homepage": [...], "blockchain_site": [...] }
}
```

### Get Markets

```
GET /coins/markets
```

Query Parameters:
- `vs_currency` (required) - Target currency (e.g., "usd")
- `ids` - Coin IDs (comma-separated)
- `order` - Sort order: `market_cap_asc`, `market_cap_desc`, `volume_asc`, `volume_desc`
- `per_page` - Results per page (1-250, default: 100)
- `page` - Page number
- `sparkline` - Include 7-day sparkline
- `price_change_percentage` - Include change for timeframes: "1h,24h,7d,14d,30d,200d,1y"
- `category` - Filter by category

Response:
```json
[
  {
    "id": "bitcoin",
    "symbol": "btc",
    "name": "Bitcoin",
    "image": "https://assets.coingecko.com/...",
    "current_price": 50000,
    "market_cap": 1000000000000,
    "market_cap_rank": 1,
    "total_volume": 50000000000,
    "high_24h": 51000,
    "low_24h": 49000,
    "price_change_24h": 1000,
    "price_change_percentage_24h": 2.0,
    "market_cap_change_percentage_24h": 1.5,
    "circulating_supply": 19000000,
    "total_supply": 21000000,
    "max_supply": 21000000,
    "ath": 69000,
    "ath_change_percentage": -27.5,
    "ath_date": "2021-11-10T00:00:00.000Z",
    "atl": 67.81,
    "atl_change_percentage": 73600,
    "last_updated": "2023-10-24T12:00:00.000Z"
  }
]
```

### Get Coin List

```
GET /coins/list
```

Returns all available coins with IDs, symbols, and names.

### Get Coin OHLC

```
GET /coins/{id}/ohlc
```

Query Parameters:
- `vs_currency` (required) - Target currency
- `days` (required) - Time range: "1", "7", "14", "30", "90", "180", "365", "max"
- `interval` - "daily" or "hourly"

Response (array of [timestamp, open, high, low, close]):
```json
[
  [1690196400000, 50000, 51000, 49000, 50500],
  [1690100000000, 49500, 50500, 49000, 50000]
]
```

### Get Coin OHLC Range

```
GET /coins/{id}/ohlc/range
```

Query Parameters:
- `vs_currency` (required)
- `from` (required) - Start date (ISO or UNIX)
- `to` (required) - End date
- `interval` - "daily" or "hourly"

### Get Market Chart

```
GET /coins/{id}/market_chart
```

Query Parameters:
- `vs_currency` (required)
- `days` (required) - Number of days
- `interval` - Data interval (optional)

Response:
```json
{
  "prices": [[timestamp, price], ...],
  "market_caps": [[timestamp, market_cap], ...],
  "total_volumes": [[timestamp, volume], ...]
}
```

### Get Coin History

```
GET /coins/{id}/history
```

Query Parameters:
- `date` (required) - Date (DD-MM-YYYY)
- `localization` - Include localized data

### Get Coin Tickers

```
GET /coins/{id}/tickers
```

Query Parameters:
- `exchange_ids` - Filter by exchange
- `include_exchange_logo`
- `page`, `order`, `depth`

Returns exchange-specific ticker data.

### Get Top Gainers/Losers

```
GET/coins/top_gainers
GET/coins/top_losers
```

Query Parameters:
- `vs_currency`
- `category` - e.g., "layer-1"
- `page`
- `per_page`

### Get Coin Categories

```
GET /coins/categories
GET /coins/categories/list
```

---

## Exchanges API

### Get Exchanges

```
GET /exchanges
```

Query Parameters:
- `per_page`, `page`
- `order`

### Get Exchange by ID

```
GET /exchanges/{id}
```

### Get Exchange Tickers

```
GET /exchanges/{id}/tickers
```

### Get Exchange Volume Chart

```
GET /exchanges/{id}/volume_chart
```

Query: `days`

---

## Derivatives API

### Get Derivatives

```
GET /derivatives
```

Query Parameters:
- `include_tickers` - "all" or "expired"

Response:
```json
[
  {
    "market": "Binance Futures",
    "symbol": "BTCUSDT",
    "price": "50000",
    "price_percentage_change_24h": 2.5,
    "contract_type": "perpetual",
    "index": "50000",
    "basis": 0.5,
    "spread": 0.01,
    "funding_rate": 0.01,
    "open_interest": 1000000000,
    "volume_24h": 50000000000
  }
]
```

### Get Derivative Exchanges

```
GET /derivatives/exchanges
GET /derivatives/exchanges/{id}
```

---

## Global API

### Get Global Data

```
GET /global
```

Response:
```json
{
  "data": {
    "active_cryptocurrencies": 10000,
    "upcoming_icos": 50,
    "ongoing_icos": 100,
    "ended_icos": 5000,
    "market_cap_change_percentage_24h_usd": 2.5,
    "total_market_cap": { "usd": 2000000000000 },
    "total_volume": { "usd": 100000000000 },
    "market_cap_percentage": { "btc": 50, "eth": 20 }
  }
}
```

### Get Global DeFi Data

```
GET /global/decentralized_finance_defi
```

---

## Exchange Rates API

### Get Exchange Rates

```
GET/exchange_rates
```

Returns BTC exchange rates to other currencies.

---

## Asset Platforms API

### Get Asset Platforms

```
GET /asset_platforms
```

Returns list of blockchain platforms (Ethereum, Solana, etc.) with IDs.

---

## Search API

### Search

```
GET /search
```

Query: `query` - Search term

### Get Trending

```
GET /search/trending
```

Returns trending coins, NFTs, and categories.

---

## Ping API

### Health Check

```
GET /ping
```

Response:
```json
{
  "gecko_says": "(V3) To the Moon!"
}
```

---

## Supported Languages

For localization (`locale` parameter):
- ar, bg, cs, da, de, el, en, es, fi, fr, he, hi, hr, hu, id, it, ja, ko, lt, nl, no, pl, pt, ro, ru, sk, sl, sv, th, tr, uk, vi, zh, zh-tw

---

## Rate Limits

| Tier | Requests/minute |
|------|-----------------|
| Free (Demo) | ~10-30 |
| Pro | Higher limits (varies by plan) |

The SDK handles retries automatically with exponential backoff.

---

## Go Implementation Notes

When implementing in Go, map the TypeScript types to Go structs:

```go
// Example: Market data response
type MarketData struct {
    CurrentPrice           map[string]float64 `json:"current_price"`
    MarketCap              map[string]float64 `json:"market_cap"`
    TotalVolume            map[string]float64 `json:"total_volume"`
    PriceChange24h         float64            `json:"price_change_24h"`
    PriceChangePercentage24h float64          `json:"price_change_percentage_24h"`
    CirculatingSupply      float64            `json:"circulating_supply"`
    TotalSupply            float64            `json:"total_supply"`
    MaxSupply              *float64           `json:"max_supply"`
    Ath                    map[string]float64 `json:"ath"`
    Atl                    map[string]float64 `json:"atl"`
}
```
