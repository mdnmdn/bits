package coingecko

import "github.com/mdnmdn/bits/internal/model"

func toModelMarketCoin(c MarketCoin) model.MarketCoin {
	return model.MarketCoin{
		ID:                       c.ID,
		Symbol:                   c.Symbol,
		Name:                     c.Name,
		CurrentPrice:             c.CurrentPrice,
		MarketCap:                c.MarketCap,
		MarketCapRank:            c.MarketCapRank,
		TotalVolume:              c.TotalVolume,
		PriceChangePercentage24h: c.PriceChangePercentage24h,
		High24h:                  c.High24h,
		Low24h:                   c.Low24h,
		ATH:                      c.ATH,
		ATHChangePercentage:      c.ATHChangePercentage,
		ATL:                      c.ATL,
		ATLChangePercentage:      c.ATLChangePercentage,
		CirculatingSupply:        c.CirculatingSupply,
		TotalSupply:              c.TotalSupply,
	}
}

func toModelMarketCoins(coins []MarketCoin) []model.MarketCoin {
	out := make([]model.MarketCoin, len(coins))
	for i, c := range coins {
		out[i] = toModelMarketCoin(c)
	}
	return out
}

func toModelSearchResponse(s *SearchResponse) *model.SearchResponse {
	coins := make([]model.SearchCoin, len(s.Coins))
	for i, c := range s.Coins {
		coins[i] = model.SearchCoin{
			ID:            c.ID,
			Name:          c.Name,
			Symbol:        c.Symbol,
			MarketCapRank: c.MarketCapRank,
		}
	}
	return &model.SearchResponse{Coins: coins}
}

func toModelTrendingResponse(t *TrendingResponse) *model.TrendingResponse {
	coins := make([]model.TrendingCoinWrapper, len(t.Coins))
	for i, c := range t.Coins {
		var data *model.TrendingCoinData
		if c.Item.Data != nil {
			data = &model.TrendingCoinData{
				Price:                    c.Item.Data.Price,
				PriceChangePercentage24h: c.Item.Data.PriceChangePercentage24h,
			}
		}
		coins[i] = model.TrendingCoinWrapper{
			Item: model.TrendingCoin{
				ID:            c.Item.ID,
				Name:          c.Item.Name,
				Symbol:        c.Item.Symbol,
				MarketCapRank: c.Item.MarketCapRank,
				Score:         c.Item.Score,
				Data:          data,
			},
		}
	}

	nfts := make([]model.TrendingNFT, len(t.NFTs))
	for i, n := range t.NFTs {
		nfts[i] = model.TrendingNFT{
			ID:                   n.ID,
			Name:                 n.Name,
			Symbol:               n.Symbol,
			FloorPriceInUSD24hPC: n.FloorPriceInUSD24hPC,
		}
	}

	cats := make([]model.TrendingCategory, len(t.Categories))
	for i, c := range t.Categories {
		cats[i] = model.TrendingCategory{
			ID:                c.ID,
			Name:              c.Name,
			MarketCap1hChange: c.MarketCap1hChange,
		}
	}

	return &model.TrendingResponse{Coins: coins, NFTs: nfts, Categories: cats}
}

func toModelHistoricalData(h *HistoricalData) *model.HistoricalData {
	var md *model.HistoricalMarket
	if h.MarketData != nil {
		md = &model.HistoricalMarket{
			CurrentPrice: h.MarketData.CurrentPrice,
			MarketCap:    h.MarketData.MarketCap,
			TotalVolume:  h.MarketData.TotalVolume,
		}
	}
	return &model.HistoricalData{
		ID:         h.ID,
		Symbol:     h.Symbol,
		Name:       h.Name,
		MarketData: md,
	}
}

func toModelMarketChartResponse(m *MarketChartResponse) *model.MarketChartResponse {
	return &model.MarketChartResponse{
		Prices:       m.Prices,
		MarketCaps:   m.MarketCaps,
		TotalVolumes: m.TotalVolumes,
	}
}

func toModelGainersLosersResponse(g *GainersLosersResponse) *model.GainersLosersResponse {
	gainers := make([]model.GainerCoin, len(g.TopGainers))
	for i, c := range g.TopGainers {
		gainers[i] = model.GainerCoin{
			ID:            c.ID,
			Symbol:        c.Symbol,
			Name:          c.Name,
			Image:         c.Image,
			MarketCapRank: c.MarketCapRank,
			Extra:         c.Extra,
		}
	}
	losers := make([]model.GainerCoin, len(g.TopLosers))
	for i, c := range g.TopLosers {
		losers[i] = model.GainerCoin{
			ID:            c.ID,
			Symbol:        c.Symbol,
			Name:          c.Name,
			Image:         c.Image,
			MarketCapRank: c.MarketCapRank,
			Extra:         c.Extra,
		}
	}
	return &model.GainersLosersResponse{TopGainers: gainers, TopLosers: losers}
}

func toModelCoinDetail(d *CoinDetail) *model.CoinDetail {
	result := &model.CoinDetail{
		ID:            d.ID,
		Symbol:        d.Symbol,
		Name:          d.Name,
		MarketCapRank: d.MarketCapRank,
	}
	result.Description.EN = d.Description.EN

	if d.MarketData != nil {
		result.MarketData = &model.CoinDetailMarket{
			CurrentPrice:             d.MarketData.CurrentPrice,
			MarketCap:                d.MarketData.MarketCap,
			TotalVolume:              d.MarketData.TotalVolume,
			High24h:                  d.MarketData.High24h,
			Low24h:                   d.MarketData.Low24h,
			PriceChangePercentage24h: d.MarketData.PriceChangePercentage24h,
			ATH:                      d.MarketData.ATH,
			ATHChangePercentage:      d.MarketData.ATHChangePercentage,
			ATHDate:                  d.MarketData.ATHDate,
			ATL:                      d.MarketData.ATL,
			ATLChangePercentage:      d.MarketData.ATLChangePercentage,
			ATLDate:                  d.MarketData.ATLDate,
			CirculatingSupply:        d.MarketData.CirculatingSupply,
			TotalSupply:              d.MarketData.TotalSupply,
		}
	}
	return result
}
