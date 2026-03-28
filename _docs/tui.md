# bits TUI Design

## Overview

The TUI (Terminal User Interface) provides an interactive, menu-driven experience for exploring cryptocurrency data. It complements the existing command-line interface with real-time data visualization, navigation, and auto-refresh capabilities.

## Activation

The TUI can be invoked in two ways:

```bash
# Interactive section selector (default)
bits tui

# Direct navigation to specific section with optional provider
bits tui exchange -p binance
bits tui prices -p coingecko
bits tui ticker -p binance -m spot
bits tui book BTCUSDT -p binance
```

## Section Architecture

Based on the existing command structure, the TUI is organized into logical sections that map to provider capabilities:

```
bits tui
├── prices      # Multi-coin price list (aggregators like CoinGecko)
├── ticker      # 24h statistics (exchanges: Binance, Bitget)
├── exchange    # Exchange info + server time
├── book        # Order book depth (Binance spot)
├── candles     # OHLCV charts
├── markets     # CoinGecko market listings
└── trending    # CoinGecko trending coins
```

### Section Mapping

| Section   | CLI Equivalent   | Primary Provider(s)          | Auto-refresh |
|-----------|------------------|-------------------------------|--------------|
| `prices`  | `bits price`     | CoinGecko, Binance, Bitget   | Yes          |
| `ticker`  | `bits ticker`    | Binance, Bitget              | Yes          |
| `exchange`| `bits time`      | Binance, Bitget              | Optional     |
| `book`    | `bits book`      | Binance spot                 | Yes          |
| `candles` | `bits candles`  | CoinGecko, Binance, Bitget   | No           |
| `markets` | `bits markets`  | CoinGecko                    | Optional     |
| `trending`| —               | CoinGecko                    | Optional     |

## UI Layout

Each section follows a consistent layout pattern inspired by CoinGecko CLI:

```
┌─────────────────────────────────────────────────────────────────┐
│  bits — TUI │ <Section Name> │ <Provider> │ <Market> │ <Status>│
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  [Header with title and current filters]                        │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  [Main content area - data table, chart, etc.]                 │
│                                                                 │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│  [Footer: keyboard shortcuts, last update time, auto-refresh]  │
└─────────────────────────────────────────────────────────────────┘
```

## Component Design

### Header Bar

Displays:
- Application branding: `bits — TUI`
- Current section name
- Active provider (with capability check indicator)
- Market type: `spot` | `futures` | `margin`
- Connection status indicator

### Filter Panel

Located below header, toggled with `F`:

```
┌─ Provider ─┬─ Market ─┬─ Filters ─────────────────────────────┐
│ [▼ Binance]│ [▼ spot] │ Symbol: [________]  [Apply] [Reset]    │
└────────────┴──────────┴────────────────────────────────────────┘
```

#### Provider Selection

Providers are filtered based on the current section's capability requirements:

- **prices**: All providers with `PriceProvider`
- **ticker**: Only `TickerProvider` (Binance, Bitget)
- **exchange**: Only `ExchangeProvider` (Binance, Bitget)
- **book**: Only `OrderBookProvider` (Binance spot)
- **candles**: All providers with `CandleProvider`
- **markets**: Only `AggregatorProvider` (CoinGecko)
- **trending**: Only CoinGecko

When multiple providers support the same capability, users can select from available options. The provider dropdown only shows providers that support the current section.

#### Market Selection

Market selection appears only when applicable:
- Exchanges (Binance, Bitget) show: `spot`, `futures`
- CoinGecko shows: N/A (aggregator)

#### Symbol/Search Filter

- Text input for filtering symbols/coins
- Supports partial matching (e.g., "BTC" matches "BTCUSDT")
- Clear button to reset filter

### Data Table

Main content area with keyboard navigation:

```
 #   │ Symbol           │ Price            │ 24h Change   │ Volume
─────┼──────────────────┼──────────────────┼──────────────┼──────────────
  1  │ BTCUSDT         │ $67,432.50       │ +2.34%       │ 1.2B
  2  │ ETHUSDT         │ $3,542.12        │ +1.87%       │ 890M
  3  │ SOLUSDT         │ $142.33          │ -0.45%       │ 234M
─────┴──────────────────┴──────────────────┴──────────────┴──────────────
```

Keyboard navigation:
- `j` / `↓`: Move cursor down
- `k` / `↑`: Move cursor up
- `Enter`: Select item (view details)
- `Space`: Toggle selection (multi-select mode)

### Detail View

Triggered by pressing `Enter` on a selected item:

