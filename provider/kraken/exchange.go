package kraken

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/model"
)

// ServerTime returns the venue clock from the serverTime field of the small
// public /feeschedules response (there is no dedicated time endpoint).
func (c *Client) ServerTime(ctx context.Context) (model.Response[model.ServerTime], error) {
	before := time.Now()
	var body struct {
		ServerTime string `json:"serverTime"`
	}
	err := c.get(ctx, "/derivatives/api/v3/feeschedules", &body)
	latency := time.Since(before)
	if err != nil {
		return model.Response[model.ServerTime]{}, err
	}
	t, err := time.Parse(time.RFC3339Nano, body.ServerTime)
	if err != nil {
		return model.Response[model.ServerTime]{}, providerErr(model.ErrKindParse, "bad serverTime "+body.ServerTime, err)
	}
	return model.Response[model.ServerTime]{
		Kind: model.KindServerTime, Provider: providerID, Market: model.MarketFutures,
		Data: model.ServerTime{Time: t.UTC(), Latency: &latency},
	}, nil
}

// ExchangeInfo lists the futures instruments (spot and margin: see spotExchangeInfo). Published per instrument: the
// tick size (Extra["tick_size"]; model.Symbol has no tick field) and the
// contract size increment 10^-contractValueTradePrecision (StepSize,
// QtyPrecision). The venue publishes no minimum quantity and no minimum
// notional, so MinQty and MinNotional stay nil. Fees come from the public
// /feeschedules endpoint, first tier of the schedule the instrument points
// to, as fractions (the venue gives percents); they are not account-specific
// and are left nil if the schedule list cannot be read.
func (c *Client) ExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	if isSpotMarket(market) {
		return c.spotExchangeInfo(ctx, market)
	}
	if err := requireFutures(market); err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}
	var body struct {
		ServerTime  string `json:"serverTime"`
		Instruments []struct {
			Symbol                      string  `json:"symbol"`
			Type                        string  `json:"type"`
			Base                        string  `json:"base"`
			Quote                       string  `json:"quote"`
			TickSize                    float64 `json:"tickSize"`
			ContractSize                float64 `json:"contractSize"`
			ContractValueTradePrecision int     `json:"contractValueTradePrecision"`
			Tradeable                   bool    `json:"tradeable"`
			IsExpired                   bool    `json:"isExpired"`
			MaxPositionSize             float64 `json:"maxPositionSize"`
			FeeScheduleUID              string  `json:"feeScheduleUid"`
		} `json:"instruments"`
	}
	if err := c.get(ctx, "/derivatives/api/v3/instruments", &body); err != nil {
		return model.Response[model.ExchangeInfo]{}, err
	}
	type fee struct{ maker, taker float64 }
	fees := map[string]fee{}
	var sched struct {
		FeeSchedules []struct {
			UID   string `json:"uid"`
			Tiers []struct {
				MakerFee float64 `json:"makerFee"`
				TakerFee float64 `json:"takerFee"`
			} `json:"tiers"`
		} `json:"feeSchedules"`
	}
	if c.get(ctx, "/derivatives/api/v3/feeschedules", &sched) == nil {
		for _, s := range sched.FeeSchedules {
			if len(s.Tiers) > 0 {
				fees[s.UID] = fee{s.Tiers[0].MakerFee / 100, s.Tiers[0].TakerFee / 100}
			}
		}
	}

	symbols := make([]model.Symbol, 0, len(body.Instruments))
	for _, in := range body.Instruments {
		status := model.SymbolStatusHalt
		if in.Tradeable && !in.IsExpired {
			status = model.SymbolStatusTrading
		}
		base, quote := in.Base, in.Quote
		s := model.Symbol{
			Symbol:           in.Symbol,
			NormalizedSymbol: base + "-" + quote,
			BaseAsset:        base,
			QuoteAsset:       quote,
			Status:           status,
			Market:           model.MarketFutures,
			Extra: map[string]any{
				"tick_size":     in.TickSize,
				"contract_size": in.ContractSize,
				"type":          in.Type,
			},
		}
		if base == "" || quote == "" {
			s.NormalizedSymbol = normalize(in.Symbol)
		}
		step := math.Pow(10, -float64(in.ContractValueTradePrecision))
		qp := max(in.ContractValueTradePrecision, 0)
		s.StepSize, s.QtyPrecision = &step, &qp
		if in.TickSize > 0 {
			pp := 0
			if _, frac, ok := strings.Cut(strconv.FormatFloat(in.TickSize, 'f', -1, 64), "."); ok {
				pp = len(frac)
			}
			s.PricePrecision = &pp
		}
		if in.MaxPositionSize > 0 {
			v := in.MaxPositionSize
			s.MaxQty = &v
		}
		if f, ok := fees[in.FeeScheduleUID]; ok {
			m, t := f.maker, f.taker
			s.MakerFee, s.TakerFee = &m, &t
		}
		symbols = append(symbols, s)
	}
	info := model.ExchangeInfo{ExchangeID: providerID, Market: model.MarketFutures, Symbols: symbols}
	if t, err := time.Parse(time.RFC3339Nano, body.ServerTime); err == nil {
		t = t.UTC()
		info.ServerTime = &t
	}
	return model.Response[model.ExchangeInfo]{Kind: model.KindExchangeInfo, Provider: providerID, Market: model.MarketFutures, Data: info}, nil
}
