package renderyaml

import (
	"io"

	"github.com/mdnmdn/bits/internal/model"
	"gopkg.in/yaml.v3"
)

// envelope is the YAML output structure with provenance fields.
type envelope[T any] struct {
	Data              T                `yaml:"data"`
	Provider          string           `yaml:"provider"`
	Market            model.MarketType `yaml:"market,omitempty"`
	Fallback          bool             `yaml:"fallback,omitempty"`
	RequestedProvider string           `yaml:"requested_provider,omitempty"`
	RequestedMarket   model.MarketType `yaml:"requested_market,omitempty"`
	Errors            []itemError      `yaml:"errors,omitempty"`
}

type itemError struct {
	Symbol string `yaml:"symbol"`
	Error  string `yaml:"error"`
}

// Render writes res as YAML to w, including all provenance fields.
func Render[T any](w io.Writer, res model.Response[T]) error {
	env := envelope[T]{
		Data:              res.Data,
		Provider:          res.Provider,
		Market:            res.Market,
		Fallback:          res.Fallback,
		RequestedProvider: res.RequestedProvider,
		RequestedMarket:   res.RequestedMarket,
	}
	for _, e := range res.Errors {
		env.Errors = append(env.Errors, itemError{Symbol: e.Symbol, Error: e.Err.Error()})
	}
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(env)
}
