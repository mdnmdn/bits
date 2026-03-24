package bitget

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mdnmdn/bits/internal/config"
)

const (
	// DefaultBaseURL is the default Bitget API base URL.
	DefaultBaseURL = "https://api.bitget.com"

	// tradingPairsCacheTTL is the duration to cache trading pairs.
	tradingPairsCacheTTL = 5 * time.Minute
)

// Client represents a Bitget API client that implements the Provider interface.
type Client struct {
	config     config.BitgetConfig
	httpClient *http.Client
	userAgent  string

	// Trading pairs cache
	tradingPairsCache     map[string]struct{}
	tradingPairsCacheTime time.Time
	tradingPairsCacheMu   sync.Mutex
}

// NewClient creates a new Bitget API client from the given config.
func NewClient(cfg config.BitgetConfig) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	return &Client{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ID returns the provider identifier.
func (c *Client) ID() string {
	return "bitget"
}

// SetUserAgent sets the User-Agent header for API requests.
func (c *Client) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// formatWithPrecision formats a value using the appropriate precision/scale from pair info.
// It truncates (not rounds) to ensure we never exceed the required decimal places.
func formatWithPrecision(value float64, precision, scale string, defaultFormat string) string {
	// Try to use scale first (more restrictive for API validation)
	if scale != "" {
		if scaleInt, err := strconv.Atoi(scale); err == nil && scaleInt >= 0 {
			return formatToScale(value, scaleInt)
		}
	}

	// Fallback to precision
	if precision != "" {
		if precInt, err := strconv.Atoi(precision); err == nil && precInt >= 0 {
			return formatToScale(value, precInt)
		}
	}

	// Fallback to default format
	return fmt.Sprintf(defaultFormat, value)
}

// formatToScale truncates a float to the specified number of decimal places.
// This ensures we never exceed the scale required by the API.
func formatToScale(value float64, scale int) string {
	if scale < 0 {
		scale = 0
	}
	if scale > 18 {
		scale = 18
	}

	// Convert to string with high precision first to avoid float rounding issues
	highPrecisionStr := fmt.Sprintf("%.18f", value)

	decimalIndex := strings.Index(highPrecisionStr, ".")
	if decimalIndex == -1 {
		if scale == 0 {
			return fmt.Sprintf("%.0f", value)
		}
		return fmt.Sprintf("%s.%s", highPrecisionStr, strings.Repeat("0", scale))
	}

	integerPart := highPrecisionStr[:decimalIndex]
	if scale == 0 {
		return integerPart
	}

	decimalPart := highPrecisionStr[decimalIndex+1:]
	if len(decimalPart) > scale {
		decimalPart = decimalPart[:scale]
	} else if len(decimalPart) < scale {
		decimalPart = decimalPart + strings.Repeat("0", scale-len(decimalPart))
	}

	result := fmt.Sprintf("%s.%s", integerPart, decimalPart)

	if scale > 0 {
		result = strings.TrimRight(result, "0")
		result = strings.TrimRight(result, ".")
	}

	return result
}

// convertGranularityFormat converts granularity from CLI format to Bitget API format.
func convertGranularityFormat(granularity string) string {
	switch granularity {
	case "1m":
		return "1min"
	case "5m":
		return "5min"
	case "15m":
		return "15min"
	case "30m":
		return "30min"
	case "1h":
		return "1h"
	case "4h":
		return "4h"
	case "6h":
		return "6h"
	case "12h":
		return "12h"
	case "1d", "daily":
		return "1day"
	case "3d":
		return "3day"
	case "1w":
		return "1week"
	case "1M":
		return "1M"
	case "hourly":
		return "1h"
	default:
		return granularity
	}
}
