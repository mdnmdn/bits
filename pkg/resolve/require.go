package resolve

import (
	"fmt"
	"github.com/mdnmdn/bits/pkg/provider"
)

// Require asserts that provider p implements interface T.
// Returns a descriptive error if not.
func Require[T any](p provider.Provider, feature string) (T, error) {
	v, ok := p.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("provider %q does not support %s", p.ID(), feature)
	}
	return v, nil
}
