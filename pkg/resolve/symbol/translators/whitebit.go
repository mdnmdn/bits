package translators

import (
	"strings"

	"github.com/mdnmdn/bits/pkg/model"
)

type WhiteBitTranslator struct{}

func NewWhiteBitTranslator() *WhiteBitTranslator {
	return &WhiteBitTranslator{}
}

func (t *WhiteBitTranslator) ProviderID() string {
	return "whitebit"
}

func (t *WhiteBitTranslator) ToNormalized(providerSymbol string, market model.MarketType) string {
	base, quote := normalizeInput(providerSymbol)
	if base == "" {
		return providerSymbol
	}
	if market == model.MarketFutures && strings.HasSuffix(base, "_PERP") {
		base = strings.TrimSuffix(base, "_PERP")
	}
	return base + "-" + quote
}

func (t *WhiteBitTranslator) NormalizeInput(input string) (string, string) {
	return normalizeInput(input)
}

func (t *WhiteBitTranslator) MatchesPattern(symbol string, market model.MarketType) bool {
	return true
}
