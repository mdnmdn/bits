package whitebit

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/model"
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
	StockPrec     string `json:"stockPrec"`
	MoneyPrec     string `json:"moneyPrec"`
	MinAmount     string `json:"minAmount"`
	MinTotal      string `json:"minTotal"`
	MaxTotal      string `json:"maxTotal"`
	MakerFee      string `json:"makerFee"`
	TakerFee      string `json:"takerFee"`
	TradesEnabled bool   `json:"tradesEnabled"`
}

type whitebitFuturesMarket struct {
	TickerID      string         `json:"ticker_id"`
	StockCurrency string         `json:"stock_currency"`
	MoneyCurrency string         `json:"money_currency"`
	Brackets      map[string]int `json:"brackets"`
	MaxLeverage   int            `json:"max_leverage"`
}

// ServerTime fetches the server time from the WhiteBit API.
func (c *Client) ServerTime(_ context.Context) (model.Response[model.ServerTime], error) {
	body, err := c.doRequest("/api/v4/public/time")
	if err != nil {
		return model.Response[model.ServerTime]{}, err
	}

	var resp whitebitTimeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.ServerTime]{}, providerErr(model.ErrKindParse, "failed to parse server time response", err)
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
		return model.Response[model.ExchangeInfo]{}, providerErr(model.ErrKindParse, "failed to parse markets response", err)
	}

	symbols := make([]model.Symbol, 0, len(resp))
	for _, m := range resp {
		status := model.SymbolStatusBreak
		if m.TradesEnabled {
			status = model.SymbolStatusTrading
		}

		qp, _ := strconv.Atoi(m.StockPrec)
		pp, _ := strconv.Atoi(m.MoneyPrec)
		minQty, _ := strconv.ParseFloat(m.MinAmount, 64)
		minPrice, _ := strconv.ParseFloat(m.MinTotal, 64)
		maxPrice, _ := strconv.ParseFloat(m.MaxTotal, 64)
		maxQty := maxPrice
		makerFee, _ := strconv.ParseFloat(m.MakerFee, 64)
		takerFee, _ := strconv.ParseFloat(m.TakerFee, 64)
		makerFee = makerFee / 100
		takerFee = takerFee / 100

		symbols = append(symbols, model.Symbol{
			Symbol:         m.Name,
			BaseAsset:      m.Stock,
			QuoteAsset:     m.Money,
			Status:         status,
			Market:         market,
			QtyPrecision:   &qp,
			PricePrecision: &pp,
			MinQty:         &minQty,
			MinPrice:       &minPrice,
			MaxPrice:       &maxPrice,
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
	body, err := c.doRequest("/api/v4/public/futures")
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}

	var resp struct {
		Success bool                    `json:"success"`
		Result  []whitebitFuturesMarket `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.ExchangeInfo]{}, providerErr(model.ErrKindParse, "failed to parse futures response", err)
	}

	symbols := make([]model.Symbol, 0, len(resp.Result))
	for _, m := range resp.Result {
		minQty := 0.0
		if v, ok := m.Brackets["1"]; ok && v > 0 {
			minQty = float64(v)
		}
		maxQty := 0.0
		if v, ok := m.Brackets["50"]; ok && v > 0 {
			maxQty = float64(v)
		} else if v, ok := m.Brackets["20"]; ok && v > 0 {
			maxQty = float64(v)
		} else if v, ok := m.Brackets["10"]; ok && v > 0 {
			maxQty = float64(v)
		}

		symbols = append(symbols, model.Symbol{
			Symbol:     m.TickerID,
			BaseAsset:  m.StockCurrency,
			QuoteAsset: m.MoneyCurrency,
			Status:     model.SymbolStatusTrading,
			Market:     market,
			MinQty:     &minQty,
			MaxQty:     &maxQty,
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
