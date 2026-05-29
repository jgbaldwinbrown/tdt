package tdt

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"slices"
	"math"

	"github.com/montanaflynn/stats"
	"github.com/gammazero/toposort"
)

func PedToNodes(ped []PedEntry, tree map[string]Node) []Node {
	out := make([]Node, 0, len(ped))
	for _, ent := range ped {
		if node, ok := tree[ent.IndividualID]; ok {
			out = append(out, node)
		}
	}
	return out
}

func familyClusterCore(cluster map[string]struct{}, node Node, tree map[string]Node) {
	cluster[node.IndividualID] = struct{}{}
	for rel := range node.RelativeNodes(tree) {
		if _, ok := cluster[rel.IndividualID]; !ok {
			familyClusterCore(cluster, rel, tree)
		}
	}
}

func FamilyCluster(node Node, tree map[string]Node) map[string]struct{} {
	out := map[string]struct{}{}
	familyClusterCore(out, node, tree)
	return out
}

func FamilyClusters(nodes []Node, tree map[string]Node) (clusters map[int]map[string]struct{}, cluster_order []int) {
	n := 0
	clusters = map[int]map[string]struct{}{}
	clusterMap := make(map[string]int, len(nodes))
	for _, node := range nodes {
		if _, ok := clusterMap[node.IndividualID]; ok {
			continue
		}
		clusters[n] = FamilyCluster(node, tree)
		for id, _ := range clusters[n] {
			clusterMap[id] = n
		}
		n++
	}

	cluster_order = make([]int, 0, len(clusters))
	for i := 0; i < len(clusters); i++ {
		cluster_order = append(cluster_order, i)
	}
	return clusters, cluster_order
}

func ClustersToTrees(clusters map[int]map[string]struct{}, cluster_order []int, tree map[string]Node) []map[string]Node {
	out := make([]map[string]Node, 0, len(cluster_order))
	for _, i := range cluster_order {
		cluster, ok := clusters[i]
		if !ok {
			continue
		}
		ped := make([]PedEntry, 0, len(cluster))
		for id, _ := range cluster {
			node, ok := tree[id]
			if !ok {
				continue
			}
			ped = append(ped, node.PedEntry)
		}
		out = append(out, BuildPedTree(ped...))
	}
	return out
}

func ToposortTree(tree map[string]Node) []Node {
	if len(tree) < 2 {
		out := []Node{}
		for _, val := range tree {
			out = append(out, val)
		}
		return out
	}

	edges := []toposort.Edge[string]{}
	for _, node := range tree {
		for child, _ := range node.ChildIDs {
			edges = append(edges, toposort.Edge[string]{node.IndividualID, child})
		}
	}
	sorted, e := toposort.Toposort(edges)
	if e != nil {
		panic(e)
	}
	out := make([]Node, 0, len(sorted))
	for _, id := range sorted {
		out = append(out, tree[id])
	}
	return out
}

func TreeLengths(tree map[string]Node) map[string]int {
	topo := ToposortTree(tree)
	lengths := map[string]int{}
	for _, node := range topo {
		longest := 0
		for parent := range node.ParentNodes(tree) {
			length, ok := lengths[parent.IndividualID]
			if ok && length > longest {
				longest = length
			}
		}
		lengths[node.IndividualID] = longest+1
	}
	return lengths
}

func TreeHeight(tree map[string]Node) int {
	lengths := TreeLengths(tree)
	longest := 0
	for _, length := range lengths {
		if length > longest {
			longest = length
		}
	}
	return longest
}

func TreeSize(tree map[string]Node) int {
	return len(tree)
}

type FamilyStatsBlock struct {
	NFamilies int
	MeanHeight float64
	SdHeight float64
	MeanSize float64
	SdSize float64
	FamilySizes []float64
}

func FamilyStats(nodes []Node, tree map[string]Node) (FamilyStatsBlock, error) {
	clusters, cluster_order := FamilyClusters(nodes, tree)
	trees := ClustersToTrees(clusters, cluster_order, tree)
	heights := make([]float64, 0, len(trees))
	sizes := make([]float64, 0, len(trees))
	for _, tree := range trees {
		heights = append(heights, float64(TreeHeight(tree)))
		sizes = append(sizes, float64(TreeSize(tree)))
	}
	var statblock FamilyStatsBlock
	var e error

	statblock.FamilySizes = slices.Clone(sizes)
	slices.Sort(statblock.FamilySizes)
	slices.Reverse(statblock.FamilySizes)

	statblock.NFamilies = len(trees)
	statblock.MeanHeight, e = stats.Mean(heights)
	if e != nil {
		return statblock, e
	}
	statblock.SdHeight, e = stats.StandardDeviation(heights)
	if e != nil {
		return statblock, e
	}
	statblock.MeanSize, e = stats.Mean(sizes)
	if e != nil {
		return statblock, e
	}
	statblock.SdSize, e = stats.StandardDeviation(sizes)
	if e != nil {
		return statblock, e
	}
	return statblock, nil
}

