package model

import (
	"encoding/json"
	"time"
)

// PriceResponse maps identifier -> currency -> value (price, 24h change, etc.).
type PriceResponse map[string]map[string]float64

// MarketCoin represents a coin in market listings.
type MarketCoin struct {
	ID                       string  `json:"id"`
	Symbol                   string  `json:"symbol"`
	Name                     string  `json:"name"`
	CurrentPrice             float64 `json:"current_price"`
	MarketCap                float64 `json:"market_cap"`
	MarketCapRank            int     `json:"market_cap_rank"`
	TotalVolume              float64 `json:"total_volume"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
	High24h                  float64 `json:"high_24h"`
	Low24h                   float64 `json:"low_24h"`
	ATH                      float64 `json:"ath"`
	ATHChangePercentage      float64 `json:"ath_change_percentage"`
	ATL                      float64 `json:"atl"`
	ATLChangePercentage      float64 `json:"atl_change_percentage"`
	CirculatingSupply        float64 `json:"circulating_supply"`
	TotalSupply              float64 `json:"total_supply"`
}

// SearchResponse wraps search results.
type SearchResponse struct {
	Coins []SearchCoin `json:"coins"`
}

// SearchCoin represents a single search result.
type SearchCoin struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	MarketCapRank int    `json:"market_cap_rank"`
}

// TrendingResponse contains trending coins, NFTs, and categories.
type TrendingResponse struct {
	Coins      []TrendingCoinWrapper `json:"coins"`
	NFTs       []TrendingNFT         `json:"nfts"`
	Categories []TrendingCategory    `json:"categories"`
}

// TrendingCoinWrapper wraps a TrendingCoin.
type TrendingCoinWrapper struct {
	Item TrendingCoin `json:"item"`
}

// TrendingCoin represents a single trending coin.
type TrendingCoin struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Symbol        string            `json:"symbol"`
	MarketCapRank int               `json:"market_cap_rank"`
	Score         int               `json:"score"`
	Data          *TrendingCoinData `json:"data"`
}

// TrendingCoinData contains pricing data for a trending coin.
type TrendingCoinData struct {
	Price                    float64            `json:"price"`
	PriceChangePercentage24h map[string]float64 `json:"price_change_percentage_24h"`
}

// TrendingNFT represents a trending NFT.
type TrendingNFT struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	Symbol               string  `json:"symbol"`
	FloorPriceInUSD24hPC float64 `json:"floor_price_24h_percentage_change"`
}

// TrendingCategory represents a trending category.
type TrendingCategory struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	MarketCap1hChange float64 `json:"market_cap_1h_change"`
}

// HistoricalData represents historical market data for a specific date.
type HistoricalData struct {
	ID         string            `json:"id"`
	Symbol     string            `json:"symbol"`
	Name       string            `json:"name"`
	MarketData *HistoricalMarket `json:"market_data"`
}

// HistoricalMarket contains market metrics at a point in time.
type HistoricalMarket struct {
	CurrentPrice map[string]float64 `json:"current_price"`
	MarketCap    map[string]float64 `json:"market_cap"`
	TotalVolume  map[string]float64 `json:"total_volume"`
}

// OHLCData represents OHLC candle data: each entry is [timestamp, open, high, low, close].
type OHLCData [][]float64

// MarketChartResponse contains time-series market data.
type MarketChartResponse struct {
	Prices       [][]float64 `json:"prices"`
	MarketCaps   [][]float64 `json:"market_caps"`
	TotalVolumes [][]float64 `json:"total_volumes"`
}

// GainersLosersResponse contains top gainers and losers.
type GainersLosersResponse struct {
	TopGainers []GainerCoin `json:"top_gainers"`
	TopLosers  []GainerCoin `json:"top_losers"`
}

// GainerCoin uses dynamic JSON keys for price fields based on the vs_currency parameter.
type GainerCoin struct {
	ID            string                 `json:"id"`
	Symbol        string                 `json:"symbol"`
	Name          string                 `json:"name"`
	Image         string                 `json:"image"`
	MarketCapRank int                    `json:"market_cap_rank"`
	Extra         map[string]interface{} `json:"-"`
}

// Price returns the price in the given vs currency.
func (g *GainerCoin) Price(vs string) float64 {
	v, _ := g.Extra[vs].(float64)
	return v
}

// PriceChange returns the 24h price change percentage in the given vs currency.
func (g *GainerCoin) PriceChange(vs string) float64 {
	v, _ := g.Extra[vs+"_24h_change"].(float64)
	return v
}

func (g *GainerCoin) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["id"].(string); ok {
		g.ID = v
	}
	if v, ok := raw["symbol"].(string); ok {
		g.Symbol = v
	}
	if v, ok := raw["name"].(string); ok {
		g.Name = v
	}
	if v, ok := raw["image"].(string); ok {
		g.Image = v
	}
	if v, ok := raw["market_cap_rank"].(float64); ok {
		g.MarketCapRank = int(v)
	}
	g.Extra = raw
	return nil
}

func (g GainerCoin) MarshalJSON() ([]byte, error) {
	if g.Extra == nil {
		type Alias GainerCoin
		return json.Marshal(Alias(g))
	}
	return json.Marshal(g.Extra)
}

// CoinDetail contains detailed information about a coin.
type CoinDetail struct {
	ID            string `json:"id"`
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	MarketCapRank int    `json:"market_cap_rank"`
	Description   struct {
		EN string `json:"en"`
	} `json:"description"`
	MarketData *CoinDetailMarket `json:"market_data"`
}

// CoinDetailMarket contains detailed market data for a coin.
type CoinDetailMarket struct {
	CurrentPrice             map[string]float64 `json:"current_price"`
	MarketCap                map[string]float64 `json:"market_cap"`
	TotalVolume              map[string]float64 `json:"total_volume"`
	High24h                  map[string]float64 `json:"high_24h"`
	Low24h                   map[string]float64 `json:"low_24h"`
	PriceChangePercentage24h float64            `json:"price_change_percentage_24h"`
	ATH                      map[string]float64 `json:"ath"`
	ATHChangePercentage      map[string]float64 `json:"ath_change_percentage"`
	ATHDate                  map[string]string  `json:"ath_date"`
	ATL                      map[string]float64 `json:"atl"`
	ATLChangePercentage      map[string]float64 `json:"atl_change_percentage"`
	ATLDate                  map[string]string  `json:"atl_date"`
	CirculatingSupply        float64            `json:"circulating_supply"`
	TotalSupply              float64            `json:"total_supply"`
}

// Ticker24h represents 24-hour ticker statistics for a trading pair.
type Ticker24h struct {
	Symbol             string    `json:"symbol"`
	LastPrice          float64   `json:"last_price"`
	PriceChange        float64   `json:"price_change"`
	PriceChangePercent float64   `json:"price_change_percent"`
	HighPrice          float64   `json:"high_price"`
	LowPrice           float64   `json:"low_price"`
	Volume             float64   `json:"volume"`
	QuoteVolume        float64   `json:"quote_volume"`
	OpenPrice          float64   `json:"open_price"`
	WeightedAvgPrice   float64   `json:"weighted_avg_price"`
	OpenTime           time.Time `json:"open_time"`
	CloseTime          time.Time `json:"close_time"`
}

// OrderBook represents the order book for a trading pair.
type OrderBook struct {
	Symbol string           `json:"symbol"`
	Bids   []OrderBookEntry `json:"bids"`
	Asks   []OrderBookEntry `json:"asks"`
}

// OrderBookEntry represents a single bid or ask in the order book.
type OrderBookEntry struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}
