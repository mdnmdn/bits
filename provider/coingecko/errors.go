package coingecko

import (
	"context"
	"net"

	"github.com/mdnmdn/bits/model"
)

// providerErr constructs a ProviderError with the given kind and message.
func providerErr(kind model.ErrorKind, msg string, cause error) *model.ProviderError {
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderMessage: msg,
		Cause:           cause,
	}
}

// httpErr constructs a ProviderError from an HTTP status code and response body.
func httpErr(status int, body string) *model.ProviderError {
	kind := httpStatusToKind(status)
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderMessage: body,
		HTTPStatus:      status,
	}
}

// httpStatusToKind maps HTTP status codes to ErrorKind categories.
func httpStatusToKind(status int) model.ErrorKind {
	switch status {
	case 401, 403:
		return model.ErrKindAuth
	case 404:
		return model.ErrKindNotFound
	case 429:
		return model.ErrKindRateLimit
	case 400:
		return model.ErrKindInvalidRequest
	default:
		if status >= 500 {
			return model.ErrKindServerError
		}
		return model.ErrKindUnknown
	}
}

// contextErr wraps context errors (Canceled, DeadlineExceeded) as ErrKindCanceled.
func contextErr(err error) *model.ProviderError {
	if err == context.Canceled {
		return providerErr(model.ErrKindCanceled, "context canceled", err)
	}
	if err == context.DeadlineExceeded {
		return providerErr(model.ErrKindCanceled, "context deadline exceeded", err)
	}
	return nil
}

// isNetworkErr reports whether the error is a network-level error.
func isNetworkErr(err error) bool {
	if err == nil || err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}
	_, ok := err.(net.Error)
	return ok
}
