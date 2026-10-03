# Kraken API Documentation (futures, spot, margin)

## Overview

- **Official exchange**: https://futures.kraken.com (futures), https://api.kraken.com (spot, margin)
- **API Docs**: https://docs.kraken.com/api/docs/futures-api/ and https://docs.kraken.com/api/docs/rest-api/get-ticker-information
- **Provider ID**: `kraken` (alias `kraken-futures`)
- **Implementation**: Manual HTTP client, public market data only (`provider/kraken/`), no streams.
- **Raw notes**: `_drafts/kraken-apis.md` (spot SDK survey, kept as draft)

## Base URL

Futures `https://futures.kraken.com`: override with `kraken.base_url` / `BITS_KRAKEN_BASE_URL`.
Spot and margin `https://api.kraken.com` (`/0/public/...`): override with `kraken.spot_base_url` / `BITS_KRAKEN_SPOT_BASE_URL`.

Market flags `kraken.{spot,margin,futures}.enabled` / `BITS_KRAKEN_{SPOT,MARGIN,FUTURES}_ENABLED` select what `Capabilities()` declares. With none set, all three are declared.

## Authentication

None.

## Products

Markets `futures`, `spot`, `margin`. Spot native symbols are pair altnames: `XBTUSD`, `ETHEUR`, `XDGUSD` (BTC is `XBT`, DOGE is `XDG`). `BTC-USD`, `BTC/USD`, `BTC_USD` and `BTCUSD` are converted (`BTC-USD` is `XBTUSD`). Output `Symbol` is `BASE-QUOTE` with Kraken codes normalised (`XXBT`/`XBT` are BTC, `ZUSD` is USD).

Margin has no endpoints of its own: Kraken spot margin trades the spot pairs (`leverage_buy` in `AssetPairs`). Every margin call reads the spot endpoints; `exchange_info(margin)` lists only pairs with margin leverage.

Futures native symbols: `PF_XBTUSD` (linear perpetual, default for `BASE-QUOTE` input; BTC is written `XBT`), `PI_` (inverse perpetual), `FF_` / `FI_` (dated futures, e.g. `FF_SOLUSD_261225`). `exchange_info` lists every instrument the venue returns, not only `PF_`.

## Capabilities

Spot, margin and futures: `server_time`, `exchange_info`, `price`, `candles`, `ticker_24h`, `order_book`. Futures also `funding_rates`. No stream, trading or account.

`Price` has no market argument: a native futures symbol (`PF_`, `PI_`, `FF_`, `FI_`) goes to the futures tickers, anything else to the spot `Ticker`. Mixing both in one call is an error. `server_time` reads the futures host and is the same for all markets.

## Exchange Info APIs

| Capability | Endpoint | Notes |
|------------|----------|-------|
| Server time | `GET /derivatives/api/v3/feeschedules` | `serverTime` field (no dedicated endpoint) |
| Exchange info | `GET /derivatives/api/v3/instruments` | plus `/feeschedules` for fees |

Mapping:

- Tick size: `Extra["tick_size"]`; `PricePrecision` is its decimals. Also `Extra["contract_size"]`, `Extra["type"]`.
- `StepSize` / `QtyPrecision` from `contractValueTradePrecision` (10^-n). `MaxQty` = `maxPositionSize`.
- `MinQty` and `MinNotional` are not published, left unset.
- Fees: first tier of the instrument's fee schedule from the public `/feeschedules`, converted from percent to a fraction. Not account-specific; left unset if the schedule cannot be read.
- Status: `tradeable` and not `isExpired` is trading, else halt.

### Spot / margin `exchange_info`

`GET /0/public/AssetPairs`. `Symbol` is the altname (`XBTUSD`), `Extra["pair_key"]` the long key (`XXBTZUSD`).

- `StepSize` / `QtyPrecision`: `lot_decimals` (10^-n). `PricePrecision`: `pair_decimals`. Tick: `Extra["tick_size"]`.
- `MinQty` = `ordermin`, `MinNotional` = `costmin`.
- Fees: first tier of `fees` (taker) and `fees_maker`, percent to fraction. Not account-specific.
- Status: `online` is trading, any other status is halt. Dark-pool `.d` pairs are skipped.
- Margin: only pairs with `leverage_buy`; `Extra["margin_leverage"]` is the highest buy leverage.
