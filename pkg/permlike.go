package tdt

import (
	"flag"
	"fmt"
	"github.com/jgbaldwinbrown/iterh"
	"github.com/montanaflynn/stats"
	"iter"
	"log"
	"slices"
)

func Means(it iter.Seq[[]float64]) iter.Seq2[float64, error] {
	return func(y func(float64, error) bool) {
		for fs := range it {
			m, e := stats.Mean(fs)
			if !y(m, e) {
				return
			}
		}
	}
}

func Quantile(fs []float64, perc float64) float64 {
	sorted := slices.Sorted(slices.Values(fs))
	quant := int(perc * float64(len(fs)))
	return sorted[quant]
}

func PosteriorSlices(bufsiz int, it iter.Seq[iter.Seq2[Entry, error]]) iter.Seq2[[]float64, error] {
	return func(y func([]float64, error) bool) {
		for ped := range it {
			var posteriors []float64
			if bufsiz > 0 {
				posteriors = make([]float64, 0, bufsiz)
			}
			for ent, e := range ped {
				if e != nil {
					if !y(nil, e) {
						return
					}
				}
				posteriors = append(posteriors, ent.Posterior)
			}
			if !y(posteriors, nil) {
				return
			}
		}
	}
}

func FlattenEntries(itit iter.Seq[iter.Seq2[Entry, error]], ep *error) iter.Seq[Entry] {
	return func(y func(Entry) bool) {
		for it := range itit {
			for ent, e := range it {
				if e != nil {
					*ep = e
					return
				}
				if !y(ent) {
					return
				}
			}
		}
	}
}

func RunCountWARPHits() {
	var f Flags
	flag.StringVar(&f.RealPath, "r", "", "Path to output of warp for real data")
	flag.StringVar(&f.BgPathsPath, "b", "", "Path to list of paths containing warp output for background data")
	flag.BoolVar(&f.RealHeader, "rh", false, "Real data has a header line")
	flag.BoolVar(&f.BgHeader, "bh", false, "Background data has a header line")

	flag.Parse()
	if f.RealPath == "" {
		log.Fatal("missing -r")
	}
	if f.BgPathsPath == "" {
		log.Fatal("missing -b")
	}

	realEntries, e := iterh.CollectWithError(ParsePedPath(f.RealPath, f.RealHeader))
	if e != nil {
		log.Fatal(e)
	}

	bgPaths := slices.Collect(iterh.BreakOnError(iterh.PathIter(f.BgPathsPath, iterh.LineIter), &e))
	if e != nil {
		log.Fatal(e)
	}

	realLikelihoods := slices.Collect(Posteriors(slices.Values(realEntries)))
	realMeanLikelihood, e := stats.Mean(realLikelihoods)
	if e != nil {
		log.Fatal(e)
	}

	bufsiz := int(float64(len(realEntries)) * 1.25)
	bgPosteriorSlices, e := iterh.CollectWithError(PosteriorSlices(bufsiz, ParsePedPaths(f.BgHeader, bgPaths...)))
	if e != nil {
		log.Fatal(e)
	}
	bgPosteriorMeans, e := iterh.CollectWithError(Means(slices.Values(bgPosteriorSlices)))
	if e != nil {
		log.Fatal(e)
	}

	significanceThresh := Quantile(bgPosteriorMeans, 0.95)
	fmt.Printf("likelihood significance threshold: %v\n", significanceThresh)
	fmt.Printf("actual mean likelihood: %v\n", realMeanLikelihood)
	fmt.Printf("significant: %v\n", realMeanLikelihood > significanceThresh)

	// var sigs []Entry
	// for _, entry := range realEntries {
	// 	if entry.Posterior > significanceThresh {
	// 		sigs = append(sigs, entry)
	// 	}
	// }
	// fmt.Printf("num significant: %v\n", len(sigs))
	// fmt.Printf("significant individuals:\n")
	// for _, entry := range sigs {
	// 	fmt.Printf("%#v", entry)
	// }
}

func RunCountWARPHits2() {
	var f Flags
	flag.StringVar(&f.RealPath, "r", "", "Path to output of warp for real data")
	flag.StringVar(&f.BgPathsPath, "b", "", "Path to list of paths containing warp output for background data")
	flag.BoolVar(&f.RealHeader, "rh", false, "Real data has a header line")
	flag.BoolVar(&f.BgHeader, "bh", false, "Background data has a header line")

	flag.Parse()
	if f.RealPath == "" {
		log.Fatal("missing -r")
	}
	if f.BgPathsPath == "" {
		log.Fatal("missing -b")
	}

	realEntries, e := iterh.CollectWithError(ParsePedPath(f.RealPath, f.RealHeader))
	if e != nil {
		log.Fatal(e)
	}

	bgPaths := slices.Collect(iterh.BreakOnError(iterh.PathIter(f.BgPathsPath, iterh.LineIter), &e))
	if e != nil {
		log.Fatal(e)
	}

	bgHighestPosterior, e := iterh.CollectWithError(GetBiggestOutlierPaths(slices.Values(bgPaths), f.BgHeader))
	if e != nil {
		log.Fatal(e)
	}

	significanceThresh := Quantile(slices.Collect(Posteriors(slices.Values(bgHighestPosterior))), 0.95)
	fmt.Printf("likelihood significance threshold: %v\n", significanceThresh)

	var sigs []Entry
	for _, entry := range realEntries {
		if entry.Posterior > significanceThresh {
			sigs = append(sigs, entry)
		}
	}
	fmt.Printf("num significant: %v\n", len(sigs))
	fmt.Printf("significant individuals:\n")
	for _, entry := range sigs {
		fmt.Printf("%#v", entry)
	}
}

func RunCountWARPHits3() {
	var f Flags
	flag.StringVar(&f.RealPath, "r", "", "Path to output of warp for real data")
	flag.StringVar(&f.BgPathsPath, "b", "", "Path to list of paths containing warp output for background data")
	flag.BoolVar(&f.RealHeader, "rh", false, "Real data has a header line")
	flag.BoolVar(&f.BgHeader, "bh", false, "Background data has a header line")

	flag.Parse()
	if f.RealPath == "" {
		log.Fatal("missing -r")
	}
	if f.BgPathsPath == "" {
		log.Fatal("missing -b")
	}

	realEntries, e := iterh.CollectWithError(ParsePedPath(f.RealPath, f.RealHeader))
	if e != nil {
		log.Fatal(e)
	}

	bgPaths := slices.Collect(iterh.BreakOnError(iterh.PathIter(f.BgPathsPath, iterh.LineIter), &e))
	if e != nil {
		log.Fatal(e)
	}

	bgEntries := FlattenEntries(ParsePedPaths(f.BgHeader, bgPaths...), &e)
	significanceThresh := Quantile(slices.Collect(Posteriors(bgEntries)), 0.95)
	if e != nil {
		log.Fatal(e)
	}
	fmt.Printf("likelihood significance threshold: %v\n", significanceThresh)

	var sigs []Entry
	for _, entry := range realEntries {
		if entry.Posterior > significanceThresh {
			sigs = append(sigs, entry)
		}
	}
	fmt.Printf("num significant: %v\n", len(sigs))
	fmt.Printf("significant individuals:\n")
	for _, entry := range sigs {
		fmt.Printf("%#v", entry)
	}
}
