package symbol

import (
	"strings"
)

var commonQuoteAssets = map[string]bool{
	"USDT": true,
	"USDC": true,
	"USD":  true,
	"BTC":  true,
	"ETH":  true,
	"EUR":  true,
	"GBP":  true,
	"BUSD": true,
	"TRY":  true,
	"USDP": true,
	"BRL":  true,
}

var separators = []rune{'_', '-', '/', '.'}

func Normalize(input string) (base, quote string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", nil
	}

	input = strings.ToUpper(input)

	for _, sep := range separators {
		parts := strings.Split(input, string(sep))
		if len(parts) == 2 {
			base = strings.TrimSpace(parts[0])
			quote = strings.TrimSpace(parts[1])
			if base != "" && quote != "" {
				return base, quote, nil
			}
		}
	}

	for i := len(input) - 1; i >= 1; i-- {
		possibleQuote := input[i:]
		if commonQuoteAssets[possibleQuote] {
			base = input[:i]
			quote = possibleQuote
			if base != "" {
				return base, quote, nil
			}
		}
	}

	return "", "", nil
}

func FindSeparator(input string) int {
	for i, r := range input {
		for _, sep := range separators {
			if r == sep {
				return i
			}
		}
	}
	return -1
}
