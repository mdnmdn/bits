package config

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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

func (c CoinGeckoConfig) ApplyAuth(req *http.Request) {
	if c.APIKey != "" {
		key, val := c.GetAuthHeader()
		req.Header.Set(key, val)
	}
}

func (c CoinGeckoConfig) MaskedKey() string {
	apiKey := c.APIKey
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
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
	Margin     MarketConfig `mapstructure:"margin"`
	Futures    MarketConfig `mapstructure:"futures"`
}

// WhiteBitConfig holds WhiteBit API credentials.
type WhiteBitConfig struct {
	APIKey    string       `mapstructure:"api_key"`
	APISecret string       `mapstructure:"api_secret"`
	BaseURL   string       `mapstructure:"base_url"`
	Spot      MarketConfig `mapstructure:"spot"`
}

// SymbolConfig holds symbol resolution settings.
type SymbolConfig struct {
	CacheTTL time.Duration `mapstructure:"cache_ttl"`
	CacheDir string        `mapstructure:"cache_dir"`
}

func (c SymbolConfig) GetCacheTTL() time.Duration {
	if c.CacheTTL <= 0 {
		return 5 * time.Minute
	}
	return c.CacheTTL
}

func (c SymbolConfig) GetCacheDir() string {
	if c.CacheDir != "" {
		return c.CacheDir
	}
	return filepath.Join(os.TempDir(), "bits")
}

func (c WhiteBitConfig) IsSpotEnabled() bool { return c.Spot.Enabled }

func (c BinanceConfig) IsSpotEnabled() bool    { return c.Spot.Enabled }
func (c BinanceConfig) IsMarginEnabled() bool  { return c.Margin.Enabled }
func (c BinanceConfig) IsFuturesEnabled() bool { return c.Futures.Enabled }

func (c BitgetConfig) IsSpotEnabled() bool    { return c.Spot.Enabled }
func (c BitgetConfig) IsMarginEnabled() bool  { return c.Margin.Enabled }
func (c BitgetConfig) IsFuturesEnabled() bool { return c.Futures.Enabled }

// Config holds all provider configurations.
// It embeds the public config.Config for extensibility.
type Config struct {
	Provider  string          `mapstructure:"provider"`
	CoinGecko CoinGeckoConfig `mapstructure:"coingecko"`
	Binance   BinanceConfig   `mapstructure:"binance"`
	Bitget    BitgetConfig    `mapstructure:"bitget"`
	WhiteBit  WhiteBitConfig  `mapstructure:"whitebit"`
	Symbol    SymbolConfig    `mapstructure:"symbol"`
}

func ConfigDirs() []string {
	var dirs []string

	// 1. Local directory (current working directory)
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}

	// 2. Platform-specific config directories
	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\bits (roaming) and %LOCALAPPDATA%\bits (local)
		if appData := os.Getenv("APPDATA"); appData != "" {
			dirs = append(dirs, filepath.Join(appData, "bits"))
		}
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			dirs = append(dirs, filepath.Join(localAppData, "bits"))
		}
	case "darwin":
		// macOS: ~/Library/Application Support/bits-cli
		if appSupport, err := os.UserConfigDir(); err == nil {
			dirs = append(dirs, filepath.Join(appSupport, "bits-cli"))
		}
		fallthrough
	default:
		// Linux/Unix: XDG_CONFIG_HOME/bits or ~/.config/bits
		if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
			dirs = append(dirs, filepath.Join(configHome, "bits"))
		} else if userConfig, err := os.UserConfigDir(); err == nil {
			dirs = append(dirs, filepath.Join(userConfig, "bits"))
		}
	}

	// Also add fallback for macOS in case fallthrough didn't work
	if runtime.GOOS == "darwin" {
		if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
			dirs = append(dirs, filepath.Join(configHome, "bits"))
		} else if userConfig, err := os.UserConfigDir(); err == nil {
			dirs = append(dirs, filepath.Join(userConfig, "bits"))
		}
	}

	return dirs
}

func configDir() string {
	dirs := ConfigDirs()
	if len(dirs) > 1 {
		return dirs[1]
	}
	if len(dirs) > 0 {
		return dirs[0]
	}
	return ""
}

