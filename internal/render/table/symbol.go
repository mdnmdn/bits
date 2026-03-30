package rendertable

import (
	"strings"
)

var commonQuoteAssets = map[string]bool{
	"USDT": true, "USDC": true, "USD": true, "BTC": true, "ETH": true,
	"EUR": true, "GBP": true, "BUSD": true, "TRY": true, "USDP": true, "BRL": true,
}

var symbolSeparators = []rune{'_', '-', '/', '.'}

func NormalizeSymbol(symbol string) string {
	if symbol == "" {
		return symbol
	}

	symbol = strings.ToUpper(symbol)

	for _, sep := range symbolSeparators {
		parts := strings.Split(symbol, string(sep))
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return parts[0] + "-" + parts[1]
		}
	}

	for i := len(symbol) - 1; i >= 1; i-- {
		possibleQuote := symbol[i:]
		if commonQuoteAssets[possibleQuote] {
			base := symbol[:i]
			if base != "" {
				return base + "-" + possibleQuote
			}
		}
	}

	return symbol
}
