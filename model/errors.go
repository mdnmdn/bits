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

// Unwrap returns the underlying cause, enabling errors.Is/As chaining.
func (e *ProviderError) Unwrap() error { return e.Cause }

// Is implements errors.Is matching by Kind for sentinel comparisons.
// This ensures errors.Is(err, model.ErrUnsupportedMarket) works for any
// *ProviderError with the matching Kind, not just the exact sentinel pointer.
func (e *ProviderError) Is(target error) bool {
	t, ok := target.(*ProviderError)
	if !ok {
		return false
	}
	// Match sentinels by Kind when the target has no ProviderID.
	if t.ProviderID == "" {
		return e.Kind == t.Kind
	}
	return e == t
}

var (
	ErrUnsupportedMarket  = &ProviderError{Kind: ErrKindUnsupportedMarket, ProviderMessage: "unsupported market type"}
	ErrUnsupportedFeature = &ProviderError{Kind: ErrKindUnsupportedFeature, ProviderMessage: "unsupported feature"}
)

// WrapError wraps an arbitrary error as ErrKindUnknown. Use only as a
// temporary shim during migration; replace with a typed providerErr call.
func WrapError(providerID string, err error) *ProviderError {
	if err == nil {
		return nil
	}
	var pe *ProviderError
	if errors.As(err, &pe) {
		return pe
	}
	return &ProviderError{
		Kind:            ErrKindUnknown,
		ProviderID:      providerID,
		ProviderMessage: err.Error(),
		Cause:           err,
	}
}
