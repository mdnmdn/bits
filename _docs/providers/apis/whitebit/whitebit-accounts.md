# WhiteBit Account REST API Documentation

## Reference

- **Official API Docs**: https://docs.whitebit.com/api-reference/overview
- **Account & Wallet**: https://docs.whitebit.com/api-reference/account-wallet/overview
- **Sub-Accounts**: https://docs.whitebit.com/api-reference/sub-accounts/overview
- **Authentication**: https://docs.whitebit.com/api-reference/authentication
- **Base URL**: `https://whitebit.com` (global) or `https://whitebit.eu` (EU region)
- **API Version**: V4
- **API Prefix**: `/api/v4/`

## Authentication

All private endpoints require authentication via HMAC-SHA512 signed requests.

### Required Headers

| Header | Value | Description |
|--------|-------|-------------|
| `Content-type` | `application/json` | Specifies JSON format |
| `X-TXC-APIKEY` | Your API key | The public WhiteBIT API key |
| `X-TXC-PAYLOAD` | Base64-encoded request body | Base64 encoding of the JSON request body |
| `X-TXC-SIGNATURE` | HMAC-SHA512 hex signature | `hex(HMAC_SHA512(payload, key=api_secret))` |

### Request Body Format

Every authenticated request must include these fields in the JSON body:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `request` | string | Yes | Request path without domain (e.g., `/api/v4/main-account/balance`) |
| `nonce` | string/integer | Yes | Incrementing number larger than previous requests (Unix timestamp in ms recommended) |
| `nonceWindow` | boolean | No | Enable time-based nonce validation (±5 seconds) |

### Signature Method

```
signature = hex(HMAC_SHA512(raw_payload_string, api_secret))
```

### Notes

- All authenticated requests use the `POST` HTTP method
- Nonce must be larger than previous requests; use Unix timestamp in milliseconds
- When `nonceWindow` is enabled, timestamp must be within ±5 seconds of server time
- API keys can be generated at https://whitebit.com/settings/api

## Account APIs

### Get Account Balances

Retrieves the main account balance by currency ticker or all balances.

