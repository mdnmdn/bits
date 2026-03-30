package rendermarkdown

import (
	"fmt"
	"io"

	"github.com/mdnmdn/bits/internal/render"
	"github.com/mdnmdn/bits/pkg/model"
	"gopkg.in/yaml.v3"
)

// envelope mirrors the JSON/YAML envelope for consistent structured output.
type envelope[T any] struct {
	Kind              string           `yaml:"kind"`
	Provider          string           `yaml:"provider"`
	Market            model.MarketType `yaml:"mkt,omitempty"`
	Fallback          bool             `yaml:"fallback,omitempty"`
	RequestedProvider string           `yaml:"req_provider,omitempty"`
	RequestedMarket   model.MarketType `yaml:"req_mkt,omitempty"`
	Metadata          map[string]any   `yaml:"metadata,omitempty"`
	Errors            []itemError      `yaml:"errors,omitempty"`
	Data              T                `yaml:"data"`
}

type itemError struct {
	Symbol string `yaml:"sym"`
	Error  string `yaml:"err"`
}

// Render writes res as a Markdown document to w.
// The document contains a heading with provider/market, an optional fallback
// blockquote, and a fenced YAML code block with the full provenance envelope.
func Render[T any](w io.Writer, res model.Response[T]) error {
	_, _ = fmt.Fprintf(w, "# %s\n\n", render.ProviderLabel(res))
	if note := render.FallbackFootnote(res); note != "" {
		_, _ = fmt.Fprintf(w, "> %s\n\n", note)
	}
	env := envelope[T]{
		Kind:              res.Kind,
		Data:              res.Data,
		Provider:          res.Provider,
		Market:            res.Market,
		Fallback:          res.Fallback,
		RequestedProvider: res.RequestedProvider,
		RequestedMarket:   res.RequestedMarket,
		Metadata:          res.Metadata,
	}
	for _, e := range res.Errors {
		env.Errors = append(env.Errors, itemError{Symbol: e.Symbol, Error: e.Err.Error()})
	}
	_, _ = fmt.Fprintln(w, "```yaml")
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(env); err != nil {
		return err
	}
	_, _ = fmt.Fprint(w, "```\n")
	return nil
}
