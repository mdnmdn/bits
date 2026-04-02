package bitget

import (
	"context"
	"net"

	"github.com/mdnmdn/bits/model"
)

// providerErr creates a ProviderError with the given kind, message, and cause.
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
		ProviderMessage: body,
		HTTPStatus:      status,
	}
}

// apiErr creates a ProviderError from a Bitget API code and message.
func apiErr(code, msg string) *model.ProviderError {
	kind := apiCodeToKind(code)
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderCode:    code,
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

// apiCodeToKind maps Bitget API error codes to ErrorKind.
// Bitget uses string codes like "00000" (success), "40001", "40003" (auth), etc.
func apiCodeToKind(code string) model.ErrorKind {
	switch code {
	case "40001", "40003":
		return model.ErrKindAuth
	case "40004":
		return model.ErrKindNotFound
	default:
		// Codes starting with "429" are rate limit errors
		if len(code) >= 3 && code[:3] == "429" {
			return model.ErrKindRateLimit
		}
		return model.ErrKindUnknown
	}
}

// wrapNetError wraps network errors as ErrKindNetwork.
func wrapNetError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(net.Error); ok {
		return providerErr(model.ErrKindNetwork, err.Error(), err)
	}
	return err
}

// wrapContextError wraps context errors as ErrKindCanceled.
func wrapContextError(err error) error {
	if err == nil {
		return nil
	}
	if err == context.Canceled || err == context.DeadlineExceeded {
		return providerErr(model.ErrKindCanceled, err.Error(), err)
	}
	return err
}
