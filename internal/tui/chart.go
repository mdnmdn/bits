package tui

import (
	"fmt"
	"strings"

	"github.com/coingecko/coingecko-cli/internal/model"
	"github.com/coingecko/coingecko-cli/internal/display"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/graph"
	"github.com/charmbracelet/lipgloss"
)

func renderBrailleChart(ohlc model.OHLCData, width, height int, vs string) string {
	if len(ohlc) == 0 || width < 10 || height < 5 {
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

	// Y-axis labels: high, mid, low
	midP := (minP + maxP) / 2
	yHigh := display.FormatPrice(maxP, vs)
	yMid := display.FormatPrice(midP, vs)
	yLow := display.FormatPrice(minP, vs)

	// Find the widest Y label for padding (display width, not byte length).
	yWidth := max(lipgloss.Width(yHigh), lipgloss.Width(yMid), lipgloss.Width(yLow)) + 1

	// Chart area dimensions (subtract Y-axis width and X-axis row)
	chartW := width - yWidth
	chartH := height - 2 // leave room for x-axis label row
	if chartW < 4 {
		chartW = 4
	}
	if chartH < 3 {
		chartH = 3
	}

	bg := graph.NewBrailleGrid(chartW, chartH,
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

	// Color the chart line based on 7-day performance.
	var chartStyle lipgloss.Style
	if prices[len(prices)-1] >= prices[0] {
		chartStyle = GreenStyle
	} else {
		chartStyle = RedStyle
	}

	// Pre-compute styled Y-axis labels (right-aligned to yWidth).
	pad := strings.Repeat(" ", yWidth)
	styledHigh := DimStyle.Render(fmt.Sprintf("%*s", yWidth, yHigh))
	styledMid := DimStyle.Render(fmt.Sprintf("%*s", yWidth, yMid))
	styledLow := DimStyle.Render(fmt.Sprintf("%*s", yWidth, yLow))

	var b strings.Builder
	for i, row := range patterns {
		switch {
		case i == 0:
			b.WriteString(styledHigh)
		case i == len(patterns)/2:
			b.WriteString(styledMid)
		case i == len(patterns)-1:
			b.WriteString(styledLow)
		default:
			b.WriteString(pad)
		}
		b.WriteString(chartStyle.Render(string(row)))
		b.WriteString("\n")
	}

	// X-axis labels: Day 1, Day 4, Day 7
	xLeft := "Day 1"
	xMid := "Day 4"
	xRight := "Day 7"
	gap := chartW - len(xLeft) - len(xMid) - len(xRight)
	if gap < 2 {
		gap = 2
	}
	leftGap := gap / 2
	rightGap := gap - leftGap
	xAxis := strings.Repeat(" ", yWidth) +
		xLeft + strings.Repeat(" ", leftGap) +
		xMid + strings.Repeat(" ", rightGap) +
		xRight
	b.WriteString(DimStyle.Render(xAxis))

	return b.String()
}
