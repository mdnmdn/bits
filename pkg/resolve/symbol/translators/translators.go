package translators

import (
	"strings"

	"github.com/mdnmdn/bits/pkg/model"
)

var commonQuoteAssets = map[string]bool{
	"USDT": true, "USDC": true, "USD": true, "BTC": true, "ETH": true,
	"EUR": true, "GBP": true, "BUSD": true, "TRY": true, "USDP": true, "BRL": true,
}

var separators = []rune{'_', '-', '/', '.'}

type SymbolTranslator interface {
	ProviderID() string
	ToNormalized(providerSymbol string, market model.MarketType) string
	NormalizeInput(input string) (base, quote string)
	MatchesPattern(symbol string, market model.MarketType) bool
}

func NormalizeSymbol(symbol string) string {
	if symbol == "" {
		return symbol
	}

	symbol = strings.ToUpper(symbol)

	for _, sep := range separators {
		parts := strings.Split(symbol, string(sep))
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return parts[0] + "-" + parts[1]
		}
	}

	for i := len(symbol) - 1; i > 0; i-- {
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

func normalizeInput(input string) (base, quote string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ""
	}

	input = strings.ToUpper(input)

	for _, sep := range separators {
		parts := strings.Split(input, string(sep))
		if len(parts) == 2 {
			base = strings.TrimSpace(parts[0])
			quote = strings.TrimSpace(parts[1])
			if base != "" && quote != "" {
				return base, quote
			}
		}
	}

	for i := len(input) - 1; i > 0; i-- {
		possibleQuote := input[i:]
		if commonQuoteAssets[possibleQuote] {
			base = input[:i]
			quote = possibleQuote
			if base != "" {
				return base, quote
			}
		}
	}

	return "", ""
}
