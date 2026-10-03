# Bitvavo API Documentation

## Overview

- **Official exchange**: https://bitvavo.com
- **API Docs**: https://docs.bitvavo.com
- **Provider ID**: `bitvavo` (alias `bv`)
- **Implementation**: Manual HTTP client, public spot data only (`provider/bitvavo/`)

## Base URL

`https://api.bitvavo.com/v2`. Override with `bitvavo.base_url` / `BITS_BITVAVO_BASE_URL`.

## Authentication

None. Public endpoints only. Fees are published only by the authenticated `/account/fees`, so `MakerFee` and `TakerFee` stay unset; the market `feeCategory` is kept in `Extra["fee_category"]`.

## Products

Spot only (`BTC-EUR` style, already the canonical bits form). Bitvavo offers no margin and no futures or perpetuals (the public REST API at https://docs.bitvavo.com has spot endpoints only).

## Capabilities

`server_time`, `exchange_info`, `price`, `ticker_24h`, `order_book`, `candles` (spot). No stream, trading or account.

## Exchange Info APIs

| Capability | Endpoint |
|------------|----------|
| Server time | `GET /time` |
| Exchange info | `GET /markets` |

Mapping:

- Tick size: `MinPrice` (smallest price increment), not `Extra["tick_size"]`. `PricePrecision` is the tick's decimals (the API `pricePrecision` is always null). The tick depends on the price magnitude and can change.
- `QtyPrecision` = `quantityDecimals`; `StepSize` = 10^-`quantityDecimals`.
- `MinQty` / `MaxQty` = `minOrderInBaseAsset` / `maxOrderInBaseAsset`.
- `MinNotional` = `minOrderInQuoteAsset` (available). Max notional in `Extra["max_notional"]`.
- Other `Extra`: `fee_category`, `notional_decimals`, `order_types`, `status_raw`.
