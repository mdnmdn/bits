package provider

import "github.com/mdnmdn/bits/pkg/capability"

// Provider is the base interface implemented by all providers.
// It defines the core behavior shared by all crypto data sources.
type Provider interface {
	// ID returns the unique identifier for the provider (e.g., "binance", "coingecko").
	ID() string
	SetUserAgent(string)
	Capabilities() capability.CapabilityMatrix
}
