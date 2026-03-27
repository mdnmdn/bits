package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/mdnmdn/bits/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestFormatError_JSONMode_WritesStderr(t *testing.T) {
	// Set up a command with -o json flag, then call formatError directly.
	// This verifies JSON is written to stderr in JSON output mode.
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// Prepare a command with -o json set.
	resetAllFlags(RootCmd)
	RootCmd.SetArgs([]string{"search", "-o", "json"})
	cmd, _, _ := RootCmd.Find([]string{"search", "-o", "json"})
	require.NotNil(t, cmd)
	_ = cmd.Flags().Set("output", "json")

	err := formatError(cmd, model.ErrInvalidAPIKey)
	require.Error(t, err) // formatError returns the original error

	_ = wErr.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(rErr)
	os.Stderr = oldStderr

	var cliErr CLIError
	require.NoError(t, json.Unmarshal(buf.Bytes(), &cliErr), "stderr should contain valid JSON: %s", buf.String())
	assert.Equal(t, "invalid_api_key", cliErr.Error)
}

func TestFormatError_TableMode_NoJSON(t *testing.T) {
	// In table mode, formatError should return the error without writing JSON.
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	resetAllFlags(RootCmd)
	RootCmd.SetArgs([]string{"search"})
	cmd, _, _ := RootCmd.Find([]string{"search"})
	require.NotNil(t, cmd)
	_ = cmd.Flags().Set("output", "table")

	err := formatError(cmd, model.ErrInvalidAPIKey)
	require.Error(t, err)

	_ = wErr.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(rErr)
	os.Stderr = oldStderr

	// stderr should be empty — formatError only writes JSON in JSON mode.
	assert.Empty(t, buf.String(), "formatError should not write to stderr in table mode")
}
