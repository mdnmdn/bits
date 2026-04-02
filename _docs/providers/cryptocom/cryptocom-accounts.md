# Crypto.com Account REST API Documentation

## Reference
- **Official API Docs**: https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#account-balance-and-position-api
- **Base URL**: `https://api.crypto.com/exchange/v1`
- **UAT Sandbox**: `https://uat-api.3ona.co/exchange/v1`

## Authentication

All account endpoints are **private** methods requiring authentication.

### Required Fields

| Field     | Type   | Required | Description                                      |
|-----------|--------|----------|--------------------------------------------------|
| `api_key` | string | Yes      | API key generated from User Center → API         |
| `sig`     | string | Yes      | HMAC-SHA256 digital signature                    |
| `nonce`   | long   | Yes      | Current timestamp in milliseconds since epoch    |

### Signature Method

The signature is computed as:

```
sig = HMAC-SHA256(method + id + api_key + parameter_string + nonce, api_secret)
```

Where `parameter_string` is constructed by:
1. Sorting all parameter keys in ascending order
2. Concatenating each key + value with no delimiters
3. For nested objects/arrays, recursively flatten to depth 3

### Timestamp Requirements

- `nonce` must be within **60 seconds** of server time
- Error code `40102` (`INVALID_NONCE`) returned if nonce differs by more than 60 seconds

## Account APIs

### Get User Balance

