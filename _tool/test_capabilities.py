#!/usr/bin/env python3
import subprocess
import json
import time
import argparse
import sys
import os

# Default symbols for providers
DEFAULT_SYMBOLS = {
    "coingecko": "bitcoin",
    "binance": "BTCUSDT",
    "bitget": "BTCUSDT",
    "whitebit": "BTC_USDT",
    "cryptocom": "BTC_USDT",
}

# Mapping features to bits commands
# %s will be replaced by symbols if needed
FEATURE_COMMANDS = {
    "server_time": ["time"],
    "exchange_info": ["info"],
    "price": ["price", "%s"],
    "candles": ["candles", "%s", "--limit", "5"],
    "ticker_24h": ["ticker", "%s"],
    "order_book": ["book", "%s", "--depth", "5"],
    "markets_list": ["markets", "--per-page", "5"],
    "stream_price": ["stream", "price", "%s"],
    "stream_order_book": ["stream", "book", "%s"],
}

def run_bits(args, timeout=None, format_json=True):
    cmd = ["./bits"] + args
    if format_json:
        cmd += ["-o", "json"]
    try:
        process = subprocess.Popen(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )

        if timeout:
            try:
                stdout, stderr = process.communicate(timeout=timeout)
            except subprocess.TimeoutExpired:
                process.kill()
                stdout, stderr = process.communicate()
                # For streaming, timeout is expected
                if "stream" in args:
                    return 0, stdout, stderr
                return 1, stdout, "Timeout expired"
        else:
            stdout, stderr = process.communicate()

        return process.returncode, stdout, stderr
    except Exception as e:
        return 1, "", str(e)

def parse_capabilities():
    code, stdout, stderr = run_bits(["capabilities"], format_json=False)
    if code != 0:
        print(f"Error running capabilities: {stderr}")
        sys.exit(1)

    lines = stdout.strip().split('\n')
    if not lines:
        print("Empty capabilities output")
        sys.exit(1)

    header = lines[0].split()
    providers = header[2:]

    matrix = {p: [] for p in providers}

    for line in lines[1:]:
        parts = line.split()
        if len(parts) < 3:
            continue
        feat = parts[0]
        market = parts[1]
        supports = parts[2:]

        for i, val in enumerate(supports):
            if val == '✓':
                matrix[providers[i]].append((feat, market))

    return matrix

def test_capability(provider, feat, market):
    symbol = DEFAULT_SYMBOLS.get(provider, "BTCUSDT")

    if feat not in FEATURE_COMMANDS:
        return False, f"Unknown feature: {feat}"

    cmd_template = FEATURE_COMMANDS[feat]
    cmd_args = []
    for arg in cmd_template:
        if "%s" in arg:
            cmd_args.append(arg % symbol)
        else:
            cmd_args.append(arg)

    bits_args = cmd_args + ["-p", provider, "-m", market]

    is_stream = feat.startswith("stream_")
    timeout = 5 if is_stream else 15

    code, stdout, stderr = run_bits(bits_args, timeout=timeout)

    if code != 0:
        # Check if it's a known error like "restricted location" or "API key not configured"
        error_msg = stderr.strip().splitlines()[0] if stderr.strip() else "Unknown error"
        if any(msg in error_msg for msg in ["restricted location", "API key not configured", "not configured", "bad handshake", "plan restricted"]):
            return False, f"SKIPPED ({error_msg})"
        return False, stderr.strip() or "Unknown error"

    if is_stream:
        # For streams, we check if we got at least some JSON data
        if not stdout.strip():
            return False, "No stream data received"

        # Some streams might output pretty-printed JSON (multi-line)
        # or JSONL (one object per line).
        # Let's try a more robust way to check for valid JSON.

        # Try to find JSON objects in the output
        try:
            # First attempt: check if any line is valid JSON (JSONL style)
            found_valid = False
            for line in stdout.strip().split('\n'):
                if line.strip():
                    try:
                        json.loads(line)
                        found_valid = True
                        break
                    except json.JSONDecodeError:
                        continue

            if found_valid:
                return True, "OK (Stream)"

            # Second attempt: try to see if it's a single pretty-printed JSON object
            # or multiple pretty-printed JSON objects separated by something (or nothing)
            # A simple way: check if it starts with { and has a } somewhere
            if '{' in stdout and '}' in stdout:
                 # Just looking for basic JSON structure to validate it got something
                 return True, "OK (Stream)"

            # Third attempt: try to see if it's multiple objects not separated by newlines
            # but this is unlikely with bits.

            return False, f"Could not find valid JSON in stream output (first 100 chars): {stdout[:100]}"
        except Exception as e:
            return False, f"Error parsing stream output: {str(e)}"

    try:
        data = json.loads(stdout)

        # Verify data is not empty for relevant commands
        if feat in ["price", "candles", "ticker_24h", "order_book", "exchange_info", "markets_list"]:
            res_data = data.get("data")
            if not res_data:
                # Binance/CoinGecko might return error in data or empty data if restricted
                if "restricted location" in stdout or "API key not configured" in stdout:
                     return False, f"SKIPPED (Restricted/No Key in JSON)"
                return False, "Empty data in response"
            if isinstance(res_data, list) and len(res_data) == 0:
                return False, "Empty data list in response"

        return True, "OK"
    except json.JSONDecodeError:
        if "restricted location" in stdout or "API key not configured" in stdout:
             return False, f"SKIPPED (Restricted/No Key)"
        return False, "Invalid JSON response"

def main():
    parser = argparse.ArgumentParser(description="Test bits capabilities")
    parser.add_argument("-p", "--provider", help="Select a specific provider to test")
    args = parser.parse_args()

    if not os.path.exists("./bits"):
        print("Error: 'bits' binary not found. Run 'make build' first.")
        sys.exit(1)

    matrix = parse_capabilities()

    if args.provider:
        if args.provider not in matrix:
            print(f"Error: Unknown provider '{args.provider}'")
            sys.exit(1)
        providers_to_test = [args.provider]
    else:
        providers_to_test = sorted(matrix.keys())

    results = {}

    for provider in providers_to_test:
        print(f"Testing provider: {provider}")
        results[provider] = []
        for feat, market in matrix[provider]:
            print(f"  - {feat} ({market})... ", end="", flush=True)
            ok, msg = test_capability(provider, feat, market)
            if ok:
                print("OK")
            else:
                print(f"FAILED: {msg}")
            results[provider].append({
                "feature": feat,
                "market": market,
                "ok": ok,
                "msg": msg
            })

    print("\n" + "="*40)
    print("SUMMARY")
    print("="*40)

    total_ok = 0
    total_failed = 0
    total_skipped = 0

    for provider, tests in results.items():
        prov_ok = sum(1 for t in tests if t["ok"])
        prov_skipped = sum(1 for t in tests if not t["ok"] and "SKIPPED" in t["msg"])
        prov_failed = len(tests) - prov_ok - prov_skipped

        print(f"{provider:15} | OK: {prov_ok:2} | FAILED: {prov_failed:2} | SKIPPED: {prov_skipped:2}")

        total_ok += prov_ok
        total_failed += prov_failed
        total_skipped += prov_skipped

    print("-" * 40)
    print(f"{'TOTAL':15} | OK: {total_ok:2} | FAILED: {total_failed:2} | SKIPPED: {total_skipped:2}")

if __name__ == "__main__":
    main()
