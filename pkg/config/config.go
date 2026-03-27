package config

import (
	"net/http"
	"strings"
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

type MarketConfig struct {
	Enabled    bool `yaml:"enabled" toml:"enabled" mapstructure:"enabled"`
	UseTestnet bool `yaml:"use_testnet" toml:"use_testnet" mapstructure:"use_testnet"`
}

type CoinGeckoConfig struct {
	APIKey  string `yaml:"api_key" toml:"api_key" mapstructure:"api_key"`
	Tier    string `yaml:"tier" toml:"tier" mapstructure:"tier"`
	BaseURL string `yaml:"base_url" toml:"base_url" mapstructure:"base_url"`
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

type BinanceConfig struct {
	APIKey    string       `yaml:"api_key" toml:"api_key" mapstructure:"api_key"`
	APISecret string       `yaml:"api_secret" toml:"api_secret" mapstructure:"api_secret"`
	BaseURL   string       `yaml:"base_url" toml:"base_url" mapstructure:"base_url"`
	Spot      MarketConfig `yaml:"spot" toml:"spot" mapstructure:"spot"`
	Margin    MarketConfig `yaml:"margin" toml:"margin" mapstructure:"margin"`
	Futures   MarketConfig `yaml:"futures" toml:"futures" mapstructure:"futures"`
}

func (c BinanceConfig) IsSpotEnabled() bool    { return c.Spot.Enabled }
func (c BinanceConfig) IsMarginEnabled() bool  { return c.Margin.Enabled }
func (c BinanceConfig) IsFuturesEnabled() bool { return c.Futures.Enabled }

type BitgetConfig struct {
	APIKey     string       `yaml:"api_key" toml:"api_key" mapstructure:"api_key"`
	APISecret  string       `yaml:"api_secret" toml:"api_secret" mapstructure:"api_secret"`
	Passphrase string       `yaml:"passphrase" toml:"passphrase" mapstructure:"passphrase"`
	BaseURL    string       `yaml:"base_url" toml:"base_url" mapstructure:"base_url"`
	Spot       MarketConfig `yaml:"spot" toml:"spot" mapstructure:"spot"`
	Futures    MarketConfig `yaml:"futures" toml:"futures" mapstructure:"futures"`
}

func (c BitgetConfig) IsSpotEnabled() bool    { return c.Spot.Enabled }
func (c BitgetConfig) IsFuturesEnabled() bool { return c.Futures.Enabled }

type Config struct {
	Provider  string          `yaml:"provider" toml:"provider" mapstructure:"provider"`
	CoinGecko CoinGeckoConfig `yaml:"coingecko" toml:"coingecko" mapstructure:"coingecko"`
	Binance   BinanceConfig   `yaml:"binance" toml:"binance" mapstructure:"binance"`
	Bitget    BitgetConfig    `yaml:"bitget" toml:"bitget" mapstructure:"bitget"`
}

func (c *Config) ActiveProvider() string {
	if c.Provider == "" {
		return "coingecko"
	}
	return c.Provider
}

func (c *Config) ApplyAuth(req *http.Request) {
	if c.CoinGecko.APIKey != "" {
		key, val := c.CoinGecko.GetAuthHeader()
		req.Header.Set(key, val)
	}
}

func (c *Config) IsPaid() bool {
	return c.CoinGecko.IsPaid()
}

func (c *Config) BaseURL() string {
	return c.CoinGecko.GetBaseURL()
}

func (c *Config) AuthHeader() (string, string) {
	return c.CoinGecko.GetAuthHeader()
}