- **Description**: Returns the user's wallet balance including margin, collateral, and position information.
- **Endpoint**: `private/user-balance`
- **Official Docs**: [private/user-balance](https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#private-user-balance)
- **Applies To**: REST, WebSocket (User API)
- **REST Method**: POST

#### Parameters

No parameters required, but an empty `params: {}` block must be included.

#### Response Fields

**Top-level result fields:**

| Field                         | Type    | Description                                                                 |
|-------------------------------|---------|-----------------------------------------------------------------------------|
| `instrument_name`             | string  | Instrument name of the balance, e.g. `USD`                                  |
| `total_available_balance`     | string  | Balance available to open new orders (Margin Balance - Initial Margin)      |
| `total_margin_balance`        | string  | Positive cash on eligible collateral + negative balance + unrealized PnL - fee reserves |
| `total_initial_margin`        | string  | Total margin requirement for positions and open orders (position_im + haircut) |
| `total_position_im`           | string  | Initial margin requirement to support open positions and orders             |
| `total_haircut`               | string  | Total haircut on eligible collateral token assets                           |
| `total_maintenance_margin`    | string  | Total maintenance margin requirement for all positions                      |
| `total_position_cost`         | string  | Position value in USD                                                       |
| `total_cash_balance`          | string  | Wallet Balance (Deposits - Withdrawals + Realized PnL - Fees)               |
| `total_collateral_value`      | string  | Total collateral value                                                      |
| `total_session_unrealized_pnl`| string  | Unrealized PnL from all open positions                                      |
| `total_session_realized_pnl`  | string  | Realized PnL from all open positions                                        |
| `is_liquidating`              | boolean | Whether the account is under liquidation                                    |
| `total_effective_leverage`    | string  | Actual leverage used (position size / margin balance)                       |
| `position_limit`              | string  | Maximum position size allowed                                               |
| `used_position_limit`         | string  | Combined position size of all open positions + order exposure               |
| `total_isolated_cash_balance` | string  | Total cash balance in isolated margin positions (added 2026-01-08)          |
| `position_balances`           | array   | Array of collateral balances (see below)                                    |
| `isolated_positions`          | array   | Array of isolated margin positions (added 2026-01-08)                       |

**`position_balances` array fields:**

| Field                   | Type    | Description                                               |
|-------------------------|---------|-----------------------------------------------------------|
| `instrument_name`       | string  | Instrument name of the collateral (e.g. `CRO`, `USDT`)    |
| `quantity`              | string  | Quantity of the collateral                                |
| `market_value`          | string  | Market value of the collateral                            |
| `collateral_eligible`   | boolean | Whether the token is eligible as collateral               |
| `haircut`               | string  | Haircut applied to eligible collateral token              |
| `collateral_amount`     | string  | Collateral amount (market_value minus haircut)            |
| `max_withdrawal_balance`| string  | Maximum withdrawal balance for the collateral             |
| `reserved_qty`          | string  | Balance in use, not available for new orders              |

#### Sample Response

```json
{
  "id": 11,
  "method": "private/user-balance",
  "code": 0,
  "result": {
    "data": [
      {
        "total_available_balance": "4721.05898582",
        "total_margin_balance": "7595.42571782",
        "total_initial_margin": "2874.36673202",
        "total_position_im": "486.31273202",
        "total_haircut": "2388.054",
        "total_maintenance_margin": "1437.18336601",
        "total_position_cost": "14517.54641301",
        "total_cash_balance": "7890.00320721",
        "total_collateral_value": "7651.18811483",
        "total_session_unrealized_pnl": "-55.76239701",
        "instrument_name": "USD",
        "total_session_realized_pnl": "0.00000000",
        "is_liquidating": false,
        "total_effective_leverage": "1.90401230",
        "position_limit": "3000000.00000000",
        "used_position_limit": "40674.69622001",
        "total_isolated_cash_balance": "0.00000000",
        "position_balances": [
          {
            "instrument_name": "CRO",
            "quantity": "24422.72427884",
            "market_value": "4776.107959969951",
            "collateral_eligible": true,
            "haircut": "0.5",
            "collateral_amount": "4537.302561971453",
            "max_withdrawal_balance": "24422.72427884",
            "reserved_qty": "0.00000000"
          },
          {
            "instrument_name": "USD",
            "quantity": "3113.50747209",
            "market_value": "3113.50747209",
            "collateral_eligible": true,
            "haircut": "0",
            "collateral_amount": "3113.50747209",
            "max_withdrawal_balance": "3113.50747209",
            "reserved_qty": "0.00000000"
          },
          {
            "instrument_name": "USDT",
            "quantity": "0.19411607",
            "market_value": "0.19389555414448",
            "collateral_eligible": true,
            "haircut": "0.02",
            "collateral_amount": "0.18904816529086801",
            "max_withdrawal_balance": "0.19411607",
            "reserved_qty": "0.00000000"
          }
        ],
        "isolated_positions": []
      }
    ]
  }
}
```

### Get User Balance History

- **Description**: Returns the user's balance history. May temporarily have discrepancies with the GUI.
- **Endpoint**: `private/user-balance-history`
- **Official Docs**: [private/user-balance-history](https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#private-user-balance-history)
- **Applies To**: REST
- **REST Method**: POST

#### Parameters

| Parameter   | Type   | Required | Description                                                                 |
|-------------|--------|----------|-----------------------------------------------------------------------------|
| `timeframe` | string | No       | `H1` for hourly, `D1` for daily. Defaults to `D1` if omitted                |
| `end_time`  | number | No       | Exclusive end time in milliseconds or nanoseconds. Defaults to current time  |
| `limit`     | int    | No       | Max `30` for `D1`, max `120` for `H1`                                       |

#### Response Fields

| Field            | Type   | Description                                              |
|------------------|--------|----------------------------------------------------------|
| `instrument_name`| string | Instrument name of the balance, e.g. `USD`               |
| `t`              | number | Timestamp                                                |
| `c`              | string | Total cash balance at that point                         |

#### Sample Response

```json
{
  "id": 11,
  "method": "private/user-balance-history",
  "code": 0,
  "result": {
    "instrument_name": "USD",
    "data": [
      {
        "t": 1629478800000,
        "c": "811.621851"
      }
    ]
  }
}
```

### Get Accounts

- **Description**: Get the master account and its sub-accounts with pagination support.
- **Endpoint**: `private/get-accounts`
- **Official Docs**: [private/get-accounts](https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#private/get-accounts)
- **Applies To**: REST
- **REST Method**: POST

#### Parameters

| Parameter  | Type | Required | Description                              |
|------------|------|----------|------------------------------------------|
| `page_size`| int  | No       | Number of results per page (default: 20) |
| `page`     | int  | No       | Page number (default: 0)                 |

#### Response Fields

**Top-level result fields:**

| Field             | Type   | Description                    |
|-------------------|--------|--------------------------------|
| `master_account`  | object | Master account details         |
| `sub_account_list`| array  | Array of sub-account objects   |

**Account object fields (both `master_account` and items in `sub_account_list`):**

| Field                  | Type    | Description                                             |
|------------------------|---------|---------------------------------------------------------|
| `uuid`                 | string  | Account UUID                                            |
| `master_account_uuid`  | string  | Master account UUID                                     |
| `margin_account_uuid`  | string  | (Optional) Margin account UUID                          |
| `label`                | string  | Sub-account label                                       |
| `enabled`              | boolean | Whether the account is enabled                          |
| `tradable`             | boolean | Whether the account can trade                           |
| `name`                 | string  | Account name                                            |
| `email`                | string  | Account email                                           |
| `mobile_number`        | string  | Mobile number                                           |
| `country_code`         | string  | Country code                                            |
| `address`              | string  | Address                                                 |
| `margin_access`        | string  | `DEFAULT` or `DISABLED`                                 |
| `derivatives_access`   | string  | `DEFAULT` or `DISABLED`                                 |
| `create_time`          | number  | Creation timestamp (milliseconds since epoch)           |
| `update_time`          | number  | Last update timestamp (milliseconds since epoch)        |
| `two_fa_enabled`       | boolean | Whether 2FA is enabled                                  |
| `kyc_level`            | string  | KYC level (e.g. `ADVANCED`)                             |
| `suspended`            | boolean | Whether the account is suspended                        |
| `terminated`           | boolean | Whether the account is terminated                       |

#### Sample Response

```json
{
  "id": 12,
  "method": "private/get-accounts",
  "code": 0,
  "result": {
    "master_account": {
      "uuid": "243d3f39-b193-4eb9-1d60-e98f2fc17707",
      "master_account_uuid": "291879ae-b769-4eb3-4d75-3366ebee7dd6",
      "margin_account_uuid": "69c9ab41-5b95-4d75-b769-e45f2fc16507",
      "enabled": true,
      "tradable": true,
      "name": "",
      "email": "user@example.com",
      "mobile_number": "",
      "country_code": "",
      "address": "",
      "margin_access": "DEFAULT",
      "derivatives_access": "DISABLED",
      "create_time": 1620962543792,
      "update_time": 1622019525960,
      "two_fa_enabled": true,
      "kyc_level": "ADVANCED",
      "suspended": false,
      "terminated": false
    },
    "sub_account_list": [
      {
        "uuid": "a0d206a1-6b06-47c5-9cd3-8bc6ef0915c5",
        "master_account_uuid": "291879ae-b769-4eb3-4d75-3366ebee7dd6",
        "margin_account_uuid": "69c9ab41-5b95-4d75-b769-e45f2fc16507",
        "label": "Sub Account",
        "enabled": true,
        "tradable": true,
        "name": "",
        "email": "sub@example.com",
        "mobile_number": "",
        "country_code": "",
        "address": "",
        "margin_access": "DEFAULT",
        "derivatives_access": "DISABLED",
        "create_time": 1620962543792,
        "update_time": 1622019525960,
        "two_fa_enabled": true,
        "kyc_level": "ADVANCED",
        "suspended": false,
        "terminated": false
      }
    ]
  }
}
```

### Get Positions

- **Description**: Returns the user's open positions across all instruments.
- **Endpoint**: `private/get-positions`
- **Official Docs**: [private/get-positions](https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#private-get-positions)
- **Applies To**: REST, WebSocket (User API)
- **REST Method**: POST

#### Parameters

| Parameter        | Type   | Required | Description                        |
|------------------|--------|----------|------------------------------------|
| `instrument_name`| string | No       | Filter by instrument, e.g. `BTCUSD-PERP` |

#### Response Fields

| Field                | Type   | Description                                              |
|----------------------|--------|----------------------------------------------------------|
| `account_id`         | string | Account UUID                                             |
| `instrument_name`    | string | Instrument name, e.g. `BTCUSD-PERP`                      |
| `type`               | string | Position type, e.g. `PERPETUAL_SWAP`                     |
| `quantity`           | string | Position quantity (negative for short)                   |
| `cost`               | string | Position cost or value in USD                            |
| `open_position_pnl`  | string | Profit and loss for the open position                    |
| `open_pos_cost`      | string | Open position cost                                       |
| `session_pnl`        | string | PnL in the current trading session                       |
| `update_timestamp_ms`| number | Last update time (Unix timestamp in milliseconds)        |
| `isolation_id`       | string | Isolation ID for isolated margin positions (added 2026-01-08) |
| `isolation_type`     | string | Isolation type, e.g. `ISOLATED_MARGIN` (added 2026-01-08)     |

#### Sample Response

```json
{
  "id": 1,
  "method": "private/get-positions",
  "code": 0,
  "result": {
    "data": [
      {
        "account_id": "858dbc8b-22fd-49fa-bff4-d342d98a8acb",
        "quantity": "-0.1984",
        "cost": "-10159.573500",
        "open_position_pnl": "-497.743736",
        "open_pos_cost": "-10159.352200",
        "session_pnl": "2.236145",
        "update_timestamp_ms": 1613552240770,
        "instrument_name": "BTCUSD-PERP",
        "type": "PERPETUAL_SWAP"
      },
      {
        "account_id": "858dbc8b-22fd-49fa-bff4-d342d98a8acb",
        "quantity": "-0.1984",
        "cost": "-10159.573500",
        "open_position_pnl": "-497.743736",
        "open_pos_cost": "-10159.352200",
        "session_pnl": "2.236145",
        "update_timestamp_ms": 1613552240771,
        "instrument_name": "BTCUSD-PERP",
        "type": "PERPETUAL_SWAP",
        "isolation_id": "19848526",
        "isolation_type": "ISOLATED_MARGIN"
      }
    ]
  }
}
```

## Subaccount APIs

### Get Sub-Account Balances

- **Description**: Returns the wallet balances of all sub-accounts under the master account.
- **Endpoint**: `private/get-subaccount-balances`
- **Official Docs**: [private/get-subaccount-balances](https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#private-get-subaccount-balances)
- **Applies To**: REST
- **REST Method**: POST

#### Parameters

No parameters required, but an empty `params: {}` block must be included.

#### Response Fields

**Top-level result fields:**

| Field                         | Type   | Description                                              |
|-------------------------------|--------|----------------------------------------------------------|
| `data`                        | array  | Array of sub-account balance objects                     |

**Each item in `data` array:**

| Field                         | Type    | Description                                              |
|-------------------------------|---------|----------------------------------------------------------|
| `account`                     | string  | Sub-account UUID                                         |
| `instrument_name`             | string  | Instrument name of the balance, e.g. `USD`               |
| `total_available_balance`     | string  | Balance available to open new orders                     |
| `total_margin_balance`        | string  | Margin balance                                           |
| `total_initial_margin`        | string  | Total initial margin requirement                         |
| `total_maintenance_margin`    | string  | Total maintenance margin requirement                     |
| `total_position_cost`         | string  | Position value in USD                                    |
| `total_cash_balance`          | string  | Wallet balance                                           |
| `total_collateral_value`      | string  | Total collateral value                                   |
| `total_session_unrealized_pnl`| string  | Unrealized PnL                                           |
| `total_session_realized_pnl`  | string  | Realized PnL                                             |
| `total_effective_leverage`    | string  | Actual leverage used                                     |
| `position_limit`              | string  | Maximum position size allowed                            |
| `used_position_limit`         | string  | Used position limit                                      |
| `total_isolated_cash_balance` | string  | Cash balance in isolated margin (added 2026-01-08)       |
| `is_liquidating`              | boolean | Whether the account is under liquidation                 |
| `position_balances`           | array   | Array of collateral balances (same schema as user-balance) |
| `isolated_positions`          | array   | Array of isolated margin positions (added 2026-01-08)    |

#### Sample Response

```json
{
  "id": 1,
  "method": "private/get-subaccount-balances",
  "code": 0,
  "result": {
    "data": [
      {
        "account": "a0d206a1-6b06-47c5-9cd3-8bc6ef0915c5",
        "instrument_name": "USD",
        "total_available_balance": "0.00000000",
        "total_margin_balance": "0.00000000",
        "total_initial_margin": "0.00000000",
        "total_maintenance_margin": "0.00000000",
        "total_position_cost": "0.00000000",
        "total_cash_balance": "0.00000000",
        "total_collateral_value": "0.00000000",
        "total_session_unrealized_pnl": "0.00000000",
        "total_session_realized_pnl": "0.00000000",
        "total_effective_leverage": "0.00000000",
        "position_limit": "3000000.00000000",
        "used_position_limit": "0.00000000",
        "total_isolated_cash_balance": "0.00000000",
        "is_liquidating": false,
        "position_balances": [],
        "isolated_positions": []
      },
      {
        "account": "49786818-6ead-40c4-a008-ea6b0fa5cf96",
        "instrument_name": "USD",
        "total_available_balance": "20823.62250000",
        "total_margin_balance": "20823.62250000",
        "total_initial_margin": "0.00000000",
        "total_maintenance_margin": "0.00000000",
        "total_position_cost": "0.00000000",
        "total_cash_balance": "21919.55000000",
        "total_collateral_value": "20823.62250000",
        "total_session_unrealized_pnl": "0.00000000",
        "total_session_realized_pnl": "0.00000000",
        "total_effective_leverage": "0.00000000",
        "position_limit": "3000000.00000000",
        "used_position_limit": "0.00000000",
        "total_isolated_cash_balance": "0.00000000",
        "is_liquidating": false,
        "position_balances": [
          {
            "instrument_name": "BTC",
            "quantity": "1.0000000000",
            "market_value": "21918.5500000000",
            "collateral_eligible": true,
            "haircut": "0.5500000000",
            "collateral_amount": "21918.0000000000",
            "max_withdrawal_balance": "1.0000000000"
          },
          {
            "instrument_name": "USD",
            "quantity": "1.00000000",
            "market_value": "1.00000000",
            "collateral_eligible": true,
            "haircut": "0.10000000",
            "collateral_amount": "0.90000000",
            "max_withdrawal_balance": "0.00000000"
          }
        ],
        "isolated_positions": []
      }
    ]
  }
}
```

## Transfer APIs

### Create Sub-Account Transfer

- **Description**: Transfer funds between sub-accounts and the master account. Transfers are internal only (no blockchain transaction).
- **Endpoint**: `private/create-subaccount-transfer`
- **Official Docs**: [private/create-subaccount-transfer](https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#private-create-subaccount-transfer)
- **Applies To**: REST
- **REST Method**: POST

#### Parameters

| Parameter  | Type   | Required | Description                                                        |
|------------|--------|----------|--------------------------------------------------------------------|
| `from`     | string | Yes      | Account UUID to be debited (master account UUID or sub-account UUID) |
| `to`       | string | Yes      | Account UUID to be credited (master account UUID or sub-account UUID) |
| `currency` | string | Yes      | Currency symbol (e.g. `CRO`, `BTC`, `USDT`)                        |
| `amount`   | string | Yes      | Amount to transfer — must be a positive number                     |

#### Response Fields

| Field  | Type   | Description                                              |
|--------|--------|----------------------------------------------------------|
| `code` | number | `0` for successful transfer, otherwise an error code     |

#### Sample Request

```json
{
  "id": 1234,
  "method": "private/create-subaccount-transfer",
  "params": {
    "from": "12345678-0000-0000-0000-000000000001",
    "to": "12345678-0000-0000-0000-000000000002",
    "currency": "CRO",
    "amount": "500"
  },
  "nonce": 1587846358253
}
```

#### Sample Response

```json
{
  "id": 1234,
  "method": "private/create-subaccount-transfer",
  "code": 0
}
```

## Constraints & Limits

### Rate Limits

All account endpoints fall under the "All others" category for authenticated REST calls:

| Category   | Limit                      |
|------------|----------------------------|
| Account endpoints | 3 requests per 100ms per API key |

### Nonce Window

- Nonce must be within **60 seconds** of server time
- Error code `40102` (`INVALID_NONCE`) if outside this window

### Pagination

- `private/get-accounts`: default `page_size=20`, `page=0`

### Number Formatting

- **All numeric values in responses are strings**, wrapped in double quotes (e.g. `"12.34"`, not `12.34`)
- Request parameters that represent numbers must also be strings

## Notes

### Account Types

- **Master Account**: The primary account that owns sub-accounts
- **Sub-Accounts**: Child accounts under a master account, each with their own balances and positions
- **Cross Margin**: Default margin mode where all collateral is pooled
- **Isolated Margin**: Per-position margin mode (added 2026-01-08), identified by `isolation_id` and `isolation_type`

### Key Implementation Notes

1. **Empty params block**: Even when no parameters are required, always include `params: {}` in requests for API consistency
2. **Unified wallet**: The exchange uses a unified wallet system — there are no separate spot/margin wallets
3. **Balance currency**: All balance responses are denominated in `USD` (the USD stablecoin bundle)
4. **Collateral eligibility**: Not all tokens are eligible as collateral; check `collateral_eligible` field
5. **Isolated margin fields**: Fields `total_isolated_cash_balance`, `isolated_positions`, `isolation_id`, and `isolation_type` were added on 2026-01-08 and may not be present in older API versions
6. **Transfer scope**: `private/create-subaccount-transfer` only supports internal transfers between accounts under the same master account — it does not support external deposits or withdrawals
7. **Balance history discrepancies**: The `private/user-balance-history` endpoint may temporarily show discrepancies with the GUI

### Related WebSocket Subscriptions

| Subscription              | Description                        |
|---------------------------|------------------------------------|
| `user.balance`            | Real-time balance updates          |
| `user.positions`          | Real-time position updates         |
| `user.account_risk`       | Account risk metrics updates       |
| `user.position_balance`   | Position balance updates           |
