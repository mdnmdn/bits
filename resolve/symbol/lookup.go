package symbol

import (
	"strings"

	"github.com/mdnmdn/bits/model"
)

type lookupTable struct {
	byBaseQuote map[string][]*model.Symbol
	symbols     []*model.Symbol
}

func newLookupTable(symbols []model.Symbol) *lookupTable {
	lk := &lookupTable{
		byBaseQuote: make(map[string][]*model.Symbol),
		symbols:     make([]*model.Symbol, len(symbols)),
	}

	for i := range symbols {
		sym := &symbols[i]
		lk.symbols[i] = sym

		key := normalizeKey(sym.BaseAsset, sym.QuoteAsset)
		lk.byBaseQuote[key] = append(lk.byBaseQuote[key], sym)
	}

	return lk
}

func (lk *lookupTable) find(base, quote string) *model.Symbol {
	key := normalizeKey(base, quote)

	syms, ok := lk.byBaseQuote[key]
	if !ok || len(syms) == 0 {
		return nil
	}

	if len(syms) == 1 {
		return syms[0]
	}

	for _, s := range syms {
		if s.Status == model.SymbolStatusTrading {
			return s
		}
	}

	return syms[0]
}

func normalizeKey(base, quote string) string {
	return strings.ToUpper(base) + "_" + strings.ToUpper(quote)
}
