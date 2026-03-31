package middleware

import (
	"github.com/mdnmdn/bits/internal/ws"
)

func OrderBookReconstructorMW() ws.Middleware {
	reconstructor := NewOrderBookReconstructor()
	return reconstructor.Middleware
}

func CRC32ValidatorMW() ws.Middleware {
	validator := NewCRC32Validator()
	return validator.Middleware
}
