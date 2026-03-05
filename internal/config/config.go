package config

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	TierDemo       = "demo"
	TierAnalyst    = "analyst"
	TierLite       = "lite"
	TierPro        = "pro"
	TierEnterprise = "enterprise"

	demoBaseURL = "https://api.coingecko.com/api/v3"
	proBaseURL  = "https://pro-api.coingecko.com/api/v3"

	demoHeaderKey = "x-cg-demo-api-key"
	proHeaderKey  = "x-cg-pro-api-key"
)

var ValidTiers = []string{TierDemo, TierAnalyst, TierLite, TierPro, TierEnterprise}

type Config struct {
	APIKey string `mapstructure:"api_key"`
	Tier   string `mapstructure:"tier"`
}

func configDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}
	return filepath.Join(dir, "coingecko-cli"), nil
}

func Load() (*Config, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)
	v.SetDefault("tier", TierDemo)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return &Config{Tier: TierDemo}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	v := viper.New()
	v.Set("api_key", cfg.APIKey)
	v.Set("tier", cfg.Tier)

	path := filepath.Join(dir, "config.yaml")
	if err := v.WriteConfigAs(path); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

func (c *Config) BaseURL() string {
	if c.IsPaid() {
		return proBaseURL
	}
	return demoBaseURL
}

func (c *Config) AuthHeader() (string, string) {
	if c.IsPaid() {
		return proHeaderKey, c.APIKey
	}
	return demoHeaderKey, c.APIKey
}

func (c *Config) ApplyAuth(req *http.Request) {
	if c.APIKey != "" {
		key, val := c.AuthHeader()
		req.Header.Set(key, val)
	}
}

func (c *Config) IsPaid() bool {
	tier := strings.ToLower(c.Tier)
	return tier == TierAnalyst || tier == TierLite || tier == TierPro || tier == TierEnterprise
}

func IsValidTier(tier string) bool {
	t := strings.ToLower(tier)
	for _, v := range ValidTiers {
		if t == v {
			return true
		}
	}
	return false
}

func (c *Config) MaskedKey() string {
	if len(c.APIKey) <= 8 {
		return strings.Repeat("*", len(c.APIKey))
	}
	return c.APIKey[:4] + strings.Repeat("*", len(c.APIKey)-8) + c.APIKey[len(c.APIKey)-4:]
}
