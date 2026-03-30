# MEXC Account REST API Documentation

## Reference
- **Official API Docs**: https://www.mexc.com/api-docs/spot/account/account-information
- **Legacy Docs**: https://mexcdevelop.github.io/apidocs/spot_v3_en/
- **Sub-Account Endpoints**: https://www.mexc.com/api-docs/spot-v3/subaccount-endpoints
- **Base URL**: `https://api.mexc.com`

## Authentication

All account-related endpoints require authentication via HMAC-SHA256 signed requests.

### Required Headers

| Header | Description |
|--------|-------------|
| `X-MEXC-APIKEY` | Your API access key |
| `Content-Type` | `application/json` |

### Signature Method

- **Algorithm**: HMAC-SHA256
- **Key**: Your `secretKey`
- **Value**: The `totalParams` (query string concatenated with request body)
- **Case**: Lowercase only

### Timing Security

Requests must include a `timestamp` parameter (millisecond epoch). The server validates:

```
if (timestamp < (serverTime + 1000) && (serverTime - timestamp) <= recvWindow)
```

| Parameter | Type | Mandatory | Description |
|-----------|------|-----------|-------------|
| `timestamp` | long | YES | Millisecond timestamp of request creation |
| `recvWindow` | long | NO | Max validity window after timestamp. Default: 5000ms, Max: 60000ms |
| `signature` | string | YES | HMAC-SHA256 signature |

### Example Signature

```bash
echo -n "symbol=BTCUSDT&side=BUY&type=LIMIT&quantity=1&price=11&recvWindow=5000&timestamp=1644489390087" \
  | openssl dgst -sha256 -hmac "YOUR_SECRET_KEY"
```

## Account APIs (Spot)

### Account Information

- **Description**: Get current account information including balances, permissions, and trading capabilities.
- **Endpoint**: `GET /api/v3/account`
- **Permission**: `SPOT_ACCOUNT_READ`
- **Weight (IP)**: 10
- **Rate Limit**: 2 times/s

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `recvWindow` | long | NO | Request validity window (default: 5000ms) |
| `timestamp` | long | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `canTrade` | boolean | Whether trading is enabled |
| `canWithdraw` | boolean | Whether withdrawals are enabled |
| `canDeposit` | boolean | Whether deposits are enabled |
| `updateTime` | long | Last update time (may be null) |
| `accountType` | string | Account type (e.g., "SPOT") |
| `balances` | array | List of asset balances |
| `balances[].asset` | string | Asset coin symbol |
| `balances[].free` | string | Available (unlocked) balance |
| `balances[].locked` | string | Frozen (locked) balance |
| `permissions` | array | Account permissions (e.g., ["SPOT"]) |

#### Sample Response

```json
{
    "canTrade": true,
    "canWithdraw": true,
    "canDeposit": true,
    "updateTime": null,
    "accountType": "SPOT",
    "balances": [
        {
            "asset": "USDT",
            "free": "10200.00",
            "locked": "500.00"
        },
        {
            "asset": "BTC",
            "free": "0.50000000",
            "locked": "0.00000000"
        }
    ],
    "permissions": ["SPOT"]
}
```

### Account Trade List

- **Description**: Get trades for a specific account and symbol. Only transaction records from the past 1 month can be queried via API.
- **Endpoint**: `GET /api/v3/myTrades`
- **Permission**: `SPOT_ACCOUNT_READ`
- **Weight (IP)**: 10

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | string | YES | Trading pair symbol |
| `orderId` | string | NO | Filter by order ID |
| `startTime` | long | NO | Start timestamp |
| `endTime` | long | NO | End timestamp |
| `limit` | int | NO | Default: 100, Max: 100 |
| `recvWindow` | long | NO | Request validity window |
| `timestamp` | long | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `symbol` | string | Trading pair |
| `id` | string | Trade ID |
| `orderId` | string | Order ID |
| `orderListId` | long | Order list ID (-1 if none) |
| `price` | string | Trade price |
| `qty` | string | Trade quantity |
| `quoteQty` | string | Quote asset quantity |
| `commission` | string | Commission amount |
| `commissionAsset` | string | Commission asset |
| `time` | long | Trade timestamp |
| `isBuyer` | boolean | Whether the account was the buyer |
| `isMaker` | boolean | Whether the account was the maker |
| `isBestMatch` | boolean | Whether it was the best price match |
| `isSelfTrade` | boolean | Whether it was a self-trade |
| `clientOrderId` | string | Client order ID |

