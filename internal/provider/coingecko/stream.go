package coingecko

import (
	"context"

	"github.com/mdnmdn/bits/internal/model"
	legacyws "github.com/mdnmdn/bits/internal/legacy/ws"
)

// WatchPrices streams live price updates for the given coin IDs.
func (c *Client) WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	wsClient := legacyws.NewClient(c.cfg, ids)
	wsClient.UserAgent = c.UserAgent

	legacyCh, err := wsClient.Connect(ctx)
	if err != nil {
		return nil, err
	}

	out := make(chan *model.CoinPrice, 64)
	go func() {
		defer close(out)
		for update := range legacyCh {
			change := update.Change24h
			cp := &model.CoinPrice{
				ID:        update.CoinID,
				Symbol:    update.CoinID,
				Currency:  "usd",
				Price:     update.Price,
				Change24h: &change,
			}
			select {
			case out <- cp:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}
