package model

import "github.com/mdnmdn/bits/internal/capability"

// MarketType is a type alias for capability.MarketType so callers only import model.
type MarketType = capability.MarketType

const (
	MarketSpot    MarketType = capability.MarketSpot
	MarketFutures MarketType = capability.MarketFutures
	MarketMargin  MarketType = capability.MarketMargin
)
