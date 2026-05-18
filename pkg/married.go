package tdt

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"maps"
	"math/rand"
	"os"
)

func SingleMatch(paternalIdx int, pok bool, maternalIdx int, mok bool) (int, bool) {
	if pok && !mok {
		return paternalIdx, true
	}
	if mok && !pok {
		return maternalIdx, true
	}
	if pok && mok && (paternalIdx == maternalIdx) {
		return paternalIdx, true
	}
	return 0, false
}

func CombineClusters(c0, c1 map[string]struct{}) map[string]struct{} {
	c2 := make(map[string]struct{}, len(c0)+len(c1))
	for id, _ := range c0 {
		c2[id] = struct{}{}
	}
	for id, _ := range c1 {
		c2[id] = struct{}{}
	}
	return c2
}

func ClusterParents(tree map[string]Node) map[int]map[string]struct{} {
	n := 0
	clusters := map[int]map[string]struct{}{}
	clusterMap := make(map[string]int, len(tree))
	for _, node := range tree {
		pi, pok := clusterMap[node.PaternalID]
		mi, mok := clusterMap[node.MaternalID]
		if !pok && !mok {
			clusters[n] = map[string]struct{}{node.PaternalID: struct{}{}, node.MaternalID: struct{}{}}
			clusterMap[node.PaternalID] = n
			clusterMap[node.MaternalID] = n
			n++
		}
		if ini, inok := SingleMatch(pi, pok, mi, mok); inok {
			clusters[ini][node.PaternalID] = struct{}{}
			clusters[ini][node.MaternalID] = struct{}{}
		}
		if pok && mok && pok != mok {
			newcluster := CombineClusters(clusters[pi], clusters[mi])
			newcluster[node.PaternalID] = struct{}{}
			newcluster[node.MaternalID] = struct{}{}
			clusters[n] = newcluster
			delete(clusters, pi)
			delete(clusters, mi)
			for name, _ := range newcluster {
				clusterMap[name] = n
			}
			n++
		}
	}
	return clusters
}

func PrintClusters(w io.Writer, clusters map[int]map[string]struct{}) error {
	for _, cluster := range clusters {
		i := 0
		for id, _ := range cluster {
			if i < 1 {
				if _, e := fmt.Fprintf(w, "%v", id); e != nil {
					return e
				}
			} else {
				if _, e := fmt.Fprintf(w, "\t%v", id); e != nil {
					return e
				}
			}
			i++
		}
		if _, e := fmt.Fprintf(w, "\n"); e != nil {
			return e
		}
	}
	return nil
}

func FlipClustersSexes(clusters map[int]map[string]struct{}, tree map[string]Node, rng *rand.Rand) map[string]Node {
	out := maps.Clone(tree)
	for _, cluster := range clusters {
		flip := rng.Float64() < 0.5
		for id, _ := range cluster {
			node, ok := out[id]
			if !ok {
				continue
			}
			if flip {
				node.Phenotype = (node.Phenotype%2)+1
				out[id] = node
			}
		}
	}
	return out
}

type flipParentClustersFlags struct {
	RngSeed int64
}

func FullFlipParentClusters() {
	var f flipParentClustersFlags
	flag.Int64Var(&f.RngSeed, "s", 0, "64-bit integer for RNG seed")
	flag.Parse()
	rng := rand.New(rand.NewSource(f.RngSeed))

	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer func() {
		if e := w.Flush(); e != nil {
			log.Fatal(e)
		}
	}()

	ped, e := ParsePedFromReader(r)
	if e != nil {
		log.Fatal(e)
	}
	tree := BuildPedTree(ped...)
	clusters := ClusterParents(tree)
	tree2 := FlipClustersSexes(clusters, tree, rng)
	for _, node := range tree2 {
		if e := PrintPedEntry(w, node.PedEntry); e != nil {
			log.Fatal(e)
		}
	}
}
