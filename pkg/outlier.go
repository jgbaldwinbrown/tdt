package tdt

import (
	"log"
	"errors"
	"fmt"
	"github.com/jgbaldwinbrown/csvh"
	"github.com/jgbaldwinbrown/iterh"
	"flag"
	"io"
	"iter"
	"github.com/montanaflynn/stats"
)

type Entry struct {
	FamilyID string
	IndividualID string
	FatherID string
	MotherID string
	Sex string
	Phenotype string
	Prior float64
	Posterior float64
	PhenoRisk float64
	GenoRisk string
}

var ParseError = errors.New("entry parsing error")

func ParseLineToEntry(line []string) (Entry, error) {
	if len(line) != 11 {
		return Entry{}, ParseError
	}
	var gap string
	var ent Entry
	_, e := csvh.Scan(line,
		&ent.FamilyID,
		&ent.IndividualID,
		&ent.FatherID,
		&ent.MotherID,
		&ent.Sex,
		&ent.Phenotype,
		&gap,
		&ent.Prior,
		&ent.Posterior,
		&ent.PhenoRisk,
		&ent.GenoRisk,
	)
	return ent, e
}

func ParsePed(r io.Reader, header bool) iter.Seq2[Entry, error] {
	return func(y func(Entry, error) bool) {
		h := csvh.Handle0("ParsePed: %w")
		hl := func(e error, l []string) error {
			return fmt.Errorf("ParsePed: line %v; %w", l, e)
		}
		cr := csvh.CsvIn(r)

		if header {
			_, e := cr.Read()
			if e != nil {
				if !y(Entry{}, h(e)) {
					return
				}
			}
		}

		for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
			if e != nil {
				if !y(Entry{}, hl(e, l)) {
					return
				}
			}
			ent, e := ParseLineToEntry(l)
			if e != nil {
				if !y(ent, hl(e, l)) {
					return
				}
			}
			if !y(ent, nil) {
				return
			}
		}
	}
}

func ParsePedPath(path string, header bool) iter.Seq2[Entry, error] {
	return func(y func(Entry, error) bool) {
		r, e := csvh.OpenMaybeGz(path)
		if e != nil {
			if !y(Entry{}, e) {
				return
			}
		}
		defer r.Close()
		it := ParsePed(r, header)
		for ent, e := range it {
			if !y(ent, e) {
				return
			}
		}
	}
}

func ParsePedPaths(header bool, paths ...string) iter.Seq[iter.Seq2[Entry, error]] {
	return func(y func(iter.Seq2[Entry, error]) bool) {
		for _, path := range paths {
			if !y(ParsePedPath(path, header)) {
				return
			}
		}
	}
}

func GetBiggestOutlier(it iter.Seq[Entry]) Entry {
	var best Entry
	started := false
	for ent := range it {
		if !started || best.Posterior < ent.Posterior {
			best = ent
			started = true
		}
	}
	return best
}

func GetBiggestOutlierPath(path string, header bool) (Entry, error) {
	r, e := csvh.OpenMaybeGz(path)
	if e != nil {
		return Entry{}, e
	}
	defer r.Close()

	it := iterh.BreakOnError(ParsePed(r, header), &e)
	best := GetBiggestOutlier(it)
	return best, e
}

func GetBiggestOutlierPaths(paths iter.Seq[string], header bool) iter.Seq2[Entry, error] {
	return func(y func(Entry, error) bool) {
		for path := range paths {
			bgOutlier, e := GetBiggestOutlierPath(path, header)
			if !y(bgOutlier, e) {
				return
			}
		}
	}
}

func BiggestOutlierPerc(realOutlier Entry, bgOutliers iter.Seq[Entry]) float64 {
	totalCount := 0
	biggerCount := 0
	for bgOutlier := range bgOutliers {
		totalCount++
		if bgOutlier.Posterior > realOutlier.Posterior {
			biggerCount++
		}
	}
	return float64(biggerCount) / float64(totalCount)
}

func GetBiggestOutlierAvg(bgOutliers iter.Seq[Entry]) float64 {
	sum := 0.0
	count := 0.0
	for entry := range bgOutliers {
		sum += entry.Posterior
		count++
	}
	return sum / count
}

func GetZscores(ents []Entry) ([]float64, error) {
	fs := make([]float64, 0, len(ents))
	for _, ent := range ents {
		fs = append(fs, ent.Posterior)
	}
	return Zscores(fs)
}

func GetAllBestZscores(its iter.Seq[iter.Seq2[Entry, error]]) ([]float64, error) {
	var bests []float64
	for it := range its {
		ents, e := iterh.CollectWithError(it)
		if e != nil {
			return nil, e
		}
		zs, e := GetZscores(ents)
		if e != nil {
			return nil, e
		}
		best := iterh.Max(iterh.SliceIter(zs))
		bests = append(bests, best)
	}
	return bests, nil
}

