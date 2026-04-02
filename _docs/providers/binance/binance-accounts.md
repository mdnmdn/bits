# Binance Account REST API Documentation

## Reference

- **Official API Docs**:
  - Spot: https://developers.binance.com/docs/binance-spot-api-docs/rest-api/account-endpoints
  - USDT-M Futures: https://developers.binance.com/docs/derivatives/usds-margined-futures/account/rest-api
  - Sub-Account: https://developers.binance.com/docs/sub_account/account-management
- **Base URLs**:
  - Spot: `https://api.binance.com`
  - USDT-M: `https://fapi.binance.com`

## Authentication

All account-related endpoints are `USER_DATA` or `SIGNED` endpoints requiring authentication.

### Required Headers/Params

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `X-MBX-APIKEY` | STRING | YES | API key sent in request header |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |
| `signature` | STRING | YES | HMAC-SHA256 signature of the query string + body |
| `recvWindow` | LONG | NO | Request validity window in ms (default: 5000, max: 60000) |

### Signature Method (HMAC-SHA256)

1. Construct the payload: `param1=value1&param2=value2...`
2. Percent-encode any non-ASCII characters before signing
3. Compute HMAC-SHA256 using the `secretKey` as the signing key
4. Append the hex-encoded signature as the `signature` parameter

### Timestamp Requirements

- The server rejects requests where `timestamp` differs from `serverTime` by more than `recvWindow`
- Default `recvWindow` is 5000ms; maximum is 60000ms
- `recvWindow` supports up to 3 decimal places for microsecond precision (e.g., `6000.346`)
- Recommended: use a `recvWindow` of 5000ms or less

## Account APIs (Spot)

### Account Information

- **Description**: Get current account information including commission rates, permissions, and all asset balances.
- **Endpoint**: `GET /api/v3/account`
- **Weight**: 20
- **Security Type**: `USER_DATA`
- **Docs**: https://developers.binance.com/docs/binance-spot-api-docs/rest-api/account-endpoints#account-information-user_data

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `omitZeroBalances` | BOOLEAN | NO | If `true`, returns only non-zero balances. Default: `false` |
| `recvWindow` | DECIMAL | NO | Max value: 60000. Supports up to 3 decimal places |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `makerCommission` | INT | Maker commission rate (in basis points) |
| `takerCommission` | INT | Taker commission rate (in basis points) |
| `buyerCommission` | INT | Buyer commission rate (in basis points) |
| `sellerCommission` | INT | Seller commission rate (in basis points) |
| `commissionRates` | OBJECT | Detailed commission rates (`maker`, `taker`, `buyer`, `seller` as strings) |
| `canTrade` | BOOLEAN | Whether the account can trade |
| `canWithdraw` | BOOLEAN | Whether the account can withdraw |
| `canDeposit` | BOOLEAN | Whether the account can deposit |
| `brokered` | BOOLEAN | Whether the account is brokered |
| `requireSelfTradePrevention` | BOOLEAN | Whether STP is required |
| `preventSor` | BOOLEAN | Whether SOR is prevented |
| `updateTime` | LONG | Account update timestamp |
| `accountType` | STRING | Account type (e.g., `SPOT`) |
| `balances` | ARRAY | List of asset balances |
| `permissions` | ARRAY | Account permissions (e.g., `["SPOT"]`) |
| `uid` | LONG | User ID |

**Balance Object Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `asset` | STRING | Asset symbol (e.g., `BTC`) |
| `free` | STRING | Available balance |
| `locked` | STRING | Locked balance (in open orders) |

#### Sample Response

```json
{
  "makerCommission": 15,
  "takerCommission": 15,
  "buyerCommission": 0,
  "sellerCommission": 0,
  "commissionRates": {
    "maker": "0.00150000",
    "taker": "0.00150000",
    "buyer": "0.00000000",
    "seller": "0.00000000"
  },
  "canTrade": true,
  "canWithdraw": true,
  "canDeposit": true,
  "brokered": false,
  "requireSelfTradePrevention": false,
  "preventSor": false,
  "updateTime": 123456789,
  "accountType": "SPOT",
  "balances": [
    {
      "asset": "BTC",
      "free": "4723846.89208129",
      "locked": "0.00000000"
    },
    {
      "asset": "LTC",
      "free": "4763368.68006011",
      "locked": "0.00000000"
    }
  ],
  "permissions": ["SPOT"],
  "uid": 354937868
}
```

