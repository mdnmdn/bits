package mexc

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/model"
)

// ServerTime implements provider.ExchangeProvider.
// Note: MEXC uses the same endpoint for all market types.
func (c *Client) ServerTime(ctx context.Context) (model.Response[model.ServerTime], error) {
	start := time.Now()
	data, err := c.doRequest(ctx, model.MarketSpot, "/time", "")
	latency := time.Since(start)

	resp := model.Response[model.ServerTime]{
		Provider: providerID,
		Market:   model.MarketSpot,
		Kind:     model.KindServerTime,
	}
	if err != nil {
		return resp, err
	}

	var t struct {
		ServerTime int64 `json:"serverTime"`
	}
	if err := json.Unmarshal(data, &t); err != nil {
		return resp, err
	}

	resp.Data = model.ServerTime{
		Time:    time.UnixMilli(t.ServerTime),
		Latency: &latency,
	}
	return resp, nil
}

// ExchangeInfo implements provider.ExchangeProvider.
// Note: MEXC Margin market data is served via Spot REST endpoints.
func (c *Client) ExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	resp := model.Response[model.ExchangeInfo]{
		Provider: providerID,
		Market:   market,
		Kind:     model.KindExchangeInfo,
	}

	if market == model.MarketFutures {
		data, err := c.doRequest(ctx, market, "/detail", "")
		if err != nil {
			return resp, err
		}

		var infoResp mexcFuturesExchangeInfoResponse
		if err := json.Unmarshal(data, &infoResp); err != nil {
			return resp, err
		}

		makerFee, takerFee := c.getFeeConfig(market)
		var symbols []model.Symbol
		for _, s := range infoResp.Data {
			pp := s.PricePrecision
			symbols = append(symbols, model.Symbol{
				Symbol:         s.Symbol,
				Status:         model.SymbolStatusTrading,
				BaseAsset:      s.BaseCoin,
				QuoteAsset:     s.QuoteCoin,
				PricePrecision: &pp,
				MinQty:         &s.VolUnit,
				Market:         market,
				MakerFee:       &makerFee,
				TakerFee:       &takerFee,
			})
		}
		resp.Data = model.ExchangeInfo{
			ExchangeID: providerID,
			Market:     market,
			Symbols:    symbols,
		}
		return resp, nil
	}

	// Spot / Margin
	data, err := c.doRequest(ctx, market, "/exchangeInfo", "")
	if err != nil {
		return resp, err
	}

	var info mexcSpotExchangeInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return resp, err
	}

	makerFee, takerFee := c.getFeeConfig(market)
	var symbols []model.Symbol
	for _, s := range info.Symbols {
		status := model.SymbolStatusTrading
		if s.Status != "ENABLED" {
			status = model.SymbolStatusHalt
		}

		pp := s.QuotePrecision
		qp := 0
		if q, err := strconv.Atoi(s.BaseSizePrecision); err == nil {
			qp = q
		}

		symbols = append(symbols, model.Symbol{
			Symbol:         s.Symbol,
			Status:         status,
			BaseAsset:      s.BaseAsset,
			QuoteAsset:     s.QuoteAsset,
			PricePrecision: &pp,
			QtyPrecision:   &qp,
			Market:         market,
			MakerFee:       &makerFee,
			TakerFee:       &takerFee,
		})
	}

	st := time.UnixMilli(info.ServerTime)
	resp.Data = model.ExchangeInfo{
		ExchangeID: providerID,
		Market:     market,
		Symbols:    symbols,
		ServerTime: &st,
	}

	return resp, nil
}

func (c *Client) getFeeConfig(market model.MarketType) (makerFee, takerFee float64) {
	switch market {
	case model.MarketFutures:
		if c.cfg.Futures.MakerFee > 0 {
			return c.cfg.Futures.MakerFee, c.cfg.Futures.TakerFee
		}
	default: // spot/margin
		if c.cfg.Spot.MakerFee > 0 {
			return c.cfg.Spot.MakerFee, c.cfg.Spot.TakerFee
		}
	}
	return 0.002, 0.002 // MEXC default
}
