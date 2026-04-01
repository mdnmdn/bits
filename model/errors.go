package model

import (
	"errors"
	"fmt"
)

// ErrorKind is a normalized error category across all providers.
type ErrorKind int

const (
	ErrKindUnknown          ErrorKind = iota // catch-all
	ErrKindAuth                              // 401 / 403 / invalid API key
	ErrKindRateLimit                         // 429 / quota exceeded
	ErrKindNotFound                          // 404 / symbol not found
	ErrKindInvalidRequest                    // 400 / bad parameters
	ErrKindServerError                       // 502/503/504 / provider-side transient failure
	ErrKindNetwork                           // connection refused, DNS failure
	ErrKindCanceled                          // context.Canceled or context.DeadlineExceeded
	ErrKindParse                             // unexpected response shape
	ErrKindUnsupportedMarket                 // market type not supported
	ErrKindUnsupportedFeature                // feature not supported
)

// Retryable reports whether errors of this kind are worth retrying with backoff.
// Callers are responsible for applying exponential backoff; this method only signals intent.
func (k ErrorKind) Retryable() bool {
	return k == ErrKindRateLimit || k == ErrKindServerError || k == ErrKindNetwork
}

// ProviderError is a typed error that crosses provider boundaries.
// It preserves both normalized error kind and raw provider-specific information.
type ProviderError struct {
	// Normalized — callers use these without knowing which provider.
	Kind       ErrorKind
	ProviderID string // "binance", "bitget", … empty for resolution-layer errors

	// Raw — always preserved, never dropped.
	ProviderCode    string // provider's native error code ("00000", "400", …)
	ProviderMessage string // provider's native message text

	// Optional HTTP layer information.
	HTTPStatus int // 0 when not an HTTP error

	// Underlying cause (network error, json error, …).
	Cause error
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.ProviderCode != "" {
		return fmt.Sprintf("[%s] %s (code %s)", e.ProviderID, e.ProviderMessage, e.ProviderCode)
	}
	if e.HTTPStatus != 0 {
		return fmt.Sprintf("[%s] HTTP %d: %s", e.ProviderID, e.HTTPStatus, e.ProviderMessage)
	}
	if e.ProviderID != "" {
		return fmt.Sprintf("[%s] %s", e.ProviderID, e.ProviderMessage)
	}
	return e.ProviderMessage
}

var (
	ErrUnsupportedMarket  = errors.New("unsupported market type")
	ErrUnsupportedFeature = errors.New("unsupported feature")
)
