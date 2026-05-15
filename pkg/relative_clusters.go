package tdt

import (
	"os"
	"slices"
	"iter"
	"strconv"
	"fmt"
	"log"
	"bufio"
	"flag"

	"github.com/jgbaldwinbrown/iterh"
)

// ind all relatives within dist steps from focalID along the tree. The outer slice is the distance.
func Relatives(tree map[string]Node, focalID string, dist int) [][]string {
	ret := make([][]string, dist + 1)
	found := map[string]struct{}{}
	ret[0] = append(ret[0], focalID)

	for d := 1; d < dist + 1; d++ {
		neighbors := []string{}
		for _, looking := range ret[d - 1] {
			node := tree[looking]
			if node.PaternalID != "0" && node.PaternalID != "999999" {
				neighbors = append(neighbors, node.PaternalID)
			}
			if node.MaternalID != "0" && node.MaternalID != "999999" {
				neighbors = append(neighbors, node.MaternalID)
			}
			for id, _ := range node.ChildIDs {
				neighbors = append(neighbors, id)
			}
		}
		for _, n := range neighbors {
			if _, ok := found[n]; !ok {
				found[n] = struct{}{}
				ret[d] = append(ret[d], n)
			}
		}
	}
	return ret
}

// If a set of clusters share some individuals, merge them into superclusters
func MergeClusters(clusters *[]map[string]struct{}, hits []int) []int {
	newcl := map[string]struct{}{}
	for _, hit := range hits {
		for id, _ := range (*clusters)[hit] {
			newcl[id] = struct{}{}
		}
	}
	slices.SortFunc(hits, func(i, j int) int {
		return j - i
	})
	for _, hit := range hits {
		slices.Delete(*clusters, hit, hit + 1)
	}
	*clusters = append(*clusters, newcl)
	return []int{len(*clusters) - 1}
}

// For a set of focalIDs, generate clusters of relatives that are within dist of them
func RelativeClusters(tree map[string]Node, dist int, focalIDs ...string) []map[string]struct{} {
	clusters := []map[string]struct{}{}
	for _, u := range focalIDs {
		rels := Relatives(tree, u, dist)
		clhits := []int{}
		for _, relset := range rels {
			for _, rel := range relset {
				for i, cl := range clusters {
					if _, ok := cl[rel]; ok {
						clhits = append(clhits, i)
					}
				}
			}
		}
		if len(clhits) > 1 {
			clhits = MergeClusters(&clusters, clhits)
		}
		if len(clhits) < 1 {
			clusters = append(clusters, map[string]struct{}{})
			clhits = []int{len(clusters) - 1}
		}

		clusters[clhits[0]][u] = struct{}{}
		for _,relset := range rels {
			for _, rel := range relset {
				clusters[clhits[0]][rel] = struct{}{}
			}
		}
	}
	return clusters
}

// 	res := TDTTest(BuildFamiliesY(f.IndividualID, peds...)...)
// 	r, e := csvh.OpenMaybeGz(*pedPath)
// 	Must(e)
// 	defer r.Close()
// 	peds, e := ParsePedSafe(r)
// 	Must(e)

// Filter a sequence of Entries based on posterior probability
func FilterPed(it iter.Seq[Entry], minimum float64) iter.Seq[Entry] {
	return func(y func(Entry) bool) {
		for ent := range it {
			if ent.Posterior >= minimum {
				if !y(ent) {
					return
				}
			}
		}
	}
}

// Get the IndividualID from each Entry in "it".
func ToIndividualID(it iter.Seq[Entry]) iter.Seq[string] {
	return func(y func(string) bool) {
		for ent := range it {
			if !y(ent.IndividualID) {
				return
			}
		}
	}
}

// Convert one Entry to one PedEntry
func ToPedEntry(e Entry) PedEntry {
	sex, err := strconv.ParseInt(e.Sex, 0, 64)
	if err != nil {
		panic(err)
	}
	phen, err := strconv.ParseInt(e.Phenotype, 0, 64)
	if err != nil {
		panic(err)
	}
	return PedEntry{
		FamilyID: e.FamilyID,
		IndividualID: e.IndividualID,
		PaternalID: e.FatherID,
		MaternalID: e.MotherID,
		Sex: sex,
		Phenotype: phen,
	}
}

// Convert Entries into PedEntries
func ToPedEntries(it iter.Seq[Entry]) iter.Seq[PedEntry] {
	return func(y func(PedEntry) bool) {
		for ent := range it {
			if !y(ToPedEntry(ent)) {
				return
			}
		}
	}
}

// Flags for the master FullCluster() script
type ClusterFlags struct {
	Header bool
	Steps int
	Thresh float64
}

// Run all clustering code on the command line
func FullCluster() {
	var f ClusterFlags
	flag.BoolVar(&f.Header, "h", false, "Parse a WARP output file with a header")
	flag.IntVar(&f.Steps, "s", 1, "Number of steps allowed between family members")
	flag.Float64Var(&f.Thresh, "t", 1.0, "Likelihood threshold for an individual counting as significant")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("usage: %v [-h] warpout.ped", os.Args[0])
		return
	}
	warp, errp := iterh.BreakWithError(ParsePedPath(args[0], f.Header))
	tree := BuildPedTree(slices.Collect(ToPedEntries(warp))...)
	if *errp != nil {
		log.Fatal(*errp)
	}
	goods := slices.Collect(ToIndividualID(FilterPed(warp, f.Thresh)))
	if *errp != nil {
		log.Fatal(*errp)
	}
	clusters := RelativeClusters(tree, f.Steps, goods...)

	w := bufio.NewWriter(os.Stdout)
	defer func() {
		e := w.Flush()
		if e != nil {
			log.Fatal(e)
		}
	}()
	for _, c := range clusters {
		i := 0
		for id, _ := range c {
			if i < 1 {
				fmt.Fprintf(w, "%v", id)
			} else {
				fmt.Fprintf(w, " %v", id)
			}
			i++
		}
		if len(c) > 0 {
			fmt.Fprintf(w, "\n")
		}
	}
}