#### Sample Response

```json
[
  {
    "symbol": "BTCUSDT",
    "id": "fad2af9e942049b6adbda1a271f990c6",
    "orderId": "bb41e5663e124046bd9497a3f5692f39",
    "orderListId": -1,
    "price": "45000.00",
    "qty": "0.00100000",
    "quoteQty": "45.00",
    "commission": "0.04500000",
    "commissionAsset": "USDT",
    "time": 1644489549590,
    "isBuyer": true,
    "isMaker": false,
    "isBestMatch": true,
    "isSelfTrade": false,
    "clientOrderId": null
  }
]
```

### Query KYC Status

- **Description**: Query the KYC verification status of the account.
- **Endpoint**: `GET /api/v3/kyc/status`
- **Permission**: `SPOT_ACCOUNT_READ`
- **Weight (IP)**: 1

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `timestamp` | long | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `status` | string | KYC status: "0" = unverified, "1" = verified |

#### Sample Response

```json
{
    "status": "1"
}
```

### Query Symbol Commission

- **Description**: Query the commission rate for a specific trading pair.
- **Endpoint**: `GET /api/v3/symbolCommission`
- **Permission**: `SPOT_ACCOUNT_READ`
- **Weight (IP)**: 1

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | string | YES | Trading pair symbol |
| `timestamp` | long | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `symbol` | string | Trading pair |
| `makerCommission` | string | Maker commission rate |
| `takerCommission` | string | Taker commission rate |

## Subaccount APIs

> **Note**: All sub-account endpoints only support **main account API keys**. Sub-account API keys cannot access these endpoints.

### Create Virtual Sub-Account

- **Description**: Create a new virtual sub-account from the master account. Sub-accounts created via API cannot be logged in on the web interface.
- **Endpoint**: `POST /api/v3/sub-account/virtualSubAccount`
- **Permission**: `SPOT_ACCOUNT_READ`
- **Weight (IP)**: 1

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `subAccount` | string | YES | Sub-account name (8-32 letters and numbers) |
| `note` | string | YES | Sub-account notes/remarks |
| `recvWindow` | long | NO | Request validity window |
| `timestamp` | long | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `subAccount` | string | Created sub-account name |
| `note` | string | Sub-account notes |

#### Sample Response

```json
{
    "subAccount": "trading_bot_01",
    "note": "Automated trading sub-account"
}
```

### List Sub-Accounts

- **Description**: Get details of the sub-account list with pagination support.
- **Endpoint**: `GET /api/v3/sub-account/list`
- **Permission**: `SPOT_ACCOUNT_READ`
- **Weight (IP)**: 1

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `subAccount` | string | NO | Filter by sub-account name |
| `isFreeze` | string | NO | Filter by freeze status: "true" or "false" |
| `page` | int | NO | Page number (default: 1) |
| `limit` | int | NO | Results per page (default: 10, max: 200) |
| `recvWindow` | long | NO | Request validity window |
| `timestamp` | long | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `subAccounts` | array | List of sub-accounts |
| `subAccounts[].subAccount` | string | Sub-account name |
| `subAccounts[].isFreeze` | boolean | Whether the sub-account is frozen |
| `subAccounts[].createTime` | long | Creation timestamp |
| `subAccounts[].uid` | string | Sub-account UID |

#### Sample Response

```json
{
    "subAccounts": [
        {
            "subAccount": "trading_bot_01",
            "isFreeze": false,
            "createTime": 1644433328000,
            "uid": "49910511"
        },
        {
            "subAccount": "arb_strategy_02",
            "isFreeze": false,
            "createTime": 1644433328000,
            "uid": "91921059"
        }
    ]
}
```