### Order Rate Limits

- **Description**: Displays the user's unfilled order count for all intervals.
- **Endpoint**: `GET /api/v3/rateLimit/order`
- **Weight**: 40
- **Security Type**: `USER_DATA`
- **Docs**: https://developers.binance.com/docs/binance-spot-api-docs/rest-api/account-endpoints#query-unfilled-order-count-user_data

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `recvWindow` | DECIMAL | NO | Max value: 60000 |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `rateLimitType` | STRING | Type of rate limit (e.g., `ORDERS`) |
| `interval` | STRING | Time interval (`SECOND`, `DAY`) |
| `intervalNum` | INT | Number of intervals |
| `limit` | INT | Maximum allowed count |
| `count` | INT | Current count |

#### Sample Response

```json
[
  {
    "rateLimitType": "ORDERS",
    "interval": "SECOND",
    "intervalNum": 10,
    "limit": 50,
    "count": 0
  },
  {
    "rateLimitType": "ORDERS",
    "interval": "DAY",
    "intervalNum": 1,
    "limit": 160000,
    "count": 0
  }
]
```

## Subaccount APIs (Spot)

### Create Virtual Sub-Account

- **Description**: Create a virtual sub-account under the master account. A virtual email is generated automatically.
- **Endpoint**: `POST /sapi/v1/sub-account/virtualSubAccount`
- **Weight**: 1 (IP)
- **Security Type**: `USER_DATA`
- **Docs**: https://developers.binance.com/docs/sub_account/account-management

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `subAccountString` | STRING | YES | A string used to generate a virtual email for registration |
| `recvWindow` | LONG | NO | Request validity window |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

> **Note**: The API key must have the `trade` permission enabled.

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `email` | STRING | Generated virtual sub-account email |

#### Sample Response

```json
{
  "email": "addsdd_virtual@aasaixwqnoemail.com"
}
```

### Query Sub-Account List

- **Description**: Query the list of sub-accounts under the master account.
- **Endpoint**: `GET /sapi/v1/sub-account/list`
- **Weight**: 1 (IP)
- **Security Type**: `USER_DATA`
- **Docs**: https://developers.binance.com/docs/sub_account/account-management/Query-Sub-account-List

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `email` | STRING | NO | Filter by sub-account email |
| `isFreeze` | STRING | NO | Filter by freeze status (`true` or `false`) |
| `page` | INT | NO | Page number. Default: 1 |
| `limit` | INT | NO | Results per page. Default: 1, Max: 200 |
| `recvWindow` | LONG | NO | Request validity window |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `subAccounts` | ARRAY | List of sub-account objects |

**Sub-Account Object Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `subUserId` | LONG | Sub-account user ID |
| `email` | STRING | Sub-account email |
| `remark` | STRING | Sub-account remark |
| `isFreeze` | BOOLEAN | Whether the sub-account is frozen |
| `createTime` | LONG | Creation timestamp |
| `isManagedSubAccount` | BOOLEAN | Whether it is a managed sub-account |
| `isAssetManagementSubAccount` | BOOLEAN | Whether it is an asset management sub-account |

#### Sample Response

```json
{
  "subAccounts": [
    {
      "subUserId": 123456,
      "email": "testsub@gmail.com",
      "remark": "remark",
      "isFreeze": false,
      "createTime": 1544433328000,
      "isManagedSubAccount": false,
      "isAssetManagementSubAccount": false
    },
    {
      "subUserId": 1234567,
      "email": "virtual@oxebmvfonoemail.com",
      "remark": "remarks",
      "isFreeze": false,
      "createTime": 1544433328000,
      "isManagedSubAccount": false,
      "isAssetManagementSubAccount": false
    }
  ]
}
```

### Query Sub-Account Spot Asset Transfer History

- **Description**: Query spot asset transfer history between sub-accounts and master account.
- **Endpoint**: `GET /sapi/v1/sub-account/sub/transfer/history`
- **Weight**: 1 (IP)
- **Security Type**: `USER_DATA`
- **Docs**: https://developers.binance.com/docs/sub_account/asset-management/Query-Sub-account-Spot-Asset-Transfer-History

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `fromEmail` | STRING | NO | Source email. Cannot be sent together with `toEmail` |
| `toEmail` | STRING | NO | Destination email. Cannot be sent together with `fromEmail` |
| `startTime` | LONG | NO | Start timestamp |
| `endTime` | LONG | NO | End timestamp |
| `page` | INT | NO | Page number. Default: 1 |
| `limit` | INT | NO | Results per page. Default: 500 |
| `recvWindow` | LONG | NO | Request validity window |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

