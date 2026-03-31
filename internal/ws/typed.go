package ws

import (
	"github.com/mdnmdn/bits/pkg/model"
)

// TypedChan filters a raw output channel into typed responses and errors.
// Data messages that do not match *model.Response[T] are silently dropped.
//
// IMPORTANT: the caller MUST drain the returned error channel, typically in a
// background goroutine. If the error channel fills up, the forwarding goroutine
// blocks and data messages stop flowing — effectively a deadlock. Errors from
// a healthy stream are rare; a buffer of 8–16 is sufficient for bursts.
//
// Example usage:
//
//	data, errs := ws.TypedChan[model.CoinPrice](out, 100)
//	go func() {
//	    for err := range errs {
//	        log.Warn("stream error", err)
//	    }
//	}()
//	for price := range data {
//	    fmt.Printf("%s: %.2f\n", price.Symbol, price.Price)
//	}
func TypedChan[T any](
	in <-chan StreamResponse[any],
	bufSize int,
) (<-chan *model.Response[T], <-chan *model.ProviderError) {
	out := make(chan *model.Response[T], bufSize)
	errs := make(chan *model.ProviderError, 16) // small buffer; errors are rare

	go func() {
		defer close(out)
		defer close(errs)
		for sr := range in {
			if sr.Error != nil {
				if pe, ok := sr.Error.(*model.ProviderError); ok {
					select {
					case errs <- pe:
					default:
						// error channel full: drop error, never block data flow
						// log a warning or count the dropped error
					}
				}
				continue
			}
			if resp, ok := sr.Response.(*model.Response[T]); ok {
				out <- resp
			}
		}
	}()

	return out, errs
}
