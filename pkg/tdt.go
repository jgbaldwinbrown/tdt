package tdt

import (
	"log"
	"encoding/json"
	"regexp"
	"io"
	"flag"
	"math"
	"bufio"
	"os"
	"strings"
	"fmt"
	"gonum.org/v1/gonum/stat/distuv"
	"github.com/jgbaldwinbrown/csvh"
)

func ChiSqTrio(b, c float64) float64 {
	chisq := ( (b-c) * (b-c) ) / (b+c)
	return chisq
}

func ChiSqExtended(i, j, h float64) float64 {
	return 4 * (i - j) * (i - j) / h
}

type Family struct {
	MaleF1 float64
	FemaleF1 float64
}

func ChiSqTrioMultiFamily(fams ...Family) float64 {
	sums := CondenseFamilies(fams...)
	return ChiSqTrio(sums.MaleF1, sums.FemaleF1)
}

func CondenseFamilies(fams ...Family) Family {
	var sums Family
	for _, f := range fams {
		sums.MaleF1 += f.MaleF1
		sums.FemaleF1 += f.FemaleF1
	}
	return sums
}

type PedEntry struct {
	FamilyID string
	IndividualID string
	PaternalID string
	MaternalID string
	Sex int64
	Phenotype int64
}

type Node struct {
	PedEntry
	ChildIDs map[string]struct{}
}

func BuildPedTree(ps ...PedEntry) map[string]Node {
	tree := make(map[string]Node, len(ps))
	for _, p := range ps {
		tree[p.IndividualID] = Node{PedEntry: p, ChildIDs: map[string]struct{}{}}
	}

	for _, p := range ps {
		if dad, ok := tree[p.PaternalID]; ok {
			dad.ChildIDs[p.IndividualID] = struct{}{}
			tree[p.PaternalID] = dad
		}
		if mom, ok := tree[p.MaternalID]; ok {
			mom.ChildIDs[p.IndividualID] = struct{}{}
			tree[p.MaternalID] = mom
		}
	}
	return tree
}

func AddFam(fams []Family, indiv PedEntry, tree map[string]Node) []Family {
	var fam Family
	node, ok := tree[indiv.IndividualID]
	if !ok {
		panic(fmt.Errorf("indiv %v not in tree %v", indiv, tree))
	}
	for childID, _ := range node.ChildIDs {
		child, ok := tree[childID]
		if !ok {
			panic(fmt.Errorf("child %v not in tree %v", child, tree))
		}
		if child.Sex == 1 {
			fam.MaleF1++
		}
		if child.Sex == 2 {
			fam.FemaleF1++
		}
	}
	return append(fams, fam)
}

func HasY(p PedEntry, focalID string, tree map[string]Node) bool {
	if p.IndividualID == focalID {
		return true
	}

	if p.Sex != 1 {
		return false
	}

	if p.PaternalID == focalID {
		return true
	}

	if dad, ok := tree[p.PaternalID]; ok {
		return HasY(dad.PedEntry, focalID, tree)
	}
	return false
}

func DadHasY(p PedEntry, focalID string, tree map[string]Node) bool {
	if dad, ok := tree[p.PaternalID]; ok {
		return HasY(dad.PedEntry, focalID, tree)
	}
	return false
}

func HasX(p PedEntry, focalID string, tree map[string]Node) bool {
	if p.IndividualID == focalID {
		return true
	}

	if p.Sex == 1 {
		if p.MaternalID == focalID {
			return true
		}
		if mom, ok := tree[p.MaternalID]; ok {
			return HasX(mom.PedEntry, focalID, tree)
		}
		return false
	}

	if p.Sex == 2 {
		if p.PaternalID == focalID {
			return true
		}
		if dad, ok := tree[p.PaternalID]; ok {
			if HasX(dad.PedEntry, focalID, tree) {
				return true
			}
		}
		if p.MaternalID == focalID {
			return true
		}
		if mom, ok := tree[p.MaternalID]; ok {
			if HasX(mom.PedEntry, focalID, tree) {
				return true
			}
		}
		return false
	}

	return false
}