func Load() (*Config, string, error) {
	dirs := ConfigDirs()

	v := viper.New()
	v.SetConfigName("config")
	for _, dir := range dirs {
		v.AddConfigPath(dir)
	}
	v.SetDefault("coingecko.tier", TierDemo)
	v.SetDefault("provider", "coingecko")
	v.SetDefault("binance.spot.enabled", true)
	v.SetDefault("symbol.cache_ttl", "5m")
	v.SetDefault("symbol.cache_dir", "/tmp/bits")

	if err := v.ReadInConfig(); err != nil {
		// Ignore "config file not found" errors - use defaults
		if !os.IsNotExist(err) && !strings.Contains(err.Error(), "Not Found") {
			return nil, "", fmt.Errorf("failed to read config: %w", err)
		}
	}

	configFile := v.ConfigFileUsed()

	// Check for .env files in config directories
	var envVars map[string]string
	for _, dir := range dirs {
		envPath := filepath.Join(dir, ".env")
		if data, err := os.ReadFile(envPath); err == nil {
			envVars = parseEnvFileToMap(string(data))
			if configFile == "" {
				configFile = envPath
			}
			// Found .env, no need to check other directories
			break
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, "", fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply .env file overrides
	applyEnvMap(envVars, &cfg)

	// Apply env var overrides (BITS_* prefix, env vars take priority)
	applyEnvOverrides(&cfg)

	// Default CoinGecko tier if not set
	if cfg.CoinGecko.Tier == "" {
		cfg.CoinGecko.Tier = TierDemo
	}

	return &cfg, configFile, nil
}

var ConfigTemplate = `# bits Configuration
# Supported formats: .yaml, .yml, .toml, .json

# Active provider (coingecko, binance, bitget)
provider = "coingecko"

# CoinGecko configuration
[coingecko]
api_key = ""           # Your CoinGecko API key
tier = "demo"          # demo or paid
# base_url = ""        # optional custom endpoint

# Binance configuration (shared API key for all markets)
[binance]
api_key = ""
api_secret = ""
# base_url = "https://api.binance.com"

[binance.spot]
enabled = true

[binance.margin]
enabled = false

[binance.futures]
enabled = false
use_testnet = false

# Bitget configuration (shared API key for all markets)
[bitget]
api_key = ""
api_secret = ""
# passphrase = ""      # required for authenticated endpoints
# base_url = "https://api.bitget.com"

[bitget.spot]
enabled = false

[bitget.futures]
enabled = false

# WhiteBit configuration
[whitebit]
api_key = ""
api_secret = ""
# base_url = "https://whitebit.com"

[whitebit.spot]
enabled = false

# Symbol resolution cache settings
[symbol]
# cache_ttl = "5m"       # TTL for symbol cache (e.g., 5m, 10m, 1h)
# cache_dir = "/tmp/bits"  # Cache directory for symbol data
`

func defaultSaveDir() string {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "bits")
		}
	case "darwin":
		if appSupport, err := os.UserConfigDir(); err == nil {
			return filepath.Join(appSupport, "bits")
		}
	default:
		if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
			return filepath.Join(configHome, "bits")
		} else if userConfig, err := os.UserConfigDir(); err == nil {
			return filepath.Join(userConfig, "bits")
		}
	}
	return ""
}

func Init(local bool) (string, error) {
	dirs := ConfigDirs()
	var dir string

	if local {
		dir = dirs[0] // Local directory
	} else {
		// Use platform-specific default save directory
		dir = defaultSaveDir()
		if dir == "" {
			// Fallback to second directory in ConfigDirs
			if len(dirs) > 1 {
				dir = dirs[1]
			} else if len(dirs) > 0 {
				dir = dirs[0]
			}
		}
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check for existing config file with any supported extension
	for _, ext := range []string{".yaml", ".yml", ".toml", ".json"} {
		configFile := filepath.Join(dir, "config"+ext)
		if _, err := os.Stat(configFile); err == nil {
			return configFile, nil // File already exists
		}
	}

	configFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(ConfigTemplate), 0o644); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configFile, nil
}

func parseEnvFileToMap(data string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		key = strings.ToLower(key)
		key = strings.TrimPrefix(key, "bits_")
		// Replace __ with . for nested keys
		key = strings.ReplaceAll(key, "__", ".")
		// For keys without dots, split on first underscore only:
		// COINGECKO_API_KEY -> coingecko.api_key
		// BITGET_API_KEY -> bitget.api_key
		if !strings.Contains(key, ".") {
			if idx := strings.Index(key, "_"); idx > 0 {
				key = key[:idx] + "." + key[idx+1:]
			}
		}
		result[key] = value
	}
	return result
}