> **Note**: `fromEmail` and `toEmail` cannot be sent at the same time. `fromEmail` defaults to master account email.

#### Response Fields

Returns an array of transfer records.

| Field | Type | Description |
|-------|------|-------------|
| `from` | STRING | Source email |
| `to` | STRING | Destination email |
| `asset` | STRING | Asset symbol |
| `qty` | STRING | Transfer quantity |
| `status` | STRING | Transfer status (`SUCCESS`, etc.) |
| `tranId` | LONG | Transaction ID |
| `time` | LONG | Transfer timestamp |

#### Sample Response

```json
[
  {
    "from": "aaa@test.com",
    "to": "bbb@test.com",
    "asset": "BTC",
    "qty": "10",
    "status": "SUCCESS",
    "tranId": 6489943656,
    "time": 1544433328000
  },
  {
    "from": "bbb@test.com",
    "to": "ccc@test.com",
    "asset": "ETH",
    "qty": "2",
    "status": "SUCCESS",
    "tranId": 6489938713,
    "time": 1544433328000
  }
]
```

### Sub-Account Spot Asset Transfer (Universal Transfer)

- **Description**: Universal transfer between master and sub-accounts across account types (SPOT, USDT_FUTURE, COIN_FUTURE, MARGIN, ISOLATED_MARGIN).
- **Endpoint**: `POST /sapi/v1/sub-account/universalTransfer`
- **Weight**: 1 (IP) / 360 (UID)
- **Security Type**: `USER_DATA`
- **Docs**: https://developers.binance.com/docs/sub_account/asset-management/Universal-Transfer

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `fromEmail` | STRING | NO | Source email. Defaults to master account |
| `toEmail` | STRING | NO | Destination email. Defaults to master account |
| `fromAccountType` | STRING | YES | Source account type: `SPOT`, `USDT_FUTURE`, `COIN_FUTURE`, `MARGIN`, `ISOLATED_MARGIN` |
| `toAccountType` | STRING | YES | Destination account type: `SPOT`, `USDT_FUTURE`, `COIN_FUTURE`, `MARGIN`, `ISOLATED_MARGIN` |
| `clientTranId` | STRING | NO | Unique client transaction ID |
| `symbol` | STRING | NO | Only required for `ISOLATED_MARGIN` transfers |
| `asset` | STRING | YES | Asset to transfer |
| `amount` | DECIMAL | YES | Transfer amount |
| `recvWindow` | LONG | NO | Request validity window |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

> **Notes**:
> - The API key must have `internal transfer` permission enabled.
> - At least one of `fromEmail` or `toEmail` must be sent when `fromAccountType` equals `toAccountType`.
> - Supported transfers: SPOT <-> SPOT/USDT_FUTURE/COIN_FUTURE, SPOT <-> MARGIN/ISOLATED_MARGIN (master <-> sub), MARGIN <-> MARGIN (sub <-> sub)

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `tranId` | LONG | Transaction ID |
| `clientTranId` | STRING | Client transaction ID (if provided) |

#### Sample Response

```json
{
  "tranId": 11945860693,
  "clientTranId": "test"
}
```

### Query Sub-Account Spot Assets Summary

- **Description**: Get BTC-valued asset summary of sub-accounts.
- **Endpoint**: `GET /sapi/v1/sub-account/spotSummary`
- **Weight**: 1 (IP)
- **Security Type**: `USER_DATA`
- **Docs**: https://developers.binance.com/docs/sub_account/asset-management/Query-Sub-account-Spot-Assets-Summary

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `email` | STRING | NO | Sub-account email to filter |
| `page` | LONG | NO | Page number. Default: 1 |
| `size` | LONG | NO | Results per page. Default: 10, Max: 20 |
| `recvWindow` | LONG | NO | Request validity window |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `totalCount` | INT | Total number of sub-accounts |
| `masterAccountTotalAsset` | STRING | Master account total asset in BTC |
| `spotSubUserAssetBtcVoList` | ARRAY | List of sub-account asset summaries |

