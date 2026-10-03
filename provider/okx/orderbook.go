package okx

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/model"
)

const (
	// maxBookDepth is the largest sz accepted by /market/books.
	maxBookDepth = 400
	// defaultBookDepth is used when the caller passes no depth.
	defaultBookDepth = 20
)

type okxBook struct {
	// Each level is [price, size, deprecated "0", number of orders].
	Asks [][]string `json:"asks"`
	Bids [][]string `json:"bids"`
	Ts   string     `json:"ts"`
}

// OrderBook fetches a depth snapshot via /market/books. depth is the number of
// levels per side, capped at 400 (default 20). For SWAP the quantity is in
// contracts (see ExchangeInfo Extra "ct_val"). Margin uses the spot book.
func (c *Client) OrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error) {
	id := instID(symbol, market)
	if depth <= 0 {
		depth = defaultBookDepth
	}
	if depth > maxBookDepth {
		depth = maxBookDepth
	}
	data, err := c.get(ctx, "/api/v5/market/books", fmt.Sprintf("instId=%s&sz=%d", id, depth))
	if err != nil {
		return model.Response[model.OrderBook]{}, err
	}
	var rows []okxBook
	if err := json.Unmarshal(data, &rows); err != nil {
		return model.Response[model.OrderBook]{}, providerErr(model.ErrKindParse, "failed to parse order book", err)
	}
	if len(rows) == 0 {
		return model.Response[model.OrderBook]{}, providerErr(model.ErrKindNotFound, "no order book for "+id, nil)
	}
	r := rows[0]

	bids, err := parseLevels(r.Bids)
	if err != nil {
		return model.Response[model.OrderBook]{}, err
	}
	asks, err := parseLevels(r.Asks)
	if err != nil {
		return model.Response[model.OrderBook]{}, err
	}
	book := model.OrderBook{
		Symbol:         normalizedFromInstID(id),
		OriginalSymbol: id,
		Market:         market,
		Bids:           bids,
		Asks:           asks,
	}
	if ms, err := strconv.ParseInt(r.Ts, 10, 64); err == nil {
		t := time.UnixMilli(ms).UTC()
		book.Time = &t
	}
	return model.Response[model.OrderBook]{
		Kind:     model.KindOrderBook,
		Provider: providerID,
		Market:   market,
		Data:     book,
	}, nil
}

func parseLevels(raw [][]string) ([]model.OrderBookEntry, error) {
	out := make([]model.OrderBookEntry, 0, len(raw))
	for _, l := range raw {
		if len(l) < 2 {
			return nil, providerErr(model.ErrKindParse, fmt.Sprintf("expected at least 2 fields in book level, got %d", len(l)), nil)
		}
		p, err := strconv.ParseFloat(l[0], 64)
		if err != nil {
			return nil, providerErr(model.ErrKindParse, fmt.Sprintf("invalid book price %q", l[0]), err)
		}
		q, err := strconv.ParseFloat(l[1], 64)
		if err != nil {
			return nil, providerErr(model.ErrKindParse, fmt.Sprintf("invalid book size %q", l[1]), err)
		}
		out = append(out, model.OrderBookEntry{Price: p, Quantity: q})
	}
	return out, nil
}
