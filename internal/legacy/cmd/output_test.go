package legacycmd

import (
	"testing"

	"github.com/mdnmdn/bits/internal/legacy/model"

	"github.com/stretchr/testify/assert"
)

func TestClassifyError_RateLimited(t *testing.T) {
	rle := &model.RateLimitError{RetryAfter: 30}
	ce := classifyError(rle)
	assert.Equal(t, "rate_limited", ce.Error)
	assert.NotNil(t, ce.RetryAfter)
	assert.Equal(t, 30, *ce.RetryAfter)
}

func TestClassifyError_RateLimitedNoRetryAfter(t *testing.T) {
	rle := &model.RateLimitError{}
	ce := classifyError(rle)
	assert.Equal(t, "rate_limited", ce.Error)
	assert.Nil(t, ce.RetryAfter)
}

func TestClassifyError_InvalidAPIKey(t *testing.T) {
	ce := classifyError(model.ErrInvalidAPIKey)
	assert.Equal(t, "invalid_api_key", ce.Error)
}

func TestClassifyError_PlanRestricted(t *testing.T) {
	ce := classifyError(model.ErrPlanRestricted)
	assert.Equal(t, "plan_restricted", ce.Error)
}

func TestClassifyError_GenericError(t *testing.T) {
	ce := classifyError(assert.AnError)
	assert.Equal(t, "error", ce.Error)
	assert.Nil(t, ce.RetryAfter)
}

