package rendertable

import (
	"fmt"
	"io"

	"github.com/mdnmdn/bits/internal/render"
	"github.com/mdnmdn/bits/model"
)

// printProvenance writes a provider/market header line and, if a fallback
// occurred, a footnote after the table body. Call printHeader at the top and
// printFooter at the bottom of every table renderer.
func printHeader[T any](w io.Writer, res model.Response[T]) {
	_, _ = fmt.Fprintf(w, "provider: %s\n\n", render.ProviderLabel(res))
}

func printFooter[T any](w io.Writer, res model.Response[T]) {
	if note := render.FallbackFootnote(res); note != "" {
		_, _ = fmt.Fprintf(w, "\n%s\n", note)
	}
}
