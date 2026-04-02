package resolve

import (
	"context"
	"github.com/mdnmdn/bits/model"
	"sync"
)

// FanOut calls fn for each symbol in parallel and collects results into a
// single Response[[]T]. Partial failures go into Response.Errors; successes
// are collected into Response.Data. The returned Response.Provider and
// Response.Market are taken from the first successful result (or left empty).
func FanOut[T any](
	ctx context.Context,
	symbols []string,
	fn func(ctx context.Context, symbol string) (model.Response[T], error),
) model.Response[[]T] {
	type result struct {
		res model.Response[T]
		err error
		sym string
	}

	ch := make(chan result, len(symbols))
	var wg sync.WaitGroup

	for _, sym := range symbols {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			res, err := fn(ctx, s)
			ch <- result{res: res, err: err, sym: s}
		}(sym)
	}

	wg.Wait()
	close(ch)

	out := model.Response[[]T]{}
	for r := range ch {
		if r.err != nil {
			out.Errors = append(out.Errors, model.ItemError{Symbol: r.sym, Err: model.WrapError("", r.err)})
			continue
		}
		out.Data = append(out.Data, r.res.Data)
		out.Errors = append(out.Errors, r.res.Errors...)
		if out.Provider == "" {
			out.Provider = r.res.Provider
			out.Market = r.res.Market
			out.Fallback = r.res.Fallback
			out.RequestedProvider = r.res.RequestedProvider
			out.RequestedMarket = r.res.RequestedMarket
		}
	}
	return out
}
