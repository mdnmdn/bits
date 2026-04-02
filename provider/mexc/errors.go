package mexc

import (
	"context"
	"net/http"

	"github.com/mdnmdn/bits/model"
)

// providerErr creates a ProviderError with the MEXC provider ID and given kind.
func providerErr(kind model.ErrorKind, msg string, cause error) *model.ProviderError {
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderMessage: msg,
		Cause:           cause,
	}
}

// httpErr converts an HTTP status code and response body to a typed ProviderError.
func httpErr(status int, body string) *model.ProviderError {
	kind := httpStatusToKind(status)
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderCode:    "",
		ProviderMessage: body,
		HTTPStatus:      status,
		Cause:           nil,
	}
}

// httpStatusToKind maps HTTP status codes to ErrorKind.
func httpStatusToKind(status int) model.ErrorKind {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden: // 401, 403
		return model.ErrKindAuth
	case http.StatusNotFound: // 404
		return model.ErrKindNotFound
	case http.StatusTooManyRequests: // 429
		return model.ErrKindRateLimit
	case http.StatusBadRequest: // 400
		return model.ErrKindInvalidRequest
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout: // 502, 503, 504
		return model.ErrKindServerError
	default:
		if status >= http.StatusInternalServerError {
			return model.ErrKindServerError
		}
		return model.ErrKindUnknown
	}
}

// classifyError wraps various error types into appropriate ProviderError categories.
func classifyError(err error) *model.ProviderError {
	if err == nil {
		return nil
	}

	// Check if it's already a ProviderError
	if pe, ok := err.(*model.ProviderError); ok {
		return pe
	}

	// Check for context cancellation
	if err == context.Canceled {
		return providerErr(model.ErrKindCanceled, "request was canceled", err)
	}

	// Check for context deadline exceeded
	if err == context.DeadlineExceeded {
		return providerErr(model.ErrKindCanceled, "request deadline exceeded", err)
	}

	// Default: treat as network/unknown error
	return providerErr(model.ErrKindNetwork, err.Error(), err)
}
