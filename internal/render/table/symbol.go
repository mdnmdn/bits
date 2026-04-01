package rendertable

import (
	"github.com/mdnmdn/bits/resolve/symbol/translators"
)

func NormalizeSymbol(symbol string) string {
	return translators.NormalizeSymbol(symbol)
}
