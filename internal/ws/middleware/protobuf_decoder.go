package middleware

import (
	"google.golang.org/protobuf/proto"
)

// ProtobufDecoder creates a middleware that decodes binary protobuf messages.
// The newMsg function should return a new empty proto.Message instance to unmarshal into.
func ProtobufDecoder(newMsg func() proto.Message) func(ctx interface{}, msg any, next func(any) (any, error)) (any, error) {
	return func(ctx interface{}, msg any, next func(any) (any, error)) (any, error) {
		data, ok := msg.([]byte)
		if !ok {
			return next(msg)
		}

		m := newMsg()
		if err := proto.Unmarshal(data, m); err != nil {
			return nil, err
		}

		return next(m)
	}
}
