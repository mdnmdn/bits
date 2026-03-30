package translators

import (
	"github.com/mdnmdn/bits/pkg/model"
)

type DefaultTranslator struct{}

func NewDefaultTranslator() *DefaultTranslator {
	return &DefaultTranslator{}
}

func (t *DefaultTranslator) ProviderID() string {
	return ""
}

func (t *DefaultTranslator) ToNormalized(providerSymbol string, market model.MarketType) string {
	base, quote := normalizeInput(providerSymbol)
	if base == "" {
		return providerSymbol
	}
	return base + "-" + quote
}

func (t *DefaultTranslator) NormalizeInput(input string) (string, string) {
	return normalizeInput(input)
}

func (t *DefaultTranslator) MatchesPattern(symbol string, market model.MarketType) bool {
	return true
}
