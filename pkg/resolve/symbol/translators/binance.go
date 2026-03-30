package translators

import (
	"github.com/mdnmdn/bits/pkg/model"
)

type BinanceTranslator struct{}

func NewBinanceTranslator() *BinanceTranslator {
	return &BinanceTranslator{}
}

func (t *BinanceTranslator) ProviderID() string {
	return "binance"
}

func (t *BinanceTranslator) ToNormalized(providerSymbol string, market model.MarketType) string {
	base, quote, _ := normalizeInput(providerSymbol)
	if base == "" {
		return providerSymbol
	}
	return base + "-" + quote
}

func (t *BinanceTranslator) NormalizeInput(input string) (string, string, error) {
	return normalizeInput(input)
}
