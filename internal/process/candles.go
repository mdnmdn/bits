package process

import "github.com/mdnmdn/bits/pkg/model"

// CandleStats computes VWAP, typical price, and body/wick ratios into each Candle.Extra.
func CandleStats(res model.Response[[]model.Candle]) model.Response[[]model.Candle] {
	var totalVolume, vwapNumerator float64
	for i, c := range res.Data {
		typical := (c.High + c.Low + c.Close) / 3
		bodySize := abs(c.Close - c.Open)
		totalRange := c.High - c.Low
		var bodyRatio, wickRatio float64
		if totalRange > 0 {
			bodyRatio = bodySize / totalRange
			wickRatio = 1 - bodyRatio
		}
		if c.Extra == nil {
			res.Data[i].Extra = make(map[string]any)
		}
		res.Data[i].Extra["typical_price"] = typical
		res.Data[i].Extra["body_ratio"] = bodyRatio
		res.Data[i].Extra["wick_ratio"] = wickRatio

		if c.Volume != nil {
			vwapNumerator += typical * *c.Volume
			totalVolume += *c.Volume
		}
	}

	if totalVolume > 0 {
		vwap := vwapNumerator / totalVolume
		// Attach VWAP to the last candle's Extra as a summary.
		if len(res.Data) > 0 {
			last := len(res.Data) - 1
			if res.Data[last].Extra == nil {
				res.Data[last].Extra = make(map[string]any)
			}
			res.Data[last].Extra["vwap"] = vwap
		}
	}
	return res
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