```
┌─────────────────────────────────────────────────────────────────┐
│  BTCUSDT — Ticker Detail                    [← Back: Esc]     │
├─────────────────────────────────────────────────────────────────┤
│  Price      │ $67,432.50    │ 24h High   │ $68,100.00          │
│  24h Low    │ $66,200.00    │ 24h Change │ +2.34%              │
│  Volume     │ 1,234,567,890 │ Quote Vol  │ $83.4B              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│                    [Price Chart Area]                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Multi-Provider View

When viewing data from multiple providers side-by-side (toggle with `M`):

```
┌─────────────────────────────────────────────────────────────────┐
│  Symbol    │ Binance Price  │ Bitget Price   │ Diff            │
────────────┼────────────────┼────────────────┼─────────────────
│ BTCUSDT   │ $67,432.50     │ $67,428.30     │ -0.006%         │
│ ETHUSDT   │ $3,542.12      │ $3,541.80      │ -0.009%         │
└────────────┴────────────────┴────────────────┴─────────────────┘
```

This view is automatically enabled when:
- The selected section supports multiple providers
- User explicitly enables multi-provider mode
- Only 1-3 items are selected (for better comparison)

## Auto-Refresh

### Configuration

Toggle auto-refresh with `R`. Available intervals:

| Interval | Key | Use Case |
|----------|-----|----------|
| Off | `0` | Static data (charts) |
| 5s | `1` | High-frequency trading |
| 10s | `2` | General monitoring |
| 30s | `3` | Default for most views |
| 1m | `4` | Low-frequency updates |

Key bindings:
- `R`: Cycle through refresh intervals
- `1-4`: Directly select refresh interval
- `0`: Disable auto-refresh

### Visual Indicator

```
┌─────────────────────────────────────────────────────────────────┐
│ ...Header...                                    [Auto: 30s ↻]   │
└─────────────────────────────────────────────────────────────────┘
```

The refresh indicator shows:
- Current interval (or "Off" when disabled)
- Next refresh countdown
- Last successful update timestamp
- Connection status (green = connected, yellow = reconnecting)

## Keyboard Shortcuts

Global shortcuts (available in all sections):

| Key | Action |
|-----|--------|
| `q` / `Esc` | Quit TUI |
| `?` | Show help overlay |
| `F` | Toggle filter panel |
| `R` | Cycle auto-refresh interval |
| `0-4` | Set specific auto-refresh interval |
| `p` | Open provider selector |
| `m` | Open market selector |
| `Tab` | Cycle through sections |
| `g` + `1-9` | Jump to page (g1 = first, g$ = last) |
| `/` | Focus search/filter input |

Section-specific shortcuts:

| Section | Key | Action |
|---------|-----|--------|
| All | `Enter` | View item details |
| All | `Space` | Multi-select toggle |
| All | `a` | Select all visible |
| All | `n` | Deselect all |
| Prices | `+` | Add symbol to watchlist |
| Ticker | `c` | View candles for selected |
| Book | `s` | Switch bid/ask view |
| Markets | `t` | Sort by ticker |

## Technical Implementation

### Package Structure

```
internal/
├── tui/
│   ├── app.go           # Main TUI application (tea.Model)
│   ├── model.go         # Section model definitions
│   ├── view.go          # View rendering logic
│   ├── input.go         # Input handling and keymaps
│   ├── provider.go      # TUI-specific provider wrappers
│   └── components/
│       ├── table.go     # Reusable table component
│       ├── chart.go     # Candle/line chart rendering
│       ├── filter.go    # Filter panel component
│       └── statusbar.go # Status bar component
```

### State Management

```go
type TUIState struct {
    Section    string         // Current section
    Provider   string         // Selected provider
    Market     MarketType     // spot/futures/margin
    Filters    map[string]any // Section-specific filters
    Items      []Item         // Displayed data
    Selected   int            // Cursor position
    MultiSelect []int         // Multi-selected indices
    AutoRefresh RefreshConfig
    LastUpdate time.Time
    Error      error
}
```

### Provider Integration

The TUI reuses existing provider implementations:

1. Use `resolve.Resolver` to select provider based on capability
2. Call existing provider methods (same as CLI commands)
3. Wrap responses in TUI-friendly format
4. Handle streaming providers for real-time updates

### Stream Handling

For sections with auto-refresh:
- Use existing `PriceStreamProvider` / `OrderBookStreamProvider`
- Buffer updates and render at refresh interval
- Show loading indicator during initial fetch
- Gracefully handle disconnections with reconnect

## Command-Line Flags

The TUI command supports additional flags:

```bash
bits tui [section] [flags]

Flags:
  -p, --provider string    Provider to use
  -m, --market string      Market type (spot/futures/margin)
  -s, --symbol string      Initial symbol filter
  -r, --refresh string     Auto-refresh interval (5s/10s/30s/1m)
  -h, --help               Help for tui
```

Examples:

```bash
# Open TUI at ticker section with Binance futures
bits tui ticker -p binance -m futures

# Open TUI with 5-second auto-refresh
bits tui prices -r 5s

# Open TUI with BTC symbol pre-selected
bits tui ticker -s BTC
```

## Design Decisions

### Why These Sections?

Based on capability matrix analysis:
- **prices**: Core functionality, all providers support
- **ticker**: High-demand for exchange data, Binance/Bitget only
- **book**: Unique to Binance, valuable for depth analysis
- **candles**: Chart data, all providers support
- **markets**: CoinGecko aggregator-specific
- **trending**: CoinGecko aggregator-specific

### Why Bubble Tea?

- Well-maintained Go TUI library
- Supports complex layouts with `lipgloss`
- Built-in component library (bubbles)
- Active community and documentation

### Auto-Refresh Limits

- Minimum interval: 5 seconds (API rate limit protection)
- Maximum visible providers in comparison: 3
- Reason: Avoid overwhelming the terminal and API