**Sub-Account Asset Summary Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `email` | STRING | Sub-account email |
| `totalAsset` | STRING | Total asset value in BTC |

#### Sample Response

```json
{
  "totalCount": 2,
  "masterAccountTotalAsset": "0.23231201",
  "spotSubUserAssetBtcVoList": [
    {
      "email": "sub123@test.com",
      "totalAsset": "9999.00000000"
    },
    {
      "email": "test456@test.com",
      "totalAsset": "0.00000000"
    }
  ]
}
```

### Query Sub-Account Assets

- **Description**: Fetch detailed asset balances of a specific sub-account.
- **Endpoint**: `GET /sapi/v4/sub-account/assets`
- **Weight**: 60 (UID)
- **Security Type**: `USER_DATA`
- **Docs**: https://developers.binance.com/docs/sub_account/asset-management/Query-Sub-account-Assets-V4

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `email` | STRING | YES | Sub-account email |
| `recvWindow` | LONG | NO | Request validity window |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `balances` | ARRAY | List of asset balance objects |

**Balance Object Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `asset` | STRING | Asset symbol |
| `free` | STRING | Available balance |
| `locked` | STRING | Locked balance |
| `freeze` | STRING | Frozen balance |
| `withdrawing` | STRING | Withdrawing balance |

#### Sample Response

```json
{
  "balances": [
    {
      "freeze": "0",
      "withdrawing": "0",
      "asset": "ADA",
      "free": "10000",
      "locked": "0"
    },
    {
      "freeze": "0",
      "withdrawing": "0",
      "asset": "BNB",
      "free": "10003",
      "locked": "0"
    },
    {
      "freeze": "0",
      "withdrawing": "0",
      "asset": "BTC",
      "free": "11467.6399",
      "locked": "0"
    }
  ]
}
```

## Futures Account APIs

### Futures Account Information (USDT-M)

- **Description**: Get current futures account information. Response differs between single-asset and multi-assets mode.
- **Endpoint**: `GET /fapi/v2/account`
- **Weight**: 5
- **Security Type**: `USER_DATA`
- **Docs**: https://developers.binance.com/docs/derivatives/usds-margined-futures/account/rest-api/Account-Information-V2

> **Note**: A newer V3 endpoint (`GET /fapi/v3/account`) exists that returns only symbols with positions/open orders for better performance. V2 returns all symbols.

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `recvWindow` | LONG | NO | Request validity window |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `feeTier` | INT | Account commission tier |
| `feeBurn` | BOOLEAN | Fee discount status (`true` = on, `false` = off) |
| `canTrade` | BOOLEAN | Whether the account can trade |
| `canDeposit` | BOOLEAN | Whether the account can deposit |
| `canWithdraw` | BOOLEAN | Whether the account can withdraw |
| `updateTime` | LONG | Reserved (ignore) |
| `multiAssetsMargin` | BOOLEAN | Whether multi-assets margin is enabled |
| `tradeGroupId` | INT | Trade group ID |
| `totalInitialMargin` | STRING | Total initial margin required |
| `totalMaintMargin` | STRING | Total maintenance margin required |
| `totalWalletBalance` | STRING | Total wallet balance |
| `totalUnrealizedProfit` | STRING | Total unrealized profit |
| `totalMarginBalance` | STRING | Total margin balance |
| `totalPositionInitialMargin` | STRING | Initial margin for positions |
| `totalOpenOrderInitialMargin` | STRING | Initial margin for open orders |
| `totalCrossWalletBalance` | STRING | Crossed wallet balance |
| `totalCrossUnPnl` | STRING | Unrealized profit of crossed positions |
| `availableBalance` | STRING | Available balance |
| `maxWithdrawAmount` | STRING | Maximum amount for transfer out |
| `assets` | ARRAY | Asset details |
| `positions` | ARRAY | Position details for all symbols |

**Asset Object Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `asset` | STRING | Asset name |
| `walletBalance` | STRING | Wallet balance |
| `unrealizedProfit` | STRING | Unrealized profit |
| `marginBalance` | STRING | Margin balance |
| `maintMargin` | STRING | Maintenance margin required |
| `initialMargin` | STRING | Total initial margin |
| `positionInitialMargin` | STRING | Initial margin for positions |
| `openOrderInitialMargin` | STRING | Initial margin for open orders |
| `crossWalletBalance` | STRING | Crossed wallet balance |
| `crossUnPnl` | STRING | Unrealized profit of crossed positions |
| `availableBalance` | STRING | Available balance |
| `maxWithdrawAmount` | STRING | Maximum transfer out amount |
| `marginAvailable` | BOOLEAN | Whether asset can be used as margin in Multi-Assets mode |
| `updateTime` | LONG | Last update time |