func GetIDPosterior(id string, ents iter.Seq[Entry]) (float64, bool) {
	for ent := range ents {
		if ent.IndividualID == id {
			return ent.Posterior, true
		}
	}
	return 0, false
}

func GetAllIDPosteriors(id string, its iter.Seq[iter.Seq2[Entry, error]]) (vals []float64, nmissing int, err error) {
	for it := range its {
		val, ok := GetIDPosterior(id, iterh.BreakOnError(it, &err))
		if err != nil {
			return nil, 0, err
		}
		if ok {
			vals = append(vals, val)
		} else {
			nmissing++
		}
	}
	return vals, nmissing, nil
}

func RankStats(id string, realPed iter.Seq[Entry], bgPeds iter.Seq[iter.Seq2[Entry, error]]) (chosenRank, chosenInternalRank float64, bgRanks []float64, err error) {
	idx, realIDVal := iterh.IndexFunc(realPed, func(ent Entry) bool {
		return ent.IndividualID == id
	})
	if idx == -1 {
		return 0, 0, nil, fmt.Errorf("could not find id %v in realPed", id)
	}

	chosenInternalRank, _, _ = iterh.Rank(realIDVal.Posterior, iterh.Transform(realPed, func(ent Entry) float64 {
		return ent.Posterior
	}))

	bgIDVals, nmissing, err := GetAllIDPosteriors(id, bgPeds)
	if err != nil {
		return 0, 0, nil, err
	}
	if nmissing > 0 {
		return 0, 0, nil, fmt.Errorf("RankStats: nmissing %v", nmissing)
	}
	chosenRank, _, _ = iterh.Rank(realIDVal.Posterior, iterh.SliceIter(bgIDVals))

	for i, bgPed := range iterh.Enumerate(bgPeds) {
		bgRank, _, _ := iterh.Rank(bgIDVals[i], iterh.Transform(iterh.BreakOnError(bgPed, &err), func(ent Entry) float64 {
			return ent.Posterior
		}))
		if err != nil {
			return 0, 0, nil, err
		}
		bgRanks = append(bgRanks, bgRank)
	}

	return chosenRank, chosenInternalRank, bgRanks, nil
}

type Flags struct {
	RealPath string
	RealHeader bool
	BgPathsPath string
	BgHeader bool
	Chosen string
}

func RunOutlier() {
	var f Flags
	flag.StringVar(&f.RealPath, "r", "", "Path to output of warp for real data")
	flag.StringVar(&f.BgPathsPath, "b", "", "Path to list of paths containing warp output for background data")
	flag.StringVar(&f.Chosen, "c", "", "Chosen individual ID to run rank order statistics on")
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

	realOutlier := GetBiggestOutlier(iterh.SliceIter(realEntries))

	bgPaths := iterh.Collect(iterh.BreakOnError(iterh.PathIter(f.BgPathsPath, iterh.LineIter), &e))
	if e != nil {
		log.Fatal(e)
	}

	bgOutliersSeq := iterh.BreakOnError(GetBiggestOutlierPaths(iterh.SliceIter[[]string](bgPaths), f.BgHeader), &e)
	bgOutliers := iterh.Collect(bgOutliersSeq)
	if e != nil {
		log.Fatal(e)
	}

	frac := BiggestOutlierPerc(realOutlier, iterh.SliceIter(bgOutliers))
	fmt.Println("biggest outlier percentage:", frac)

	bgAvg := GetBiggestOutlierAvg(iterh.SliceIter(bgOutliers))
	fmt.Println("backgrount biggest average:", bgAvg)


	if f.Chosen != "" {
		chosenRank, chosenInternalRank, bgRanks, e := RankStats(f.Chosen, iterh.SliceIter(realEntries), ParsePedPaths(f.BgHeader, bgPaths...))
		if e != nil {
			log.Fatal(e)
		}
		meanBgRank, e := stats.Mean(bgRanks)
		if e != nil {
			log.Fatal(e)
		}
		fmt.Printf("chosenRank %v; chosenInternalRank %v; meanBgRank %v\n", chosenRank, chosenInternalRank, meanBgRank)
	}

	realZs, e := GetZscores(realEntries)
	if e != nil {
		log.Fatal(e)
	}
	realHighestZ := iterh.Max(iterh.SliceIter(realZs))
	bgZScores, e := GetAllBestZscores(ParsePedPaths(f.BgHeader, bgPaths...))
	if e != nil {
		log.Fatal(e)
	}

	zrankperc, zhigher, ztotal := iterh.Rank(realHighestZ, iterh.SliceIter(bgZScores))
	fmt.Printf("realHighestZ %v; zrankperc %v; zhigher %v; ztotal %v\n", realHighestZ, zrankperc, zhigher, ztotal)
}

// family - Family ID
// ind - Individual ID
// father - Father ID
// mother - Mother ID
// sex - Individual sex
// phenotype - provided phenotype status
// prior - provided prior probability
// posterior - calculated posterior probability
// pheno_risk - calculated phenotype posterior probability
// geno_risk - calculated genotype posterior probability
