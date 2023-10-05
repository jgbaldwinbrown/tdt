package tdt

import (
	"io"
	"flag"
	"math"
	"bufio"
	"os"
	"strings"
	"fmt"
	"gonum.org/v1/gonum/stat/distuv"
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

func ChiSqTrioMultiFamily(fams []Family) float64 {
	var b, c float64
	for _, f := range fams {
		b += f.MaleF1
		c += f.FemaleF1
	}
	return ChiSqTrio(b, c)
}

type PedEntry struct {
	FamilyID string
	IndividualID int64
	PaternalID int64
	MaternalID int64
	Sex int64
	Phenotype int64
}

type Node struct {
	PedEntry
	ChildIDs map[int64]struct{}
}

func BuildPedTree(ps ...PedEntry) map[int64]Node {
	tree := make(map[int64]Node, len(ps))
	for _, p := range ps {
		tree[p.IndividualID] = Node{PedEntry: p, ChildIDs: map[int64]struct{}{}}
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

func AddFam(fams []Family, indiv PedEntry, tree map[int64]Node) []Family {
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

func HasY(p PedEntry, focalID int64, tree map[int64]Node) bool {
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

func HasX(p PedEntry, focalID int64, tree map[int64]Node) bool {
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

func HasAuto(p PedEntry, focalID int64, tree map[int64]Node) bool {
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

func BuildFamiliesY(focalID int64, ps ...PedEntry) []Family {
	tree := BuildPedTree(ps...)
	var fams []Family
	for _, p := range ps {
		if HasY(p, focalID, tree) {
			fams = AddFam(fams, p, tree)
		}
	}
	return fams
}

func BuildFamiliesX(focalID int64, ps ...PedEntry) []Family {
	tree := BuildPedTree(ps...)
	var fams []Family
	for _, p := range ps {
		if HasX(p, focalID, tree) {
			fams = AddFam(fams, p, tree)
		}
	}
	return fams
}

func BuildFamiliesMaleX(focalID int64, ps ...PedEntry) []Family {
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

func BuildFamiliesFemaleX(focalID int64, ps ...PedEntry) []Family {
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

func BuildFamiliesAuto(focalID int64, ps ...PedEntry) []Family {
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
	return p, e
}

func ParsePed(r io.Reader) ([]PedEntry, error) {
	s := bufio.NewScanner(r)
	var ps []PedEntry
	for s.Scan() {
		if s.Err() != nil {
			return nil, s.Err()
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
		panic(e)
	}
}

func FullTDTTest() {
	focal := flag.Int("f", -1, "focal ID (required)")
	flag.Parse()
	if *focal == -1 {
		panic(fmt.Errorf("missing -f"))
	}

	peds, e := ParsePed(os.Stdin)
	Must(e)

	dist := distuv.ChiSquared{K: 1}

	xchi := ChiSqTrioMultiFamily(BuildFamiliesFemaleX(int64(*focal), peds...))
	xp := 1 - dist.CDF(math.Abs(xchi))
	fmt.Printf("xchi: %v; xp: %v\n", xchi, xp)

	ychi := ChiSqTrioMultiFamily(BuildFamiliesY(int64(*focal), peds...))
	yp := 1 - dist.CDF(math.Abs(ychi))
	fmt.Printf("ychi: %v; yp: %v\n", ychi, yp)

	achi := ChiSqTrioMultiFamily(BuildFamiliesAuto(int64(*focal), peds...))
	ap := 1 - dist.CDF(math.Abs(achi))
	fmt.Printf("achi: %v; ap: %v\n", achi, ap)
}