func HasXFemDescent(p PedEntry, focalID string, tree map[string]Node) bool {
	if p.IndividualID == focalID {
		return true
	}

	if p.Sex == 1 {
		return false
	}

	if p.Sex == 2 {
		if p.MaternalID == focalID {
			return true
		}
		if mom, ok := tree[p.MaternalID]; ok {
			if HasX(mom.PedEntry, focalID, tree) {
				return true
			}
		}
		return false
	}

	return false
}

func HasAuto(p PedEntry, focalID string, tree map[string]Node) bool {
	if p.IndividualID == focalID || p.PaternalID == focalID || p.MaternalID == focalID {
		return true
	}
	if dad, ok := tree[p.PaternalID]; ok {
		return HasAuto(dad.PedEntry, focalID, tree)
	}
	if mom, ok := tree[p.MaternalID]; ok {
		return HasAuto(mom.PedEntry, focalID, tree)
	}
	return false
}

func BuildFamiliesY(focalID string, ps ...PedEntry) []Family {
	tree := BuildPedTree(ps...)
	var fams []Family
	for _, p := range ps {
		if HasY(p, focalID, tree) {
			fams = AddFam(fams, p, tree)
		}
	}
	return fams
}

func BuildFamiliesX(focalID string, ps ...PedEntry) []Family {
	tree := BuildPedTree(ps...)
	var fams []Family
	for _, p := range ps {
		if HasX(p, focalID, tree) {
			fams = AddFam(fams, p, tree)
		}
	}
	return fams
}

func BuildFamiliesMaleX(focalID string, ps ...PedEntry) []Family {
	tree := BuildPedTree(ps...)
	var fams []Family
	for _, p := range ps {
		if HasX(p, focalID, tree) {
			if p.Sex == 1 {
				fams = AddFam(fams, p, tree)
			}
		}
	}
	return fams
}

func BuildFamiliesFemaleX(focalID string, ps ...PedEntry) []Family {
	tree := BuildPedTree(ps...)
	var fams []Family
	for _, p := range ps {
		if HasX(p, focalID, tree) {
			if p.Sex == 2 {
				fams = AddFam(fams, p, tree)
			}
		}
	}
	return fams
}

func BuildFamiliesFemDescentFemaleX(focalID string, ps ...PedEntry) []Family {
	tree := BuildPedTree(ps...)
	var fams []Family
	for _, p := range ps {
		if HasXFemDescent(p, focalID, tree) {
			if p.Sex == 2 {
				fams = AddFam(fams, p, tree)
			}
		}
	}
	return fams
}

func BuildFamiliesAuto(focalID string, ps ...PedEntry) []Family {
	tree := BuildPedTree(ps...)
	var fams []Family
	for _, p := range ps {
		if HasAuto(p, focalID, tree) {
			fams = AddFam(fams, p, tree)
		}
	}
	return fams
}

func Scan(line []string, ptrs ...any) (n int, err error) {
	for i, ptr := range ptrs {
		nread, e := fmt.Sscanf(line[i], "%v", ptr)
		n += nread
		if e != nil {
			return n, e
		}
	}
	return n, nil
}

func ParsePedEntry(s string) (PedEntry, error) {
	line := strings.Fields(s)
	var p PedEntry
	if len(line) < 6 {
		return p, fmt.Errorf("len(line) %v < 6", len(line))
	}
	_, e := Scan(line, &p.FamilyID, &p.IndividualID, &p.PaternalID, &p.MaternalID, &p.Sex, &p.Phenotype)
	if e != nil {
		return p, fmt.Errorf("ParsePedEntry: %w; line: %v", e, line)
	}
	return p, nil
}

var skipRe = regexp.MustCompile(`^#|^$`)

