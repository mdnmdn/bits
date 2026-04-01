package process

import (
	"time"

	"github.com/mdnmdn/bits/model"
)

func TimeEnricher(res model.Response[model.ServerTime]) model.Response[model.ServerTime] {
	now := time.Now()
	latency := now.Sub(res.Data.Time)
	skew := latency / 2
	res.Data.LocalTime = &now
	res.Data.Latency = &latency
	res.Data.ClockSkew = &skew
	return res
}
