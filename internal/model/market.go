package model

import (
	"strings"

	"github.com/mdnmdn/bits/internal/capability"
)

// MarketType is a type alias for capability.MarketType so callers only import model.
type MarketType = capability.MarketType

const (
	MarketSpot    MarketType = capability.MarketSpot
	MarketFutures MarketType = capability.MarketFutures
	MarketMargin  MarketType = capability.MarketMargin
)

// ParseMarketType parses a market type string into a MarketType, handling aliases.
func ParseMarketType(s string) MarketType {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "spot":
		return MarketSpot
	case "futures", "future", "f", "perp":
		return MarketFutures
	case "margin", "m":
		return MarketMargin
	default:
		return MarketType(s)
	}
}
