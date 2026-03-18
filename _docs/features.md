# Project Features

This document provides a comprehensive map of the features available in the CoinGecko CLI (`cg`).

## 1. Core CLI Commands

The CLI provides several standard commands to fetch cryptocurrency data directly to your terminal.

### `cg price`
- Fetch current prices for one or more coins.
- Supports both CoinGecko IDs (`--ids bitcoin`) and ticker symbols (`--symbols btc`).
- Multi-currency support via `--vs` (default: `usd`).
- Displays 24h price change.

### `cg markets`
- List top coins by market cap.
- Supports filtering by category (`--category`).
- Pagination and custom result counts (`--total`).
- Sorting options (market cap, volume, etc.).

### `cg history`
- Comprehensive historical data retrieval with three modes:
  - **Snapshot**: Get coin data for a specific date (`--date YYYY-MM-DD`).
  - **Recent**: Get data for the last N days (`--days N`).
  - **Range**: Get data for a specific date range (`--from` and `--to`).
- Support for **OHLC** (Open, High, Low, Close) data via `--ohlc`.
- Adjustable granularity via `--interval` (daily, hourly).
- **Auto-batching**: Large date ranges are automatically split into multiple API requests to overcome API limits.

### `cg search`
- Search for coins, exchanges, or NFTs by query string.
- Returns name, symbol, CoinGecko ID, and market cap rank.

### `cg trending`
- Fetch the top-7 (Demo) or top-15/30 (Paid) trending coins on CoinGecko.

### `cg top-gainers-losers`
- List coins with the highest gains and losses over a 24h period.

### `cg status`
- Display current CLI configuration, including API tier and masked API key.

---

## 2. Interactive TUI (Terminal User Interface)

Launch a rich, interactive experience using `cg tui`.

### `cg tui markets`
- Scrollable list of top coins.
- **Detail View**: Select a coin to see detailed stats, description, and a price chart.
- **Braille Chart**: Real-time rendered price chart using `ntcharts`.

### `cg tui trending`
- Interactive view of currently trending coins.

---

## 3. Real-time Monitoring

### `cg watch`
- Live price updates via WebSocket.
- Tracks multiple coins simultaneously.
- Automatic reconnection with exponential backoff.
- *Note: Requires a paid CoinGecko API plan.*

---

## 4. Output & Export

### Global Output Formats
- **Table Mode** (Default): Pretty-printed tables for human readability.
- **JSON Mode** (`-o json`): Raw API response output, ideal for piping to `jq` or other tools.

### CSV Export
- Use `--export <path>.csv` with `cg history` to save historical data for analysis in Excel/Google Sheets.

### ASCII Aesthetics
- Branded banners and welcome boxes.
- ANSI color support (detects `NO_COLOR` and TTY).

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
- `cg commands` (hidden): Outputs a machine-readable JSON catalog of all commands and their metadata.

---

## 6. Technical Architecture & Internal Features

Beyond user-facing commands, the project implements several key technical features:

### Intelligent Data Fetching
- **Pagination Helper**: `FetchAllMarkets` handles multi-page fetching (250/page) with trim-to-total logic, used by both CLI and TUI.
- **Concurrent TUI Fetching**: The coin detail view fetches coin data and OHLC history concurrently using Bubble Tea's `tea.Batch` for faster rendering.
- **Symbol Resolution**: 
  - `cg price` uses the API's native symbol lookup.
  - `cg watch` (which requires coin IDs) uses a custom resolution strategy that picks the highest-ranked coin for a given symbol.

### Security & Robustness
- **Terminal Safety**: All API-returned text (names, symbols) is sanitized via `display.SanitizeCell` to strip potential terminal escape sequence injections.
- **Pre-flight Checks**: `requirePaid()` helper ensures API calls to paid-only endpoints aren't wasted if the user is on a demo plan.
- **Wait-and-Retry**: Robust handling of API rate limits ensures reliable data retrieval during high-volume operations (like batching).
- **Graceful WebSocket Shutdown**: Uses atomic flags and state machines to ensure clean connection closure and suppress unnecessary reconnection attempts.

### Platform Support
- **Homebrew Distribution**: Automatic Homebrew formula generation via GoReleaser.
- **Cross-Platform**: Designed for macOS, Linux, and Windows terminal environments.