**Position Object Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `symbol` | STRING | Symbol name |
| `initialMargin` | STRING | Initial margin required |
| `maintMargin` | STRING | Maintenance margin required |
| `unrealizedProfit` | STRING | Unrealized profit |
| `positionInitialMargin` | STRING | Initial margin for positions |
| `openOrderInitialMargin` | STRING | Initial margin for open orders |
| `leverage` | STRING | Current initial leverage |
| `isolated` | BOOLEAN | Whether position is isolated |
| `entryPrice` | STRING | Average entry price |
| `maxNotional` | STRING | Maximum available notional |
| `bidNotional` | STRING | Bids notional (ignore) |
| `askNotional` | STRING | Ask notional (ignore) |
| `positionSide` | STRING | Position side (`BOTH`, `LONG`, `SHORT`) |
| `positionAmt` | STRING | Position amount |
| `updateTime` | LONG | Last update time |

#### Sample Response (Single-Asset Mode)

```json
{
  "feeTier": 0,
  "feeBurn": true,
  "canTrade": true,
  "canDeposit": true,
  "canWithdraw": true,
  "updateTime": 0,
  "multiAssetsMargin": false,
  "tradeGroupId": -1,
  "totalInitialMargin": "0.00000000",
  "totalMaintMargin": "0.00000000",
  "totalWalletBalance": "23.72469206",
  "totalUnrealizedProfit": "0.00000000",
  "totalMarginBalance": "23.72469206",
  "totalPositionInitialMargin": "0.00000000",
  "totalOpenOrderInitialMargin": "0.00000000",
  "totalCrossWalletBalance": "23.72469206",
  "totalCrossUnPnl": "0.00000000",
  "availableBalance": "23.72469206",
  "maxWithdrawAmount": "23.72469206",
  "assets": [
    {
      "asset": "USDT",
      "walletBalance": "23.72469206",
      "unrealizedProfit": "0.00000000",
      "marginBalance": "23.72469206",
      "maintMargin": "0.00000000",
      "initialMargin": "0.00000000",
      "positionInitialMargin": "0.00000000",
      "openOrderInitialMargin": "0.00000000",
      "crossWalletBalance": "23.72469206",
      "crossUnPnl": "0.00000000",
      "availableBalance": "23.72469206",
      "maxWithdrawAmount": "23.72469206",
      "marginAvailable": true,
      "updateTime": 1625474304765
    }
  ],
  "positions": [
    {
      "symbol": "BTCUSDT",
      "initialMargin": "0",
      "maintMargin": "0",
      "unrealizedProfit": "0.00000000",
      "positionInitialMargin": "0",
      "openOrderInitialMargin": "0",
      "leverage": "100",
      "isolated": true,
      "entryPrice": "0.00000",
      "maxNotional": "250000",
      "bidNotional": "0",
      "askNotional": "0",
      "positionSide": "BOTH",
      "positionAmt": "0",
      "updateTime": 0
    }
  ]
}
```

### Futures Account Balance (USDT-M)

- **Description**: Query futures account balance for all assets.
- **Endpoint**: `GET /fapi/v2/balance`
- **Weight**: 5
- **Security Type**: `USER_DATA`
- **Docs**: https://developers.binance.com/docs/derivatives/usds-margined-futures/account/rest-api/Futures-Account-Balance-V2

> **Note**: A newer V3 endpoint (`GET /fapi/v3/balance`) exists with identical parameters and response format but better performance.

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `recvWindow` | LONG | NO | Request validity window |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

#### Response Fields

Returns an array of balance objects.

| Field | Type | Description |
|-------|------|-------------|
| `accountAlias` | STRING | Unique account code |
| `asset` | STRING | Asset name |
| `balance` | STRING | Wallet balance |
| `crossWalletBalance` | STRING | Crossed wallet balance |
| `crossUnPnl` | STRING | Unrealized profit of crossed positions |
| `availableBalance` | STRING | Available balance |
| `maxWithdrawAmount` | STRING | Maximum amount for transfer out |
| `marginAvailable` | BOOLEAN | Whether asset can be used as margin in Multi-Assets mode |
| `updateTime` | LONG | Last update time |

