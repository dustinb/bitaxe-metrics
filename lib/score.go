package lib

import (
	"math"
)

func TemperatureScore(info Info) float64 {
	score := 65 / info.Temp
	return score
}

func HashRateScore(info Info) float64 {
	score := info.HashRate / info.ExpectedHashRate
	return math.Max(0, score)
}

func EfficiencyScore(info Info) float64 {
	var efficiency = info.Power / (info.HashRate / 1000)
	var expectedEfficiency = info.Power / (info.ExpectedHashRate / 1000)
	score := expectedEfficiency / efficiency
	return score
}
