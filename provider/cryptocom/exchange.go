package cryptocom

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/model"
)

// ServerTime estimates the server time by measuring the round-trip of a
// lightweight API call. Crypto.com v2 REST does not expose a dedicated
// server-time endpoint; uses the timestamp from ticker response as approximation.
func (c *Client) ServerTime(_ context.Context) (model.Response[model.ServerTime], error) {
	before := time.Now()
	body, err := c.doRequest("public/get-ticker", "instrument_name=BTC_USDT")
	after := time.Now()

	if err != nil {
		return model.Response[model.ServerTime]{}, err
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return model.Response[model.ServerTime]{}, fmt.Errorf("failed to parse response: %w", err)
	}
	code, err := env.GetCode()
	if err != nil {
		return model.Response[model.ServerTime]{}, fmt.Errorf("failed to parse code: %w", err)
	}
	if code != 0 {
		return model.Response[model.ServerTime]{}, fmt.Errorf("API error (code %d)", code)
	}

	latency := after.Sub(before)
	serverTime := before.Add(latency / 2)

	return model.Response[model.ServerTime]{
		Kind:     model.KindServerTime,
		Provider: providerID,
		Data: model.ServerTime{
			Time:    serverTime,
			Latency: &latency,
		},
	}, nil
}

// ExchangeInfo returns a list of SPOT or MARGIN instruments for Crypto.com.
func (c *Client) ExchangeInfo(_ context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	body, err := c.doRequest("public/get-instruments", "")
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to fetch instruments: %w", err)
	}

	var resp apiInstrumentsV1Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to parse instruments response: %w", err)
	}

	if resp.Code != 0 {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("API error (code %d)", resp.Code)
	}

	makerFee, takerFee := c.getFeeConfig(market)

	symbols := make([]model.Symbol, 0)
	for _, inst := range resp.Result.Data {
		if inst.InstType != "CCY_PAIR" {
			continue
		}

		isMargin := inst.MarginBuyEnabled || inst.MarginSellEnabled
		if market == model.MarketMargin && !isMargin {
			continue
		}
		if market == model.MarketSpot && isMargin {
			continue
		}

		status := model.SymbolStatusHalt
		if inst.Tradable {
			status = model.SymbolStatusTrading
		}

		pp := inst.QuoteDecimals
		qp := inst.QuantityDecimals
		minPrice, _ := strconv.ParseFloat(inst.PriceTickSize, 64)
		minQty, _ := strconv.ParseFloat(inst.QtyTickSize, 64)

		symbols = append(symbols, model.Symbol{
			Symbol:         inst.Symbol,
			BaseAsset:      inst.BaseCcy,
			QuoteAsset:     inst.QuoteCcy,
			Status:         status,
			Market:         market,
			PricePrecision: &pp,
			QtyPrecision:   &qp,
			MinPrice:       &minPrice,
			MinQty:         &minQty,
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

func (c *Client) getFeeConfig(market model.MarketType) (makerFee, takerFee float64) {
	if c.cfg.Spot.MakerFee > 0 {
		return c.cfg.Spot.MakerFee, c.cfg.Spot.TakerFee
	}
	return 0.001, 0.002 // Crypto.com default: 0.1% maker, 0.2% taker
}
