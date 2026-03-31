package ws

import "context"

// Middleware transforms or filters a message.
// Return (nil, nil) to drop the message. Return an error to surface it downstream.
type Middleware func(ctx context.Context, msg any, next func(any) (any, error)) (any, error)

// Pipeline is an ordered chain of Middleware functions applied to every parsed message.
type Pipeline []Middleware

// Apply runs msg through all middleware in order.
// If any middleware returns an error, that error is returned immediately.
// If any middleware returns (nil, nil), the message is dropped and Apply returns (nil, nil).
func (p Pipeline) Apply(ctx context.Context, msg any) (any, error) {
	if len(p) == 0 {
		return msg, nil
	}

	var current int
	var fn func(any) (any, error)

	fn = func(m any) (any, error) {
		if current >= len(p) {
			return m, nil
		}
		middleware := p[current]
		current++
		return middleware(ctx, m, fn)
	}

	return fn(msg)
}
