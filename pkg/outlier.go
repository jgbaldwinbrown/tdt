package tdt

import (
	"bufio"
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/jgbaldwinbrown/csvh"
	"github.com/jgbaldwinbrown/iterh"
	stat "github.com/jgbaldwinbrown/perf/pkg/stats"
	"github.com/montanaflynn/stats"
	"golang.org/x/sync/errgroup"
)

// An entry from WARP's output, an extended .ped file
type Entry struct {
	FamilyID     string
	IndividualID string
	FatherID     string
	MotherID     string
	Sex          string
	Phenotype    string
	Prior        float64
	Posterior    float64
	PhenoRisk    float64
	GenoRisk     string
}

var ParseError = errors.New("entry parsing error")

// Parse a WARP output line to an Entry
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

// Parse a WARP output .ped reader to a sequence of entries
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

// Parse WARP output .ped or .ped.gz file
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

// Parse many WARP output files sequentially
func ParsePedPaths(header bool, paths ...string) iter.Seq[iter.Seq2[Entry, error]] {
	return func(y func(iter.Seq2[Entry, error]) bool) {
		for _, path := range paths {
			if !y(ParsePedPath(path, header)) {
				return
			}
		}
	}
}

// Get the single most significant individual
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

// Get the nop n most significant individuals
func GetBiggestOutliers(it iter.Seq[Entry], n int) []Entry {
	all := slices.SortedFunc(it, iterh.Negative(func(a, b Entry) int {
		if a.Posterior < b.Posterior {
			return -1
		} else if a.Posterior > b.Posterior {
			return 1
		}
		return 0
	}))
	if len(all) < n {
		return all
	}
	return all[:n]
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

func GetBiggestOutliersPath(path string, header bool, n int) ([]Entry, error) {
	r, e := csvh.OpenMaybeGz(path)
	if e != nil {
		return nil, e
	}
	defer r.Close()

	it := iterh.BreakOnError(ParsePed(r, header), &e)
	best := GetBiggestOutliers(it, n)
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

func GetBiggestOutliersPaths(paths iter.Seq[string], header bool, n int) iter.Seq2[[]Entry, error] {
	return func(y func([]Entry, error) bool) {
		for path := range paths {
			bgOutlier, e := GetBiggestOutliersPath(path, header, n)
			if !y(bgOutlier, e) {
				return
			}
		}
	}
}

// The percentage of entries in bgOutliers more significant than realOutlier
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

// Average of all posteriors in the slice for each slice
func PosteriorMeans(it iter.Seq[[]Entry]) iter.Seq2[float64, error] {
	return func(y func(float64, error) bool) {
		for ents := range it {
			m, e := stats.Mean(slices.Collect(Posteriors(slices.Values(ents))))
			if !y(m, e) {
				return
			}
		}
	}
}

// Mean of posteriors
func GetBiggestOutlierAvg(bgOutliers iter.Seq[Entry]) float64 {
	sum := 0.0
	count := 0.0
	for entry := range bgOutliers {
		sum += entry.Posterior
		count++
	}
	return sum / count
}

// Extract and normalize all posteriors
func GetZscores(ents []Entry) ([]float64, error) {
	fs := make([]float64, 0, len(ents))
	for _, ent := range ents {
		fs = append(fs, ent.Posterior)
	}
	return Zscores(fs)
}

// Extract and normalize all posteriors for each of the entry sets, then get the highest z score for each one
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

// Like GetAllBestZscores, but get the top n
func GetAllBestZscoresTopN(its iter.Seq[iter.Seq2[Entry, error]], n int) ([][]float64, error) {
	var bests [][]float64
	for it := range its {
		ents, e := iterh.CollectWithError(it)
		if e != nil {
			return nil, e
		}
		zs, e := GetZscores(ents)
		if e != nil {
			return nil, e
		}
		best := TopN(iterh.SliceIter(zs), max(n, 1))
		bests = append(bests, best)
	}
	return bests, nil
}

// Get the posterior of a specific id and return if it was found
func GetIDPosterior(id string, ents iter.Seq[Entry]) (float64, bool) {
	for ent := range ents {
		if ent.IndividualID == id {
			return ent.Posterior, true
		}
	}
	return 0, false
}

// Get the posterior of a specific ID for each set; couns number of times it was missed.
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

// Calculate the rank of the chosen ID compared to itself in the background and compared to all other IDs in the real pedigree
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

// Get posteriors from entries
func Posteriors(it iter.Seq[Entry]) iter.Seq[float64] {
	return iterh.Transform(it, func(ent Entry) float64 {
		return ent.Posterior
	})
}

// Get the top n of anything
func TopNFunc[T any](it iter.Seq[T], cmpf func(T, T) int, n int) []T {
	sorted := slices.SortedFunc(it, cmpf)
	if len(sorted) < n {
		return sorted
	}
	return sorted[:n]
}

// Get the top n of anything ordered
func TopN[T cmp.Ordered](it iter.Seq[T], n int) []T {
	return TopNFunc(it, iterh.Negative[T](cmp.Compare), n)
}

type Flags struct {
	RealPath    string
	RealHeader  bool
	BgPathsPath string
	BgHeader    bool
	Chosen      string
	TopN        int
	BgPaths     []string
}

// bgZScoresMeans, e := GetBgZScoresMeans(bgZScores)

// For each set of scores, get the average z score
func GetBgZScoresMeans(bgZScores [][]float64) ([]float64, error) {
	out := make([]float64, 0, len(bgZScores))
	for _, set := range bgZScores {
		m, e := stats.Mean(set)
		if e != nil {
			return nil, e
		}
		out = append(out, m)
	}
	return out, nil
}

type OutlierStats struct {
	BiggestOutlierPercentage any
	BackgroundBiggestAverage any
	ChosenRank               any
	ChosenInternalRank       any
	MeanBgRank               any
	TTestResult              *stat.TTestResult
	RealHighestZ             any
	ZRankPercentile          any
	ZHigher                  int
	ZTotal                   int
}

func CleanFloat64(x *any) {
	switch v := (*x).(type) {
	case float64:
		if math.IsNaN(v) {
			*x = "NaN"
		} else if math.IsInf(v, 1) {
			*x = "Inf"
		} else if math.IsInf(v, -1) {
			*x = "-Inf"
		}
	default:
		*x = "None"
	}
}

func CleanFloat64s(x ...*any) {
	for _, ptr := range x {
		CleanFloat64(ptr)
	}
}

func PrintOutlierStats(w io.Writer, o OutlierStats) error {
	CleanFloat64s(
		&o.BiggestOutlierPercentage,
		&o.BackgroundBiggestAverage,
		&o.ChosenRank,
		&o.ChosenInternalRank,
		&o.MeanBgRank,
		&o.RealHighestZ,
		&o.ZRankPercentile,
	)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(o)
}

// Run all outlier code on the command line.
func Outlier(f Flags) (OutlierStats, error) {
	var o OutlierStats
	realEntries, e := iterh.CollectWithError(ParsePedPath(f.RealPath, f.RealHeader))
	if e != nil {
		return o, e
	}

	if f.TopN == -1 {
		realOutlier := GetBiggestOutlier(iterh.SliceIter(realEntries))

		bgOutliersSeq := iterh.BreakOnError(GetBiggestOutlierPaths(iterh.SliceIter[[]string](f.BgPaths), f.BgHeader), &e)
		bgOutliers := iterh.Collect(bgOutliersSeq)
		if e != nil {
			return o, e
		}

		o.BiggestOutlierPercentage = BiggestOutlierPerc(realOutlier, iterh.SliceIter(bgOutliers))
		o.BackgroundBiggestAverage = GetBiggestOutlierAvg(iterh.SliceIter(bgOutliers))
	} else {
		realOutliers := GetBiggestOutliers(iterh.SliceIter(realEntries), f.TopN)
		realOutliersMean, e := stats.Mean(slices.Collect(Posteriors(slices.Values(realOutliers))))
		if e != nil {
			return o, e
		}

		bgOutliersSeq := iterh.BreakOnError(GetBiggestOutliersPaths(iterh.SliceIter[[]string](f.BgPaths), f.BgHeader, f.TopN), &e)
		bgOutliers := iterh.Collect(bgOutliersSeq)
		if e != nil {
			return o, e
		}

		bgOutliersMeans, e := iterh.CollectWithError(PosteriorMeans(slices.Values(bgOutliers)))
		if e != nil {
			return o, e
		}

		o.BiggestOutlierPercentage, _, _ = iterh.Rank(realOutliersMean, slices.Values(bgOutliersMeans))

		bgAvg, e := stats.Mean(bgOutliersMeans)
		if e != nil {
			return o, e
		}
		o.BackgroundBiggestAverage = bgAvg
	}

	if f.Chosen != "" {
		chosenRank, chosenInternalRank, bgRanks, e := RankStats(f.Chosen, iterh.SliceIter(realEntries), ParsePedPaths(f.BgHeader, f.BgPaths...))
		if e != nil {
			return o, e
		}
		meanBgRank, e := stats.Mean(bgRanks)
		if e != nil {
			return o, e
		}
		o.ChosenRank = chosenRank
		o.ChosenInternalRank = chosenInternalRank
		o.MeanBgRank = meanBgRank

		sample := stat.Sample{Xs: bgRanks}
		res, e := stat.OneSampleTTest(sample, 0.5, 0)
		if e != nil {
			return o, e
		}
		o.TTestResult = res
	}

	if f.TopN == -1 {
		realZs, e := GetZscores(realEntries)
		if e != nil {
			return o, e
		}
		realHighestZ := iterh.Max(iterh.SliceIter(realZs))
		bgZScores, e := GetAllBestZscores(ParsePedPaths(f.BgHeader, f.BgPaths...))
		if e != nil {
			return o, e
		}

		zrankperc, zhigher, ztotal := iterh.Rank(realHighestZ, iterh.SliceIter(bgZScores))
		o.RealHighestZ = realHighestZ
		o.ZRankPercentile = zrankperc
		o.ZHigher = zhigher
		o.ZTotal = ztotal
	} else {
		realZs, e := GetZscores(realEntries)
		if e != nil {
			return o, e
		}
		realHighestZs := TopN(iterh.SliceIter(realZs), f.TopN)
		realHighestZsMean, e := stats.Mean(realHighestZs)
		if e != nil {
			return o, e
		}

		bgZScores, e := GetAllBestZscoresTopN(ParsePedPaths(f.BgHeader, f.BgPaths...), f.TopN)
		if e != nil {
			return o, e
		}
		bgZScoresMeans, e := GetBgZScoresMeans(bgZScores)
		if e != nil {
			return o, e
		}

		zrankperc, zhigher, ztotal := iterh.Rank(realHighestZsMean, iterh.SliceIter(bgZScoresMeans))
		o.RealHighestZ = realHighestZsMean
		o.ZRankPercentile = zrankperc
		o.ZHigher = zhigher
		o.ZTotal = ztotal
	}
	return o, nil
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

// Run all outlier code on the command line.
func RunOutlier() {
	var f Flags
	flag.StringVar(&f.RealPath, "r", "", "Path to output of warp for real data")
	flag.StringVar(&f.BgPathsPath, "b", "", "Path to list of paths containing warp output for background data")
	flag.StringVar(&f.Chosen, "c", "", "Chosen individual ID to run rank order statistics on")
	flag.BoolVar(&f.RealHeader, "rh", false, "Real data has a header line")
	flag.BoolVar(&f.BgHeader, "bh", false, "Background data has a header line")
	flag.IntVar(&f.TopN, "t", -1, "Top number of individuals to average to get score (default 1)")

	flag.Parse()
	if f.RealPath == "" {
		log.Fatal("missing -r")
	}
	if f.BgPathsPath == "" {
		log.Fatal("missing -b")
	}

	var e error
	f.BgPaths = iterh.Collect(iterh.BreakOnError(iterh.PathIter(f.BgPathsPath, iterh.LineIter), &e))
	if e != nil {
		log.Fatal(e)
	}

	out, e := Outlier(f)
	if e != nil {
		log.Fatal(e)
	}
	if e := PrintOutlierStats(os.Stdout, out); e != nil {
		log.Fatal(e)
	}
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

var tsvRe = regexp.MustCompile(`\.tsv[^/]*\.gz$`)

func WalkPaths(root string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			yield(path, err)
			return err
		})
		if err != nil {
			yield("", err)
		}
	}
}

func TsvPaths(paths iter.Seq2[string, error]) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for path, err := range paths {
			if err != nil || tsvRe.MatchString(path) {
				if !yield(path, err) {
					return
				}
			}
		}
	}
}

