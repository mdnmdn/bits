package rendermarkdown

import (
	"fmt"
	"io"

	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/render"
	"gopkg.in/yaml.v3"
)

// Render writes res as a Markdown document to w.
// The document contains a heading with provider/market, an optional fallback
// blockquote, and the data serialised as a fenced YAML code block.
func Render[T any](w io.Writer, res model.Response[T]) error {
	fmt.Fprintf(w, "# %s\n\n", render.ProviderLabel(res))
	if note := render.FallbackFootnote(res); note != "" {
		fmt.Fprintf(w, "> %s\n\n", note)
	}
	fmt.Fprintln(w, "```yaml")
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(res.Data); err != nil {
		return err
	}
	fmt.Fprint(w, "```\n")
	return nil
}
