package binance

import (
	"context"
	"fmt"
	"strings"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/mdnmdn/bits/pkg/model"
)

// ServerTime returns the server time from Binance.
// For spot/margin it uses the spot client; for futures the futures client.
func (c *Client) ServerTime(ctx context.Context) (model.Response[model.ServerTime], error) {
	var ms int64
	var err error

	if c.spotClient != nil {
		ms, err = c.spotClient.NewServerTimeService().Do(ctx)
	} else if c.futuresClient != nil {
		ms, err = c.futuresClient.NewServerTimeService().Do(ctx)
	} else {
		return model.Response[model.ServerTime]{}, fmt.Errorf("binance: no client configured")
	}

	if err != nil {
		return model.Response[model.ServerTime]{}, fmt.Errorf("binance: server time: %w", err)
	}

	t := time.UnixMilli(ms)
	return model.Response[model.ServerTime]{
		Kind:     model.KindServerTime,
		Provider: providerID,
		Data: model.ServerTime{
			Time: t,
		},
	}, nil
}

// ExchangeInfo returns exchange info from Binance for the given market.
func (c *Client) ExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	switch market {
	case model.MarketFutures:
		if c.futuresClient == nil {
			return model.Response[model.ExchangeInfo]{}, fmt.Errorf("binance: futures client not configured")
		}
		info, err := c.futuresClient.NewExchangeInfoService().Do(ctx)
		if err != nil {
			return model.Response[model.ExchangeInfo]{}, fmt.Errorf("binance: exchange info (futures): %w", err)
		}
		symbols := make([]model.Symbol, 0, len(info.Symbols))
		for _, s := range info.Symbols {
			symbols = append(symbols, convertFuturesSymbol(s))
		}
		return model.Response[model.ExchangeInfo]{
			Kind:     model.KindExchangeInfo,
			Provider: providerID,
			Market:   market,
			Data: model.ExchangeInfo{
				ExchangeID: "binance",
				Market:     market,
				Symbols:    symbols,
			},
		}, nil

	default: // spot or margin use spotClient
		if c.spotClient == nil {
			return model.Response[model.ExchangeInfo]{}, fmt.Errorf("binance: spot client not configured")
		}
		info, err := c.spotClient.NewExchangeInfoService().Do(ctx)
		if err != nil {
			return model.Response[model.ExchangeInfo]{}, fmt.Errorf("binance: exchange info (spot): %w", err)
		}
		symbols := make([]model.Symbol, 0, len(info.Symbols))
		for _, s := range info.Symbols {
			symbols = append(symbols, convertSpotSymbol(s))
		}
		return model.Response[model.ExchangeInfo]{
			Kind:     model.KindExchangeInfo,
			Provider: providerID,
			Market:   market,
			Data: model.ExchangeInfo{
				ExchangeID: "binance",
				Market:     market,
				Symbols:    symbols,
			},
		}, nil
	}
}

// convertStatus converts a Binance symbol status string to model.SymbolStatus.
func convertStatus(s string) model.SymbolStatus {
	switch strings.ToUpper(s) {
	case "TRADING":
		return model.SymbolStatusTrading
	case "BREAK":
		return model.SymbolStatusBreak
	case "HALT", "PRE_DELIVERING", "END_OF_DAY", "HALT_TRADING",
		"PRE_SETTLE", "SETTLING", "CLOSE":
		return model.SymbolStatusHalt
	default:
		return model.SymbolStatus(strings.ToLower(s))
	}
}

func convertSpotSymbol(s binance.Symbol) model.Symbol {
	sym := model.Symbol{
		Symbol:     s.Symbol,
		BaseAsset:  s.BaseAsset,
		QuoteAsset: s.QuoteAsset,
		Status:     convertStatus(s.Status),
		Market:     model.MarketSpot,
	}
	pp := s.QuotePrecision
	sym.PricePrecision = &pp
	qp := int(s.BaseAssetPrecision)
	sym.QtyPrecision = &qp

	if pf := s.PriceFilter(); pf != nil {
		if minP, ok := parseFloat(pf.MinPrice); ok {
			sym.MinPrice = minP
		}
		if maxP, ok := parseFloat(pf.MaxPrice); ok {
			sym.MaxPrice = maxP
		}
		if tick, ok := parseFloat(pf.TickSize); ok {
			sym.StepSize = tick
		}
	}
	if lsf := s.LotSizeFilter(); lsf != nil {
		if minQ, ok := parseFloat(lsf.MinQuantity); ok {
			sym.MinQty = minQ
		}
		if maxQ, ok := parseFloat(lsf.MaxQuantity); ok {
			sym.MaxQty = maxQ
		}
		if step, ok := parseFloat(lsf.StepSize); ok {
			sym.StepSize = step
		}
	}
	return sym
}

func convertFuturesSymbol(s futures.Symbol) model.Symbol {
	sym := model.Symbol{
		Symbol:     s.Symbol,
		BaseAsset:  s.BaseAsset,
		QuoteAsset: s.QuoteAsset,
		Status:     convertStatus(s.Status),
		Market:     model.MarketFutures,
	}
	pp := s.PricePrecision
	sym.PricePrecision = &pp
	qp := s.QuantityPrecision
	sym.QtyPrecision = &qp

	if pf := s.PriceFilter(); pf != nil {
		if minP, ok := parseFloat(pf.MinPrice); ok {
			sym.MinPrice = minP
		}
		if maxP, ok := parseFloat(pf.MaxPrice); ok {
			sym.MaxPrice = maxP
		}
		if tick, ok := parseFloat(pf.TickSize); ok {
			sym.StepSize = tick
		}
	}
	if lsf := s.LotSizeFilter(); lsf != nil {
		if minQ, ok := parseFloat(lsf.MinQuantity); ok {
			sym.MinQty = minQ
		}
		if maxQ, ok := parseFloat(lsf.MaxQuantity); ok {
			sym.MaxQty = maxQ
		}
		if step, ok := parseFloat(lsf.StepSize); ok {
			sym.StepSize = step
		}
	}
	return sym
}

func parseFloat(s string) (*float64, bool) {
	if s == "" || s == "0" {
		return nil, false
	}
	var v float64
	if _, err := fmt.Sscanf(s, "%f", &v); err == nil {
		return &v, true
	}
	return nil, false
}
