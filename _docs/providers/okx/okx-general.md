# OKX API Documentation

## Overview

- **Official exchange**: https://www.okx.com
- **API Docs**: https://www.okx.com/docs-v5/en/
- **Provider ID**: `okx` (alias `okex`)
- **Implementation**: Manual HTTP client, public data only (`provider/okx/`)
- **Raw notes**: `_drafts/okx-apis.md` (SDK survey, kept as draft)

## Base URLs

| Environment | URL |
|-------------|-----|
| Default (EEA) | `https://eea.okx.com` |
| Global | `https://www.okx.com` (same public data) |
| US/AU | `https://us.okx.com` |

Override with `okx.base_url` / `BITS_OKX_BASE_URL`.

## Authentication

None. Only public endpoints are used.

## Rate Limits

`/market/history-candles`: 20 req / 2 s. `/public/funding-rate-history`: 10 req / 2 s. The provider waits 110 ms between candle and funding pages (the funding limit is not enforced by the pause alone; long histories can hit it).

## Products

| Product | Market Type | Native instrument | Supported |
|---------|-------------|-------------------|-----------|
| Spot | `spot` | `BTC-USDT` | Yes |
| Perpetual swap (linear + inverse) | `futures` | `BTC-USDT-SWAP`, `BTC-USD-SWAP` | Yes |
| Margin | `margin` | spot `instId` (`BTC-USDT`) | Yes, spot data (see below) |
| Dated futures, options | | | No |

Config flags `okx.spot.enabled` / `okx.futures.enabled`; with none set, spot is enabled. There is no margin flag: margin is enabled with spot.

Margin has no separate data path: OKX margin trades the spot instruments, so price, candles, ticker and order book use the spot endpoints and the result carries `Market = margin`. Only `ExchangeInfo` differs: it lists `instType=MARGIN`.

## Capabilities

`server_time` (spot); `exchange_info`, `price`, `candles`, `ticker_24h`, `order_book` for spot, margin and futures; `funding_rates` for futures. No stream, trading or account.

## Exchange Info APIs

| Capability | Endpoint | Notes |
|------------|----------|-------|
| Server time | `GET /api/v5/public/time` | `ts` in ms |
| Exchange info | `GET /api/v5/public/instruments?instType=SPOT\|MARGIN\|SWAP` | |

Mapping:

- Tick size: `Extra["tick_size"]` (no tick field in `model.Symbol`); `PricePrecision` is its decimals.
- `StepSize` = `lotSz`, `MinQty` = `minSz`, `MaxQty` = `maxLmtSz`.
- Not published, left unset: `MinNotional`, `MakerFee`, `TakerFee`.
- SWAP quantities are in **contracts**. Contract fields in `Extra`: `ct_val` (contract size), `ct_val_ccy` (its currency), `ct_type` (`linear` / `inverse`), `settle_ccy`. Base and quote come from `uly` (`BTC-USDT` linear, `BTC-USD` inverse).
- Status: `live` is trading, `suspend` / `preopen` is halt, other is break.
