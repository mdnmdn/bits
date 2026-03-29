package cryptocom

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/internal/model"
)

// ServerTime estimates the server time by measuring the round-trip of a
// lightweight API call. Crypto.com v2 REST does not expose a dedicated
// server-time endpoint; the midpoint of the request latency is used instead.
func (c *Client) ServerTime(_ context.Context) (model.Response[model.ServerTime], error) {
	before := time.Now()
	body, err := c.doRequest("public/get-instruments", "")
	after := time.Now()

	if err != nil {
		return model.Response[model.ServerTime]{}, err
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return model.Response[model.ServerTime]{}, fmt.Errorf("failed to parse response: %w", err)
	}
	if env.Code != 0 {
		return model.Response[model.ServerTime]{}, fmt.Errorf("API error (code %d)", env.Code)
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

// ExchangeInfo fetches all SPOT instruments from public/get-instruments.
// Crypto.com v2 only exposes spot instruments via this endpoint.
func (c *Client) ExchangeInfo(_ context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	body, err := c.doRequest("public/get-instruments", "")
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to parse instruments response: %w", err)
	}
	if env.Code != 0 {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("API error (code %d)", env.Code)
	}

	var result apiInstrumentsResult
	if err := json.Unmarshal(env.Result, &result); err != nil {
		return model.Response[model.ExchangeInfo]{}, fmt.Errorf("failed to parse instruments result: %w", err)
	}

	symbols := make([]model.Symbol, 0, len(result.Instruments))
	for _, inst := range result.Instruments {
		if inst.ProductType != "SPOT" {
			continue
		}

		sym := model.Symbol{
			Symbol:     inst.InstrumentName,
			BaseAsset:  inst.BaseCurrency,
			QuoteAsset: inst.QuoteCurrency,
			Status:     model.SymbolStatusTrading,
			Market:     model.MarketSpot,
		}

		if inst.PricePrecision > 0 {
			pp := inst.PricePrecision
			sym.PricePrecision = &pp
		}
		if inst.QuantityPrecision > 0 {
			qp := inst.QuantityPrecision
			sym.QtyPrecision = &qp
		}
		if inst.MinOrderSize != "" {
			if v, err := strconv.ParseFloat(inst.MinOrderSize, 64); err == nil {
				sym.MinQty = &v
			}
		}
		if inst.MaxOrderSize != "" {
			if v, err := strconv.ParseFloat(inst.MaxOrderSize, 64); err == nil {
				sym.MaxQty = &v
			}
		}

		symbols = append(symbols, sym)
	}

	return model.Response[model.ExchangeInfo]{
		Kind:     model.KindExchangeInfo,
		Provider: providerID,
		Market:   model.MarketSpot,
		Data: model.ExchangeInfo{
			ExchangeID: providerID,
			Market:     model.MarketSpot,
			Symbols:    symbols,
		},
	}, nil
}
