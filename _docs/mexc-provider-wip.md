# MEXC Provider Implementation WIP

This document tracks the progress of the MEXC provider implementation for `bits`.

## Progress

- [x] Configuration support in `pkg/config/config.go`
- [x] Provider directory and base client (`pkg/provider/mexc/client.go`)
- [x] Internal types (`pkg/provider/mexc/types.go`)
- [x] Market data implementation (`pkg/provider/mexc/market.go`)
  - [x] Price
  - [x] Ticker
  - [x] Candles
  - [x] Order Book
- [x] Exchange info implementation (`pkg/provider/mexc/exchange.go`)
- [x] Registry integration (`pkg/provider/registry/registry.go`)
- [x] Documentation update (`_docs/provider-capabilites-wip.md`)
- [ ] Verification and smoke tests

## API Notes

MEXC has separate REST endpoints for Spot and Futures:
- Spot: `https://api.mexc.com/api/v3`
- Futures: `https://api.mexc.com/api/v1/contract`

Symbol formats:
- Spot: `BTCUSDT`
- Futures: `BTC_USDT`

## Missing Points / TODOs
- [ ] WebSocket streaming (not requested in the initial task, but good for future)
- [ ] Margin market support (MEXC margin usually uses spot endpoints with specific parameters or separate ones, need to double check)
