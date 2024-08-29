package tdt

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jgbaldwinbrown/csvh"
	"golang.org/x/exp/rand"
	"golang.org/x/exp/slices"
	"gonum.org/v1/gonum/stat/distuv"
	"io"
	"log"
	"math"
)

func Perm1(r rand.Source, totals []float64) []Family {
	out := make([]Family, 0, len(totals))
	for _, tot := range totals {
		b := distuv.Binomial{N: tot, P: 0.5, Src: r}
		males := b.Rand()
		females := tot - males
		out = append(out, Family{males, females})
	}
	return out
}

func Perm(r rand.Source, nperms int, totals []float64) [][]Family {
	out := make([][]Family, 0, nperms)
	for i := 0; i < nperms; i++ {
		out = append(out, Perm1(r, totals))
	}
	return out
}

func TDTMultipleFamilies(fams []Family) []TDTResult {
	out := make([]TDTResult, 0, len(fams))
	for _, fam := range fams {
		out = append(out, TDTTest(fam))
	}
	return out
}

func TDTReplicateFamilySets(famsets [][]Family) [][]TDTResult {
	out := make([][]TDTResult, 0, len(famsets))
	for _, famset := range famsets {
		out = append(out, TDTMultipleFamilies(famset))
	}
	return out
}

func MostSignificant(actual TDTResult, background []TDTResult) bool {
	for _, res := range background {
		if actual.P > res.P {
			return false
		}
	}
	return true
}

func MostSignificantPercentage(actual TDTResult, background [][]TDTResult) float64 {
	goods := 0
	for _, res := range background {
		if MostSignificant(actual, res) {
			goods++
		}
	}
	return float64(goods) / float64(len(background))
}

func TopSignificant(p float64, actual TDTResult, background []TDTResult) bool {
	bg := make([]float64, 0, len(background))
	for _, res := range background {
		bg = append(bg, res.P)
	}
	slices.SortFunc(bg, func(a, b float64) bool {
		return a < b
	})
	p5 := int(float64(len(bg)) * p)
	return actual.P < bg[p5]
}

func TopSignificantPercentage(p float64, actual TDTResult, background [][]TDTResult) float64 {
	goods := 0
	for _, res := range background {
		if TopSignificant(p, actual, res) {
			goods++
		}
	}
	return float64(goods) / float64(len(background))
}

func ReadResults(r io.Reader) ([]TDTResult, error) {
	dec := json.NewDecoder(r)
	var out []TDTResult
	for {
		var resj TDTResultJson
		err := dec.Decode(&resj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		res := FromJson(resj)
		out = append(out, res)
	}
	return out, nil
}

func ReadPathResults(path string) ([]TDTResult, error) {
	r, e := csvh.OpenMaybeGz(path)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	return ReadResults(r)
}

type MonteArgs struct {
	Actual     string
	Background string
	Seed       int
	Replicates int
}

func NoZeroes(rs []TDTResult) []TDTResult {
	var out []TDTResult
	for _, r := range rs {
		if r.Totals.MaleF1+r.Totals.FemaleF1 < 1.0 {
			continue
		}
		if math.IsInf(r.P, 0) {
			continue
		}
		if math.IsNaN(r.P) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func FullMonte() {
	var f MonteArgs
	flag.StringVar(&f.Actual, "a", "", "path to .json containing actual family results")
	flag.StringVar(&f.Background, "b", "", "path to .json containing background families")
	flag.IntVar(&f.Seed, "s", 0, "Random seed")
	flag.IntVar(&f.Replicates, "r", 1, "Replicates")
	flag.Parse()
	if f.Actual == "" {
		log.Fatal(fmt.Errorf("missing -a"))
	}
	if f.Background == "" {
		log.Fatal(fmt.Errorf("missing -b"))
	}

	actualSlice, e := ReadPathResults(f.Actual)
	if e != nil {
		log.Fatal(e)
	}
	actual := actualSlice[0]

	bg, e := ReadPathResults(f.Background)
	if e != nil {
		log.Fatal(e)
	}
	bg = NoZeroes(bg)

	tots := make([]float64, 0, len(bg))
	for _, bg1 := range bg {
		tots = append(tots, float64(bg1.Totals.MaleF1+bg1.Totals.FemaleF1))
	}

	rsrc := rand.NewSource(uint64(f.Seed))
	perms := Perm(rsrc, f.Replicates, tots)

	results := TDTReplicateFamilySets(perms)
	perc := MostSignificantPercentage(actual, results)
	fmt.Println(perc)
	perc2 := TopSignificantPercentage(0.05, actual, results)
	fmt.Println(perc2)
	perc3 := TopSignificantPercentage(0.01, actual, results)
	fmt.Println(perc3)
	perc4 := TopSignificantPercentage(0.001, actual, results)
	fmt.Println(perc4)
	perc5 := TopSignificantPercentage(0.0001, actual, results)
	fmt.Println(perc5)

	better := 0
	count := 0
	for _, set := range results {
		for _, res := range set {
			if actual.P < res.P {
				better++
			}
			count++
		}
	}
	fmt.Println(float64(better) / float64(count))
}
