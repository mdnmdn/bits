package process

import "github.com/mdnmdn/bits/model"

type Processor[T any] func(model.Response[T]) model.Response[T]

func Apply[T any](res model.Response[T], processors ...Processor[T]) model.Response[T] {
	for _, p := range processors {
		res = p(res)
	}
	return res
}
