# Status: bits Transformation (Phase 1)

This document tracks the progress of transforming the CoinGecko-specific CLI (`cg`) into the multi-provider crypto tool `bits`.

## ✅ Completed Tasks

### 1. Provider Architecture (Phase 1: Extraction)
- **Code Migration**: Moved `internal/api/` to `internal/provider/coingecko/`.
- **Package Refactoring**: Updated package names from `api` to `coingecko`.
- **Interface Definition**: Created `internal/provider/types.go` defining the `Provider` interface.
- **Provider Registry**: Created `internal/provider/registry.go` with `NewProvider` factory.
- **Abstraction**: Updated `cmd/client_factory.go` and all TUI models (`markets`, `detail`, `trending`) to use the `provider.Provider` interface instead of a concrete CoinGecko client.
- **Dependency Cleanup**: Updated all project imports to reflect the new provider structure.

### 2. CLI Branding & Identity
- **Command Renaming**: Changed the root command from `cg` to `bits` in `cmd/root.go`.
- **Binary Name**: Updated `install.sh` and build logic (`Makefile`) to use `bits`.
- **Version Output**: Updated `bits version` to display the new name.
- **UI Branding**: Updated the interactive TUI banner and help examples to use `bits` command syntax.

### 3. Documentation & Workflow Updates
- **README.md**: Replaced `cg` with `bits` and updated branding.
- **CLAUDE.md**: Updated project overview, build instructions, and file structure map.
- **_docs/features.md**: Updated feature descriptions to use `bits`.
- **_docs/README.md**: Updated to reflect `bits` project.
- **.goreleaser.yml**: Updated project name and binary name.
- **.github/workflows/ci.yml**: Updated build output binary name.

### 4. Verification
- **Build Success**: Verified the project compiles with `go build -o bits .` and `make build`.
- **Test Suite**: Verified all tests pass with `go test ./...`.
- **Functional Check**: Confirmed `./bits price --ids bitcoin` works correctly using the new provider abstraction.

## ⏳ Future Phases

### 1. Generic Models (Phase 3)
- Move towards generic `MarketData` models in `internal/model` to fully decouple from CoinGecko-specific response structs. Currently, the `Provider` interface still uses `coingecko` structs for pragmatism.

### 2. New Providers (Phase 4)
- Implement a second provider (e.g., Binance or CoinMarketCap) to validate the architecture's extensibility.

### 3. Trading Capabilities (Phase 5)
- Introduce `TradeProvider` interface and commands for order placement.

## Conclusion
Phase 1 and 2 of the transformation are complete. The tool is now `bits`, it uses a provider-agnostic interface internally (though still using CoinGecko-specific data types), and is ready for further expansion.
