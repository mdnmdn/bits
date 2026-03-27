package renderjson

import (
	"encoding/json"
	"io"

	"github.com/mdnmdn/bits/internal/model"
)

// envelope is the JSON output structure with provenance fields.
type envelope[T any] struct {
	Data              T                  `json:"data"`
	Provider          string             `json:"provider"`
	Market            model.MarketType   `json:"market"`
	Fallback          bool               `json:"fallback,omitempty"`
	RequestedProvider string             `json:"requested_provider,omitempty"`
	RequestedMarket   model.MarketType   `json:"requested_market,omitempty"`
	Errors            []itemError        `json:"errors,omitempty"`
}

type itemError struct {
	Symbol string `json:"symbol"`
	Error  string `json:"error"`
}

// Render writes res as JSON to w, including all provenance fields.
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
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(env)
}
