# Future Tasks

Features and capabilities that are implemented internally but not yet exposed through the CLI.

## Trading Commands

Both Binance and Bitget providers have internal trading methods ready to be wired into CLI commands.

### Binance Trading (internal/provider/binance/trading.go)
- `PlaceMarketOrder(symbol, side, quantity)` — Market buy/sell
- `PlaceLimitOrder(symbol, side, quantity, price)` — Limit orders
- `CancelOrder(symbol, orderID)` — Cancel open order
- `GetAccountInfo()` — Account balances and permissions
- `GetAssetBalance(asset)` — Single asset balance
- `ListOpenOrders(symbol)` — Open orders
- `ListOrders(symbol, start, end)` — Order history
- `GetExchangeInfo()` — Exchange rules and symbol info
- `GetSymbolsInfo(symbols)` — Filtered symbol info

### Bitget Trading (internal/provider/bitget/trading.go)
- `PlaceOrder(symbol, side, quantity, price)` — Limit order
- `PlaceMarketOrder(symbol, side, quantity)` — Market order
- `PlaceOrderWithType(symbol, side, type, quantity, price, pairInfo)` — Order with precision
- `GetOrderStatus(symbol, orderId)` — Check order status
- `CancelOrder(symbol, orderId)` — Cancel order
- `GetAssetBalance(coin)` / `GetUSDTBalance()` — Balances
- `GetAllAssets()` — All account assets
- `GetTradingPairInfo(symbol)` — Trading pair details
- `GetTradingFee()` — VIP fee rate
- `ListOpenOrders(symbol)` — Open orders
- `ListOrderHistory(symbol, start, end)` — Order history

### Proposed CLI Commands
```bash
bits balance -p binance
bits balance -p bitget
bits buy BTCUSDT 0.001 --type market -p binance
bits sell ETHUSDT 0.1 --price 3500 -p bitget
bits orders BTCUSDT -p binance
bits cancel BTCUSDT <order-id> -p binance
```

## WebSocket Streaming

Currently only CoinGecko supports live streaming via `bits watch`. Binance and Bitget have WebSocket APIs that could be integrated.

## Additional Provider Features

- **Binance Market Listings** — Implement MarketLister for Binance (24h tickers for all symbols)
- **Bitget Order Book** — Implement OrderBookProvider for Bitget
- **Provider-specific search** — Symbol search across exchanges

## Output Formatters

The architecture proposal includes support for additional output formats:
- CSV (`-o csv`)
- Markdown (`-o md`)
- YAML (`-o yaml`)
- TOON (`-o toon`) — stylized ASCII terminal format

## Portfolio Tracking

Cross-provider portfolio aggregation: track holdings across multiple exchanges.

## New Providers

The capability-based architecture supports adding more providers:
- CoinMarketCap (market data)
- Kraken (exchange)
- OKX (exchange)
- Bybit (exchange)