### Get Sub-Account Assets

- **Description**: Query the asset balances of a specific sub-account. Only supports querying a single sub-account at a time.
- **Endpoint**: `GET /api/v3/sub-account/asset`
- **Permission**: `SPOT_TRANSFER_READ`
- **Weight (IP)**: 1
- **Added**: 2024-01-12

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `subAccount` | string | YES | Sub-account name (single account only) |
| `accountType` | string | YES | Account type: "SPOT" or "FUTURES" (currently only SPOT supported) |
| `timestamp` | string | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `balances` | array | List of asset balances |
| `balances[].asset` | string | Asset coin symbol |
| `balances[].free` | string | Available (unlocked) balance |
| `balances[].locked` | string | Frozen (locked) balance |

#### Sample Response

```json
{
    "balances": [
        {
            "asset": "USDT",
            "free": "5000.00",
            "locked": "100.00"
        },
        {
            "asset": "BTC",
            "free": "0.25000000",
            "locked": "0.00000000"
        }
    ]
}
```

## Transfer APIs

### User Universal Transfer (Spot to Futures, etc.)

- **Description**: Transfer assets between different account types (SPOT ↔ FUTURES) within the same user account.
- **Endpoint**: `POST /api/v3/capital/transfer`
- **Permission**: `SPOT_TRANSFER_WRITE`
- **Weight (IP)**: 1

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `fromAccountType` | string | YES | Source account type: "SPOT" or "FUTURES" |
| `toAccountType` | string | YES | Destination account type: "SPOT" or "FUTURES" |
| `asset` | string | YES | Asset to transfer (e.g., "USDT") |
| `amount` | string | YES | Transfer amount |
| `timestamp` | string | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `tranId` | string | Transfer transaction ID |

#### Sample Response

```json
[
  {
    "tranId": "c45d800a47ba4cbc876a5cd29388319"
  }
]
```

### Query User Universal Transfer History

- **Description**: Query the history of universal transfers between account types.
- **Endpoint**: `GET /api/v3/capital/transfer`
- **Permission**: `SPOT_TRANSFER_READ`
- **Weight (IP)**: 1

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `fromAccountType` | string | YES | Source account type: "SPOT" or "FUTURES" |
| `toAccountType` | string | YES | Destination account type: "SPOT" or "FUTURES" |
| `startTime` | string | NO | Start timestamp (default: last 7 days if not provided) |
| `endTime` | string | NO | End timestamp |
| `page` | string | NO | Page number (default: 1) |
| `size` | string | NO | Results per page (default: 10, max: 100) |
| `timestamp` | string | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

**Notes**:
- Only can query data for the last six months
- If `startTime` and `endTime` are not sent, returns the last seven days' data by default

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `total` | int | Total number of records |
| `rows` | array | List of transfer records |
| `rows[].tranId` | string | Transfer ID |
| `rows[].clientTranId` | string | Client-defined transfer ID |
| `rows[].asset` | string | Asset symbol |
| `rows[].amount` | string | Transfer amount |
| `rows[].fromAccountType` | string | Source account type |
| `rows[].toAccountType` | string | Destination account type |
| `rows[].fromSymbol` | string | Source symbol |
| `rows[].toSymbol` | string | Destination symbol |
| `rows[].status` | string | Transfer status |
| `rows[].timestamp` | long | Transfer timestamp |

#### Sample Response

```json
[
  {
    "rows": [
      {
        "tranId": "11945860693",
        "clientTranId": "test",
        "asset": "BTC",
        "amount": "0.1",
        "fromAccountType": "SPOT",
        "toAccountType": "FUTURES",
        "fromSymbol": "SPOT",
        "toSymbol": "FUTURES",
        "status": "SUCCESS",
        "timestamp": 1544433325000
      }
    ],
    "total": 1
  }
]
```

### Query User Universal Transfer History (by tranId)

