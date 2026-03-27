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

	MarketTypeSpot   = "spot"
	MarketTypeMargin = "margin"
	MarketTypeFuture = "future"

	demoBaseURL = "https://api.coingecko.com/api/v3"
	proBaseURL  = "https://pro-api.coingecko.com/api/v3"

	demoHeaderKey = "x-cg-demo-api-key"
	proHeaderKey  = "x-cg-pro-api-key"
)

var ValidTiers = []string{TierDemo, TierPaid}

// MarketConfig holds settings for a specific market type.
type MarketConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	UseTestnet bool `mapstructure:"use_testnet"`
}

// CoinGeckoConfig holds CoinGecko API credentials.
type CoinGeckoConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Tier    string `mapstructure:"tier"`
	BaseURL string `mapstructure:"base_url"`
}

func (c CoinGeckoConfig) IsPaid() bool {
	return strings.ToLower(c.Tier) == TierPaid
}

func (c CoinGeckoConfig) GetBaseURL() string {
	if c.IsPaid() {
		return proBaseURL
	}
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return demoBaseURL
}

func (c CoinGeckoConfig) GetAuthHeader() (string, string) {
	if c.IsPaid() {
		return proHeaderKey, c.APIKey
	}
	return demoHeaderKey, c.APIKey
}

// BinanceConfig holds Binance API credentials for all market types.
type BinanceConfig struct {
	APIKey    string       `mapstructure:"api_key"`
	APISecret string       `mapstructure:"api_secret"`
	BaseURL   string       `mapstructure:"base_url"`
	Spot      MarketConfig `mapstructure:"spot"`
	Margin    MarketConfig `mapstructure:"margin"`
	Futures   MarketConfig `mapstructure:"futures"`
}

// BitgetConfig holds Bitget API credentials for all market types.
type BitgetConfig struct {
	APIKey     string       `mapstructure:"api_key"`
	APISecret  string       `mapstructure:"api_secret"`
	Passphrase string       `mapstructure:"passphrase"`
	BaseURL    string       `mapstructure:"base_url"`
	Spot       MarketConfig `mapstructure:"spot"`
	Futures    MarketConfig `mapstructure:"futures"`
}

func (c BinanceConfig) IsSpotEnabled() bool    { return c.Spot.Enabled }
func (c BinanceConfig) IsMarginEnabled() bool  { return c.Margin.Enabled }
func (c BinanceConfig) IsFuturesEnabled() bool { return c.Futures.Enabled }

func (c BitgetConfig) IsSpotEnabled() bool    { return c.Spot.Enabled }
func (c BitgetConfig) IsFuturesEnabled() bool { return c.Futures.Enabled }

// Config holds all provider configurations.
type Config struct {
	// Active provider (coingecko, binance, bitget)
	Provider string `mapstructure:"provider"`

	// Provider configs
	CoinGecko CoinGeckoConfig `mapstructure:"coingecko"`
	Binance   BinanceConfig   `mapstructure:"binance"`
	Bitget    BitgetConfig    `mapstructure:"bitget"`
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
	v.SetDefault("coingecko.tier", TierDemo)
	v.SetDefault("provider", "coingecko")
	v.SetDefault("binance.spot.enabled", true)

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

	// Default CoinGecko tier if not set
	if cfg.CoinGecko.Tier == "" {
		cfg.CoinGecko.Tier = TierDemo
	}

	return &cfg, nil
}

