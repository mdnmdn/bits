package translators

import (
	"strings"

	"github.com/mdnmdn/bits/pkg/model"
)

type BitgetTranslator struct{}

func NewBitgetTranslator() *BitgetTranslator {
	return &BitgetTranslator{}
}

func (t *BitgetTranslator) ProviderID() string {
	return "bitget"
}

func (t *BitgetTranslator) ToNormalized(providerSymbol string, market model.MarketType) string {
	symbol := providerSymbol
	if market == model.MarketFutures {
		symbol = strings.TrimSuffix(symbol, "_UMCBL")
		symbol = strings.TrimSuffix(symbol, "_PERP")
	}
	base, quote, _ := normalizeInput(symbol)
	if base == "" {
		return providerSymbol
	}
	return base + "-" + quote
}

func (t *BitgetTranslator) NormalizeInput(input string) (string, string, error) {
	return normalizeInput(input)
}