- **Description**: Query a specific universal transfer by its transaction ID.
- **Endpoint**: `GET /api/v3/capital/transfer/tranId`
- **Permission**: `SPOT_TRANSFER_READ`
- **Weight (IP)**: 1

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `tranId` | string | YES | Transfer transaction ID |
| `timestamp` | string | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

**Note**: Only can query data for the last six months.

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `tranId` | string | Transfer ID |
| `clientTranId` | string | Client-defined transfer ID |
| `asset` | string | Asset symbol |
| `amount` | string | Transfer amount |
| `fromAccountType` | string | Source account type |
| `toAccountType` | string | Destination account type |
| `symbol` | string | Symbol |
| `status` | string | Transfer status |
| `timestamp` | long | Transfer timestamp |

#### Sample Response

```json
{
    "tranId": "cb28c88cd20c42819e4d5148d5fb5742",
    "clientTranId": null,
    "asset": "USDT",
    "amount": "10",
    "fromAccountType": "SPOT",
    "toAccountType": "FUTURES",
    "symbol": null,
    "status": "SUCCESS",
    "timestamp": 1678603205000
}
```

### Universal Transfer (For Master Account - Sub-Accounts)

- **Description**: Transfer assets between master account and sub-accounts, or between sub-accounts. Supports SPOT and FUTURES account types.
- **Endpoint**: `POST /api/v3/capital/sub-account/universalTransfer`
- **Permission**: `SPOT_TRANSFER_WRITE`
- **Weight (IP)**: 1

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `fromAccount` | string | NO | Source account (defaults to master if not sent) |
| `toAccount` | string | NO | Destination account (defaults to master if not sent) |
| `fromAccountType` | string | YES | Source account type: "SPOT" or "FUTURES" |
| `toAccountType` | string | YES | Destination account type: "SPOT" or "FUTURES" |
| `asset` | string | YES | Asset to transfer (e.g., "USDT") |
| `amount` | string | YES | Transfer amount |
| `timestamp` | string | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `tranId` | string | Transfer transaction ID |

#### Sample Response

```json
{
    "tranId": 11945860693
}
```

### Query Universal Transfer History (For Master Account - Sub-Accounts)

- **Description**: Query the history of universal transfers between master and sub-accounts.
- **Endpoint**: `GET /api/v3/capital/sub-account/universalTransfer`
- **Permission**: `SPOT_TRANSFER_READ`
- **Weight (IP)**: 1

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `fromAccount` | string | NO | Source account (defaults to master if not sent) |
| `toAccount` | string | NO | Destination account (defaults to master if not sent) |
| `fromAccountType` | string | YES | Source account type: "SPOT" or "FUTURES" |
| `toAccountType` | string | YES | Destination account type: "SPOT" or "FUTURES" |
| `startTime` | string | NO | Start timestamp |
| `endTime` | string | NO | End timestamp |
| `page` | string | NO | Page number (default: 1) |
| `limit` | string | NO | Results per page (default: 500, max: 500) |
| `timestamp` | string | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `tranId` | string | Transfer ID |
| `fromAccount` | string | Source account email |
| `toAccount` | string | Destination account email |
| `clientTranId` | string | Client-defined transfer ID |
| `asset` | string | Asset symbol |
| `amount` | string | Transfer amount |
| `fromAccountType` | string | Source account type |
| `toAccountType` | string | Destination account type |
| `fromSymbol` | string | Source symbol |
| `toSymbol` | string | Destination symbol |
| `status` | string | Transfer status |
| `timestamp` | number | Transfer timestamp |

#### Sample Response

```json
{
    "tranId": "11945860693",
    "fromAccount": "master@example.com",
    "toAccount": "subaccount1@example.com",
    "clientTranId": "test",
    "asset": "BTC",
    "amount": "0.1",
    "fromAccountType": "SPOT",
    "toAccountType": "FUTURES",
    "fromSymbol": "SPOT",
    "toSymbol": "FUTURES",
    "status": "SUCCESS",
    "timestamp": 1544433325000
}
```

