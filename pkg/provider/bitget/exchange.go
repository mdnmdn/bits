package bitget

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/internal/logger"
	"github.com/mdnmdn/bits/pkg/model"
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
	Symbol            string `json:"symbol"`
	BaseCoin          string `json:"baseCoin"`
	QuoteCoin         string `json:"quoteCoin"`
	Status            string `json:"status"`
	PricePrecision    string `json:"pricePrecision"`
	QuantityPrecision string `json:"quantityPrecision"`
	MinTradeAmount    string `json:"minTradeAmount"`
	MaxTradeAmount    string `json:"maxTradeAmount"`
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
	ContractStatus string `json:"symbolStatus"`
	PricePlace     string `json:"pricePlace"`
	VolumePlace    string `json:"volumePlace"`
	MinTradeNum    string `json:"minTradeNum"`
	MaxOrderQty    string `json:"maxOrderQty"`
}

// bitgetFuturesContractsResponse is the API response for /api/v2/mix/market/contracts.
type bitgetFuturesContractsResponse struct {
	Code string                  `json:"code"`
	Msg  string                  `json:"msg"`
	Data []bitgetFuturesContract `json:"data"`
}

// Bitget margin API - see https://bitgetlimited.github.io/apidoc/en/margin/
type bitgetMarginSymbol struct {
	Symbol            string `json:"symbol"`
	BaseCoin          string `json:"baseCoin"`
	QuoteCoin         string `json:"quoteCoin"`
	Status            string `json:"status"`
	PricePrecision    string `json:"pricePrecision"`
	QuantityPrecision string `json:"quantityPrecision"`
	MinTradeUSDT      string `json:"minTradeUSDT"`
	MakerFeeRate      string `json:"makerFeeRate"`
	TakerFeeRate      string `json:"takerFeeRate"`
	IsBorrowable      bool   `json:"isBorrowable"`
}

type bitgetMarginSymbolsResponse struct {
	Code string               `json:"code"`
	Msg  string               `json:"msg"`
	Data []bitgetMarginSymbol `json:"data"`
}

// bitgetVIPFeeRate is the VIP fee rate from the public endpoint.
type bitgetVIPFeeRate struct {
	Level        string `json:"level"`
	DealAmount   string `json:"dealAmount"`
	AssetAmount  string `json:"assetAmount"`
	TakerFeeRate string `json:"takerFeeRate"`
	MakerFeeRate string `json:"makerFeeRate"`
}

type bitgetVIPFeeRateResponse struct {
	Code string             `json:"code"`
	Msg  string             `json:"msg"`
	Data []bitgetVIPFeeRate `json:"data"`
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
	case model.MarketMargin:
		return c.marginExchangeInfo(market)
	default:
		return c.spotExchangeInfo(market)
	}
}

