package display

import (
	"fmt"
	"math"
	"strings"
)

func FormatPrice(price float64) string {
	if price == 0 {
		return "$0.00"
	}
	abs := math.Abs(price)
	sign := ""
	if price < 0 {
		sign = "-"
	}

	switch {
	case abs >= 1:
		return sign + "$" + formatWithCommas(abs, 2)
	case abs >= 0.01:
		return fmt.Sprintf("%s$%.4f", sign, abs)
	default:
		return fmt.Sprintf("%s$%.8f", sign, abs)
	}
}

func FormatPercent(pct float64) string {
	return fmt.Sprintf("%.2f%%", pct)
}

func FormatLargeNumber(n float64) string {
	abs := math.Abs(n)
	sign := ""
	if n < 0 {
		sign = "-"
	}

	switch {
	case abs >= 1e12:
		return fmt.Sprintf("%s$%.2fT", sign, abs/1e12)
	case abs >= 1e9:
		return fmt.Sprintf("%s$%.2fB", sign, abs/1e9)
	case abs >= 1e6:
		return fmt.Sprintf("%s$%.2fM", sign, abs/1e6)
	case abs >= 1e3:
		return fmt.Sprintf("%s$%.2fK", sign, abs/1e3)
	default:
		return sign + "$" + formatWithCommas(abs, 2)
	}
}

func FormatSupply(n float64) string {
	abs := math.Abs(n)
	switch {
	case abs >= 1e12:
		return fmt.Sprintf("%.2fT", abs/1e12)
	case abs >= 1e9:
		return fmt.Sprintf("%.2fB", abs/1e9)
	case abs >= 1e6:
		return fmt.Sprintf("%.2fM", abs/1e6)
	default:
		return formatWithCommas(abs, 0)
	}
}

func formatWithCommas(n float64, decimals int) string {
	s := fmt.Sprintf("%.*f", decimals, n)
	parts := strings.Split(s, ".")
	intPart := parts[0]

	var result []byte
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}

	if len(parts) > 1 {
		return string(result) + "." + parts[1]
	}
	return string(result)
}
