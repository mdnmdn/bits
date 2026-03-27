package process

import "github.com/mdnmdn/bits/internal/model"

// SpreadCalculator adds bid-ask spread and mid price to OrderBook.Extra.
func SpreadCalculator(res model.Response[model.OrderBook]) model.Response[model.OrderBook] {
	ob := res.Data
	if len(ob.Bids) == 0 || len(ob.Asks) == 0 {
		return res
	}
	bestBid := ob.Bids[0].Price
	bestAsk := ob.Asks[0].Price
	spread := bestAsk - bestBid
	mid := (bestBid + bestAsk) / 2
	if ob.Extra == nil {
		ob.Extra = make(map[string]any)
	}
	ob.Extra["spread"] = spread
	ob.Extra["mid_price"] = mid
	res.Data = ob
	return res
}
