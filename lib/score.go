package lib

import (
	"math"
)

type Average struct {
	Score      float64
	Efficiency float64
	HashRate   float64
	Temp       float64
	Count      float64
}

func TemperatureScore(info Info) float64 {
	score := 65 / info.Temp
	return score
}

func HashRateScore(info Info) float64 {
	var expectedHashRate = ExpectedHashRate(info)
	score := info.HashRate / expectedHashRate
	return math.Max(0, score)
}

func EfficiencyScore(info Info) float64 {
	var efficiency = info.Power / (info.HashRate / 1000)
	var expectedEfficiency = info.Power / (ExpectedHashRate(info) / 1000)
	score := expectedEfficiency / efficiency
	return score
}

func ExpectedHashRate(info Info) float64 {
	return math.Floor(float64(info.Frequency) * ((float64(info.SmallCoreCount) * float64(info.AsicCount)) / 1000.0))
}
