package tdt

import (
	"bufio"
	"log"
	"maps"
	"os"
	"slices"
)

func CanonicalPed(ped []PedEntry) []PedEntry {
	tree := BuildPedTree(ped...)
	return CanonicalTree(tree)
}

func CanonicalTree(tree map[string]Node) []PedEntry {
	ids := slices.Collect(maps.Keys(tree))
	slices.Sort(ids)
	out := make([]PedEntry, 0, len(tree))
	for _, id := range ids {
		out = append(out, tree[id].PedEntry)
	}
	return out
}

func FullCanonicalTree() {
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
	ids := slices.Collect(maps.Keys(tree))
	slices.Sort(ids)
	for _, id := range ids {
		if e := PrintPedEntry(w, tree[id].PedEntry); e != nil {
			log.Fatal(e)
		}
	}
}