func ShouldSkipPedLine(line string) bool {
	return skipRe.MatchString(line)
}

func ParsePed(r io.Reader) ([]PedEntry, error) {
	s := bufio.NewScanner(r)
	var ps []PedEntry
	for s.Scan() {
		if s.Err() != nil {
			return nil, s.Err()
		}
		if ShouldSkipPedLine(s.Text()) {
			continue
		}

		p, e := ParsePedEntry(s.Text())
		if e != nil {
			return nil, e
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func Must(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

type TDTResult struct {
	Name string
	Totals Family
	Nfamilies float64
	MaleProportion float64
	MeanMalesPerFam float64
	MeanFemalesPerFam float64
	MeanChildrenPerFam float64
	Chisq float64
	P float64
	Orphan bool
}

func TDTTest(fams ...Family) TDTResult {
	var r TDTResult

	r.Totals = CondenseFamilies(fams...)
	r.Chisq = ChiSqTrio(r.Totals.MaleF1, r.Totals.FemaleF1)
	r.MaleProportion = r.Totals.MaleF1 / (r.Totals.MaleF1 + r.Totals.FemaleF1)

	dist := distuv.ChiSquared{K: 1}
	r.P = 1 - dist.CDF(math.Abs(r.Chisq))

	r.Nfamilies = float64(len(fams))
	r.MeanMalesPerFam = r.Totals.MaleF1 / r.Nfamilies
	r.MeanFemalesPerFam = r.Totals.FemaleF1 / r.Nfamilies
	r.MeanChildrenPerFam = (r.Totals.FemaleF1 + r.Totals.MaleF1) / r.Nfamilies

	return r
}

type TDTResultJson struct {
	Name string
	TotalMales any
	TotalFemales any
	Nfamilies any
	MaleProportion any
	MeanMalesPerFam any
	MeanFemalesPerFam any
	MeanChildrenPerFam any
	Chisq any
	P any
	Orphan bool
}

func FloatToJson(f float64) any {
	if math.IsNaN(f) {
		return "NaN"
	}
	if math.IsInf(f, 0) {
		return "Infinity"
	}
	return f
}

func ToJson(r TDTResult) TDTResultJson {
	var j TDTResultJson
	j.Name = r.Name
	j.TotalMales = FloatToJson(r.Totals.MaleF1)
	j.TotalFemales = FloatToJson(r.Totals.FemaleF1)
	j.Nfamilies = FloatToJson(r.Nfamilies)
	j.MaleProportion = FloatToJson(r.MaleProportion)
	j.MeanMalesPerFam = FloatToJson(r.MeanMalesPerFam)
	j.MeanFemalesPerFam = FloatToJson(r.MeanFemalesPerFam)
	j.MeanChildrenPerFam = FloatToJson(r.MeanChildrenPerFam)
	j.Chisq = FloatToJson(r.Chisq)
	j.P = FloatToJson(r.P)
	j.Orphan = r.Orphan
	return j
}

func FullTDTTestOld() {
	focal := flag.Int("f", -1, "focal ID (required)")
	flag.Parse()
	if *focal == -1 {
		panic(fmt.Errorf("missing -f"))
	}

	peds, e := ParsePed(os.Stdin)
	Must(e)

	dist := distuv.ChiSquared{K: 1}

	xchi := ChiSqTrioMultiFamily(BuildFamiliesFemaleX(string(*focal), peds...)...)
	xp := 1 - dist.CDF(math.Abs(xchi))
	fmt.Printf("xchi: %v; xp: %v\n", xchi, xp)

	xchifemdescent := ChiSqTrioMultiFamily(BuildFamiliesFemDescentFemaleX(string(*focal), peds...)...)
	xpfemdescent := 1 - dist.CDF(math.Abs(xchi))
	fmt.Printf("fem descent xchi: %v; xp: %v\n", xchifemdescent, xpfemdescent)

	ychi := ChiSqTrioMultiFamily(BuildFamiliesY(string(*focal), peds...)...)
	yp := 1 - dist.CDF(math.Abs(ychi))
	fmt.Printf("ychi: %v; yp: %v\n", ychi, yp)

	achi := ChiSqTrioMultiFamily(BuildFamiliesAuto(string(*focal), peds...)...)
	ap := 1 - dist.CDF(math.Abs(achi))
	fmt.Printf("achi: %v; ap: %v\n", achi, ap)
}


func FullTDTTest() {
	focal := flag.Int("f", -1, "focal ID (required)")
	flag.Parse()
	if *focal == -1 {
		panic(fmt.Errorf("missing -f"))
	}

	peds, e := ParsePed(os.Stdin)
	Must(e)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")

	res := TDTTest(BuildFamiliesFemaleX(string(*focal), peds...)...)
	res.Name = "FemaleX"
	err := enc.Encode(ToJson(res))
	Must(err)

	res = TDTTest(BuildFamiliesFemDescentFemaleX(string(*focal), peds...)...)
	res.Name = "FemDescentFemaleX"
	err = enc.Encode(ToJson(res))
	Must(err)

	res = TDTTest(BuildFamiliesY(string(*focal), peds...)...)
	res.Name = "Y"
	err = enc.Encode(ToJson(res))
	Must(err)

	res = TDTTest(BuildFamiliesAuto(string(*focal), peds...)...)
	res.Name = "Auto"
	err = enc.Encode(ToJson(res))
	Must(err)
}

func ReadLines(path string) ([]string, error) {
	r, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer r.Close()

	var out []string
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e9)
	for s.Scan() {
		if s.Err() != nil {
			return nil, s.Err()
		}
		out = append(out, s.Text())
	}
	return out, nil
}

func FullMultiYTDTTest() {
	focalPath := flag.String("f", "", "path to line-separated IDs for focal individuals")
	flag.Parse()
	if *focalPath == "" {
		log.Fatal(fmt.Errorf("missing -f"))
	}

	peds, e := ParsePed(os.Stdin)
	Must(e)

	w := bufio.NewWriter(os.Stdout)
	defer func() {
		e := w.Flush()
		if e != nil {
			log.Fatal(e)
		}
	}()
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")

	focals, e := ReadLines(*focalPath)
	Must(e)

	for _, f := range focals {
		res := TDTTest(BuildFamiliesY(f, peds...)...)
		res.Name = fmt.Sprint(f)
		err := enc.Encode(ToJson(res))
		Must(err)
	}
}

func FullAllYTDTTest() {
	pedPath := flag.String("i", "", "path to input .ped file")
	outPath := flag.String("o", "", "path to write output")
	flag.Parse()
	if *pedPath == "" {
		log.Fatal(fmt.Errorf("missing -i"))
	}
	if *outPath == "" {
		log.Fatal(fmt.Errorf("missing -o"))
	}

	r, e := csvh.OpenMaybeGz(*pedPath)
	Must(e)
	defer r.Close()
	peds, e := ParsePedSafe(r)
	Must(e)

	ww, e := csvh.CreateMaybeGz(*outPath)
	Must(e)
	defer func() {
		Must(ww.Close())
	}()
	w := bufio.NewWriter(ww)
	defer func() {
		Must(w.Flush())
	}()
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")

	orphanFocal, nonOrphanFocal := FindFocals(peds...)

	i := 0
	for _, f := range orphanFocal {
		res := TDTTest(BuildFamiliesY(f.IndividualID, peds...)...)
		res.Name = fmt.Sprint(i)
		res.Orphan = true
		err := enc.Encode(ToJson(res))
		Must(err)
		i++
	}

	for _, f := range nonOrphanFocal {
		res := TDTTest(BuildFamiliesY(f.IndividualID, peds...)...)
		res.Name = fmt.Sprint(i)
		res.Orphan = false
		err := enc.Encode(ToJson(res))
		Must(err)
		i++
	}
}