func Outliers(threads int, f ...Flags) ([]OutlierStats, error) {
	out := make([]OutlierStats, len(f))
	var g errgroup.Group
	if threads > 0 {
		g.SetLimit(threads)
	}
	for i, fg := range f {
		i := i
		fg := fg
		g.Go(func() error {
			var e error
			out[i], e = Outlier(fg)
			return e
		})
	}
	return out, g.Wait()
}

type OutliersFlags struct {
	Threads int
	Flags
	OutliersPathPairs string
	Flagsets          []Flags
}

func BuildOutliersFlags() OutliersFlags {
	var f OutliersFlags
	flag.StringVar(&f.Chosen, "c", "", "Chosen individual ID to run rank order statistics on")
	flag.BoolVar(&f.RealHeader, "rh", false, "Real data has a header line")
	flag.BoolVar(&f.BgHeader, "bh", false, "Background data has a header line")
	flag.IntVar(&f.TopN, "t", -1, "Top number of individuals to average to get score (default 1)")
	flag.StringVar(&f.OutliersPathPairs, "p", "", "Path to file containing tab-separated pair of paths (real data, background directory containing files of interest with *.tsv*.gz)")
	flag.IntVar(&f.Threads, "T", 1, "Threads to use (default infinite)")

	flag.Parse()

	if f.OutliersPathPairs == "" {
		log.Fatal("missing -p")
	}

	r, e := os.Open(f.OutliersPathPairs)
	if e != nil {
		log.Fatal(e)
	}
	defer func() {
		if e := r.Close(); e != nil {
			log.Fatal(e)
		}
	}()

	s := bufio.NewScanner(r)
	for s.Scan() {
		fg := f.Flags
		fields := strings.Split(s.Text(), "\t")
		if len(fields) != 2 {
			log.Printf("BuildOutlierFlags: len(fields) %v != 2; fields %v", len(fields), fields)
			continue
		}
		fg.RealPath = fields[0]
		var e error
		fg.BgPaths = iterh.Collect(iterh.BreakOnError(TsvPaths(WalkPaths(fields[1])), &e))
		if e != nil {
			log.Fatal(e)
		}

		f.Flagsets = append(f.Flagsets, fg)
	}
	return f
}

func RunOutliers() {
	flags := BuildOutliersFlags()

	outliers, e := Outliers(flags.Threads, flags.Flagsets...)
	if e != nil {
		log.Fatal(e)
	}

	for _, o := range outliers {
		if e := PrintOutlierStats(os.Stdout, o); e != nil {
			log.Fatal(e)
		}
	}
}
