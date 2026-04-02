# Bitget Account REST API Documentation

## Reference
- **Official API Docs**: https://www.bitget.com/api-doc/spot/account/Get-Account-Assets
- **Legacy Docs**: https://bitgetlimited.github.io/apidoc/en/spot/
- **Base URL**: `https://api.bitget.com`

## Authentication

All account endpoints require authentication via the following HTTP headers:

| Header | Description |
|--------|-------------|
| `ACCESS-KEY` | Your API Key |
| `ACCESS-SIGN` | HMAC-SHA256 signature (Base64 encoded) |
| `ACCESS-TIMESTAMP` | Millisecond timestamp of request (must be within 30s of server time) |
| `ACCESS-PASSPHRASE` | Passphrase set when creating the API Key |
| `Content-Type` | `application/json` |
| `locale` | Language, e.g. `en-US` |

### Signature Method

The `ACCESS-SIGN` value is computed as:

```
preHash = timestamp + method.toUpperCase() + requestPath + "?" + queryString + body
signature = Base64(HMAC-SHA256(preHash, secretKey))
```

- For GET requests with no body: `timestamp + method + requestPath + "?" + queryString`
- For POST requests: `timestamp + method + requestPath + body`

Bitget also supports RSA signatures as an alternative to HMAC-SHA256.

### Timestamp Requirements

The `ACCESS-TIMESTAMP` must be within **30 seconds** of the API server time. Use `/api/spot/v1/public/time` to check server time and adjust for clock skew.

### API Key Permissions

Account endpoints require the following API Key permissions:
- **Read-Only**: For querying account assets and bills
- **Transfer**: For internal transfers between accounts
- **Subaccount Management**: For creating/managing virtual sub-accounts

---

## Account APIs

### Get Account Information

