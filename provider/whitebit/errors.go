package whitebit

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/mdnmdn/bits/model"
)

// httpStatusToKind maps HTTP status codes to normalized error kinds.
func httpStatusToKind(status int) model.ErrorKind {
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return model.ErrKindAuth
	case status == http.StatusNotFound:
		return model.ErrKindNotFound
	case status == http.StatusTooManyRequests:
		return model.ErrKindRateLimit
	case status == http.StatusBadRequest:
		return model.ErrKindInvalidRequest
	case status >= http.StatusInternalServerError:
		return model.ErrKindServerError
	default:
		return model.ErrKindUnknown
	}
}

// providerErr creates a typed ProviderError with the given kind, message, and cause.
func providerErr(kind model.ErrorKind, msg string, cause error) *model.ProviderError {
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderMessage: msg,
		Cause:           cause,
	}
}

// httpErr creates a ProviderError from an HTTP status code and response body.
func httpErr(status int, body string) *model.ProviderError {
	kind := httpStatusToKind(status)
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderCode:    http.StatusText(status),
		ProviderMessage: body,
		HTTPStatus:      status,
	}
}

// apiErr creates a ProviderError from a WhiteBit API response with success=false.
// WhiteBit does not return numeric error codes, so we use ErrKindUnknown with the message.
func apiErr(msg string) *model.ProviderError {
	return &model.ProviderError{
		Kind:            model.ErrKindUnknown,
		ProviderID:      providerID,
		ProviderMessage: msg,
	}
}

// networkErr wraps network errors as ErrKindNetwork.
func networkErr(cause error) *model.ProviderError {
	msg := "network error"
	if cause != nil {
		msg = cause.Error()
	}
	return providerErr(model.ErrKindNetwork, msg, cause)
}

// contextErr wraps context errors as ErrKindCanceled.
func contextErr(cause error) *model.ProviderError {
	msg := "context canceled"
	if cause != nil {
		msg = cause.Error()
	}
	return providerErr(model.ErrKindCanceled, msg, cause)
}

// isNetworkError checks if an error is a network-related error.
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

// isContextError checks if an error is a context cancellation or deadline.
func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
