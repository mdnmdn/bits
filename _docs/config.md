# Configuration

The `bits` CLI supports multiple configuration formats and locations.

## Config File Formats

Supported file extensions:
- `.yaml` / `.yml`
- `.toml`
- `.json`
- `.env` (environment variables)

## Config Locations

Config files are searched in the following order (first found wins):

| Priority | Location | Description |
|----------|----------|-------------|
| 1 | `./config.*` | Local directory (current working directory) |
| 2 | Platform-specific | User config directory |
| 3 | Platform-specific | Application support directory |

### Platform-Specific Paths

**Windows:**
```
%APPDATA%\bits\
%LOCALAPPDATA%\bits\
```

**macOS:**
```
~/.config/bits/                    (if XDG_CONFIG_HOME set)
~/Library/Application Support/bits-cli/
```

**Linux/Unix:**
```
$XDG_CONFIG_HOME/bits/
~/.config/bits/
```

### Saving Config

By default, `config save` writes to the platform-specific user config directory:
- Windows: `%APPDATA%\bits\config.yaml`
- macOS: `~/Library/Application Support/bits/config.yaml`
- Linux: `~/.config/bits/config.yaml`

Use `bits config init --local` to create a config in the current directory.

## CLI Commands

### `bits config show`

Show current configuration with API keys redacted.

```bash
bits config show          # Table output
bits config show -o json  # JSON output
```

### `bits config init`

Create a new configuration file.

```bash
bits config init              # Create in user config directory
bits config init --local      # Create in current directory
```

## Configuration Options

### Top-Level Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `provider` | string | `"coingecko"` | Active provider (`coingecko`, `binance`, `bitget`) |

### CoinGecko

```toml
[coingecko]
api_key = ""           # Your CoinGecko API key
tier = "demo"          # "demo" or "paid"
base_url = ""          # Optional custom endpoint
```

### Binance

```toml
[binance]
api_key = ""           # Your Binance API key
api_secret = ""        # Your Binance API secret
base_url = ""          # Optional custom endpoint (default: https://api.binance.com)

[binance.spot]
enabled = true         # Enable spot trading
use_testnet = false    # Use testnet

[binance.margin]
enabled = false        # Enable margin trading
use_testnet = false

[binance.futures]
enabled = false        # Enable futures trading
use_testnet = false    # Use testnet
```

### Bitget

```toml
[bitget]
api_key = ""           # Your Bitget API key
api_secret = ""        # Your Bitget API secret
passphrase = ""        # Your Bitget passphrase
base_url = ""          # Optional custom endpoint (default: https://api.bitget.com)

[bitget.spot]
enabled = false        # Enable spot trading

[bitget.futures]
enabled = false        # Enable futures trading
```

### WhiteBit

```toml
[whitebit]
api_key = ""           # Your WhiteBit API key
api_secret = ""        # Your WhiteBit API secret
base_url = ""          # Optional custom endpoint (default: https://whitebit.com)

[whitebit.spot]
enabled = false        # Enable spot trading
```

### Symbol Resolution (Cache)

```toml
[symbol]
cache_ttl = "5m"       # Cache TTL (e.g., 5m, 10m, 1h)
# cache_dir = ""       # Cache directory (defaults to system temp dir)
```

## Environment Variables

### Format

- Use `BITS_` prefix for all variables
- Use double underscore `__` for nested keys: `BITS_BINANCE__SPOT__ENABLED`
- Use single underscore `_` to separate sections: `BITS_COINGECKO_API_KEY`

### Examples

```bash
# Provider
BITS_PROVIDER=binance

# CoinGecko
BITS_COINGECKO_API_KEY=your-key
BITS_COINGECKO_TIER=paid

# Binance
BITS_BINANCE_API_KEY=your-key
BITS_BINANCE_API_SECRET=your-secret
BITS_BINANCE__SPOT__ENABLED=true

# Bitget
BITS_BITGET_API_KEY=your-key
BITS_BITGET_API_SECRET=your-secret
BITS_BITGET_PASSPHRASE=your-passphrase

# Symbol cache
BITS_SYMBOL_CACHE_TTL=5m
# BITS_SYMBOL_CACHE_DIR=  # defaults to system temp dir
```

## .env File

You can also use a `.env` file in any config directory:

```bash
# .env file example
BITS_PROVIDER=binance
COINGECKO_API_KEY=your-cg-key
BINANCE_API_KEY=your-binance-key
BINANCE_API_SECRET=your-binance-secret
BITGET_API_KEY=your-bitget-key
```

Key mapping:
- `COINGECKO_API_KEY` → `coingecko.api_key`
- `BINANCE_API_SECRET` → `binance.api_secret`
- `BITGET__SPOT__ENABLED` → `bitget.spot.enabled`

## Priority Order

Configuration sources are applied in this order (later overrides earlier):

1. Config file (`.yaml`/`.toml`/`.json`)
2. `.env` file
3. Environment variables (`BITS_*`)

## Example Config Files

See sample files in the repository:
- `config.sample.yaml`
- `config.sample.toml`
- `config.sample.env`