#### Sample Response

```json
[
  {
    "accountAlias": "SgsR",
    "asset": "USDT",
    "balance": "122607.35137903",
    "crossWalletBalance": "23.72469206",
    "crossUnPnl": "0.00000000",
    "availableBalance": "23.72469206",
    "maxWithdrawAmount": "23.72469206",
    "marginAvailable": true,
    "updateTime": 1617939110373
  }
]
```

### Futures Sub-Account Transfer (Internal Transfer)

- **Description**: Transfer assets between sub-account and master account for futures.
- **Endpoint**: `POST /sapi/v1/sub-account/futures/internalTransfer`
- **Weight**: 1 (IP)
- **Security Type**: `USER_DATA`
- **Docs**: https://developers.binance.com/docs/sub_account/asset-management/Sub-account-Futures-Asset-Transfer

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `email` | STRING | YES | Sub-account email |
| `asset` | STRING | YES | Asset to transfer |
| `amount` | DECIMAL | YES | Transfer amount |
| `type` | INT | YES | Transfer type: `1` = master to sub, `2` = sub to master |
| `recvWindow` | LONG | NO | Request validity window |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `tranId` | LONG | Transaction ID |

#### Sample Response

```json
{
  "tranId": 11945860693
}
```

## Constraints & Limits

### Rate Limits

Rate limits are enforced at both IP and UID levels. Exceeding limits returns HTTP 429.

| Scope | Limit | Window |
|-------|-------|--------|
| Spot IP (RAW_REQUESTS) | 300,000 | 5 minutes |
| Spot IP (REQUEST_WEIGHT) | 6,000 | 1 minute |
| Spot Account (ORDERS) | 50 | 10 seconds |
| Spot Account (ORDERS) | 160,000 | 1 day |
| Sub-Account (UID) | Varies per endpoint | Varies |

- A `Retry-After` header is included with 429/418 responses indicating seconds to wait.
- Repeated violations result in escalating IP bans from 2 minutes to 3 days.
- IP limits are based on IP address, not API key.
- Unfilled order counts are tracked per account.

### Endpoint-Specific Weights

| Endpoint | Weight |
|----------|--------|
| `GET /api/v3/account` | 20 |
| `GET /api/v3/rateLimit/order` | 40 |
| `GET /fapi/v2/account` | 5 |
| `GET /fapi/v2/balance` | 5 |
| `POST /sapi/v1/sub-account/virtualSubAccount` | 1 (IP) |
| `GET /sapi/v1/sub-account/list` | 1 (IP) |
| `GET /sapi/v1/sub-account/sub/transfer/history` | 1 (IP) |
| `POST /sapi/v1/sub-account/universalTransfer` | 1 (IP) / 360 (UID) |
| `GET /sapi/v1/sub-account/spotSummary` | 1 (IP) |
| `GET /sapi/v4/sub-account/assets` | 60 (UID) |

## Notes

- **V2 vs V3 Futures endpoints**: V3 endpoints (`/fapi/v3/account`, `/fapi/v3/balance`) are optimized to return only symbols with active positions or open orders. V2 returns all symbols. Prefer V3 for better performance.
- **Sub-account transfers**: The `universalTransfer` endpoint (`/sapi/v1/sub-account/universalTransfer`) is the recommended approach for most transfer scenarios, supporting SPOT, USDT_FUTURE, COIN_FUTURE, MARGIN, and ISOLATED_MARGIN account types.
- **Signature encoding**: All non-ASCII characters in request parameters must be percent-encoded (UTF-8) **before** computing the HMAC-SHA256 signature.
- **Timestamp sync**: Ensure your system clock is synchronized with Binance server time. Use `GET /api/v3/time` to check server time and adjust for clock skew.
- **Sub-account API permissions**: Sub-account API keys require specific permissions (`trade`, `internal transfer`) to be explicitly enabled in the API Management page.
- **Virtual sub-accounts**: Created with auto-generated email addresses in the format `<string>_virtual@<random>noemail.com`. These cannot receive real emails.
- **Balance string types**: All numeric values in responses are returned as strings to preserve precision.
- **`recvWindow` precision**: Supports up to 3 decimal places for microsecond precision (e.g., `6000.346`), though the value is still conceptually in milliseconds.
