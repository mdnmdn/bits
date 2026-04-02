# CoinGecko Futures Order REST API Documentation

## N/A — No Order APIs

CoinGecko is a **price data aggregator only**. It does not operate an exchange, does not hold user funds, and provides no trading or order placement APIs.

### What CoinGecko Provides

- Market data (prices, tickers, OHLCV candles)
- Exchange listings and volume data
- Coin metadata and categories
- WebSocket price streams (paid tiers only)

### What CoinGecko Does NOT Provide

- Order placement, cancellation, or query
- Account balances or trading history
- Subaccount management
- Any authenticated trading endpoints
- Futures market data or derivatives tracking

### Alternative

To place orders using CoinGecko data, use the `bits` tool to fetch prices from CoinGecko, then route order execution through an actual exchange provider (Binance, Bitget, or MEXC for futures).
