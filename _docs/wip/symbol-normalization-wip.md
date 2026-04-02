# Symbol Normalization — Work in Progress

Goal: every output struct with a `Symbol` field carries both `Symbol` (normalized `BASE-QUOTE`) and `OriginalSymbol` (provider-native). `model.Symbol` in ExchangeInfo gets a `NormalizedSymbol` field.

## Model changes
- [x] `model/normalize.go` — `NormalizeSymbol` utility
- [x] `model/ticker.go` — `OriginalSymbol` field added to `Ticker24h`
- [x] `model/orderbook.go` — `OriginalSymbol` field added to `OrderBook`
- [x] `model/price.go` — `OriginalSymbol` field added to `CoinPrice`
- [x] `model/exchange.go` — `NormalizedSymbol` field added to `Symbol` struct

## Provider fixes

| Provider | market.go | stream.go | exchange.go |
|---|---|---|---|
| Binance | [x] | [x] | [x] |
| Bitget | [x] | n/a | [x] |
| WhiteBit | [x] | [x] | [x] |
| Crypto.com | [x] | [x] | [x] |
| MEXC | [x] | n/a | [x] |
| CoinGecko | [x] | n/a | n/a |

## Notes
- `Symbol` = normalized (`BTC-USDT`); `OriginalSymbol` = provider-native (`BTCUSDT`, `BTC_USDT`, etc.)
- `model.Symbol.Symbol` stays provider-native (used by symbol engine for API calls)
- `model.Symbol.NormalizedSymbol` = `BaseAsset + "-" + QuoteAsset`
- CoinGecko: aggregator — `Symbol = strings.ToUpper(ticker)`, `OriginalSymbol = raw ticker from API`
