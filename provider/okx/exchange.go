package okx

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/model"
)

// instType returns the OKX instType for a market. Futures means perpetual
// swaps. Margin trades the spot instIds; only the instruments listing differs.
func instType(market model.MarketType) string {
	switch market {
	case model.MarketFutures:
		return "SWAP"
	case model.MarketMargin:
		return "MARGIN"
	}
	return "SPOT"
}

// instID converts any accepted symbol form ("BTC-USDT", "BTCUSDT",
// "BTC-USDT-SWAP") to the native OKX instId for the market.
func instID(symbol string, market model.MarketType) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	s = strings.TrimSuffix(s, "-SWAP")
	if !strings.Contains(s, "-") {
		s = model.NormalizeSymbol(s)
	}
	if market == model.MarketFutures {
		return s + "-SWAP"
	}
	return s
}

// normalizedFromInstID returns "BASE-QUOTE" for an OKX instId.
func normalizedFromInstID(id string) string {
	return strings.TrimSuffix(id, "-SWAP")
}

type okxInstrument struct {
	InstID    string `json:"instId"`
	BaseCcy   string `json:"baseCcy"`
	QuoteCcy  string `json:"quoteCcy"`
	SettleCcy string `json:"settleCcy"`
	Uly       string `json:"uly"`
	CtType    string `json:"ctType"`
	CtVal     string `json:"ctVal"`
	CtValCcy  string `json:"ctValCcy"`
	TickSz    string `json:"tickSz"`
	LotSz     string `json:"lotSz"`
	MinSz     string `json:"minSz"`
	MaxLmtSz  string `json:"maxLmtSz"`
	MaxMktSz  string `json:"maxMktSz"`
	State     string `json:"state"`
}

type okxTime struct {
	Ts string `json:"ts"`
}

// ServerTime fetches the OKX system time.
func (c *Client) ServerTime(ctx context.Context) (model.Response[model.ServerTime], error) {
	data, err := c.get(ctx, "/api/v5/public/time", "")
	if err != nil {
		return model.Response[model.ServerTime]{}, err
	}
	var rows []okxTime
	if err := json.Unmarshal(data, &rows); err != nil || len(rows) == 0 {
		return model.Response[model.ServerTime]{}, providerErr(model.ErrKindParse, "invalid server time response", err)
	}
	ms, err := strconv.ParseInt(rows[0].Ts, 10, 64)
	if err != nil {
		return model.Response[model.ServerTime]{}, providerErr(model.ErrKindParse, "invalid server time: "+rows[0].Ts, err)
	}
	return model.Response[model.ServerTime]{
		Kind:     model.KindServerTime,
		Provider: providerID,
		Data:     model.ServerTime{Time: time.UnixMilli(ms).UTC()},
	}, nil
}

// ExchangeInfo lists the instruments of a market with their trading rules.
//
// For SWAP, quantities (min_q, max_q, step) are in contracts; the contract size
// is in Extra ("ct_val", "ct_val_ccy", "ct_type"). OKX publishes no minimum
// notional and no public fee schedule, so MinNotional, MakerFee and TakerFee
// are left unset.
func (c *Client) ExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	data, err := c.get(ctx, "/api/v5/public/instruments", "instType="+instType(market))
	if err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}
	var rows []okxInstrument
	if err := json.Unmarshal(data, &rows); err != nil {
		return model.Response[model.ExchangeInfo]{}, providerErr(model.ErrKindParse, "failed to parse instruments", err)
	}

	symbols := make([]model.Symbol, 0, len(rows))
	for _, r := range rows {
		base, quote := r.BaseCcy, r.QuoteCcy
		if market == model.MarketFutures {
			// uly is "BTC-USDT" (linear) or "BTC-USD" (inverse).
			if b, q, ok := strings.Cut(r.Uly, "-"); ok {
				base, quote = b, q
			}
		}
		sym := model.Symbol{
			Symbol:           r.InstID,
			NormalizedSymbol: normalizedFromInstID(r.InstID),
			BaseAsset:        base,
			QuoteAsset:       quote,
			Status:           convertStatus(r.State),
			Market:           market,
			PricePrecision:   decimalPlaces(r.TickSz),
			QtyPrecision:     decimalPlaces(r.LotSz),
			MinQty:           parseOpt(r.MinSz),
			MaxQty:           parseOpt(r.MaxLmtSz),
			StepSize:         parseOpt(r.LotSz),
		}
		if tick := parseOpt(r.TickSz); tick != nil {
			sym.Extra = map[string]any{"tick_size": *tick}
		}
		if market == model.MarketFutures {
			if sym.Extra == nil {
				sym.Extra = map[string]any{}
			}
			sym.Extra["ct_type"] = r.CtType
			sym.Extra["ct_val_ccy"] = r.CtValCcy
			if v := parseOpt(r.CtVal); v != nil {
				sym.Extra["ct_val"] = *v
			}
			sym.Extra["settle_ccy"] = r.SettleCcy
		}
		symbols = append(symbols, sym)
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

// convertStatus maps OKX instrument state to model.SymbolStatus.
func convertStatus(state string) model.SymbolStatus {
	switch state {
	case "live":
		return model.SymbolStatusTrading
	case "suspend", "preopen":
		return model.SymbolStatusHalt
	default:
		return model.SymbolStatusBreak
	}
}

// decimalPlaces returns the number of decimals of a step string such as "0.001".
func decimalPlaces(s string) *int {
	if s == "" {
		return nil
	}
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return nil
	}
	n := 0
	if i := strings.IndexByte(s, '.'); i >= 0 {
		n = len(strings.TrimRight(s[i+1:], "0"))
	}
	return &n
}

// parseOpt parses a float, returning nil for empty or invalid strings.
func parseOpt(s string) *float64 {
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}