- **Description**: Retrieve the authenticated account's profile information including user ID, permissions, and affiliate details.
- **Endpoint**: `GET /api/v2/spot/account/info`
- **Rate Limit**: 1 request/second per User ID
- **Docs**: [Get Account Information](https://www.bitget.com/api-doc/spot/account/Get-Account-Info)

#### Parameters

None required.

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `userId` | String | User ID |
| `inviterId` | String | Inviter's user ID |
| `ips` | String | IP whitelist |
| `authorities` | Array | Permissions (e.g. `coor`, `cpor`, `stor`, `wtor`) |
| `parentId` | Int | Main account user ID (for sub-accounts) |
| `traderType` | String | `trader` or `not_trader` |
| `channelCode` | String | Affiliate referral code |
| `channel` | String | Affiliate channel |
| `regisTime` | String | Registration time (ms timestamp) |

#### Sample Response

```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1695808949356,
  "data": {
    "userId": "**********",
    "inviterId": "**********",
    "ips": "127.0.0.1",
    "authorities": ["cpor", "coor"],
    "parentId": 1,
    "traderType": "trader",
    "channelCode": "XXX",
    "channel": "YYY",
    "regisTime": "1246566789345"
  }
}
```

---

### Get Account Assets

- **Description**: Retrieve spot account asset balances for all coins or a specific coin.
- **Endpoint**: `GET /api/v2/spot/account/assets`
- **Rate Limit**: 10 requests/second per User ID
- **Docs**: [Get Account Assets](https://www.bitget.com/api-doc/spot/account/Get-Account-Assets)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `coin` | String | No | Token name (e.g. `USDT`). Queries single coin position. |
| `assetType` | String | No | `hold_only` (default): coins with balance > 0. `all`: all coins. |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `coin` | String | Token name |
| `available` | String | Available balance |
| `frozen` | String | Frozen amount (open orders, Launchpad) |
| `locked` | String | Locked amount (fiat merchant deposits, etc.) |
| `limitAvailable` | String | Restricted availability (spot copy trading) |
| `uTime` | String | Last update time (ms timestamp) |

#### Sample Response

```json
{
  "code": "00000",
  "message": "success",
  "requestTime": 1695808949356,
  "data": [
    {
      "coin": "usdt",
      "available": "1000.50",
      "frozen": "50.00",
      "locked": "0",
      "limitAvailable": "0",
      "uTime": "1622697148000"
    },
    {
      "coin": "btc",
      "available": "0.5",
      "frozen": "0.1",
      "locked": "0",
      "limitAvailable": "0",
      "uTime": "1622697148000"
    }
  ]
}
```

---

### Get Account Assets Lite

- **Description**: Lightweight version of Get Account Assets. Returns only essential balance fields (no `limitAvailable` or `locked`).
- **Endpoint**: `GET /api/spot/v1/account/assets-lite`
- **Rate Limit**: 10 requests/second per User ID
- **Docs**: [Get Account Assets Lite](https://bitgetlimited.github.io/apidoc/en/spot/#get-account-assets-lite)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `coin` | String | No | Coin name. If omitted, returns only coins with assets > 0. |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `coinId` | String | Coin ID |
| `coinName` | String | Coin name |
| `coinDisplayName` | String | Coin display name |
| `available` | String | Available balance |
| `frozen` | String | Frozen amount |
| `lock` | String | Locked amount |
| `uTime` | String | Update time (ms timestamp) |

#### Sample Response

```json
{
  "code": "00000",
  "message": "success",
  "data": [
    {
      "coinId": "10012",
      "coinName": "usdt",
      "coinDisplayName": "usdt",
      "available": "1000.50",
      "frozen": "0",
      "lock": "0",
      "uTime": "1622697148"
    }
  ]
}
```

---

### Get Bills (Transaction History)

- **Description**: Retrieve spot account billing history (deposits, withdrawals, trades, transfers, etc.).
- **Endpoint**: `GET /api/v2/spot/account/bills`
- **Rate Limit**: 10 requests/second per User ID
- **Docs**: [Get Account Bills](https://www.bitget.com/api-doc/spot/account/Get-Account-Bills)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `coin` | String | No | Token name (e.g. `USDT`) |
| `groupType` | String | No | Billing category: `deposit`, `withdraw`, `transaction`, `transfer`, `loan`, `financial`, `fait`, `convert`, `c2c`, `pre_c2c`, `on_chain`, `strategy`, `other` |
| `businessType` | String | No | Business type: `DEPOSIT`, `WITHDRAW`, `BUY`, `SELL`, `DEDUCTION_HANDLING_FEE`, `TRANSFER_IN`, `TRANSFER_OUT`, `REBATE_REWARDS`, `AIRDROP_REWARDS`, `USDT_CONTRACT_REWARDS`, `MIX_CONTRACT_REWARDS`, `SYSTEM_LOCK`, `USER_LOCK`, `STRATEGY_TRANSFER_IN` |
| `startTime` | String | No | Start time (ms timestamp) |
| `endTime` | String | No | End time (ms timestamp). Max 90-day range. |
| `limit` | String | No | Results per page. Default 100, max 500. |
| `idLessThan` | String | No | Pagination cursor (billId for older data) |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `cTime` | String | Creation time (ms timestamp) |
| `coin` | String | Token name |
| `groupType` | String | Billing category |
| `businessType` | String | Business type |
| `size` | String | Transaction amount |
| `balance` | String | Balance after transaction |
| `fees` | String | Transaction fees |
| `billId` | String | Billing ID |

#### Sample Response

```json
{
  "code": "00000",
  "message": "success",
  "requestTime": 1695808949356,
  "data": [
    {
      "cTime": "1622697148000",
      "coin": "usdt",
      "groupType": "transfer",
      "businessType": "TRANSFER_IN",
      "size": "100",
      "balance": "1100.50",
      "fees": "0",
      "billId": "1291"
    }
  ]
}
```

---

## Subaccount APIs

### Create Virtual Sub Account

- **Description**: Create one or more virtual sub-accounts under the main account. Sub-account names must be unique 8-character English strings.
- **Endpoint**: `POST /api/user/v1/sub/virtual-create`
- **Rate Limit**: 5 requests/second per User ID
- **Docs**: [Create Virtual Sub Account](https://bitgetlimited.github.io/apidoc/en/spot/#create-virtual-sub-account)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `subName` | Array | Yes | List of virtual nicknames (8-character English letters) |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `failAccounts` | Array | List of failed account names |
| `successAccounts` | Array | List of successfully created accounts |
| `successAccounts[].subUid` | String | Virtual sub-account UID |
| `successAccounts[].subName` | String | Virtual email address |
| `successAccounts[].status` | String | Account status: `normal`, `freeze` |
| `successAccounts[].auth` | Array | Permissions: `readonly`, `spot_trade`, `contract_trade` |
| `successAccounts[].remark` | String | Account label |
| `successAccounts[].cTime` | String | Creation time (ms timestamp) |

#### Sample Response

```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1682660169412,
  "data": {
    "failAccounts": ["duplicate@virtual-bitget.com"],
    "successAccounts": [
      {
        "subUid": "9837924274",
        "subName": "trading01@virtual-bitget.com",
        "status": "normal",
        "auth": ["contract_trade", "spot_trade"],
        "remark": null,
        "cTime": "1682660169573"
      }
    ]
  }
}
```

---

### Get Virtual Account List

- **Description**: Retrieve a paginated list of all virtual sub-accounts under the main account.
- **Endpoint**: `GET /api/user/v1/sub/virtual-list`
- **Rate Limit**: 10 requests/second per User ID
- **Docs**: [Get Virtual Account List](https://bitgetlimited.github.io/apidoc/en/spot/#get-virtual-account-list)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `pageSize` | String | No | Results per page. Default 20, max 100. |
| `pageNo` | String | No | Page number. Default 1. |
| `status` | String | No | Filter by status: `normal`, `freeze` |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `hasNextPage` | Boolean | Whether more pages exist |
| `lastEndId` | Int | Cursor for next page |
| `list` | Array | List of sub-accounts |
| `list[].subUid` | String | Virtual sub-account UID |
| `list[].subName` | String | Virtual email |
| `list[].status` | String | Status: `normal`, `freeze` |
| `list[].auth` | Array | Permissions |
| `list[].remark` | String | Account label |
| `list[].cTime` | String | Creation time (ms timestamp) |

#### Sample Response

```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1656589586807,
  "data": {
    "hasNextPage": false,
    "lastEndId": 51,
    "list": [
      {
        "subUid": "7713789662",
        "subName": "mySub01@bgbroker6314497154",
        "status": "normal",
        "auth": ["readonly", "spot_trade", "contract_trade"],
        "remark": "mySub01",
        "cTime": "1653287983475"
      }
    ]
  }
}
```

---

### Get Sub-Account Spot Assets (V2)

- **Description**: Retrieve spot asset balances for all sub-accounts (only returns sub-accounts with assets > 0). **ND Brokers cannot call this endpoint.**
- **Endpoint**: `GET /api/v2/spot/account/subaccount-assets`
- **Rate Limit**: 10 requests/second per User ID
- **Docs**: [Get Sub-accounts Assets](https://www.bitget.com/api-doc/spot/account/Get-Subaccount-Assets)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `idLessThan` | String | No | Cursor ID for pagination. Omit on first request. |
| `limit` | String | No | Results per page. Default 10, max 50. |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | String | Cursor ID |
| `userId` | String | Sub-account user ID |
| `assetsList` | Array | List of spot assets |
| `assetsList[].coin` | String | Token name |
| `assetsList[].available` | String | Available balance |
| `assetsList[].limitAvailable` | String | Restricted availability (copy trading) |
| `assetsList[].frozen` | String | Frozen amount |
| `assetsList[].locked` | String | Locked amount |
| `assetsList[].uTime` | String | Update time (ms timestamp) |

#### Sample Response

```json
{
  "code": "00000",
  "message": "success",
  "requestTime": 1695808949356,
  "data": [
    {
      "id": 1111,
      "userId": 1234567890,
      "assetsList": [
        {
          "coin": "BTC",
          "available": "1.1",
          "limitAvailable": "12.1",
          "frozen": "0",
          "locked": "1.1",
          "uTime": "1337654897651"
        }
      ]
    },
    {
      "id": 2222,
      "userId": 1234567890,
      "assetsList": [
        {
          "coin": "ETH",
          "available": "12.1",
          "limitAvailable": "12.1",
          "frozen": "0",
          "locked": "1.1",
          "uTime": "1337654897651"
        }
      ]
    }
  ]
}
```

---

### Get Sub-Account Spot Assets (V1)

- **Description**: Legacy endpoint for retrieving sub-account spot assets. Main account only.
- **Endpoint**: `POST /api/spot/v1/account/sub-account-spot-assets`
- **Rate Limit**: 1 request per 10 seconds per User ID
- **Docs**: [Get sub Account Spot Assets](https://bitgetlimited.github.io/apidoc/en/spot/#get-sub-account-spot-assets)

#### Parameters

None required. Send empty JSON body `{}`.

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `userId` | Int | Sub-account user ID |
| `spotAssetsList` | Array | List of spot assets |
| `spotAssetsList[].coinId` | Int | Coin ID |
| `spotAssetsList[].coinName` | String | Coin name |
| `spotAssetsList[].available` | String | Available balance |
| `spotAssetsList[].frozen` | String | Frozen amount |
| `spotAssetsList[].lock` | String | Locked amount |

#### Sample Response

```json
{
  "code": "00000",
  "message": "success",
  "data": [
    {
      "userId": 9165454769,
      "spotAssetsList": [
        {
          "coinId": 1,
          "coinName": "BTC",
          "available": "1.1",
          "frozen": "0",
          "lock": "1.1"
        }
      ]
    }
  ]
}
```

---

## Transfer APIs

### Transfer (V2)

- **Description**: Transfer assets between different account types within the same user. Supports spot, futures, margin, and P2P accounts.
- **Endpoint**: `POST /api/v2/spot/wallet/transfer`
- **Rate Limit**: 10 requests/second per User ID
- **Docs**: [Transfer](https://www.bitget.com/api-doc/spot/account/Wallet-Transfer)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `fromType` | String | Yes | Source account: `spot`, `p2p`, `coin_futures`, `usdt_futures`, `usdc_futures`, `crossed_margin`, `isolated_margin` |
| `toType` | String | Yes | Destination account (same options as `fromType`) |
| `amount` | String | Yes | Transfer amount |
| `coin` | String | Yes | Currency to transfer |
| `symbol` | String | Yes | Required for isolated margin transfers |
| `clientOid` | String | No | Custom order ID (unique; duplicates return existing result) |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `transferId` | String | Transfer ID |
| `clientOid` | String | Custom order ID |

#### Sample Response

```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1683875302853,
  "data": {
    "transferId": "123456",
    "clientOid": "x123"
  }
}
```

---

### Transfer V2 (Legacy)

- **Description**: Legacy transfer endpoint using older account type naming conventions. Main account only.
- **Endpoint**: `POST /api/spot/v1/wallet/transfer-v2`
- **Rate Limit**: 5 requests/second per User ID
- **Docs**: [Transfer V2](https://bitgetlimited.github.io/apidoc/en/spot/#transfer-v2)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `fromType` | String | Yes | Source account: `spot`, `mix_usdt`, `mix_usd`, `mix_usdc`, `margin_cross`, `margin_isolated` |
| `toType` | String | Yes | Destination account (same options) |
| `amount` | String | Yes | Transfer amount |
| `coin` | String | Yes | Currency to transfer |
| `symbol` | String | No | Required when `fromType` or `toType` = `margin_isolated` |
| `clientOid` | String | No | Custom order ID |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `transferId` | String | Transfer ID |
| `clientOrderId` | String | Client order ID |

#### Sample Response

```json
{
  "code": "00000",
  "msg": "success",
  "data": {
    "transferId": "123456",
    "clientOrderId": "x123"
  }
}
```

#### Account Type Restrictions

| fromType | toType | Restriction |
|----------|--------|-------------|
| `spot` | `mix_usdt` | Only USDT |
| `spot` | `mix_usd` | Only margin coins (BTC, ETH, EOS, XRP, USDC) |
| `spot` | `mix_usdc` | Only USDC |
| `mix_usdt` | `spot` | Only USDT |
| `mix_usd` | `spot` | Only margin coins |
| `mix_usdt` | `mix_usd` | Not allowed |
| `mix_usdt` | `mix_usdc` | Not allowed |

---

### Sub Transfer

- **Description**: Transfer assets between parent and sub-accounts, or between sub-accounts under the same parent. Also supports internal transfers within a sub-account (e.g., spot to futures).
- **Endpoint**: `POST /api/v2/spot/wallet/subaccount-transfer`
- **Rate Limit**: 20 requests/second per User ID
- **Docs**: [Sub Transfer](https://www.bitget.com/api-doc/spot/account/Sub-Transfer)

**Important**: Only the **parent account API Key** can use this endpoint, and the API Key **must have an IP whitelist bound**.

#### Supported Transfer Types

- Parent account → Sub-account
- Sub-account → Parent account
- Sub-account → Sub-account (same parent)
- Sub-account internal transfer (e.g., spot → futures; `fromUserId` == `toUserId`)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `fromUserId` | String | Yes | Source account UID |
| `toUserId` | String | Yes | Destination account UID |
| `fromType` | String | Yes | Source account type: `spot`, `p2p`, `coin_futures`, `usdt_futures`, `usdc_futures`, `crossed_margin`, `isolated_margin` |
| `toType` | String | Yes | Destination account type (same options) |
| `amount` | String | Yes | Transfer amount |
| `coin` | String | Yes | Currency to transfer |
| `symbol` | String | No | Required for isolated margin transfers |
| `clientOid` | String | No | Custom order ID |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `transferId` | String | Transfer ID |
| `clientOid` | String | Custom order ID |

#### Sample Response

```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1683875302853,
  "data": {
    "transferId": "123456",
    "clientOid": "x123"
  }
}
```

---

### Get Transfer List (V2)

- **Description**: Retrieve transfer history for the authenticated account.
- **Endpoint**: `GET /api/v2/spot/account/transferRecords`
- **Rate Limit**: 20 requests/second per User ID
- **Docs**: [Get Transfer Record](https://www.bitget.com/api-doc/spot/account/Get-Account-TransferRecords)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `coin` | String | Yes | Token name |
| `fromType` | String | No | Source account type filter |
| `startTime` | String | No | Start time (ms timestamp) |
| `endTime` | String | No | End time (ms timestamp). Max 90-day range. |
| `clientOid` | String | No | Filter by custom order ID |
| `pageNum` | String | No | **Deprecated**. Page number (max 1000) |
| `limit` | String | No | Results per page. Default 100, max 500. |
| `idLessThan` | String | No | Pagination cursor (transferId for older data) |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `coin` | String | Token name |
| `status` | String | Transfer status: `Successful`, `Failed`, `Processing` |
| `toType` | String | Destination account type |
| `toSymbol` | String | Trading pair for destination (isolated margin only) |
| `fromType` | String | Source account type |
| `fromSymbol` | String | Trading pair for source (isolated margin only) |
| `size` | String | Transfer amount |
| `ts` | String | Transfer time (ms timestamp) |
| `clientOid` | String | Custom order ID |
| `transferId` | String | Transfer order ID |

#### Sample Response

```json
{
  "code": "00000",
  "data": [
    {
      "coin": "btc",
      "status": "Successful",
      "toType": "usdt_futures",
      "toSymbol": "",
      "fromType": "spot",
      "fromSymbol": "BTC/USD",
      "size": "1000.00000000",
      "ts": "1631070374488",
      "clientOid": "1",
      "transferId": "1"
    }
  ],
  "msg": "success",
  "requestTime": 1631608142260
}
```

---

### Get Transfer List (V1 Legacy)

- **Description**: Legacy endpoint for transfer history using older account type naming.
- **Endpoint**: `GET /api/spot/v1/account/transferRecords`
- **Rate Limit**: 20 requests/second per User ID
- **Docs**: [Get Transfer List](https://bitgetlimited.github.io/apidoc/en/spot/#get-transfer-list)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `coinId` | Integer | Yes | Coin ID |
| `fromType` | String | Yes | Source account: `exchange` (spot), `usdt_mix`, `usdc_mix`, `usd_mix`, `margin_cross`, `margin_isolated` |
| `before` | String | Yes | Start time (seconds timestamp) |
| `after` | String | Yes | End time (seconds timestamp) |
| `clientOid` | String | No | Filter by custom order ID |
| `limit` | Integer | No | Results per page. Default 100, max 500. |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `coinName` | String | Coin name |
| `status` | String | Transfer status: `Successful`, `Failed`, `Processing` |
| `toType` | String | Destination account type |
| `toSymbol` | String | Destination symbol |
| `fromType` | String | Source account type |
| `fromSymbol` | String | Source symbol |
| `amount` | String | Transfer amount |
| `tradeTime` | String | Transfer time (ms timestamp) |
| `clientOid` | String | Custom order ID |
| `transferId` | String | Transfer ID |

#### Sample Response

```json
{
  "code": "00000",
  "data": [
    {
      "coinName": "btc",
      "status": "Successful",
      "toType": "USD_MIX",
      "toSymbol": "",
      "fromType": "CONTRACT",
      "fromSymbol": "BTC/USD",
      "amount": "1000.00000000",
      "tradeTime": "1631070374488",
      "clientOid": "1",
      "transferId": "997381107487641600"
    }
  ],
  "msg": "success"
}
```

---

## Constraints & Limits

### Rate Limits Summary

| Endpoint | Rate Limit | Scope |
|----------|-----------|-------|
| `GET /api/v2/spot/account/info` | 1 req/s | User ID |
| `GET /api/v2/spot/account/assets` | 10 req/s | User ID |
| `GET /api/v2/spot/account/bills` | 10 req/s | User ID |
| `GET /api/v2/spot/account/subaccount-assets` | 10 req/s | User ID |
| `POST /api/spot/v1/account/assets-lite` | 10 req/s | User ID |
| `POST /api/spot/v1/account/sub-account-spot-assets` | 1 req/10s | User ID |
| `POST /api/spot/v1/account/bills` | 10 req/s | User ID |
| `POST /api/v2/spot/wallet/transfer` | 10 req/s | User ID |
| `POST /api/spot/v1/wallet/transfer-v2` | 5 req/s | User ID |
| `POST /api/v2/spot/wallet/subaccount-transfer` | 20 req/s | User ID |
| `GET /api/v2/spot/account/transferRecords` | 20 req/s | User ID |
| `GET /api/spot/v1/account/transferRecords` | 20 req/s | User ID |
| `POST /api/user/v1/sub/virtual-create` | 5 req/s | User ID |
| `GET /api/user/v1/sub/virtual-list` | 10 req/s | User ID |

### General Constraints

- **Timestamp window**: Requests must be within 30 seconds of server time
- **HTTP 429**: Returned when rate limit exceeded
- **Pagination**: Most list endpoints support cursor-based pagination via `idLessThan`
- **Time range**: Bills and transfer records queries are limited to 90-day ranges
- **All numeric values**: Returned as strings to preserve precision

### Account Types

| V2 Account Type | V1/Legacy Account Type | Description |
|-----------------|----------------------|-------------|
| `spot` | `exchange` | Spot trading account |
| `p2p` | - | P2P/Funding account |
| `usdt_futures` | `usdt_mix` | USDT-M futures account |
| `usdc_futures` | `usdc_mix` | USDC-M futures account |
| `coin_futures` | `usd_mix` | Coin-M futures account |
| `crossed_margin` | `margin_cross` | Cross margin account |
| `isolated_margin` | `margin_isolated` | Isolated margin account |

---

## Notes

### API Version Differences

- **V2 endpoints** (`/api/v2/...`) are the current recommended API. They use more descriptive account type names and return millisecond timestamps.
- **V1 endpoints** (`/api/spot/v1/...`) are legacy. Some V1 endpoints use different account type naming (`exchange` vs `spot`, `usdt_mix` vs `usdt_futures`) and some use second-based timestamps.

### Sub-Account Restrictions

- Only **main accounts** can create and manage sub-accounts
- **ND Brokers** cannot call the sub-account assets endpoint
- Sub-transfer endpoint requires the parent API Key to have an **IP whitelist** bound
- Sub-accounts can only transfer to/from their parent account or sibling sub-accounts under the same parent

### Transfer Restrictions

- Transfers between different futures account types (e.g., `mix_usdt` → `mix_usd`) are **not allowed**
- Isolated margin transfers require the `symbol` parameter
- `clientOid` must be unique per transfer; duplicate `clientOid` values return the existing transfer result

### Balance Fields

- `available`: Free balance available for trading/withdrawal
- `frozen`: Balance frozen due to open orders or Launchpad participation
- `locked`: Balance locked for fiat merchant deposits, etc.
- `limitAvailable`: Restricted availability for spot copy trading (V2 only)

### Error Response Format

All endpoints return a standard envelope:

```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1695808949356,
  "data": { ... }
}
```

- `code: "00000"` indicates success
- Non-zero `code` values indicate errors (see [Error Codes](https://www.bitget.com/api-doc/spot/error-code/restapi))