type TreeStatsBlock struct {
	MeanKidsPerParent float64
	SdKidsPerParent float64
	NumHalfOrphans float64
	NumOrphans float64
	NumParented float64
	NumKids float64
	MeanParentsPerKid float64
}

func CountOrphans(tree map[string]Node) (orphans, halforphans int64) {
	for _, node := range tree {
		dadOrphan := IsOrphan(node.PaternalID)
		momOrphan := IsOrphan(node.MaternalID)
		count := 0
		if !dadOrphan {
			count++
		}
		if !momOrphan {
			count++
		}
		if count < 1 {
			orphans++
		}
		if count == 1 {
			halforphans++
		}
	}
	return orphans, halforphans
}

func TreeStats(tree map[string]Node) (TreeStatsBlock, error) {
	counts := make([]float64, 0, len(tree))
	for _, node := range tree {
		if len(node.ChildIDs) > 0 {
			counts = append(counts, float64(len(node.ChildIDs)))
		}
	}
	var ts TreeStatsBlock
	var e error
	ts.MeanKidsPerParent, e = stats.Mean(counts)
	if e != nil {
		return ts, e
	}
	ts.SdKidsPerParent, e = stats.StandardDeviation(counts)
	if e != nil {
		return ts, e
	}
	orphans, halforphans := CountOrphans(tree)
	ts.NumKids = float64(len(tree))
	ts.NumOrphans = float64(orphans)
	ts.NumHalfOrphans = float64(halforphans)
	ts.NumParented = ts.NumKids - (ts.NumOrphans + ts.NumHalfOrphans)
	ts.MeanParentsPerKid = ((2*ts.NumParented) + ts.NumHalfOrphans) / ts.NumKids
	return ts, nil
}

type MarriagesStatBlock struct {
	NMarriageWebs float64
	NMarriedPeople float64
	NEdges float64
	MeanPeoplePerWeb float64
	SdPeoplePerWeb float64
	MeanEdgesPerPerson float64
	SdEdgesPerPerson float64
	MarriageSizes []float64
	MarriageEdges []float64
	MarriageEdgesPerPerson []float64
}

func Weigh(values []float64, weights []float64) []float64 {
	length := 0
	iweights := make([]int, 0, len(weights))
	for _, w := range weights {
		r := int(math.Round(w))
		length += r
		iweights = append(iweights, r)
	}
	out := make([]float64, 0, length)
	for i, v := range values {
		w := iweights[i]
		for _ = range w {
			out = append(out, v)
		}
	}
	return out
}

func MarriagesStats(marriages map[int]map[string]struct{}, marriagesOrder []int, ped []PedEntry, tree map[string]Node) (MarriagesStatBlock, error) {
	var s MarriagesStatBlock
	s.MarriageSizes = make([]float64, 0, len(marriages))
	s.MarriageEdges = make([]float64, 0, len(marriages))
	s.MarriageEdgesPerPerson = make([]float64, 0, len(marriages))
	for _, i := range marriagesOrder {
		mar := marriages[i]
		s.NMarriageWebs++
		s.NMarriedPeople += float64(len(mar))
		s.MarriageSizes = append(s.MarriageSizes, float64(len(mar)))
		s.MarriageEdges = append(s.MarriageEdges, float64(len(mar)-1))
		s.MarriageEdgesPerPerson = append(s.MarriageEdgesPerPerson, (2*float64(len(mar)-1)) / float64(len(mar)))
		s.NEdges += float64(len(mar)-1)
	}
	var e error
	s.MeanPeoplePerWeb, e = stats.Mean(s.MarriageSizes)
	if e != nil {
		return s, e
	}
	s.SdPeoplePerWeb, e = stats.StandardDeviation(s.MarriageSizes)
	if e != nil {
		return s, e
	}

	weighted := Weigh(s.MarriageEdgesPerPerson, s.MarriageSizes)
	s.MeanEdgesPerPerson, e = stats.Mean(weighted)
	if e != nil {
		return s, e
	}
	s.SdEdgesPerPerson, e = stats.StandardDeviation(weighted)
	if e != nil {
		return s, e
	}
	return s, nil
}

func FullFamilyStats() {
	r := bufio.NewReader(os.Stdin)
	ped, e := ParsePedFromReader(r)
	if e != nil {
		log.Fatal(e)
	}
	ped = UniqPed(ped...)

	tree := BuildPedTree(ped...)
	treestats, e := TreeStats(tree)
	if e != nil {
		log.Fatal(e)
	}
	fmt.Printf("%#v\n", treestats)

	marriages, marriagesOrder := ClusterParents(ped)
	marriagesStats, e := MarriagesStats(marriages, marriagesOrder, ped, tree)
	if e != nil {
		log.Fatal(e)
	}
	fmt.Printf("%#v\n", marriagesStats)

	nodes := PedToNodes(ped, tree)
	stats, e := FamilyStats(nodes, tree)
	if e != nil {
		log.Fatal(e)
	}
	fmt.Printf("%#v\n", stats)
}
