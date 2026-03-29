package process

import "github.com/mdnmdn/bits/pkg/model"

// Processor is a function that enriches a Response[T] and returns it.
type Processor[T any] func(model.Response[T]) model.Response[T]

// Apply runs processors in order on res, returning the final enriched Response.
func Apply[T any](res model.Response[T], processors ...Processor[T]) model.Response[T] {
	for _, p := range processors {
		res = p(res)
	}
	return res
}
