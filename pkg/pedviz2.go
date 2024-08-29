package tdt

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func PointAes() string {
	return "shape = point, width = 0.06, height = 0.06"
}

type Pedviz2Flags struct {
	FocalID string
}

type IndivNode struct {
	PedEntry
	ChildRels map[Parents]struct{}
}

type Parents struct {
	Father string
	Mother string
}

type RelNode struct {
	Parents
	Children map[string]struct{}
}

type RelTree struct {
	Indivs map[string]IndivNode
	Rels   map[Parents]RelNode
}

func BuildRelTree(tree map[string]Node) RelTree {
	var t RelTree
	t.Indivs = make(map[string]IndivNode)
	t.Rels = make(map[Parents]RelNode)

	for id, node := range tree {
		if _, ok := t.Indivs[id]; !ok {
			t.Indivs[id] = IndivNode{PedEntry: node.PedEntry, ChildRels: map[Parents]struct{}{}}
		}
		parents := Parents{node.PaternalID, node.MaternalID}
		if _, ok := t.Rels[parents]; !ok {
			t.Rels[parents] = RelNode{Parents: parents, Children: map[string]struct{}{}}
		}
		rel := t.Rels[parents]
		rel.Parents = parents
		rel.Children[id] = struct{}{}
	}

	for _, node := range tree {
		parents := Parents{node.PaternalID, node.MaternalID}
		if parents.Father != "" {
			if dad, ok := t.Indivs[node.PaternalID]; ok {
				dad.PedEntry = tree[node.PaternalID].PedEntry
				dad.ChildRels[Parents{node.PaternalID, node.MaternalID}] = struct{}{}
			}
		}
		if parents.Mother != "" {
			if mom, ok := t.Indivs[node.MaternalID]; ok {
				mom.PedEntry = tree[node.MaternalID].PedEntry
				mom.ChildRels[Parents{node.PaternalID, node.MaternalID}] = struct{}{}
			}
		}
	}
	return t
}

func SexAes(p PedEntry) string {
	if p.Sex == 1 {
		return "style = filled" + MaleAes()
	} else if p.Sex == 2 {
		return "style = filled" + FemAes()
	}
	panic(fmt.Errorf("impossible sex %v", p))
	return ""
}

func PedvizSingleIndiv(w io.Writer, in IndivNode) (n int, err error) {
	return fmt.Fprintf(w, "p%v [%v]\n", in.IndividualID, SexAes(in.PedEntry))
}

func PedvizSingleRel(w io.Writer, r RelNode) (n int, err error) {
	return fmt.Fprintf(w, "r%vx%v [%v]\n", r.Parents.Father, r.Parents.Mother, PointAes())
}

func PedvizRelAndDirectChildren(w io.Writer, t RelTree, r RelNode) (n int, err error) {
	nw, e := PedvizSingleRel(w, r)
	n += nw
	if e != nil {
		return n, e
	}
	for childId, _ := range r.Children {
		child := t.Indivs[childId]
		nw, e := PedvizSingleIndiv(w, child)
		n += nw
		if e != nil {
			return n, e
		}

		nw, e = fmt.Fprintf(w, "r%vx%v -> p%v\n", r.Parents.Father, r.Parents.Mother, childId)
		n += nw
		if e != nil {
			return n, e
		}
	}
	return n, e
}

func CollectIndivChildren(t RelTree, in IndivNode) []PedEntry {
	var out []PedEntry
	for rname, _ := range in.ChildRels {
		rel := t.Rels[rname]
		for childname, _ := range rel.Children {
			child := t.Indivs[childname]
			out = append(out, child.PedEntry)
		}
	}
	return out
}

func PedvizIndivRecursive(w io.Writer, t RelTree, in IndivNode) (n int, err error) {
	if in.Sex != 1 {
		return 0, nil
	}
	children := CollectIndivChildren(t, in)
	malefrac := MaleFrac(children)

	nwritten, e := fmt.Fprintf(w, `subgraph cluster_%v {
label = "%v"
labeljust = "l"
labelloc = "l"
graph [color = "#888888"]
`, in.IndividualID, Percentify(malefrac))
	n += nwritten
	if e != nil {
		return n, e
	}

	for rname, _ := range in.ChildRels {
		nwritten, e = PedvizRelAndDirectChildren(w, t, t.Rels[rname])
		n += nwritten
		if e != nil {
			return n, e
		}
	}

	nwritten, e = fmt.Fprintf(w, "}\n")
	n += nwritten
	if e != nil {
		return n, e
	}

	for rname, _ := range in.ChildRels {
		rel := t.Rels[rname]
		nw, e := fmt.Fprintf(w, "p%v -> r%vx%v\n", in.IndividualID, rel.Father, rel.Mother)
		n += nw
		if e != nil {
			return n, e
		}

		for childname, _ := range rel.Children {
			child := t.Indivs[childname]
			nw, e := PedvizIndivRecursive(w, t, child)
			n += nw
			if e != nil {
				return n, e
			}
		}
	}
	return n, nil
}

func Pedviz2(w io.Writer, f Pedviz2Flags, tree map[string]Node) (n int, err error) {
	rtree := BuildRelTree(tree)
	focal, ok := rtree.Indivs[f.FocalID]
	if !ok {
		return n, fmt.Errorf("focal ID %v not in rtree %v", f.FocalID, rtree)
	}
	nw, e := fmt.Fprintf(w, "digraph full {\n")
	n += nw
	if e != nil {
		return n, e
	}

	nw, e = PedvizSingleIndiv(w, focal)
	n += nw
	if e != nil {
		return n, e
	}
	nw, e = PedvizIndivRecursive(w, rtree, focal)
	n += nw
	if e != nil {
		return n, e
	}

	nw, e = fmt.Fprintf(w, "}\n")
	n += nw
	if e != nil {
		return n, e
	}
	return n, e
}

func FullPedviz2() {
	var f Pedviz2Flags
	flag.StringVar(&f.FocalID, "f", "", "focal ID")
	flag.Parse()
	ps, e := ParsePedSafe(os.Stdin)
	if e != nil {
		panic(e)
	}

	tree := BuildPedTree(ps...)

	_, e = Pedviz2(os.Stdout, f, tree)
	if e != nil {
		log.Fatal(e)
	}
}
