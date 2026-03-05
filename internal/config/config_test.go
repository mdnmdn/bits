package config

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidTier(t *testing.T) {
	tests := []struct {
		tier  string
		valid bool
	}{
		{"demo", true},
		{"analyst", true},
		{"lite", true},
		{"pro", true},
		{"enterprise", true},
		{"Demo", true},
		{"PRO", true},
		{"free", false},
		{"", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.valid, IsValidTier(tt.tier), "tier=%q", tt.tier)
	}
}

func TestIsPaid(t *testing.T) {
	tests := []struct {
		tier string
		paid bool
	}{
		{"demo", false},
		{"analyst", true},
		{"lite", true},
		{"pro", true},
		{"enterprise", true},
	}
	for _, tt := range tests {
		cfg := &Config{Tier: tt.tier}
		assert.Equal(t, tt.paid, cfg.IsPaid(), "tier=%q", tt.tier)
	}
}

func TestBaseURL(t *testing.T) {
	demo := &Config{Tier: TierDemo}
	assert.Equal(t, demoBaseURL, demo.BaseURL())

	pro := &Config{Tier: TierPro}
	assert.Equal(t, proBaseURL, pro.BaseURL())
}

func TestAuthHeader(t *testing.T) {
	demo := &Config{APIKey: "demo-key-123", Tier: TierDemo}
	key, val := demo.AuthHeader()
	assert.Equal(t, demoHeaderKey, key)
	assert.Equal(t, "demo-key-123", val)

	pro := &Config{APIKey: "pro-key-456", Tier: TierPro}
	key, val = pro.AuthHeader()
	assert.Equal(t, proHeaderKey, key)
	assert.Equal(t, "pro-key-456", val)
}

func TestApplyAuth(t *testing.T) {
	cfg := &Config{APIKey: "test-key", Tier: TierDemo}
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	cfg.ApplyAuth(req)
	assert.Equal(t, "test-key", req.Header.Get(demoHeaderKey))

	// No key set — should not add header
	cfg2 := &Config{Tier: TierDemo}
	req2, _ := http.NewRequest("GET", "https://example.com", nil)
	cfg2.ApplyAuth(req2)
	assert.Empty(t, req2.Header.Get(demoHeaderKey))
}

func TestLoadMissingConfigReturnsDefault(t *testing.T) {
	// Point HOME to a temp dir so os.UserConfigDir() finds no config
	t.Setenv("HOME", t.TempDir())
	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, TierDemo, cfg.Tier)
	assert.Empty(t, cfg.APIKey)
}

func TestMaskedKey(t *testing.T) {
	tests := []struct {
		key    string
		expect string
	}{
		{"", ""},
		{"abcd", "****"},
		{"abcdefgh", "********"},
		{"abcdefghij", "abcd**ghij"},
		{"CG-abc123def456ghi", "CG-a**********6ghi"},
	}
	for _, tt := range tests {
		cfg := &Config{APIKey: tt.key}
		assert.Equal(t, tt.expect, cfg.MaskedKey(), "key=%q", tt.key)
	}
}
