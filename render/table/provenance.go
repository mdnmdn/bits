package table

import (
	"fmt"
	"io"

	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/render"
	"github.com/mdnmdn/bits/resolve/symbol/translators"
)

func NormalizeSymbol(s string) string {
	return translators.NormalizeSymbol(s)
}

func printHeader[T any](w io.Writer, res model.Response[T]) {
	_, _ = fmt.Fprintf(w, "provider: %s\n\n", res.Provider)
}

func printFooter[T any](w io.Writer, res model.Response[T]) {
	if note := render.FallbackFootnote(res); note != "" {
		_, _ = fmt.Fprintf(w, "\n%s\n", note)
	}
}
