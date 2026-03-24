package model

import (
	"fmt"
	"time"
)

var (
	ErrInvalidAPIKey  = fmt.Errorf("invalid API key")
	ErrPlanRestricted = fmt.Errorf("this endpoint requires a paid plan")
	ErrRateLimited    = fmt.Errorf("rate limited — please wait and try again")
	ErrNotSupported   = fmt.Errorf("not supported by this provider")
)

// RateLimitError carries rate-limit metadata from a 429 response.
type RateLimitError struct {
	RetryAfter int       // seconds from Retry-After header; 0 if absent
	ResetAt    time.Time // from rate-limit reset header; zero if absent
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limited — retry after %d seconds", e.RetryAfter)
	}
	if !e.ResetAt.IsZero() {
		return fmt.Sprintf("rate limited — resets at %s", e.ResetAt.UTC().Format("15:04:05 UTC"))
	}
	return "rate limited — please wait and try again"
}

func (e *RateLimitError) Is(target error) bool {
	return target == ErrRateLimited
}
