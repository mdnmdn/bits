package binance

import "github.com/mdnmdn/bits/model"

// providerErr creates a typed ProviderError with the binance provider ID.
func providerErr(kind model.ErrorKind, msg string, cause error) *model.ProviderError {
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderMessage: msg,
		Cause:           cause,
	}
}