- **Endpoint**: `POST /api/v4/main-account/balance`
- **Rate Limit**: 1000 requests / 10 sec
- **Docs**: https://docs.whitebit.com/api-reference/account-wallet/main-balance

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ticker` | string | No | Currency ticker (e.g., `BTC`). Omit to get all balances |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `{ticker}` | object | Keyed by currency ticker |
| `main_balance` | string | Main balance volume for the currency |

#### Sample Response

```json
{
  "BSV": { "main_balance": "0" },
  "BTC": { "main_balance": "0" },
  "BTG": { "main_balance": "0" },
  "BTT": { "main_balance": "0" },
  "XLM": { "main_balance": "36.48" }
}
```

### Get Account Info

WhiteBit does not expose a dedicated "account info" endpoint in the V4 REST API. Account details (KYC status, email, etc.) are available through the sub-account listing endpoint for sub-accounts, but the main account profile is managed via the web UI.

- **Status**: Not available via REST API
- **Workaround**: Use the web interface at https://whitebit.com/settings/profile

## Subaccount APIs

### Create Sub-Account

Creates a new sub-account under the main account.

- **Endpoint**: `POST /api/v4/sub-account/create`
- **Rate Limit**: 1000 requests / 10 sec
- **Docs**: https://docs.whitebit.com/api-reference/sub-accounts/create-sub-account

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `alias` | string | Yes | Name for the sub-account |
| `email` | string | Conditional | Required when `shareKyc` is `false` or not provided |
| `shareKyc` | boolean | No | If `true`, KYC is shared with main account (makes `email` optional) |
| `permissions.spotEnabled` | boolean | Yes | Enable transfers to trade (spot) balance |
| `permissions.collateralEnabled` | boolean | Yes | Enable transfers to collateral balance |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Sub-account identifier (UUID) |
| `alias` | string | Sub-account alias/name |
| `userId` | string | User identifier associated with account |
| `email` | string | Sub-account email (masked) |
| `status` | string | Sub-account status (e.g., `active`) |
| `color` | string | Sub-account color |
| `kyc.shareKyc` | boolean | Whether KYC is shared with main account |
| `kyc.kycStatus` | string | KYC status |
| `permissions.spotEnabled` | boolean | Spot trading enabled |
| `permissions.collateralEnabled` | boolean | Collateral trading enabled |

#### Sample Response

```json
{
  "id": "8e667b4a-0b71-4988-8af5-9474dbfaeb51",
  "alias": "trading_bot",
  "userId": "u-12345",
  "email": "s***@example.com",
  "status": "active",
  "color": "#FF5733",
  "kyc": {
    "shareKyc": false,
    "kycStatus": "verified"
  },
  "permissions": {
    "spotEnabled": true,
    "collateralEnabled": false
  }
}
```

### List Sub-Accounts

Returns a paginated list of all sub-accounts for the current user.

- **Endpoint**: `POST /api/v4/sub-account/list`
- **Rate Limit**: 1000 requests / 10 sec
- **Docs**: https://docs.whitebit.com/api-reference/sub-accounts/list-of-sub-accounts

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `search` | string | No | - | Search term to filter sub-accounts |
| `limit` | integer | No | 30 | Number of results (1-100) |
| `offset` | integer | No | 0 | Offset for pagination (0-10000) |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `offset` | integer | Current offset |
| `limit` | integer | Current limit |
| `data` | array | Array of sub-account objects (see Create Sub-Account response schema) |

#### Sample Response

```json
{
  "offset": 0,
  "limit": 30,
  "data": [
    {
      "id": "8e667b4a-0b71-4988-8af5-9474dbfaeb51",
      "alias": "trading_bot",
      "userId": "u-12345",
      "email": "s***@example.com",
      "status": "active",
      "color": "#FF5733",
      "kyc": {
        "shareKyc": false,
        "kycStatus": "verified"
      },
      "permissions": {
        "spotEnabled": true,
        "collateralEnabled": false
      }
    }
  ]
}
```

### Get Sub-Account Balances

Returns balances for a specific sub-account across main, spot, and collateral wallets.

- **Endpoint**: `POST /api/v4/sub-account/balances`
- **Rate Limit**: 1000 requests / 10 sec
- **Docs**: https://docs.whitebit.com/api-reference/sub-accounts/sub-account-balances

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Sub-account ID (UUID) |
| `ticker` | string | No | Currency ticker. Omit to get all balances |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `{ticker}` | array | Keyed by currency ticker, contains balance objects |
| `main` | string | Balance in main wallet |
| `spot` | string | Balance in spot (trade) wallet |
| `collateral` | string | Balance in collateral wallet |

#### Sample Response

```json
{
  "USDC": [
    {
      "main": "42",
      "spot": "10",
      "collateral": "14"
    }
  ],
  "BTC": [
    {
      "main": "0.005",
      "spot": "0.01",
      "collateral": "0"
    }
  ]
}
```

## Transfer APIs

### Internal Transfer (Between Balances)

Transfers funds between main, spot (trade), and collateral balances within the main account.

- **Endpoint**: `POST /api/v4/main-account/transfer`
- **Rate Limit**: 1000 requests / 10 sec
- **Docs**: https://docs.whitebit.com/api-reference/account-wallet/transfer-between-balances

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ticker` | string | Yes | Currency ticker (e.g., `BTC`) |
| `amount` | string | Yes | Amount to transfer (max 8 decimal precision) |
| `from` | string | Conditional | Source balance: `main`, `spot`, or `collateral` |
| `to` | string | Conditional | Destination balance: `main`, `spot`, or `collateral` |
| `method` | string | Conditional | **Deprecated**. Use `from`/`to` instead. Values: `deposit`, `withdraw`, `collateral-deposit`, `collateral-withdraw` |

#### Response Fields

Returns an empty array `[]` on success.

#### Sample Response

```json
[]
```

#### Error Response (422)

