package okx

import (
	"context"
	"errors"
	"net"

	"github.com/mdnmdn/bits/model"
)

func providerErr(kind model.ErrorKind, msg string, cause error) *model.ProviderError {
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderMessage: msg,
		Cause:           cause,
	}
}

func httpErr(status int, body string) *model.ProviderError {
	return &model.ProviderError{
		Kind:            httpStatusToKind(status),
		ProviderID:      providerID,
		ProviderMessage: body,
		HTTPStatus:      status,
	}
}

func apiErr(code, msg string) *model.ProviderError {
	return &model.ProviderError{
		Kind:            apiCodeToKind(code),
		ProviderID:      providerID,
		ProviderCode:    code,
		ProviderMessage: msg,
	}
}

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

// apiCodeToKind maps OKX string codes ("0" is success) to ErrorKind.
func apiCodeToKind(code string) model.ErrorKind {
	switch code {
	case "50011", "50061": // rate limit reached
		return model.ErrKindRateLimit
	case "51001": // instrument does not exist
		return model.ErrKindNotFound
	case "50013", "50026", "50001", "50002":
		return model.ErrKindServerError
	case "51000", "50014":
		return model.ErrKindInvalidRequest
	default:
		return model.ErrKindUnknown
	}
}

func wrapTransportError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return providerErr(model.ErrKindCanceled, err.Error(), err)
	}
	var ne net.Error
	if errors.As(err, &ne) {
		return providerErr(model.ErrKindNetwork, err.Error(), err)
	}
	return err
}
