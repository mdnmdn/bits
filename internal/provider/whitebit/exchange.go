package whitebit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdnmdn/bits/internal/model"
)

// whitebitTimeResponse is the API response for /api/v4/public/time.
type whitebitTimeResponse struct {
	Time int64 `json:"time"`
}

// whitebitMarket is a single market from the markets endpoint.
type whitebitMarket struct {
	Name          string `json:"name"`
	Stock         string `json:"stock"`
	Money         string `json:"money"`
	TradesEnabled bool   `json:"tradesEnabled"`
}

type whitebitFuturesMarket struct {
	TickerID      string `json:"ticker_id"`
	StockCurrency string `json:"stock_currency"`
	MoneyCurrency string `json:"money_currency"`
}

// ServerTime fetches the server time from the WhiteBit API.
func (c *Client) ServerTime(_ context.Context) (model.Response[model.ServerTime], error) {
	body, err := c.doRequest("/api/v4/public/time")
	if err != nil {
		return model.Response[model.ServerTime]{}, err
	}

	var resp whitebitTimeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.ServerTime]{}, fmt.Errorf("failed to parse server time response: %w", err)
	}

	serverTime := time.Unix(resp.Time, 0)

	return model.Response[model.ServerTime]{
		Kind:     model.KindServerTime,
		Provider: providerID,
		Data: model.ServerTime{
			Time: serverTime,
		},
	}, nil
}

// ExchangeInfo fetches exchange symbol information for the specified market.
func (c *Client) ExchangeInfo(_ context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	if market == model.MarketFutures {
		return c.futuresExchangeInfo(market)
	}

	body, err := c.doRequest("/api/v4/public/markets")
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}

	var resp []whitebitMarket
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to parse markets response: %w", err)
	}

	symbols := make([]model.Symbol, 0, len(resp))
	for _, m := range resp {
		status := model.SymbolStatusBreak
		if m.TradesEnabled {
			status = model.SymbolStatusTrading
		}

		symbols = append(symbols, model.Symbol{
			Symbol:     m.Name,
			BaseAsset:  m.Stock,
			QuoteAsset: m.Money,
			Status:     status,
			Market:     market,
		})
	}

	return model.Response[model.ExchangeInfo]{
		Kind:     model.KindExchangeInfo,
		Provider: providerID,
		Market:   market,
		Data: model.ExchangeInfo{
			ExchangeID: providerID,
			Market:     market,
			Symbols:    symbols,
		},
	}, nil
}

func (c *Client) futuresExchangeInfo(market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	body, err := c.doRequest("/api/v4/public/futures")
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}

	var resp struct {
		Success bool                    `json:"success"`
		Result  []whitebitFuturesMarket `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to parse futures response: %w", err)
	}

	symbols := make([]model.Symbol, 0, len(resp.Result))
	for _, m := range resp.Result {
		symbols = append(symbols, model.Symbol{
			Symbol:     m.TickerID,
			BaseAsset:  m.StockCurrency,
			QuoteAsset: m.MoneyCurrency,
			Status:     model.SymbolStatusTrading,
			Market:     market,
		})
	}

	return model.Response[model.ExchangeInfo]{
		Kind:     model.KindExchangeInfo,
		Provider: providerID,
		Market:   market,
		Data: model.ExchangeInfo{
			ExchangeID: providerID,
			Market:     market,
			Symbols:    symbols,
		},
	}, nil
}
