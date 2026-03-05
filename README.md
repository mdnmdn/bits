# cg — CoinGecko CLI

A command-line tool for accessing CoinGecko cryptocurrency market data, built with Go.

## Install

### Homebrew (macOS/Linux)

> Coming soon — tap repo pending.

### Shell script

```sh
curl -sSfL https://raw.githubusercontent.com/coingecko/coingecko-cli/main/install.sh | sh
```

### Go install

```sh
go install coingecko-cli@latest
```

### Manual

Download the binary for your platform from [Releases](https://github.com/coingecko/coingecko-cli/releases), extract, and place `cg` in your `$PATH`.

## Setup

```sh
# Interactive setup (recommended — input is masked)
cg auth

# Or with flags (note: key may be visible in shell history)
cg auth --key YOUR_API_KEY --tier demo
```

Tiers: `demo`, `analyst`, `lite`, `pro`, `enterprise`

```sh
# Verify configuration
cg status
```

## Commands

### Price lookup

```sh
# By CoinGecko ID
cg price --ids bitcoin,ethereum

# By ticker symbol (auto-resolves to IDs)
cg price --symbols btc,eth

# Different target currency
cg price --ids bitcoin --vs eur
```

### Markets

```sh
# Top 100 coins by market cap
cg markets

# Top 20, filtered by category
cg markets --total 20 --category decentralized-finance-defi

# Export to CSV
cg markets --total 50 --export markets.csv
```

### Search

```sh
cg search solana
cg search dog --limit 5
```

### Trending

```sh
cg trending
```

### Historical data

```sh
# Price on a specific date
cg history bitcoin --date 2024-01-01

# OHLC for last 30 days
cg history bitcoin --days 30

# Price range with CSV export
cg history bitcoin --from 2024-01-01 --to 2024-06-30 --export btc-h1-2024.csv
```

### Top gainers/losers (paid plans only)

```sh
cg top-gainers-losers
cg top-gainers-losers --losers --duration 7d
```

### Interactive TUI

```sh
# Markets browser
cg tui

# Trending coins browser
cg tui-trending
```

Navigate with `j`/`k` or arrow keys, press `Enter` for detail view with a 7-day price chart, `Esc` to go back, `q` to quit.

## Development

```sh
# Build
make build

# Test
make test

# Lint (requires golangci-lint)
make lint
```

## License

MIT
