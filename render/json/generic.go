package renderjson

import (
	"encoding/json"
	"io"

	"github.com/mdnmdn/bits/model"
)

type envelope[T any] struct {
	Kind              string           `json:"kind"`
	Provider          string           `json:"provider"`
	Market            model.MarketType `json:"mkt"`
	Fallback          bool             `json:"fallback,omitempty"`
	RequestedProvider string           `json:"req_provider,omitempty"`
	RequestedMarket   model.MarketType `json:"req_mkt,omitempty"`
	Metadata          map[string]any   `json:"metadata,omitempty"`
	Errors            []itemError      `json:"errors,omitempty"`
	Data              T                `json:"data"`
}

type itemError struct {
	Symbol string `json:"sym"`
	Error  string `json:"err"`
}

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
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(env)
}