func (c *Client) marginExchangeInfo(market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	body, err := c.doRequest("GET", "/api/v2/margin/currencies", "")
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}

	var resp bitgetMarginSymbolsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to parse margin symbols response: %w", err)
	}
	if resp.Code != "00000" {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("API error: %s", resp.Msg)
	}

	symbols := make([]model.Symbol, 0, len(resp.Data))
	for _, s := range resp.Data {
		if !s.IsBorrowable {
			continue
		}

		status := model.SymbolStatusTrading
		if s.Status == "offline" {
			status = model.SymbolStatusBreak
		}

		pp, _ := strconv.Atoi(s.PricePrecision)
		qp, _ := strconv.Atoi(s.QuantityPrecision)
		minPrice, _ := strconv.ParseFloat(s.MinTradeUSDT, 64)
		makerFee, _ := strconv.ParseFloat(s.MakerFeeRate, 64)
		takerFee, _ := strconv.ParseFloat(s.TakerFeeRate, 64)

		symbols = append(symbols, model.Symbol{
			Symbol:         s.Symbol,
			BaseAsset:      s.BaseCoin,
			QuoteAsset:     s.QuoteCoin,
			Status:         status,
			Market:         market,
			PricePrecision: &pp,
			QtyPrecision:   &qp,
			MinPrice:       &minPrice,
			MakerFee:       &makerFee,
			TakerFee:       &takerFee,
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

func (c *Client) spotExchangeInfo(market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	symbolsBody, err := c.doRequest("GET", "/api/v2/spot/public/symbols", "")
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}

	var symbolsResp bitgetSpotSymbolsResponse
	if err := json.Unmarshal(symbolsBody, &symbolsResp); err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to parse spot symbols response: %w", err)
	}
	if symbolsResp.Code != "00000" {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("API error: %s", symbolsResp.Msg)
	}

	feeRates, err := c.getSpotFeeRates()
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to get fee rates: %w", err)
	}

	symbols := make([]model.Symbol, 0, len(symbolsResp.Data))
	for _, s := range symbolsResp.Data {
		pp, _ := strconv.Atoi(s.PricePrecision)
		qp, _ := strconv.Atoi(s.QuantityPrecision)
		minQty, _ := strconv.ParseFloat(s.MinTradeAmount, 64)
		maxQty, _ := strconv.ParseFloat(s.MaxTradeAmount, 64)

		makerFee, takerFee := getDefaultFees(feeRates)

		symbols = append(symbols, model.Symbol{
			Symbol:         s.Symbol,
			BaseAsset:      s.BaseCoin,
			QuoteAsset:     s.QuoteCoin,
			Status:         convertSpotStatus(s.Status),
			Market:         market,
			PricePrecision: &pp,
			QtyPrecision:   &qp,
			MinQty:         &minQty,
			MaxQty:         &maxQty,
			MakerFee:       &makerFee,
			TakerFee:       &takerFee,
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
	feeRates, err := c.getFuturesFeeRates()
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to get futures fee rates: %w", err)
	}

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

	makerFee, takerFee := getDefaultFees(feeRates)

	symbols := make([]model.Symbol, 0, len(resp.Data))
	for _, s := range resp.Data {
		pp, _ := strconv.Atoi(s.PricePlace)
		qp, _ := strconv.Atoi(s.VolumePlace)
		minQty, _ := strconv.ParseFloat(s.MinTradeNum, 64)
		maxQty, _ := strconv.ParseFloat(s.MaxOrderQty, 64)

		symbols = append(symbols, model.Symbol{
			Symbol:         s.Symbol,
			BaseAsset:      s.BaseCoin,
			QuoteAsset:     s.QuoteCoin,
			Status:         convertFuturesStatus(s.ContractStatus),
			Market:         market,
			PricePrecision: &pp,
			QtyPrecision:   &qp,
			MinQty:         &minQty,
			MaxQty:         &maxQty,
			MakerFee:       &makerFee,
			TakerFee:       &takerFee,
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

func (c *Client) getSpotFeeRates() ([]bitgetVIPFeeRate, error) {
	logger.Default.Debug("fetching VIP fee rates from Bitget")
	body, err := c.doRequest("GET", "/api/v2/spot/market/vip-fee-rate", "")
	if err != nil {
		logger.Default.Debug("failed to fetch fee rates", "error", err)
		return nil, err
	}

	var resp bitgetVIPFeeRateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		logger.Default.Debug("failed to parse fee rate response", "error", err)
		return nil, fmt.Errorf("failed to parse VIP fee rate response: %w", err)
	}
	if resp.Code != "00000" {
		logger.Default.Debug("API error fetching fee rates", "code", resp.Code, "msg", resp.Msg)
		return nil, fmt.Errorf("API error: %s", resp.Msg)
	}

	logger.Default.Debug("fetched VIP fee rates", "levels", len(resp.Data))
	return resp.Data, nil
}

func getDefaultFees(rates []bitgetVIPFeeRate) (makerFee, takerFee float64) {
	for _, r := range rates {
		if r.Level == "0" {
			makerFee, _ = strconv.ParseFloat(r.MakerFeeRate, 64)
			takerFee, _ = strconv.ParseFloat(r.TakerFeeRate, 64)
			return makerFee, takerFee
		}
	}
	return 0.001, 0.001
}

func (c *Client) getFuturesFeeRates() ([]bitgetVIPFeeRate, error) {
	logger.Default.Debug("fetching VIP fee rates from Bitget (futures)")
	body, err := c.doRequest("GET", "/api/v2/mix/market/vip-fee-rate", "")
	if err != nil {
		logger.Default.Debug("failed to fetch futures fee rates", "error", err)
		return nil, err
	}

	var resp bitgetVIPFeeRateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		logger.Default.Debug("failed to parse futures fee rate response", "error", err)
		return nil, fmt.Errorf("failed to parse futures VIP fee rate response: %w", err)
	}
	if resp.Code != "00000" {
		logger.Default.Debug("API error fetching futures fee rates", "code", resp.Code, "msg", resp.Msg)
		return nil, fmt.Errorf("API error: %s", resp.Msg)
	}

	logger.Default.Debug("fetched futures VIP fee rates", "levels", len(resp.Data))
	return resp.Data, nil
}
