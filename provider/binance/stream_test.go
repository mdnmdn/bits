package binance

import (
	"context"
	"testing"

	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider"
	"github.com/stretchr/testify/assert"
)

func TestGetStreamBaseURL(t *testing.T) {
	tests := []struct {
		name             string
		cfg              config.BinanceConfig
		market           model.MarketType
		expectedURL      string
		description      string
	}{
		{
			name:        "spot production",
			cfg:         config.BinanceConfig{},
			market:      model.MarketSpot,
			expectedURL: "wss://stream.binance.com:9443/stream",
			description: "Spot market uses production host",
		},
		{
			name: "spot testnet",
			cfg: config.BinanceConfig{
				Spot: config.MarketConfig{UseTestnet: true},
			},
			market:      model.MarketSpot,
			expectedURL: "wss://testnet.binance.vision/stream",
			description: "Spot market testnet uses testnet.binance.vision",
		},
		{
			name:        "margin production",
			cfg:         config.BinanceConfig{},
			market:      model.MarketMargin,
			expectedURL: "wss://stream.binance.com:9443/stream",
			description: "Margin market uses same endpoint as spot",
		},
		{
			name: "margin testnet",
			cfg: config.BinanceConfig{
				Spot: config.MarketConfig{UseTestnet: true},
			},
			market:      model.MarketMargin,
			expectedURL: "wss://testnet.binance.vision/stream",
			description: "Margin market testnet uses testnet.binance.vision",
		},
		{
			name:        "futures production",
			cfg:         config.BinanceConfig{},
			market:      model.MarketFutures,
			expectedURL: "wss://fstream.binance.com/stream",
			description: "Futures market uses fstream.binance.com",
		},
		{
			name: "futures testnet",
			cfg: config.BinanceConfig{
				Futures: config.MarketConfig{UseTestnet: true},
			},
			market:      model.MarketFutures,
			expectedURL: "wss://stream.binancefuture.com/stream",
			description: "Futures market testnet uses stream.binancefuture.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{cfg: tt.cfg}
			url := client.getStreamBaseURL(tt.market)
			assert.Equal(t, tt.expectedURL, url, tt.description)
		})
	}
}

func TestParseTradeEvent(t *testing.T) {
	spot := []byte(`{"e":"trade","E":1700000000100,"s":"BTCUSDT","t":12345,"p":"42000.50","q":"0.25","T":1700000000050,"m":true}`)
	tr, ok := parseTradeEvent(spot)
	assert.True(t, ok)
	assert.Equal(t, "12345", tr.ID)
	assert.Equal(t, 42000.5, tr.Price)
	assert.Equal(t, 0.25, tr.Quantity)
	assert.Equal(t, model.TradeSideSell, tr.Side, "buyer is maker: taker sold")
	assert.Equal(t, int64(1700000000050), tr.Time.UnixMilli())
	assert.Equal(t, "UTC", tr.Time.Location().String())
	assert.Equal(t, "BTCUSDT", tr.OriginalSymbol)

	agg := []byte(`{"e":"aggTrade","E":1700000000100,"s":"BTCUSDT","a":777,"p":"1","q":"2","f":1,"l":2,"T":1700000000060,"m":false}`)
	tr, ok = parseTradeEvent(agg)
	assert.True(t, ok)
	assert.Equal(t, "777", tr.ID)
	assert.Equal(t, model.TradeSideBuy, tr.Side)

	_, ok = parseTradeEvent([]byte(`{"s":"BTCUSDT","p":"x","q":"1","T":1}`))
	assert.False(t, ok)
	_, ok = parseTradeEvent([]byte(`not json`))
	assert.False(t, ok)
}

func TestHandleTradeStreams(t *testing.T) {
	h := &binanceHandler{providerID: "binance"}
	for _, stream := range []string{"btcusdt@trade", "btcusdt@aggTrade"} {
		raw := []byte(`{"stream":"` + stream + `","data":{"e":"trade","s":"BTCUSDT","t":1,"a":1,"p":"1","q":"1","T":1700000000000,"m":false}}`)
		res, err := h.Handle(context.Background(), raw)
		assert.NoError(t, err)
		r, ok := res.(*model.Response[model.Trade])
		assert.True(t, ok, stream)
		assert.Equal(t, model.KindTrade, r.Kind)
	}
}

func TestStartTradeStreamNeedsOneSymbol(t *testing.T) {
	c := &Client{}
	_, err := c.StartTradeStream(context.Background(), []string{"A", "B"}, model.MarketSpot)
	assert.Error(t, err)
}

var _ provider.TradeStreamProvider = (*Client)(nil)
