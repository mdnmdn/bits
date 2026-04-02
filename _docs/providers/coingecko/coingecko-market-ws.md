# CoinGecko WebSocket Market Documentation (Spot)

## Reference
- **Official API Docs**: https://docs.coingecko.com/reference/introduction
- **WebSocket Docs**: https://docs.coingecko.com/websocket
- **CGSimplePrice Channel**: https://docs.coingecko.com/websocket/cgsimpleprice
- **AsyncAPI Spec**: https://docs.coingecko.com/websocket/asyncapi.json
- **WebSocket Base URL**: `wss://stream.coingecko.com/v1`

## Protocol Overview
- **Protocol type**: ActionCable-style WebSocket (subscribe/message/unsubscribe commands with JSON-string identifiers)
- **Connection limits**: 10 concurrent socket connections (Analyst plan & above)
- **Keep-alive / ping-pong mechanism**: Server sends a **ping every 10 seconds**. Client must respond with pong within **20 seconds** or the connection is automatically terminated. Most WebSocket libraries handle this automatically.
- **Reconnection guidelines**: Implement exponential backoff for reconnection attempts. Planned disconnections occur during deployments/reboots.
- **Authentication**: Requires CoinGecko Pro API key (paid plan: Analyst, Lite, Pro, or Pro+). Pass via query parameter `x_cg_pro_api_key=YOUR_KEY` or header `x-cg-pro-api-key: YOUR_KEY`.

## Message Format

### Request Format
All requests use a uniform ActionCable-style envelope:

```json
{
  "command": "<subscribe|message|unsubscribe>",
  "identifier": "{\"channel\":\"<ChannelName>\"}",
  "data": "<JSON string with channel-specific payload>"
}
```

- `command`: One of `subscribe`, `message`, or `unsubscribe`
- `identifier`: JSON-encoded string containing the channel name
- `data`: JSON-encoded string with the subscription/streaming parameters (required for `message` commands)

### Response Format
Responses vary by type:

**Subscription confirmation:**
```json
{"type": "confirm_subscription", "identifier": "{\"channel\":\"CGSimplePrice\"}"}
```

**Subscription success/error:**
```json
{"code": 2000, "message": "Subscription is successful for bitcoin"}
```

**Data payload:** Compact field names (see field tables below). Keys may arrive in random order.

### Error Format
Errors are returned as status messages with a code and description:

```json
{"code": 2000, "message": "Subscription is successful for bitcoin"}
```

Note: CoinGecko uses code 2000 for both success and informational messages. Check the `message` field for context.

## WebSocket Endpoints

### Price Stream (CGSimplePrice)

- **Description**: Real-time price updates for coins as seen on CoinGecko.com. Uses CoinGecko coin IDs (e.g., "bitcoin", "ethereum").
- **Channel/Stream Name**: `CGSimplePrice` (channel code: `C1`)
- **Update Frequency**: As fast as ~10 seconds for large-cap and actively traded coins
- **Official Docs**: https://docs.coingecko.com/websocket/cgsimpleprice

#### Subscription Flow

**Step 1: Establish connection**
```
wss://stream.coingecko.com/v1?x_cg_pro_api_key=YOUR_KEY
```

**Step 2: Subscribe to channel**
```json
{"command":"subscribe","identifier":"{\"channel\":\"CGSimplePrice\"}"}
```

**Step 3: Subscribe to specific tokens**
```json
{"command":"message","identifier":"{\"channel\":\"CGSimplePrice\"}","data":"{\"coin_id\":[\"bitcoin\",\"ethereum\"],\"vs_currencies\":[\"usd\",\"eur\"],\"action\":\"set_tokens\"}"}
```