### Internal Transfer (User to User)

- **Description**: Transfer assets internally to another MEXC user identified by email, UID, or mobile number.
- **Endpoint**: `POST /api/v3/capital/transfer/internal`
- **Permission**: `SPOT_WITHDRAW_WRITE`
- **Weight (IP)**: 1
- **Added**: 2023-11-10

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `toAccountType` | string | YES | Recipient identifier type: "EMAIL", "UID", or "MOBILE" |
| `toAccount` | string | YES | Recipient account (email, UID, or mobile) |
| `areaCode` | string | NO | Area code for mobile number |
| `asset` | string | YES | Asset to transfer |
| `amount` | string | YES | Transfer amount |
| `timestamp` | string | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `tranId` | string | Transfer transaction ID |

#### Sample Response

```json
{
    "tranId": "c45d800a47ba4cbc876a5cd29388319"
}
```

### Query Internal Transfer History

- **Description**: Query the history of internal transfers to other MEXC users.
- **Endpoint**: `GET /api/v3/capital/transfer/internal`
- **Permission**: `SPOT_WITHDRAW_READ`
- **Weight (IP)**: 1
- **Added**: 2023-11-10

#### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `startTime` | long | NO | Start timestamp (default: last 7 days if not provided) |
| `endTime` | long | NO | End timestamp |
| `page` | int | NO | Page number (default: 1) |
| `limit` | int | NO | Results per page (default: 10) |
| `tranId` | string | NO | Filter by transfer ID |
| `timestamp` | string | YES | Millisecond timestamp |
| `signature` | string | YES | HMAC-SHA256 signature |

**Note**: If `startTime` and `endTime` are not provided, defaults to returning data from the last 7 days.

#### Response Fields

| Name | Type | Description |
|------|------|-------------|
| `page` | int | Current page number |
| `totalRecords` | int | Total number of records |
| `totalPageNum` | int | Total number of pages |
| `data` | array | List of transfer records |
| `data[].tranId` | string | Transfer ID |
| `data[].asset` | string | Asset symbol |
| `data[].amount` | string | Transfer amount |
| `data[].toAccountType` | string | Recipient identifier type |
| `data[].toAccount` | string | Recipient account |
| `data[].fromAccount` | string | Sender account |
| `data[].status` | string | Transfer status: "SUCCESS", "FAILED", "WAIT" |
| `data[].timestamp` | long | Transfer timestamp |

#### Sample Response

```json
{
    "page": 1,
    "totalRecords": 2,
    "totalPageNum": 1,
    "data": [
        {
            "tranId": "11945860693",
            "asset": "BTC",
            "amount": "0.1",
            "toAccountType": "EMAIL",
            "toAccount": "user@example.com",
            "fromAccount": "sender@example.com",
            "status": "SUCCESS",
            "timestamp": 1544433325000
        },
        {
            "tranId": "",
            "asset": "BTC",
            "amount": "0.8",
            "toAccountType": "UID",
            "toAccount": "87658765",
            "fromAccount": "sender@example.com",
            "status": "SUCCESS",
            "timestamp": 1544433325000
        }
    ]
}
```

## Constraints & Limits

### Account Limits

| Constraint | Value |
|------------|-------|
| Max API Keys per account | 30 |
| API Key validity (no IP) | 90 days |
| Max IPs per API Key | 10 |
| Max sub-accounts per main account | 30 |
| Max valid orders per account | 500 |
| Max WebSocket channels per connection | 30 |

### Rate Limits

| Endpoint | Weight | Rate Limit |
|----------|--------|------------|
| `GET /api/v3/account` | 10 | 2 times/s |
| `GET /api/v3/myTrades` | 10 | - |
| `POST /api/v3/sub-account/virtualSubAccount` | 1 | - |
| `GET /api/v3/sub-account/list` | 1 | - |
| `GET /api/v3/sub-account/asset` | 1 | - |
| `POST /api/v3/capital/transfer` | 1 | - |
| `GET /api/v3/capital/transfer` | 1 | - |
| `POST /api/v3/capital/transfer/internal` | 1 | - |
| `GET /api/v3/capital/transfer/internal` | 1 | - |
| `POST /api/v3/capital/sub-account/universalTransfer` | 1 | - |
| `GET /api/v3/capital/sub-account/universalTransfer` | 1 | - |

