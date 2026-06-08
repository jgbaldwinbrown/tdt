package tdt

import (
	"log"
	"fmt"
	"bufio"
	"os"
	"errors"
	"flag"
	"math/rand"

	"golang.org/x/sync/errgroup"
	"github.com/jgbaldwinbrown/zfile"
)

// func ToposortTree(tree map[string]Node) []Node
// func ClusterParents(ped []PedEntry) (clusters map[int]map[string]struct{}, cluster_order []int)

type NodePair struct {
	N0 Node
	N1 Node
}

func FindSexConflicts(nodes []Node, tree map[string]Node) []NodePair {
	var out []NodePair
	for _, n := range nodes {
		hasPat := !IsOrphan(n.PaternalID)
		hasMat := !IsOrphan(n.MaternalID)
		pat := tree[n.PaternalID]
		mat := tree[n.MaternalID]
		if hasPat && hasMat && pat.Phenotype == mat.Phenotype {
			out = append(out, NodePair{pat, mat})
		}
	}
	return out
}

func NodesToPed(nodes []Node) []PedEntry {
	peds := make([]PedEntry, 0, len(nodes))
	for _, node := range nodes {
		peds = append(peds, node.PedEntry)
	}
	return peds
}

func CoinflipTree(tree map[string]Node, maleProb float64, rng *rand.Rand, noconflict bool) []Node {
	log.Println("pre-ToposortTree tree length:", len(tree))
	topo := ToposortTree(tree)
	log.Println("topo length:", len(topo))
	i := 0
	for _, node := range topo {
		if _, ok := tree[node.IndividualID]; !ok {
			log.Printf("extra node %v: %#v\n", i, node)
			i++
		}
	}
	CoinflipNodes(topo, maleProb, rng, noconflict)
	log.Println("post-CoinflipNodes topo length:", len(topo), noconflict)
	return topo
}

func ResolveConflicts(conflicts []NodePair, topoNodes []Node, tree map[string]Node, maleProb float64, rng *rand.Rand) {
	for _, conflict := range conflicts {
		p := tree[conflict.N0.IndividualID]
		m := tree[conflict.N1.IndividualID]
		if p.Phenotype == m.Phenotype {
			newsex := int64(1)
			if rng.Float64() < maleProb {
				newsex = 2
			}
			if rng.Float64() < 0.5 {
				p.Phenotype = newsex
				tree[p.IndividualID] = p
			} else {
				m.Phenotype = newsex
				tree[m.IndividualID] = m
			}
		}
	}
	for i, _ := range topoNodes {
		topoNodes[i] = tree[topoNodes[i].IndividualID]
	}
}

const (
	SexMale = 1
	SexFemale = 2
)

func CoinflipNodes(topoNodes []Node, maleProb float64, rng *rand.Rand, noconflict bool) {
	for i, _ := range topoNodes {
		node := &topoNodes[i]
		flip := rng.Float64()
		if flip < maleProb {
			node.Phenotype = SexMale
		} else {
			node.Phenotype = SexFemale
		}
	}
	tree := BuildPedTree(NodesToPed(topoNodes)...)
	if !noconflict {
		conflicts := FindSexConflicts(topoNodes, tree)
		log.Printf("len(conflicts): %v; conflicts: %v\n", len(conflicts), conflicts)
		for len(conflicts) > 0 {
			ResolveConflicts(conflicts, topoNodes, tree, maleProb, rng)
			conflicts = FindSexConflicts(topoNodes, tree)
			log.Printf("len(conflicts): %v; conflicts: %v\n", len(conflicts), conflicts)
		}
	}
}

func CoinflipParentClustersPath(tree map[string]Node, outpath string, seed int64, maleprob float64, noconflict bool) (err error) {
	rng := rand.New(rand.NewSource(seed))

	w, e := zfile.Create(outpath)
	if e != nil {
		return e
	}
	defer func() {
		err = errors.Join(err, w.Close())
	}()

	nodes := CoinflipTree(tree, maleprob, rng, noconflict)
	log.Print("post-coinflip nodes length:", len(nodes))
	for _, node := range nodes {
		if e := PrintPedEntry(w, node.PedEntry); e != nil {
			return e
		}
	}
	return nil
}

type coinflipParentClustersFlags struct {
	Replicates int
	Outpre string
	RngSeed int64
	Threads int
	MaleProb float64
	NoConflictFix bool
}

func FullCoinflipMulti() {
	var f coinflipParentClustersFlags
	flag.Int64Var(&f.RngSeed, "s", 0, "64-bit integer for RNG seed")
	flag.IntVar(&f.Threads, "t", 1, "Threads")
	flag.IntVar(&f.Replicates, "r", 1, "Replicates")
	flag.StringVar(&f.Outpre, "o", "out", "Prefix for output files")
	flag.Float64Var(&f.MaleProb, "m", 0.5, "Proportion of males to put in permuted files")
	flag.BoolVar(&f.NoConflictFix, "c", false, "Do not fix same-sex mating conflicts")
	flag.Parse()
	rng := rand.New(rand.NewSource(f.RngSeed))

	r := bufio.NewReader(os.Stdin)

	ped, e := ParsePedFromReader(r)
	if e != nil {
		log.Fatal(e)
	}
	ped = CanonicalPed(ped)
	log.Print("ped length:", len(ped))
	tree := BuildPedTree(ped...)
	log.Print("tree length:", len(tree))

	var g errgroup.Group
	if f.Threads > 0 {
		g.SetLimit(f.Threads)
	}
	for i := 0; i < f.Replicates; i++ {
		i := i
		seed := rng.Int63()
		outpath := fmt.Sprintf("%v_%05v.ped.gz", f.Outpre, i)
		g.Go(func() error {
			return CoinflipParentClustersPath(tree, outpath, seed, f.MaleProb, f.NoConflictFix)
		})
	}
	if e := g.Wait(); e != nil {
		log.Fatal(e)
	}
}