```json
{
  "code": 3,
  "message": "Inner validation failed",
  "errors": {
    "amount": [
      "You don't have such amount for transfer (available 34.68, in amount: 1000000)"
    ]
  }
}
```

### Sub-Account Transfer

Creates a transfer between the main account and a sub-account.

- **Endpoint**: `POST /api/v4/sub-account/transfer`
- **Rate Limit**: 1000 requests / 10 sec
- **Docs**: https://docs.whitebit.com/api-reference/sub-accounts/sub-account-transfer

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Sub-account ID (UUID) |
| `direction` | string | Yes | Transfer direction: `main_to_sub` or `sub_to_main` |
| `amount` | string | Yes | Transfer amount (min 0.00000001) |
| `ticker` | string | Yes | Currency ticker |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `transaction_id` | string | External identifier of the transaction |

#### Sample Response

```json
{
  "transaction_id": "tx-abc123-def456"
}
```

### Transfer History (Main Account)

Retrieves the history of deposits and withdrawals for the main account.

- **Endpoint**: `POST /api/v4/main-account/history`
- **Rate Limit**: 200 requests / 10 sec
- **Docs**: https://docs.whitebit.com/api-reference/account-wallet/get-deposit-withdraw-history

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `transactionMethod` | integer | No | - | `1` for deposits, `2` for withdrawals. Omit for both |
| `ticker` | string | No | - | Filter by currency ticker |
| `address` | string | No | - | Filter by specific address |
| `memo` | string | No | - | Filter by memo/destination tag |
| `addresses` | array | No | - | Filter by array of addresses (max 20) |
| `unique_id` | string | No | - | Filter by unique ID |
| `limit` | integer | No | 50 | Number of results (1-100) |
| `offset` | integer | No | 0 | Offset for pagination (0-10000) |
| `status` | array | No | - | Filter by status codes |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `limit` | integer | Current limit |
| `offset` | integer | Current offset |
| `total` | integer | Total number of transactions |
| `records` | array | Array of transaction records |

**Transaction Record Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `address` | string | Deposit/Withdraw address |
| `unique_id` | string | Unique ID of deposit/withdraw |
| `createdAt` | integer | Unix timestamp |
| `currency` | string | Currency full name |
| `ticker` | string | Currency ticker |
| `method` | integer | `1` = deposit, `2` = withdraw |
| `amount` | string | Transaction amount |
| `description` | string | Transaction description |
| `memo` | string | Memo/destination tag |
| `fee` | string | Transaction fee |
| `status` | integer | Status code (see below) |
| `network` | string | Network (for multi-network currencies) |
| `transactionHash` | string | Blockchain transaction hash |
| `transaction_id` | string | Internal transaction ID |
| `confirmations.actual` | integer | Actual confirmations |
| `confirmations.required` | integer | Required confirmations |

**Deposit Status Codes:**
- `3`, `7` — Successful
- `4`, `9` — Canceled
- `5` — Unconfirmed by user
- `15` — Pending
- `21` — Additional data required
- `22` — Uncredited

**Withdraw Status Codes:**
- `3`, `7` — Successful
- `4` — Canceled
- `5` — Unconfirmed by user
- `1`, `2`, `6`, `10`, `11`, `12`, `13`, `14`, `15`, `16`, `17` — Pending
- `18` — Partially successful
- `21` — Additional data required

#### Sample Response

```json
{
  "limit": 100,
  "offset": 0,
  "total": 300,
  "records": [
    {
      "address": "3ApEASLcrQtZpg1TsssFgYF5V5YQJAKvuE",
      "unique_id": null,
      "createdAt": 1593437922,
      "currency": "Bitcoin",
      "ticker": "BTC",
      "method": 1,
      "amount": "0.0006",
      "description": "",
      "memo": "",
      "fee": "0",
      "status": 15,
      "network": null,
      "transactionHash": "a275a514013e4e0f927fd0d1bed215e7f6f2c4c6ce762836fe135ec22529d886",
      "transaction_id": "5e112b38-9652-11ed-a1eb-0242ac120002",
      "details": {
        "partial": {
          "requestAmount": "50000",
          "processedAmount": "39000",
          "processedFee": "273",
          "normalizeTransaction": ""
        }
      },
      "confirmations": {
        "actual": 1,
        "required": 2
      }
    }
  ]
}
```

