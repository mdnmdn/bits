package cryptocom

import (
	"fmt"

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
// Uses httpStatusToKind to map the status to an ErrorKind.
func httpErr(status int, body string) *model.ProviderError {
	kind := httpStatusToKind(status)
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderCode:    fmt.Sprintf("%d", status),
		ProviderMessage: body,
		HTTPStatus:      status,
	}
}

// apiErr constructs a ProviderError from a Crypto.com API error code and message.
// Crypto.com API codes are integers, not strings.
// Uses apiCodeToKind to map the code to an ErrorKind.
func apiErr(code int, msg string) *model.ProviderError {
	kind := apiCodeToKind(code)
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderCode:    fmt.Sprintf("%d", code),
		ProviderMessage: msg,
	}
}

// httpStatusToKind maps HTTP status codes to ErrorKind.
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

// apiCodeToKind maps Crypto.com API error codes to ErrorKind.
func apiCodeToKind(code int) model.ErrorKind {
	switch code {
	case 40001:
		return model.ErrKindAuth
	case 10006:
		return model.ErrKindRateLimit
	default:
		return model.ErrKindUnknown
	}
}

