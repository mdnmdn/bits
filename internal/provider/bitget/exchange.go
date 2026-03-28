package bitget

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/internal/model"
)

// bitgetServerTimeResponse is the API response for /api/v2/public/time.
type bitgetServerTimeResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ServerTime string `json:"serverTime"`
	} `json:"data"`
}

// bitgetSpotSymbol is a single symbol from the spot symbols endpoint.
type bitgetSpotSymbol struct {
	Symbol    string `json:"symbol"`
	BaseCoin  string `json:"baseCoin"`
	QuoteCoin string `json:"quoteCoin"`
	Status    string `json:"status"`
}

// bitgetSpotSymbolsResponse is the API response for /api/v2/spot/public/symbols.
type bitgetSpotSymbolsResponse struct {
	Code string             `json:"code"`
	Msg  string             `json:"msg"`
	Data []bitgetSpotSymbol `json:"data"`
}

// bitgetFuturesContract is a single contract from the futures contracts endpoint.
type bitgetFuturesContract struct {
	Symbol         string `json:"symbol"`
	BaseCoin       string `json:"baseCoin"`
	QuoteCoin      string `json:"quoteCoin"`
	ContractStatus string `json:"contractStatus"`
}

// bitgetFuturesContractsResponse is the API response for /api/v2/mix/market/contracts.
type bitgetFuturesContractsResponse struct {
	Code string                  `json:"code"`
	Msg  string                  `json:"msg"`
	Data []bitgetFuturesContract `json:"data"`
}

// ServerTime fetches the server time from the Bitget API.
func (c *Client) ServerTime(_ context.Context) (model.Response[model.ServerTime], error) {
	body, err := c.doRequest("GET", "/api/v2/public/time", "")
	if err != nil {
		return model.Response[model.ServerTime]{}, err
	}

	var resp bitgetServerTimeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.ServerTime]{}, fmt.Errorf("failed to parse server time response: %w", err)
	}
	if resp.Code != "00000" {
		return model.Response[model.ServerTime]{}, fmt.Errorf("API error: %s", resp.Msg)
	}

	tsMs, err := strconv.ParseInt(resp.Data.ServerTime, 10, 64)
	if err != nil {
		return model.Response[model.ServerTime]{}, fmt.Errorf("failed to parse server timestamp: %w", err)
	}
	serverTime := time.UnixMilli(tsMs)

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
	switch market {
	case model.MarketFutures:
		return c.futuresExchangeInfo(market)
	default:
		return c.spotExchangeInfo(market)
	}
}

func (c *Client) spotExchangeInfo(market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	body, err := c.doRequest("GET", "/api/v2/spot/public/symbols", "")
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}

	var resp bitgetSpotSymbolsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to parse spot symbols response: %w", err)
	}
	if resp.Code != "00000" {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("API error: %s", resp.Msg)
	}

	symbols := make([]model.Symbol, 0, len(resp.Data))
	for _, s := range resp.Data {
		symbols = append(symbols, model.Symbol{
			Symbol:     s.Symbol,
			BaseAsset:  s.BaseCoin,
			QuoteAsset: s.QuoteCoin,
			Status:     convertSpotStatus(s.Status),
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
	body, err := c.doRequest("GET", "/api/v2/mix/market/contracts", "productType=USDT-FUTURES")
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}

	var resp bitgetFuturesContractsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to parse futures contracts response: %w", err)
	}
	if resp.Code != "00000" {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("API error: %s", resp.Msg)
	}

	symbols := make([]model.Symbol, 0, len(resp.Data))
	for _, s := range resp.Data {
		symbols = append(symbols, model.Symbol{
			Symbol:     s.Symbol,
			BaseAsset:  s.BaseCoin,
			QuoteAsset: s.QuoteCoin,
			Status:     convertFuturesStatus(s.ContractStatus),
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

// convertSpotStatus maps Bitget spot symbol status strings to model.SymbolStatus.
func convertSpotStatus(status string) model.SymbolStatus {
	switch status {
	case "online":
		return model.SymbolStatusTrading
	case "halt":
		return model.SymbolStatusHalt
	case "offline":
		return model.SymbolStatusBreak
	default:
		return model.SymbolStatusBreak
	}
}

// convertFuturesStatus maps Bitget futures contract status strings to model.SymbolStatus.
func convertFuturesStatus(status string) model.SymbolStatus {
	switch status {
	case "normal":
		return model.SymbolStatusTrading
	case "halt":
		return model.SymbolStatusHalt
	default:
		return model.SymbolStatusBreak
	}
}