#### Subscription Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `coin_id` | string[] | Yes | Array of CoinGecko coin IDs (e.g., `["bitcoin", "ethereum"]`) |
| `vs_currencies` | string[] | No | Target currencies for price data. Defaults to `["usd"]`. See [/simple/supported_vs_currencies](https://docs.coingecko.com/reference/simple-supported-currencies) |
| `action` | string | Yes | Must be `"set_tokens"` to subscribe |

#### Response Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `c` | string | Channel type identifier | `"C1"` |
| `i` | string | CoinGecko coin ID | `"ethereum"` |
| `vs` | string | Target currency (vs_currency) | `"usd"` |
| `p` | float | Current token price in vs_currency | `2591.080889351465` |
| `pp` | float | 24h price change percentage | `1.3763793110454519` |
| `m` | float | Market cap in vs_currency | `312938652962.8005` |
| `v` | float | 24h trading volume in vs_currency | `20460612214.801384` |
| `t` | float | Last updated timestamp (UNIX with milliseconds) | `1747808150.269067` |

#### Sample Response

```json
{
    "c": "C1",
    "i": "ethereum",
    "vs": "usd",
    "m": 312938652962.8005,
    "p": 2591.080889351465,
    "pp": 1.3763793110454519,
    "t": 1747808150.269067,
    "v": 20460612214.801384
}
```

#### Unsubscribe

**Unsubscribe from specific token:**
```json
{"command":"message","identifier":"{\"channel\":\"CGSimplePrice\"}","data":"{\"coin_id\":[\"ethereum\"],\"action\":\"unset_tokens\"}"}
```

**Unsubscribe from channel entirely:**
```json
{"command":"unsubscribe","identifier":"{\"channel\":\"CGSimplePrice\"}"}
```

### 24h Ticker / Market Data Stream

CoinGecko WebSocket does **not** have a dedicated 24h ticker stream channel for spot markets. The `CGSimplePrice` channel (C1) includes 24h market data fields (`pp`, `m`, `v`) in every price update, effectively providing ticker data alongside prices.

For onchain/DEX pool data (GeckoTerminal), the following channels are available but are **not** spot market streams:

| Channel | Code | Description | Docs |
|---------|------|-------------|------|
| `OnchainSimpleTokenPrice` | G1 | Onchain token prices by network + contract address | [Docs](https://docs.coingecko.com/websocket/onchainsimpletokenprice) |
| `OnchainTrade` | G2 | Real-time trade/swap transactions for DEX pools | [Docs](https://docs.coingecko.com/websocket/wss-onchain-trade) |
| `OnchainOHLCV` | G3 | Real-time OHLCV candle data for DEX pools | [Docs](https://docs.coingecko.com/websocket/wssonchainohlcv) |

## Constraints & Limits

| Constraint | Value | Notes |
|------------|-------|-------|
| **Max concurrent connections** | 10 | Per account (Analyst plan & above) |
| **Max subscriptions per channel** | 100 | Per socket, per channel |
| **Credit charge** | 0.1 credit per response | Deducted from monthly API plan credits |
| **Channel access** | All 4 channels | C1, G1, G2, G3 |
| **Update frequency** | ~10s | For large-cap and actively traded coins |
| **Plan requirement** | Analyst plan & above | Free tier has no WebSocket access |

### API Tier Details

| Plan | WebSocket Access | Notes |
|------|-----------------|-------|
| Free / Demo | No | REST API only |
| Analyst | Yes | Standard limits (10 connections, 100 subs, 0.1 credit/response) |
| Lite | Yes | Standard limits |
| Pro | Yes | Standard limits |
| Pro+ | Yes | Standard limits |
| Enterprise | Yes | Higher limits available via Customer Success Manager |

## Notes

- **CoinGecko WebSocket is in Beta**: Features and limits may change. Provide feedback via [survey form](https://forms.gle/gNE1Txc9FCV55s7ZA) or email soonaik@coingecko.com.
- **Coin ID format**: Uses CoinGecko API IDs (e.g., `"bitcoin"`, `"ethereum"`, `"usd-coin"`), **not** trading symbols. Obtain IDs via [/coins/list](https://docs.coingecko.com/reference/coins-list) endpoint or from the coin's page on CoinGecko.com.
- **Null values**: Fields may return `null` when data is unavailable. Ensure your application handles null gracefully.
- **Random key order**: Response JSON keys may arrive in any order. Do not rely on key ordering.
- **vs_currencies**: Defaults to `["usd"]` if not specified in subscription data.
- **Enterprise clients**: Contact your Customer Success Manager for higher limits (more connections, more subscriptions, lower credit charges).
- **No spot exchange-level streams**: CoinGecko WebSocket provides aggregated market data, not per-exchange order book or trade streams. For exchange-level data, use exchange-specific WebSocket APIs (Binance, Bitget, etc.).
- **Current implementation note**: The existing `bits` codebase (`pkg/provider/coingecko/stream.go`) uses a legacy ActionCable client (`internal/ws/client.go`). The new official WebSocket API at `stream.coingecko.com` replaces the older endpoint and should be migrated to for production use.
