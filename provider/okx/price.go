package okx

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mdnmdn/bits/model"
)

// Price fetches the last traded price of each symbol via /market/ticker.
//
// ids are trading symbols ("BTC-USDT", "BTCUSDT"); currency is echoed in the
// result and not used for lookup. The market is not passed, so a symbol with a
// "-SWAP" suffix ("BTC-USDT-SWAP") is a perpetual swap and any other symbol is
// spot. Margin trades the spot instIds, so it has the same prices as spot.
// A failing symbol is reported in Response.Errors and does not fail the batch.
func (c *Client) Price(ctx context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error) {
	prices := make([]model.CoinPrice, 0, len(ids))
	var itemErrors []model.ItemError

	for _, symbol := range ids {
		market := model.MarketSpot
		if strings.HasSuffix(strings.ToUpper(strings.TrimSpace(symbol)), "-SWAP") {
			market = model.MarketFutures
		}
		r, err := c.fetchTicker(ctx, instID(symbol, market))
		if err == nil {
			var cp model.CoinPrice
			cp, err = coinPrice(r, symbol, currency, market)
			if err == nil {
				prices = append(prices, cp)
				continue
			}
		}
		itemErrors = append(itemErrors, model.ItemError{Symbol: symbol, Err: model.WrapError(providerID, err)})
	}

	return model.Response[[]model.CoinPrice]{
		Kind:     model.KindPrice,
		Provider: providerID,
		Data:     prices,
		Errors:   itemErrors,
	}, nil
}

func coinPrice(r okxTicker, id, currency string, market model.MarketType) (model.CoinPrice, error) {
	last, err := strconv.ParseFloat(r.Last, 64)
	if err != nil {
		return model.CoinPrice{}, providerErr(model.ErrKindParse, fmt.Sprintf("invalid last price %q", r.Last), err)
	}
	cp := model.CoinPrice{
		ID:             id,
		Symbol:         normalizedFromInstID(r.InstID),
		OriginalSymbol: r.InstID,
		Currency:       currency,
		Price:          last,
		High24h:        parseOpt(r.High24h),
		Low24h:         parseOpt(r.Low24h),
		Open24h:        parseOpt(r.Open24h),
		BidPrice:       parseOpt(r.BidPx),
		AskPrice:       parseOpt(r.AskPx),
	}
	if market == model.MarketFutures {
		cp.Volume24h = parseOpt(r.VolCcy24h) // base coin; vol24h is contracts
	} else {
		cp.Volume24h = parseOpt(r.Vol24h)
	}
	if cp.Open24h != nil && *cp.Open24h != 0 {
		pct := (last - *cp.Open24h) / *cp.Open24h * 100
		cp.Change24h = &pct
	}
	return cp, nil
}