**General Rate Limits**:
- IP-based: 500 requests per 10 seconds per endpoint
- UID-based: 500 requests per 10 seconds per endpoint
- HTTP 429 returned when rate limit exceeded
- IP bans scale from 2 minutes to 3 days for repeat offenders

### Error Codes

| Code | Description |
|------|-------------|
| `-2011` | Unknown order sent |
| `26` | Operation not allowed |
| `400` | API key required |
| `401` | No authority |
| `403` | Access Denied |
| `429` | Too Many Requests |
| `500` | Internal error |
| `503` | Service not available |
| `504` | Gateway Time-out |
| `602` | Signature verification failed |
| `10099` | User sub account does not open |
| `10100` | This currency transfer is not supported |
| `10101` | Insufficient balance |
| `10102` | Amount cannot be zero or negative |
| `10103` | This account transfer is not supported |
| `10200` | Transfer operation processing |
| `10201` | Transfer in failed |
| `10202` | Transfer out failed |
| `10206` | Transfer is disabled |
| `10211` | Transfer is forbidden |
| `700001` | API-key format invalid |
| `700002` | Signature for this request is not valid |
| `700003` | Timestamp outside of recvWindow |
| `700005` | recvWindow must be less than 60000 |
| `700006` | IP not in whitelist |
| `700007` | No permission to access the endpoint |
| `730600` | Sub-account Name cannot be null |
| `730601` | Sub-account Name must be 8-32 letters and numbers |
| `730602` | Sub-account remarks cannot be null |
| `730705` | At most 30 API Keys allowed |
| `140001` | Sub account does not exist |
| `140002` | Sub account is forbidden |

## Notes

### Account Types Supported
- **SPOT**: Spot trading account (fully supported)
- **FUTURES**: Futures/contract trading account (supported for transfers)

### Implementation Notes

1. **Sub-account login**: Sub-accounts created via API **cannot** be logged in on the web interface. They are designed for programmatic access only.

2. **Sub-account rates**: Sub-accounts automatically inherit the main account's trading fee rates.

3. **Master-only endpoints**: All sub-account management endpoints require a **master account API key**. Sub-account API keys cannot access these endpoints.

4. **Transfer directions**:
   - `SPOT` ↔ `FUTURES`: Universal transfers within the same account
   - Master ↔ Sub-account: Universal transfers for sub-account management
   - User ↔ User: Internal transfers via email, UID, or mobile

5. **Signature case**: The HMAC-SHA256 signature must be **lowercase only**.

6. **Timestamp validation**: Server time should be synced. Use `GET /api/v3/time` to check server time and adjust for clock skew.

7. **recvWindow recommendation**: Use a small `recvWindow` of 5000ms or less. The maximum allowed is 60000ms.

8. **Query limitations**:
   - Trade history: Only past 1 month via API (web export supports up to 3 years)
   - Transfer history: Only past 6 months
   - Internal transfer history: Defaults to last 7 days if no date range provided

9. **API Key permissions**: Available permission scopes include:
   - `SPOT_ACCOUNT_READ`, `SPOT_ACCOUNT_WRITE`
   - `SPOT_DEAL_READ`, `SPOT_DEAL_WRITE`
   - `CONTRACT_ACCOUNT_READ`, `CONTRACT_ACCOUNT_WRITE`
   - `CONTRACT_DEAL_READ`, `CONTRACT_DEAL_WRITE`
   - `SPOT_TRANSFER_READ`, `SPOT_TRANSFER_WRITE`

10. **IP restrictions**: Each API key can be bound to up to 10 IP addresses. It is recommended to set IP restrictions for security.
