package tdt

import (
	"bufio"
	"log"
	"maps"
	"os"
	"slices"
)

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
