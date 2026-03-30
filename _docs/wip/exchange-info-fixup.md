# Exchange Info Fixup

Testing exchange info for all providers and markets to ensure symbols and trading limits are populated correctly.

## Test Commands

```sh
go run . info -p <provider> -m <market>
```

## Results

### WhiteBit (wb)

| Market | Symbols | MIN/MAX PRICE | MIN/MAX QTY | Status |
|--------|---------|---------------|-------------|--------|
| spot   | 1060    | ✅            | ✅          | Fixed - now parsing all fields |
| futures| 294     | ✅            | ✅          | Fixed - now parsing bracket-based min/max qty |

**Finding**: Spot and futures use DIFFERENT endpoints:
- Spot: `/api/v4/public/markets` → symbols like `BTC_USDT`
- Futures: `/api/v4/public/futures` → symbols like `BTC_PERP`

Code correctly calls different endpoints but doesn't parse precision/limit fields.

### Binance

| Market | Symbols | MIN/MAX PRICE | MIN/MAX QTY | Status |
|--------|---------|---------------|-------------|--------|
| spot   | 3549    | ✅            | ✅          | Working correctly |
| futures| 702     | ✅            | ✅          | Working correctly |

### Bitget

| Market | Symbols | MIN/MAX PRICE | MIN/MAX QTY | Status |
|--------|---------|---------------|-------------|--------|
| spot   | 716     | ✅            | ✅          | Fixed - now parsing precision and limits |
| futures| 538     | ✅            | ✅          | Fixed - now parsing precision and limits |
| margin | 180     | ✅            | ✅          | Added - now parsing margin data with fee rates |

**Finding**: Both endpoints return precision/limit data:
- Spot: `/api/v2/spot/public/symbols` has `pricePrecision`, `quantityPrecision`, `minTradeAmount`, `maxTradeAmount`
- Futures: `/api/v2/mix/market/contracts?productType=USDT-FUTURES` has `pricePlace`, `volumePlace`, `minTradeNum`, `maxOrderQty`

### MEXC

| Market | Symbols | MIN/MAX PRICE | MIN/MAX QTY | Status |
|--------|---------|---------------|-------------|--------|
| spot   | 2377    | ✅            | ✅          | Fixed - now parsing all precision fields |
| futures| 846     | ✅            | ⚠️          | PricePrecision populated, no qty limits in API |

**Finding**: 
- Spot: Has `baseAssetPrecision`, `quoteAssetPrecision`, `quotePrecision` but only `QuotePrecision` parsed
- Futures: Has `pricePrecision`, no min/max qty fields in API

### Crypto.com

| Market | Symbols | MIN/MAX PRICE | MIN/MAX QTY | Status |
|--------|---------|---------------|-------------|--------|
| spot   | 608     | ✅            | ✅          | Fixed - now using API with precision and limits |
| margin | 162     | ✅            | ✅          | Fixed - now using API with precision and limits for margin |
| futures| -       | ❌            | ❌          | No futures support (not implemented in API) |

**Finding**: Crypto.com uses a HARDCODED list of 25 popular pairs instead of fetching from API (comment says "public/get-instruments endpoint is currently returning errors")

## Action Items

- [ ] WhiteBit: Parse exchange info API to extract MIN_PRICE, MAX_PRICE, MIN_QTY, MAX_QTY
  - Spot: `/api/v4/public/markets` has `stockPrec`, `moneyPrec`, `minAmount`, `minTotal`, `maxTotal`
  - Futures: `/api/v4/public/futures` has `brackets` for tiered limits, no explicit min/max
- [ ] Bitget: Parse exchange info API to extract precision and limits
- [ ] Crypto.com: Add QTY_PREC parsing
- [ ] MEXC: Add MIN/MAX QTY and PRICE parsing

## API Verification

### WhiteBit
Spot endpoint `/api/v4/public/markets` returns:
```json
{
  "name": "ETH_BTC",
  "stockPrec": "4",    // QTY_PREC
  "moneyPrec": "6",    // PRICE_PREC  
  "minAmount": "0.0001",  // MIN_QTY
  "minTotal": "0.000012", // MIN_PRICE
  "maxTotal": "100000",   // MAX_PRICE/MAX_QTY
  "makerFee": "0.1",
  "takerFee": "0.1"
}
```

Futures endpoint `/api/v4/public/futures` returns:
```json
{
  "ticker_id": "BTC_PERP",
  "brackets": {"1":0,"10":0,"100":0,"2":0,"20":4000,"3":0,"5":0,"50":800},
  "max_leverage": 50
}
```

### Bitget
Spot endpoint `/api/v2/spot/public/symbols` returns:
```json
{
  "symbol": "BTCUSDT",
  "baseCoin": "BTC",
  "quoteCoin": "USDT",
  "pricePrecision": 2,
  "quantityPrecision": 6,
  "minTradeAmount": "1",
  "maxTradeAmount": "999999999999999999"
}
```

Futures endpoint `/api/v2/mix/market/contracts?productType=USDT-FUTURES` returns:
```json
{
  "symbol": "BTCUSDT",
  "baseCoin": "BTC",
  "quoteCoin": "USDT",
  "pricePlace": 1,
  "volumePlace": 4,
  "minTradeNum": "0.0001",
  "maxOrderQty": "1200"
}
```

## Summary

All providers except MEXC futures now have working ExchangeInfo with at least precision data:
- ✅ WhiteBit: Spot and futures both working (spot has full limits, futures has bracket-based limits)
- ✅ Bitget: Spot, futures, and margin all working with precision and limits
- ✅ Crypto.com: Spot and margin working via API (replaced hardcoded list)
- ⚠️ MEXC: Spot has precision but missing min/max limits, futures has only price precision
- ✅ Binance: Reference implementation - all fields working

## Completed Fixes

1. **WhiteBit**: Added parsing of `stockPrec` (QTY_PREC), `moneyPrec` (PREC_PREC), `minAmount` (MIN_QTY), `minTotal` (MIN_PRICE), `maxTotal` (MAX_PRICE/MAX_QTY) for spot; and bracket-based min/max qty for futures

2. **Bitget**: Added parsing of `pricePrecision`, `quantityPrecision`, `minTradeAmount`, `maxTradeAmount` for spot; and `pricePlace`, `volumePlace`, `minTradeNum`, `maxOrderQty` for futures; plus margin data with fee rates

3. **Crypto.com**: Replaced hardcoded popular pairs list with live API call to `/exchange/v1/public/get-instruments`, filtering by `inst_type` and parsing `quote_decimals` (PREC_PREC), `quantity_decimals` (QTY_PREC), `price_tick_size` (MIN_PRICE), `qty_tick_size` (MIN_QTY)

4. **MEXC**: Improved spot parsing to use both `baseAssetPrecision` (QTY_PREC) and `quoteAssetPrecision` (PREC_PREC); futures already had `pricePrecision`

## Remaining Work

- MEXC futures: API doesn't provide min/max quantity fields, only price precision
- Crypto.com futures: Would require adding market type support for futures (currently only spot/margin)
