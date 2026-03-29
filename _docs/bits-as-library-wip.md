# Bits as a Library

Analysis and proposals for improving the ergonomics of `bits` when used as an external package.

## Current Analysis

The `bits` codebase is currently structured primarily as a CLI application. While it has a modular provider-based architecture, most of its core components are located within the `internal/` directory, making them inaccessible to external Go packages.

### Key Observations:
- **Encapsulation in `internal/`**: Core interfaces like `Provider`, `ExchangeProvider`, and `AggregatorProvider` are in `internal/provider`. Data models are in `internal/model`. The registry for creating providers is in `internal/registry`. None of these can be imported by external tools.
- **Config Duplication**: There is a `pkg/config` which seems to be a partially synced or older version of `internal/config`. This creates confusion about which one to use.
- **Provider Creation**: `registry.NewProvider` is the main entry point for getting a provider instance, but it's also internal.
- **Dependency on `viper`**: The configuration system is tightly coupled with `viper` and CLI-specific logic (like searching for config files in specific OS directories).
- **Inconsistent Market Handling**: Some providers handle market types (Spot, Futures) differently in their internal state vs. method parameters.

## Potential Improvements

To improve ergonomics for use as a library, the following refactoring steps are proposed:

### 1. Migration of Core Packages to `pkg/`

Move the following packages from `internal/` to `pkg/` to make them importable by external tools:
- `internal/model` -> `pkg/model`: Contains all data structures (e.g., `CoinPrice`, `OrderBook`, `Candle`).
- `internal/capability` -> `pkg/capability`: Defines the feature/market matrix.
- `internal/provider` -> `pkg/provider`: Defines the base and specialized provider interfaces (`ExchangeProvider`, `AggregatorProvider`).

### 2. Configuration Consolidation

Consolidate `internal/config` and `pkg/config` into a single, comprehensive `pkg/config` package.
- **Library Mode**: The `Config` struct should be easily constructible via code (not just via `viper` or files).
- **CLI Mode**: Keep the `viper` and file discovery logic in a separate package (e.g., `cmd/bits/config`) or as an optional extension in `pkg/config`.
- **Decoupling**: Remove dependencies on CLI-specific tools (like `lipgloss` or TUI-related logic) from the core configuration structures.

### 3. Public Registry and Factory

Move `internal/registry` to `pkg/provider/registry` (or similar).
- **Simpler Initialization**: Provide a high-level factory method that can create any provider with a simple configuration.
- **Custom Providers**: Allow external users to register their own providers that implement the `Provider` interface.

### 4. High-Level Facade (`pkg/bits`)

Introduce a new `pkg/bits` package that acts as a facade for the entire library.
- Provide a `Client` struct that manages multiple providers.
- Implement high-level methods like `GetPrice(ctx, symbol, providerID)` or `ComparePrices(ctx, symbol, providers)`.
- Handle internal details like symbol resolution and provider selection automatically.

### 5. Ergonomic Enhancements

- **Functional Options**: Use functional options for provider configuration (e.g., `binance.NewClient(config, binance.WithTestnet())`).
- **Consistent Context**: Ensure every network-facing method accepts `context.Context` for cancellation and timeouts.
- **Graceful Error Handling**: Return structured errors (e.g., `model.ItemError`) more consistently to help library users diagnose partial failures.

## Proposed `pkg/bits` API

The `pkg/bits` package will provide a high-level API to interact with various crypto providers.

### The `Client` Interface

The `Client` struct acts as a manager for multiple providers. It simplifies the process of interacting with different exchanges by handling provider initialization and selection.

```go
type Client struct {
    Config *config.Config
}

func NewClient(cfg *config.Config) *Client {
    return &Client{Config: cfg}
}

// GetPrice retrieves the price for a symbol from a specific provider.
func (c *Client) GetPrice(ctx context.Context, symbol string, providerID string) (model.Response[model.CoinPrice], error) {
    // 1. Get provider from registry
    // 2. Fetch price
    // 3. Return result
}

// ComparePrices retrieves the price for a symbol from multiple providers.
func (c *Client) ComparePrices(ctx context.Context, symbol string, providerIDs []string) ([]model.Response[model.CoinPrice], error) {
    // 1. Concurrently fetch prices from each provider
    // 2. Aggregate results
    // 3. Return comparison
}
```

## Use Case: Multi-Exchange Quote Comparison

An external tool that compares quotes from different exchanges would need to:
1. Initialize multiple providers (e.g., Binance, Bitget, WhiteBit).
2. Fetch prices for the same symbol from each.
3. Normalize and compare the results.

### Current (Impossible) Code:
```go
import (
    "github.com/mdnmdn/bits/internal/registry"
    "github.com/mdnmdn/bits/internal/config"
)

// This fails because internal/ cannot be imported
cfg := &config.Config{...}
binance, _ := registry.NewProvider("binance", cfg)
bitget, _ := registry.NewProvider("bitget", cfg)
```

### Proposed Ergonomics:
```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mdnmdn/bits/pkg/bits"
	"github.com/mdnmdn/bits/pkg/config"
)

func main() {
	ctx := context.Background()

	// Initialize the library with a manual configuration
	cfg := &config.Config{
		Binance: config.BinanceConfig{
			Spot: config.MarketConfig{Enabled: true},
		},
		Bitget: config.BitgetConfig{
			Spot: config.MarketConfig{Enabled: true},
		},
	}

	client := bits.NewClient(cfg)

	// Compare BTC prices across multiple exchanges
	exchanges := []string{"binance", "bitget", "whitebit"}
	results, err := client.ComparePrices(ctx, "BTCUSDT", exchanges)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Price Comparison for BTCUSDT:\n")
	for _, res := range results {
		if len(res.Errors) > 0 {
			fmt.Printf("- %s: Error: %v\n", res.Provider, res.Errors[0].Err)
			continue
		}
		fmt.Printf("- %s: %.2f\n", res.Provider, res.Data.Price)
	}
}
```

## Summary of Benefits

- **Wider Adoption**: Developers can use `bits` as a building block for their own trading bots, dashboards, and research tools.
- **Code Reuse**: The logic for interacting with multiple exchanges (REST, WebSockets, normalization) is centralized and maintained once.
- **Improved Maintainability**: Clear separation between core logic (`pkg/`) and CLI-specific presentation (`cmd/` and `internal/tui/`).
- **Standardized Crypto API**: A common set of interfaces and data models for diverse providers.

## Implementation Roadmap

1. **Phase 1: Foundation (Current)**:
   - Finalize the analysis and get community feedback.
   - Audit all current `internal/` packages for library readiness.

2. **Phase 2: Refactoring**:
   - Move `model`, `capability`, and `provider` to `pkg/`.
   - Implement the consolidated `pkg/config`.
   - Update existing CLI commands to use the new `pkg/` structure.

3. **Phase 3: Ergonomics**:
   - Implement the high-level `pkg/bits` facade.
   - Add functional options and refine error handling.
   - Provide comprehensive documentation and examples.

4. **Phase 4: Community & Ecosystem**:
   - Encourage third-party provider contributions.
   - Build a showcase of tools (like the Quote Comparator) built on top of `bits`.
