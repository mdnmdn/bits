package rendertoon

import (
	"io"

	"github.com/mdnmdn/bits/internal/model"
	toon "github.com/toon-format/toon-go"
)

// envelope mirrors the JSON/YAML envelope for consistent structured output.
type envelope[T any] struct {
	Kind              string           `toon:"kind"`
	Provider          string           `toon:"provider"`
	Market            model.MarketType `toon:"mkt,omitempty"`
	Fallback          bool             `toon:"fallback,omitempty"`
	RequestedProvider string           `toon:"req_provider,omitempty"`
	RequestedMarket   model.MarketType `toon:"req_mkt,omitempty"`
	Metadata          map[string]any   `toon:"metadata,omitempty"`
	Errors            []itemError      `toon:"errors,omitempty"`
	Data              T                `toon:"data"`
}

type itemError struct {
	Symbol string `toon:"sym"`
	Error  string `toon:"err"`
}

// Render writes res as a TOON document to w, including all provenance fields.
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
	out, err := toon.MarshalString(env)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, out)
	return err
}
