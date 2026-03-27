package provider

import (
	"fmt"

	"github.com/mdnmdn/bits/internal/config"
)

// NewProvider constructs a provider by name using the given config.
// Implementations are registered in Phase 5.
func NewProvider(name string, cfg *config.Config) (Provider, error) {
	return nil, fmt.Errorf("provider %q not yet implemented", name)
}

// RegisteredProviderIDs returns the IDs of all registered providers.
// Populated in Phase 5.
func RegisteredProviderIDs() []string {
	return nil
}
