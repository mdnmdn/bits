package mexc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMapInterval(t *testing.T) {
	tests := []struct {
		interval string
		futures  bool
		expected string
	}{
		{"1m", false, "1m"},
		{"5m", false, "5m"},
		{"1h", false, "60m"},
		{"1m", true, "Min1"},
		{"1h", true, "Min60"},
		{"4h", true, "Hour4"},
		{"unknown", false, "60m"},
		{"unknown", true, "Min60"},
	}

	for _, tt := range tests {
		result := mapInterval(tt.interval, tt.futures)
		assert.Equal(t, tt.expected, result)
	}
}

func TestParseFuturesOrders(t *testing.T) {
	raw := [][]float64{
		{100.0, 10.0, 1.0}, // price, orders, quantity
		{101.0, 5.0, 0.5},
		{102.0}, // too short
	}

	entries := parseFuturesOrders(raw)
	assert.Len(t, entries, 2)
	assert.Equal(t, 100.0, entries[0].Price)
	assert.Equal(t, 1.0, entries[0].Quantity)
	assert.Equal(t, 101.0, entries[1].Price)
	assert.Equal(t, 0.5, entries[1].Quantity)
}

func TestParseSpotOrders(t *testing.T) {
	raw := [][]string{
		{"100.0", "1.0"},
		{"101.0", "0.5"},
	}

	entries := parseSpotOrders(raw)
	assert.Len(t, entries, 2)
	assert.Equal(t, 100.0, entries[0].Price)
	assert.Equal(t, 1.0, entries[0].Quantity)
}

func TestSpotCandleParsing(t *testing.T) {
	now := time.Now().UnixMilli()
	rawCandles := [][]interface{}{
		{float64(now), "100.0", "105.0", "95.0", "102.0", "10.0"},
		{float64(now + 60000), 102.0, 110.0, 100.0, 108.0, 5.0}, // mixed types
		{float64(now + 120000)},                                 // too short
	}

	candles := parseSpotCandles(rawCandles)
	assert.Len(t, candles, 2)

	assert.Equal(t, time.UnixMilli(now).UTC(), candles[0].OpenTime.UTC())
	assert.Equal(t, 100.0, candles[0].Open)
	assert.Equal(t, 10.0, *candles[0].Volume)

	assert.Equal(t, time.UnixMilli(now+60000).UTC(), candles[1].OpenTime.UTC())
	assert.Equal(t, 102.0, candles[1].Open)
	assert.Equal(t, 5.0, *candles[1].Volume)
}
