# Provider Capabilities Test Tool

The `_tool/test_capabilities.py` script is a utility to automatically verify that each provider's declared capabilities are actually functional. It invokes the `bits` CLI for each supported feature and market, ensuring that the commands return valid data and do not error out.

## How it works

1. It runs `bits capabilities` to discover which features and markets are supported by each provider.
2. For each supported capability, it constructs the appropriate `bits` command.
3. It executes the command with `-o json` to allow for automated verification.
4. It verifies:
   - The command exits with a success code (0).
   - The output is valid JSON.
   - The `data` field in the response is not empty (for snapshot commands).
   - For streaming commands, it waits for up to 5 seconds and ensures some valid JSON data was received.

## Prerequisites

- Python 3.x
- The `bits` binary must be built and present in the root directory. Run `make build` before running the test script.

## Usage

Run the script from the root of the repository:

```sh
python3 _tool/test_capabilities.py
```

### Options

- `-p, --provider <id>`: Test only a specific provider (e.g., `binance`, `whitebit`).

Example:
```sh
python3 _tool/test_capabilities.py --provider whitebit
```

## Interpreting Results

The script outputs the status of each test as it runs:

- **OK**: The command worked as expected.
- **FAILED**: The command failed or returned invalid/empty data. The error message is displayed.
- **SKIPPED**: The command failed due to a known environment issue, such as a missing API key or a restricted IP location (common for Binance).

At the end, a summary table shows the total number of OK, FAILED, and SKIPPED tests per provider.

## Adding/Modifying Tests

To update how a specific feature is tested, modify the `FEATURE_COMMANDS` dictionary in `_tool/test_capabilities.py`.

To update the symbols used for testing, modify the `DEFAULT_SYMBOLS` dictionary.
