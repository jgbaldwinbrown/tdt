package tdt

import (
	"github.com/montanaflynn/stats"
)

func Zscores(fs stats.Float64Data) ([]float64, error) {
	mean, e := stats.Mean(fs)
	if e != nil {
		return nil, e
	}
	sd, e := stats.StandardDeviation(fs)
	if e != nil {
		return nil, e
	}
	out := make([]float64, 0, len(fs))
	for _, f := range fs {
		out = append(out, (f-mean)/sd)
	}
	return out, nil
}
