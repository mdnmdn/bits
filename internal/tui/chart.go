package tui

import (
	"fmt"
	"strings"

	"github.com/coingecko/coingecko-cli/internal/api"
	"github.com/coingecko/coingecko-cli/internal/display"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/graph"
)

func renderBrailleChart(ohlc api.OHLCData, width, height int) string {
	if len(ohlc) == 0 || width < 4 || height < 3 {
		return "No data"
	}

	// Extract close prices (index 4)
	prices := make([]float64, 0, len(ohlc))
	for _, d := range ohlc {
		if len(d) >= 5 {
			prices = append(prices, d[4])
		}
	}
	if len(prices) == 0 {
		return "No data"
	}

	minP, maxP := prices[0], prices[0]
	for _, p := range prices {
		if p < minP {
			minP = p
		}
		if p > maxP {
			maxP = p
		}
	}

	if maxP == minP {
		maxP = minP + 1
	}

	bg := graph.NewBrailleGrid(width, height,
		0, float64(len(prices)-1),
		minP, maxP,
	)

	for i := 1; i < len(prices); i++ {
		p1 := canvas.Float64Point{X: float64(i - 1), Y: prices[i-1]}
		p2 := canvas.Float64Point{X: float64(i), Y: prices[i]}
		gp1 := bg.GridPoint(p1)
		gp2 := bg.GridPoint(p2)

		points := graph.GetLinePoints(gp1, gp2)
		for _, pt := range points {
			bg.Set(pt)
		}
	}

	patterns := bg.BraillePatterns()
	var lines []string
	for _, row := range patterns {
		lines = append(lines, string(row))
	}

	chart := strings.Join(lines, "\n")

	// Add price labels
	chart += fmt.Sprintf("\n%s  High: %s  Low: %s",
		DimStyle.Render(""),
		display.FormatPrice(maxP),
		display.FormatPrice(minP),
	)

	return chart
}
