package model

import "strings"

var commonQuoteAssets = map[string]bool{
	"USDT": true, "USDC": true, "USD": true, "BTC": true, "ETH": true,
	"EUR": true, "GBP": true, "BNB": true, "BUSD": true, "TRY": true,
	"USDP": true, "BRL": true,
}

var symbolSeparators = []string{"_", "-", "/", "."}

// NormalizeSymbol converts a provider-native trading symbol to BASE-QUOTE format.
// Examples: "BTCUSDT" → "BTC-USDT", "BTC_USDT" → "BTC-USDT", "btc_usdt" → "BTC-USDT".
// Returns s unchanged if the symbol cannot be parsed.
func NormalizeSymbol(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	upper := strings.ToUpper(s)
	for _, sep := range symbolSeparators {
		if parts := strings.SplitN(upper, sep, 2); len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return parts[0] + "-" + parts[1]
		}
	}
	for i := len(upper) - 1; i > 0; i-- {
		if commonQuoteAssets[upper[i:]] && i > 0 {
			return upper[:i] + "-" + upper[i:]
		}
	}
	return s
}
