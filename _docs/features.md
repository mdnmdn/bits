# Project Features

This document provides a comprehensive map of the features available in the Multi-Provider Crypto CLI (`bits`).

## 1. Multi-Provider Support

`bits` supports multiple data providers, allowing you to fetch market data from different sources:

- **CoinGecko** (Default): Comprehensive market data, historical charts, and trending coins.
- **Binance**: Real-time prices, 24h ticker stats, and order book depth from the world's largest exchange.
- **Bitget**: Real-time prices and 24h ticker stats from Bitget.

Use the `--provider` or `-p` flag to select a provider:
```bash
bits price --ids bitcoin                    # CoinGecko (default)
bits price --ids BTCUSDT -p binance         # Binance
bits ticker BTCUSDT -p binance              # Binance 24h stats
bits orderbook BTCUSDT -p binance           # Binance order book
```

## 2. Core CLI Commands

### `bits price`
- Fetch current prices for one or more coins.
- Supports both CoinGecko IDs (`--ids bitcoin`) and ticker symbols (`--symbols btc`).
- Multi-currency support via `--vs` (default: `usd`).
- Displays 24h price change.
- Works across all providers (CoinGecko, Binance, Bitget).

### `bits ticker` [New]
- Fetch 24-hour ticker statistics for a trading pair.
- Displays last price, 24h change, high/low, and volume.
- Supported by: Binance, Bitget.

### `bits orderbook` [New]
- Fetch real-time order book depth (bids and asks).
- Customizable depth via `--limit`.
- Supported by: Binance.

### `bits markets`
- List top coins by market cap.
- Supports filtering by category (`--category`).
- Pagination and custom result counts (`--total`).
- Sorting options (market cap, volume, etc.).
- Supported by: CoinGecko.

### `bits history`
- Comprehensive historical data retrieval with three modes:
  - **Snapshot**: Get coin data for a specific date (`--date YYYY-MM-DD`).
  - **Recent**: Get data for the last N days (`--days N`).
  - **Range**: Get data for a specific date range (`--from` and `--to`).
- Support for **OHLC** (Open, High, Low, Close) data via `--ohlc`.
- Adjustable granularity via `--interval` (daily, hourly).
- **Auto-batching**: Large date ranges are automatically split into multiple API requests to overcome API limits.
- Supported by: CoinGecko.

### `bits search`
- Search for coins, exchanges, or NFTs by query string.
- Returns name, symbol, CoinGecko ID, and market cap rank.
- Supported by: CoinGecko.

### `bits trending`
- Fetch the top-7 (Demo) or top-15/30 (Paid) trending coins on CoinGecko.
- Supported by: CoinGecko.

### `bits top-gainers-losers`
- List coins with the highest gains and losses over a 24h period.
- Supported by: CoinGecko.

### `bits capabilities` (alias: `bits caps`)
- Display a matrix of which features are supported by each provider and market type.
- Rows show features × market types (spot, futures, margin); columns show providers.
- Filter to a single provider with `--provider` / `-p`.
- Machine-readable JSON output via `-o json`.
- Requires no API key — reads static provider declarations.
- Example:
  ```
  bits capabilities
  bits caps -p binance
  bits capabilities -o json
  ```

### `bits status`
- Display current CLI configuration, including active provider, API tiers, and masked keys.

---

## 3. Interactive TUI (Terminal User Interface)

Launch a rich, interactive experience using `bits tui`.

### `bits tui markets`
- Scrollable list of top coins.
- **Detail View**: Select a coin to see detailed stats, description, and a price chart.
- **Braille Chart**: Real-time rendered price chart using `ntcharts`.

### `bits tui trending`
- Interactive view of currently trending coins.

---

## 4. Real-time Monitoring

### `bits watch`
- Live price updates via WebSocket.
- Tracks multiple coins simultaneously.
- Automatic reconnection with exponential backoff.
- Supported by: CoinGecko (Requires a paid API plan).

---

## 5. Advanced Features

### Smart Rate Limiting
- Automatic retries on HTTP 429 (Rate Limit) errors.
- Respects `Retry-After` and `x-ratelimit-reset` headers.
- Uses exponential backoff with jitter when no headers are provided.

### Dry Run Mode
- Use `--dry-run` on any data command to see exactly what API request would be made without executing it.

### Batching & Granularity
- **Hourly Data**: Automatically routes through range endpoints to provide hourly granularity even for recent periods.
- **Data Guardrails**: Prevents requesting hourly data before it was available (Jan 2018).

### Environment & Security
- API Key priority: `CG_API_KEY` environment variable > `--key` flag > Interactive prompt.
- Configuration stored securely in `~/.config/coingecko-cli/config.yaml`.
- Masked API key display in status.

### Command Catalog
- `bits commands` (hidden): Outputs a machine-readable JSON catalog of all commands and their metadata.

---

## 6. Technical Architecture & Internal Features

Beyond user-facing commands, the project implements several key technical features:

### Intelligent Data Fetching
- **Pagination Helper**: `FetchAllMarkets` handles multi-page fetching (250/page) with trim-to-total logic, used by both CLI and TUI.
- **Concurrent TUI Fetching**: The coin detail view fetches coin data and OHLC history concurrently using Bubble Tea's `tea.Batch` for faster rendering.
- **Symbol Resolution**: 
  - `bits price` uses the API's native symbol lookup.
  - `bits watch` (which requires coin IDs) uses a custom resolution strategy that picks the highest-ranked coin for a given symbol.

### Security & Robustness
- **Terminal Safety**: All API-returned text (names, symbols) is sanitized via `display.SanitizeCell` to strip potential terminal escape sequence injections.
- **Pre-flight Checks**: `requirePaid()` helper ensures API calls to paid-only endpoints aren't wasted if the user is on a demo plan.
- **Wait-and-Retry**: Robust handling of API rate limits ensures reliable data retrieval during high-volume operations (like batching).
- **Graceful WebSocket Shutdown**: Uses atomic flags and state machines to ensure clean connection closure and suppress unnecessary reconnection attempts.

### Platform Support
- **Homebrew Distribution**: Automatic Homebrew formula generation via GoReleaser.
- **Cross-Platform**: Designed for macOS, Linux, and Windows terminal environments.
