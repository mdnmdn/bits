package translators

import (
	"strings"

	"github.com/mdnmdn/bits/pkg/model"
)

type CryptoComTranslator struct{}

func NewCryptoComTranslator() *CryptoComTranslator {
	return &CryptoComTranslator{}
}

func (t *CryptoComTranslator) ProviderID() string {
	return "cryptocom"
}

func (t *CryptoComTranslator) ToNormalized(providerSymbol string, market model.MarketType) string {
	symbol := providerSymbol
	if market == model.MarketFutures {
		symbol = strings.TrimSuffix(symbol, "-PERP")
	}
	base, quote, _ := normalizeInput(symbol)
	if base == "" {
		return providerSymbol
	}
	return base + "-" + quote
}

func (t *CryptoComTranslator) NormalizeInput(input string) (string, string, error) {
	return normalizeInput(input)
}
