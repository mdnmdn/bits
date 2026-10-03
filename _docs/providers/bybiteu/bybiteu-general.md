# Bybit EU API Documentation

## Overview

- **Official exchange**: https://www.bybit.eu
- **API Docs**: https://bybit-exchange.github.io/docs/ (V5)
- **Provider ID**: `bybiteu` (alias `bybit`)
- **Implementation**: Manual HTTP client, public spot and spot margin data (`provider/bybiteu/`)
- **Raw notes**: `_drafts/bybit-apis.md` (SDK survey, kept as draft)

## Base URL

`https://api.bybit.eu` (EEA host; same data endpoints as `api.bybit.com`). Override with `bybiteu.base_url` / `BITS_BYBITEU_BASE_URL`.

## Authentication

None. `/v5/account/fee-rate` needs credentials, so `MakerFee` and `TakerFee` stay unset.

## Products

Spot and spot margin (`category=spot`, native `BTCUSDT`). Margin has no separate data path: it reads the same `category=spot` endpoints and only labels the result `margin`; margin `exchange_info` keeps the instruments whose `marginTrading` is not `none`.

Not offered: derivatives. The EU entity (MiCA authorisation) lists spot, spot margin (up to 10x, since Aug 2025), Earn and the card; perpetuals and futures wait for a MiFID II licence. The V5 guide names `api.bybit.eu` for EEA users but does not list categories. Sources: https://leodex.io/learn/country-restrictions/bybit-eu-mica-migration, https://www.financemagnates.com/cryptocurrency/bybit-pursues-mifid-license-for-eu-derivatives-phases-out-mt5-for-in-house-trafi/, https://bybit-exchange.github.io/docs/v5/guide. So there is no `category=linear` code and no `funding_rates`; if the EU entity gets perpetuals, add futures then (the same V5 shapes as `api.bybit.com` apply).

## Capabilities

`server_time`, `exchange_info`, `price`, `ticker_24h`, `order_book`, `candles`, for spot and margin. No futures, funding rates, stream, trading or account.

## Exchange Info APIs

| Capability | Endpoint |
|------------|----------|
| Server time | `GET /v5/market/time` (`timeNano`) |
| Exchange info | `GET /v5/market/instruments-info?category=spot&limit=1000` (cursor paged) |

Mapping:

- Tick size: `Extra["tick_size"]`; `PricePrecision` is its decimals.
- `StepSize` = `basePrecision`, `MinQty` / `MaxQty` = `minOrderQty` / `maxOrderQty`.
- `MinNotional` = `minOrderAmt` (available when set).