func applyEnvMap(envVars map[string]string, cfg *Config) {
	if envVars == nil {
		return
	}
	if v, ok := envVars["provider"]; ok {
		cfg.Provider = v
	}
	if v, ok := envVars["coingecko.api_key"]; ok {
		cfg.CoinGecko.APIKey = v
	}
	if v, ok := envVars["coingecko.tier"]; ok {
		cfg.CoinGecko.Tier = v
	}
	if v, ok := envVars["coingecko.base_url"]; ok {
		cfg.CoinGecko.BaseURL = v
	}
	if v, ok := envVars["binance.api_key"]; ok {
		cfg.Binance.APIKey = v
	}
	if v, ok := envVars["binance.api_secret"]; ok {
		cfg.Binance.APISecret = v
	}
	if v, ok := envVars["binance.base_url"]; ok {
		cfg.Binance.BaseURL = v
	}
	if v, ok := envVars["binance.spot.enabled"]; ok {
		cfg.Binance.Spot.Enabled = v == "true" || v == "1"
	}
	if v, ok := envVars["binance.margin.enabled"]; ok {
		cfg.Binance.Margin.Enabled = v == "true" || v == "1"
	}
	if v, ok := envVars["binance.futures.enabled"]; ok {
		cfg.Binance.Futures.Enabled = v == "true" || v == "1"
	}
	if v, ok := envVars["binance.futures.use_testnet"]; ok {
		cfg.Binance.Futures.UseTestnet = v == "true" || v == "1"
	}
	if v, ok := envVars["bitget.api_key"]; ok {
		cfg.Bitget.APIKey = v
	}
	if v, ok := envVars["bitget.api_secret"]; ok {
		cfg.Bitget.APISecret = v
	}
	if v, ok := envVars["bitget.passphrase"]; ok {
		cfg.Bitget.Passphrase = v
	}
	if v, ok := envVars["bitget.base_url"]; ok {
		cfg.Bitget.BaseURL = v
	}
	if v, ok := envVars["bitget.spot.enabled"]; ok {
		cfg.Bitget.Spot.Enabled = v == "true" || v == "1"
	}
	if v, ok := envVars["bitget.margin.enabled"]; ok {
		cfg.Bitget.Margin.Enabled = v == "true" || v == "1"
	}
	if v, ok := envVars["bitget.futures.enabled"]; ok {
		cfg.Bitget.Futures.Enabled = v == "true" || v == "1"
	}
	if v, ok := envVars["whitebit.api_key"]; ok {
		cfg.WhiteBit.APIKey = v
	}
	if v, ok := envVars["whitebit.api_secret"]; ok {
		cfg.WhiteBit.APISecret = v
	}
	if v, ok := envVars["whitebit.base_url"]; ok {
		cfg.WhiteBit.BaseURL = v
	}
	if v, ok := envVars["whitebit.spot.enabled"]; ok {
		cfg.WhiteBit.Spot.Enabled = v == "true" || v == "1"
	}
	// Symbol config
	if v, ok := envVars["symbol.cache_ttl"]; ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Symbol.CacheTTL = d
		}
	}
	if v, ok := envVars["symbol.cache_dir"]; ok {
		cfg.Symbol.CacheDir = v
	}
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
	if v := os.Getenv("BITS_BITGET_MARGIN_ENABLED"); v != "" {
		cfg.Bitget.Margin.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("BITS_BITGET_FUTURES_ENABLED"); v != "" {
		cfg.Bitget.Futures.Enabled = v == "true" || v == "1"
	}
	// Legacy env var support for Bitget passphrase
	if v := os.Getenv("BITS_BITGET_PASSPHRASE"); v != "" {
		cfg.Bitget.Passphrase = v
	}
	// WhiteBit
	if v := os.Getenv("BITS_WHITEBIT_API_KEY"); v != "" {
		cfg.WhiteBit.APIKey = v
	}
	if v := os.Getenv("BITS_WHITEBIT_API_SECRET"); v != "" {
		cfg.WhiteBit.APISecret = v
	}
	if v := os.Getenv("BITS_WHITEBIT_BASE_URL"); v != "" {
		cfg.WhiteBit.BaseURL = v
	}
	if v := os.Getenv("BITS_WHITEBIT_SPOT_ENABLED"); v != "" {
		cfg.WhiteBit.Spot.Enabled = v == "true" || v == "1"
	}
	// Symbol resolution cache
	if v := os.Getenv("BITS_SYMBOL_CACHE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Symbol.CacheTTL = d
		}
	}
	if v := os.Getenv("BITS_SYMBOL_CACHE_DIR"); v != "" {
		cfg.Symbol.CacheDir = v
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
	dir := defaultSaveDir()
	if dir == "" {
		// Fallback to ConfigDirs
		dirs := ConfigDirs()
		if len(dirs) > 0 {
			dir = dirs[0]
		}
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

func IsValidTier(tier string) bool {
	t := strings.ToLower(tier)
	for _, v := range ValidTiers {
		if t == v {
			return true
		}
	}
	return false
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
			Margin:     c.Bitget.Margin,
			Futures:    c.Bitget.Futures,
		},
		WhiteBit: WhiteBitConfig{
			APIKey:    maskedLong(c.WhiteBit.APIKey),
			APISecret: maskedLong(c.WhiteBit.APISecret),
			BaseURL:   c.WhiteBit.BaseURL,
			Spot:      c.WhiteBit.Spot,
		},
		Symbol: SymbolConfig{
			CacheTTL: c.Symbol.CacheTTL,
			CacheDir: c.Symbol.CacheDir,
		},
	}
}