// applyEnvOverrides applies BITS_* environment variable overrides to the config.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("BITS_PROVIDER"); v != "" {
		cfg.Provider = v
	}
	// CoinGecko
	if v := os.Getenv("BITS_COINGECKO_API_KEY"); v != "" {
		cfg.CoinGecko.APIKey = v
	}
	if v := os.Getenv("BITS_COINGECKO_TIER"); v != "" {
		cfg.CoinGecko.Tier = v
	}
	if v := os.Getenv("BITS_COINGECKO_BASE_URL"); v != "" {
		cfg.CoinGecko.BaseURL = v
	}
	// Binance (shared key for all markets)
	if v := os.Getenv("BITS_BINANCE_API_KEY"); v != "" {
		cfg.Binance.APIKey = v
	}
	if v := os.Getenv("BITS_BINANCE_API_SECRET"); v != "" {
		cfg.Binance.APISecret = v
	}
	if v := os.Getenv("BITS_BINANCE_BASE_URL"); v != "" {
		cfg.Binance.BaseURL = v
	}
	// Binance market-specific settings
	if v := os.Getenv("BITS_BINANCE_SPOT_ENABLED"); v != "" {
		cfg.Binance.Spot.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("BITS_BINANCE_MARGIN_ENABLED"); v != "" {
		cfg.Binance.Margin.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("BITS_BINANCE_FUTURES_ENABLED"); v != "" {
		cfg.Binance.Futures.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("BITS_BINANCE_FUTURES_USE_TESTNET"); v != "" {
		cfg.Binance.Futures.UseTestnet = v == "true" || v == "1"
	}
	// Bitget (shared key for all markets)
	if v := os.Getenv("BITS_BITGET_API_KEY"); v != "" {
		cfg.Bitget.APIKey = v
	}
	if v := os.Getenv("BITS_BITGET_API_SECRET"); v != "" {
		cfg.Bitget.APISecret = v
	}
	if v := os.Getenv("BITS_BITGET_PASSPHRASE"); v != "" {
		cfg.Bitget.Passphrase = v
	}
	if v := os.Getenv("BITS_BITGET_BASE_URL"); v != "" {
		cfg.Bitget.BaseURL = v
	}
	// Bitget market-specific settings
	if v := os.Getenv("BITS_BITGET_SPOT_ENABLED"); v != "" {
		cfg.Bitget.Spot.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("BITS_BITGET_FUTURES_ENABLED"); v != "" {
		cfg.Bitget.Futures.Enabled = v == "true" || v == "1"
	}
	// Legacy env var support for Bitget passphrase
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
	v.Set("coingecko.api_key", cfg.CoinGecko.APIKey)
	v.Set("coingecko.tier", cfg.CoinGecko.Tier)
	v.Set("coingecko.base_url", cfg.CoinGecko.BaseURL)
	v.Set("binance.api_key", cfg.Binance.APIKey)
	v.Set("binance.api_secret", cfg.Binance.APISecret)
	v.Set("binance.base_url", cfg.Binance.BaseURL)
	v.Set("bitget.api_key", cfg.Bitget.APIKey)
	v.Set("bitget.api_secret", cfg.Bitget.APISecret)
	v.Set("bitget.passphrase", cfg.Bitget.Passphrase)
	v.Set("bitget.base_url", cfg.Bitget.BaseURL)

	path := filepath.Join(dir, "config.yaml")
	if err := v.WriteConfigAs(path); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

func (c *Config) BaseURL() string {
	return c.CoinGecko.GetBaseURL()
}

func (c *Config) AuthHeader() (string, string) {
	return c.CoinGecko.GetAuthHeader()
}

func (c *Config) ApplyAuth(req *http.Request) {
	if c.CoinGecko.APIKey != "" {
		key, val := c.AuthHeader()
		req.Header.Set(key, val)
	}
}

func (c *Config) IsPaid() bool {
	return c.CoinGecko.IsPaid()
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
	apiKey := c.CoinGecko.APIKey
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}

func (c *Config) Redacted() *Config {
	maskedLong := func(s string) string {
		if s == "" {
			return ""
		}
		if len(s) <= 8 {
			return strings.Repeat("*", len(s))
		}
		return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
	}

	return &Config{
		Provider: c.Provider,
		CoinGecko: CoinGeckoConfig{
			APIKey:  maskedLong(c.CoinGecko.APIKey),
			Tier:    c.CoinGecko.Tier,
			BaseURL: c.CoinGecko.BaseURL,
		},
		Binance: BinanceConfig{
			APIKey:    maskedLong(c.Binance.APIKey),
			APISecret: maskedLong(c.Binance.APISecret),
			BaseURL:   c.Binance.BaseURL,
			Spot:      c.Binance.Spot,
			Margin:    c.Binance.Margin,
			Futures:   c.Binance.Futures,
		},
		Bitget: BitgetConfig{
			APIKey:     maskedLong(c.Bitget.APIKey),
			APISecret:  maskedLong(c.Bitget.APISecret),
			Passphrase: maskedLong(c.Bitget.Passphrase),
			BaseURL:    c.Bitget.BaseURL,
			Spot:       c.Bitget.Spot,
			Futures:    c.Bitget.Futures,
		},
	}
}
