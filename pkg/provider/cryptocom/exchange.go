package cryptocom

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdnmdn/bits/pkg/model"
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

// ExchangeInfo returns a list of popular SPOT instruments for Crypto.com.
// Note: The public/get-instruments endpoint is currently returning errors,
// so we return a curated list of popular trading pairs.
func (c *Client) ExchangeInfo(_ context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	popularPairs := []struct {
		Symbol         string
		BaseAsset      string
		QuoteAsset     string
		PricePrecision int
	}{
		{"BTC_USDT", "BTC", "USDT", 2},
		{"ETH_USDT", "ETH", "USDT", 2},
		{"XRP_USDT", "XRP", "USDT", 4},
		{"DOGE_USDT", "DOGE", "USDT", 5},
		{"SOL_USDT", "SOL", "USDT", 3},
		{"ADA_USDT", "ADA", "USDT", 4},
		{"DOT_USDT", "DOT", "USDT", 3},
		{"AVAX_USDT", "AVAX", "USDT", 3},
		{"LINK_USDT", "LINK", "USDT", 3},
		{"MATIC_USDT", "MATIC", "USDT", 4},
		{"LTC_USDT", "LTC", "USDT", 2},
		{"UNI_USDT", "UNI", "USDT", 3},
		{"ATOM_USDT", "ATOM", "USDT", 3},
		{"XLM_USDT", "XLM", "USDT", 4},
		{"NEAR_USDT", "NEAR", "USDT", 3},
		{"APT_USDT", "APT", "USDT", 3},
		{"ARB_USDT", "ARB", "USDT", 3},
		{"OP_USDT", "OP", "USDT", 3},
		{"SHIB_USDT", "SHIB", "USDT", 6},
		{"FIL_USDT", "FIL", "USDT", 3},
		{"TRX_USDT", "TRX", "USDT", 4},
		{"PEPE_USDT", "PEPE", "USDT", 8},
		{"BNB_USDT", "BNB", "USDT", 2},
		{"ETH_BTC", "ETH", "BTC", 5},
		{"BNB_BTC", "BNB", "BTC", 4},
	}

	symbols := make([]model.Symbol, 0, len(popularPairs))
	for _, p := range popularPairs {
		pp := p.PricePrecision
		symbols = append(symbols, model.Symbol{
			Symbol:         p.Symbol,
			BaseAsset:      p.BaseAsset,
			QuoteAsset:     p.QuoteAsset,
			Status:         model.SymbolStatusTrading,
			Market:         model.MarketSpot,
			PricePrecision: &pp,
		})
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
