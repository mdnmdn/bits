package translators

import (
	"strings"

	"github.com/mdnmdn/bits/pkg/model"
)

type MEXCTranslator struct{}

func NewMEXCTranslator() *MEXCTranslator {
	return &MEXCTranslator{}
}

func (t *MEXCTranslator) ProviderID() string {
	return "mexc"
}

func (t *MEXCTranslator) ToNormalized(providerSymbol string, market model.MarketType) string {
	symbol := providerSymbol
	if market == model.MarketFutures && strings.HasPrefix(symbol, "CCM") {
		symbol = strings.TrimPrefix(symbol, "CCM")
	}
	base, quote := normalizeInput(symbol)
	if base == "" {
		return providerSymbol
	}
	return base + "-" + quote
}

func (t *MEXCTranslator) NormalizeInput(input string) (string, string) {
	return normalizeInput(input)
}

func (t *MEXCTranslator) MatchesPattern(symbol string, market model.MarketType) bool {
	return true
}
