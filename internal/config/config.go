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
	TierDemo = "demo"
	TierPaid = "paid"

	demoBaseURL = "https://api.coingecko.com/api/v3"
	proBaseURL  = "https://pro-api.coingecko.com/api/v3"

	demoHeaderKey = "x-cg-demo-api-key"
	proHeaderKey  = "x-cg-pro-api-key"
)

var ValidTiers = []string{TierDemo, TierPaid}

// BinanceConfig holds Binance API credentials.
type BinanceConfig struct {
	APIKey     string `mapstructure:"api_key"`
	APISecret  string `mapstructure:"api_secret"`
	BaseURL    string `mapstructure:"base_url"`
	UseTestnet bool   `mapstructure:"use_testnet"`
}

// BitgetConfig holds Bitget API credentials.
type BitgetConfig struct {
	Key        string `mapstructure:"key"`
	Secret     string `mapstructure:"secret"`
	Passphrase string `mapstructure:"passphrase"`
	BaseURL    string `mapstructure:"base_url"`
}

type Config struct {
	// Active provider (coingecko, binance, bitget)
	Provider string `mapstructure:"provider"`

	// CoinGecko credentials (backward compatible)
	APIKey string `mapstructure:"api_key"`
	Tier   string `mapstructure:"tier"`

	// Exchange provider configs
	Binance BinanceConfig `mapstructure:"binance"`
	Bitget  BitgetConfig  `mapstructure:"bitget"`
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
	v.SetConfigFile(filepath.Join(dir, "config.yaml"))
	v.SetDefault("tier", TierDemo)
	v.SetDefault("provider", "coingecko")

	if err := v.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply env var overrides (BITS_ prefix, env vars take priority)
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

// applyEnvOverrides applies BITS_* environment variable overrides to the config.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("BITS_PROVIDER"); v != "" {
		cfg.Provider = v
	}
	// CoinGecko
	if v := os.Getenv("BITS_COINGECKO_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("BITS_COINGECKO_TIER"); v != "" {
		cfg.Tier = v
	}
	// Binance
	if v := os.Getenv("BITS_BINANCE_API_KEY"); v != "" {
		cfg.Binance.APIKey = v
	}
	if v := os.Getenv("BITS_BINANCE_API_SECRET"); v != "" {
		cfg.Binance.APISecret = v
	}
	// Bitget
	if v := os.Getenv("BITS_BITGET_KEY"); v != "" {
		cfg.Bitget.Key = v
	}
	if v := os.Getenv("BITS_BITGET_SECRET"); v != "" {
		cfg.Bitget.Secret = v
	}
	if v := os.Getenv("BITS_BITGET_PASSPHRASE"); v != "" {
		cfg.Bitget.Passphrase = v
	}
}

// ActiveProvider returns the configured provider name, defaulting to "coingecko".
func (c *Config) ActiveProvider() string {
	if c.Provider == "" {
		return "coingecko"
	}
	return c.Provider
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
	return strings.ToLower(c.Tier) == TierPaid
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
