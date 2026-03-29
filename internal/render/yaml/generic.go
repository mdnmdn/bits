package renderyaml

import (
	"io"

	"github.com/mdnmdn/bits/pkg/model"
	"gopkg.in/yaml.v3"
)

// envelope is the YAML output structure with provenance fields.
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

// Render writes res as YAML to w, including all provenance fields.
func Render[T any](w io.Writer, res model.Response[T]) error {
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
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(env)
}