### Sub-Account Transfer History

Returns the history of transfers between the main account and a sub-account.

- **Endpoint**: `POST /api/v4/sub-account/transfer/history`
- **Rate Limit**: 1000 requests / 10 sec
- **Docs**: https://docs.whitebit.com/api-reference/sub-accounts/get-sub-account-transfer-history

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `id` | string | Yes | - | Sub-account ID (UUID) |
| `direction` | string | No | - | Filter by direction: `main_to_sub` or `sub_to_main` |
| `limit` | integer | No | 30 | Number of results (1-100) |
| `offset` | integer | No | 0 | Offset for pagination (0-10000) |

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `offset` | integer | Current offset |
| `limit` | integer | Current limit |
| `data` | array | Array of transfer records |

**Transfer Record Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `transaction_id` | string | Transaction identifier |
| `id` | string | Transfer identifier (deprecated, same as `transaction_id`) |
| `direction` | string | `main_to_sub` or `sub_to_main` |
| `currency` | string | Currency ticker |
| `amount` | string | Transfer amount |
| `createdAt` | integer | Unix timestamp |

#### Sample Response

```json
{
  "offset": 0,
  "limit": 30,
  "data": [
    {
      "transaction_id": "tx-abc123",
      "id": "tx-abc123",
      "direction": "main_to_sub",
      "currency": "ETH",
      "amount": "0.5",
      "createdAt": 1641081600
    }
  ]
}
```

## Constraints & Limits

### Rate Limits

| Endpoint | Rate Limit |
|----------|------------|
| `POST /api/v4/main-account/balance` | 1000 requests / 10 sec |
| `POST /api/v4/main-account/transfer` | 1000 requests / 10 sec |
| `POST /api/v4/main-account/history` | 200 requests / 10 sec |
| `POST /api/v4/sub-account/create` | 1000 requests / 10 sec |
| `POST /api/v4/sub-account/list` | 1000 requests / 10 sec |
| `POST /api/v4/sub-account/balances` | 1000 requests / 10 sec |
| `POST /api/v4/sub-account/transfer` | 1000 requests / 10 sec |
| `POST /api/v4/sub-account/transfer/history` | 1000 requests / 10 sec |

### General Constraints

- All private endpoints require `POST` method with JSON body
- Nonce values must be strictly incrementing
- Amount precision: maximum 8 decimal places
- Pagination: `offset` max 10000, `limit` max 100
- Responses are not cached by the API
- All timestamps are in Unix-time format (seconds)

## Notes

### Account Types / Balance Types

WhiteBit supports three balance types:

| Type | Description |
|------|-------------|
| `main` | Main wallet — for deposits and withdrawals |
| `spot` | Spot (trade) wallet — for trading on spot markets |
| `collateral` | Collateral wallet — for futures/margin trading |

### Implementation Notes

1. **No dedicated account info endpoint**: Main account profile information is not exposed via the V4 REST API. Manage via the web UI.

2. **Balance response format**: Main account balances return a flat object keyed by ticker. Sub-account balances return an array per ticker with `main`, `spot`, and `collateral` breakdowns.

3. **Transfer methods**: The `method` field (`deposit`/`withdraw`) for internal transfers is deprecated. Use `from`/`to` fields instead for clarity and future compatibility.

4. **Fiat currency restrictions**: Fiat currencies cannot be transferred between balances without KYC verification.

5. **Empty balances omitted**: The main balance endpoint only returns currencies with non-zero balances.

6. **Sub-account email requirement**: When creating a sub-account with `shareKyc: true`, the `email` field becomes optional since KYC is inherited from the main account.

7. **Error format**: Private endpoints return errors with `code`, `message`, and `errors` fields. The `errors` object maps field names to arrays of error messages.

8. **Transfer correlation**: Use `transaction_id` from the sub-account transfer response to correlate with entries in the transfer history endpoint.
