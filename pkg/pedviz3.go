package tdt

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Relation struct {
	Parent string
	Child string
}

func Pedviz3One(w io.Writer, tree map[string]Node, r Relation, parentSex int64) (n int, err error) {
	nw, e := fmt.Fprintf(w, "%v -> %v\n", r.Parent, r.Child);
	n += nw
	if e != nil {
		return n, e
	}

	aes := MaleAes()
	if parentSex == 2 {
		aes = FemAes()
	}
	nw, e = fmt.Fprintf(w, "%v [style = filled%v]\n", r.Parent, aes);
	n += nw
	if e != nil {
		return n, e
	}

	aes = MaleAes()
	if tree[r.Child].Sex == 2 {
		aes = FemAes()
	}
	nw, e = fmt.Fprintf(w, "%v [style = filled%v]\n", r.Child, aes);
	n += nw
	return n, e
}

func Pedviz3(w io.Writer, tree map[string]Node) (n int, err error) {
	nw, e := fmt.Fprintf(w, "digraph full {\n");
	n += nw
	if e != nil {
		return n, e
	}
	seen := map[Relation]struct{}{}
	for _, node := range tree {
		father := Relation{node.PaternalID, node.IndividualID};
		if _, ok := seen[father]; !ok && !IsOrphan(node.PaternalID) {
			nw, e := Pedviz3One(w, tree, father, 1)
			n += nw
			if e != nil {
				return n, e
			}
			seen[father] = struct{}{}
		}
		mother := Relation{node.MaternalID, node.IndividualID};
		if _, ok := seen[mother]; !ok && !IsOrphan(node.MaternalID){
			nw, e := Pedviz3One(w, tree, mother, 2)
			n += nw
			if e != nil {
				return n, e
			}
			seen[mother] = struct{}{}
		}
		for childid, _ := range node.ChildIDs {
			child := Relation{node.IndividualID, childid}
			if _, ok := seen[child]; !ok {
				nw, e := Pedviz3One(w, tree, child, node.Sex)
				n += nw
				if e != nil {
					return n, e
				}
				seen[child] = struct{}{}
			}
		}
	}
	nw, e = fmt.Fprintf(w, "}\n");
	n += nw
	if e != nil {
		return n, e
	}
	return n, nil
}

func RunPedviz3() {
	ped, e := ParsePedFromReader(os.Stdin)
	if e != nil {
		log.Fatal(e)
	}
	_, e = Pedviz3(os.Stdout, BuildPedTree(ped...))
	if e != nil {
		log.Fatal(e)
	}
}
