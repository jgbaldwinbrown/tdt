package tdt

import (
	"io"
	"fmt"
	"os"
)

func ClusterXs(focalID int64, tree map[int64]Node, ps ...PedEntry) (unclustered []PedEntry, clusters []NamedCluster) {
	mclusters := make(map[int64][]PedEntry)
	for _, p := range ps {
		if HasX(tree[p.MaternalID].PedEntry, focalID, tree) {
			mclusters[p.MaternalID] = append(mclusters[p.MaternalID], p)
		} else {
			unclustered = append(unclustered, p)
		}
	}
	return unclustered, SortClusters(mclusters)
}

func PedEntryToGraphVizX(w io.Writer, focalID int64, tree map[int64]Node, p PedEntry) (n int, err error) {
	if p.PaternalID != 0 {
		extra := " [style=dotted]"
		if HasX(tree[p.PaternalID].PedEntry, focalID, tree) && p.Sex == 2 {
			extra = ""
		}
		nwritten, e := fmt.Fprintf(w, "p%v -> p%v%v\np%v [style=filled; fillcolor=%v]\n", p.PaternalID, p.IndividualID, extra, p.PaternalID, Blue())
		n += nwritten
		if e != nil {
			return n, e
		}
	}
	if p.MaternalID != 0 {
		extra := " [style=dotted]"
		if HasX(tree[p.MaternalID].PedEntry, focalID, tree) {
			extra = ""
		}
		nwritten, e := fmt.Fprintf(w, "p%v -> p%v%v\np%v [style=filled; fillcolor=%v]\n", p.MaternalID, p.IndividualID, extra, p.MaternalID, Red())
		n += nwritten
		if e != nil {
			return n, e
		}
	}
	if p.Sex == 1 {
		nwritten, e := fmt.Fprintf(w, "p%v [style=filled; fillcolor=%v]\n", p.IndividualID, Blue())
		n += nwritten
		if e != nil {
			return n, e
		}
	} else if p.Sex == 2 {
		nwritten, e := fmt.Fprintf(w, "p%v [style=filled; fillcolor=%v]\n", p.IndividualID, Red())
		n += nwritten
		if e != nil {
			return n, e
		}
	} else if p.Sex == 0 {
	} else {
		fmt.Fprintf(os.Stderr, "weird sex: %v; %v\n", p.Sex, p)
	}

	return n, nil
}

func ToGraphVizX(w io.Writer, opts GraphVizOpts, ps ...PedEntry) (n int, err error) {
	f := opts.FocalID
	tree := BuildPedTree(ps...)

	unclustered, clusters := ClusterXs(f, tree, ps...)

	nwritten, e := fmt.Fprintf(w, "digraph full {\n")
	n += nwritten
	if e != nil { return n, e }

	for _, p := range unclustered {
		nwritten, e := fmt.Fprintf(w, "p%v\n", p.IndividualID)
		n += nwritten
		if e != nil { return n, e }
	}

	for _, cps := range clusters {
		cid := cps.ID
		nwritten, e := fmt.Fprintf(w, "subgraph cluster_%v {\n", cid)
		n += nwritten
		if e != nil { return n, e }

		for _, p := range cps.Cluster {
			nwritten, e := PedEntryToGraphVizX(w, f, tree, p)
			n += nwritten
			if e != nil { return n, e }
		}

		nwritten, e = fmt.Fprintf(w, "}\n")
		n += nwritten
		if e != nil { return n, e }
	}

	for _, p := range unclustered {
		nwritten, e := PedEntryToGraphVizX(w, f, tree, p)
		n += nwritten
		if e != nil { return n, e }
	}

	nwritten, e = fmt.Fprintf(w, "}\n")
	n += nwritten
	if e != nil { return n, e }

	return n, nil
}

