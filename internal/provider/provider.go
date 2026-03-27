package provider

import "github.com/mdnmdn/bits/internal/capability"

// Provider is the base interface implemented by all providers.
type Provider interface {
	ID() string
	SetUserAgent(string)
	Capabilities() capability.CapabilityMatrix
}
