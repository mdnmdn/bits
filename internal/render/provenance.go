package render

import (
	"fmt"
	"github.com/mdnmdn/bits/pkg/model"
)

// FallbackFootnote returns a human-readable note when a fallback occurred.
// Returns empty string when res.Fallback is false.
func FallbackFootnote[T any](res model.Response[T]) string {
	if !res.Fallback {
		return ""
	}
	return fmt.Sprintf("† served by %s (requested: %s)", res.Provider, res.RequestedProvider)
}

// ProviderLabel returns a short "provider/market" label for the response.
func ProviderLabel[T any](res model.Response[T]) string {
	if res.Market == "" {
		return res.Provider
	}
	return fmt.Sprintf("%s/%s", res.Provider, res.Market)
}
