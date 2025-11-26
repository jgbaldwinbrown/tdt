package tdt

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

type Relation struct {
	Parent string
	Child string
}

func GetFocalAes(id string, focalSet map[string]struct{}) string {
	if _, ok := focalSet[id]; !ok {
		return ""
	}
	return `; color = "#007700"; penwidth = 10`
}

func Pedviz3One(w io.Writer, tree map[string]Node, r Relation, parentSex int64, focalSet map[string]struct{}) (n int, err error) {
	nw, e := fmt.Fprintf(w, "%v -> %v\n", r.Parent, r.Child);
	n += nw
	if e != nil {
		return n, e
	}

	aes := MaleAes()
	if parentSex == 2 {
		aes = FemAes()
	}
	focalAes := GetFocalAes(r.Parent, focalSet)
	nw, e = fmt.Fprintf(w, "%v [style = filled%v%v]\n", r.Parent, aes, focalAes);
	n += nw
	if e != nil {
		return n, e
	}

	aes = MaleAes()
	if tree[r.Child].Sex == 2 {
		aes = FemAes()
	}
	focalAes = GetFocalAes(r.Child, focalSet)
	nw, e = fmt.Fprintf(w, "%v [style = filled%v%v]\n", r.Child, aes, focalAes);
	n += nw
	return n, e
}

func Pedviz3(w io.Writer, tree map[string]Node, focals []string) (n int, err error) {
	focalSet := make(map[string]struct{}, len(focals))
	for _, foc := range focals {
		focalSet[foc] = struct{}{}
	}
	nw, e := fmt.Fprintf(w, "digraph full {\n");
	n += nw
	if e != nil {
		return n, e
	}
	seen := map[Relation]struct{}{}
	for _, node := range tree {
		father := Relation{node.PaternalID, node.IndividualID};
		if _, ok := seen[father]; !ok && !IsOrphan(node.PaternalID) {
			nw, e := Pedviz3One(w, tree, father, 1, focalSet)
			n += nw
			if e != nil {
				return n, e
			}
			seen[father] = struct{}{}
		}
		mother := Relation{node.MaternalID, node.IndividualID};
		if _, ok := seen[mother]; !ok && !IsOrphan(node.MaternalID){
			nw, e := Pedviz3One(w, tree, mother, 2, focalSet)
			n += nw
			if e != nil {
				return n, e
			}
			seen[mother] = struct{}{}
		}
		for childid, _ := range node.ChildIDs {
			child := Relation{node.IndividualID, childid}
			if _, ok := seen[child]; !ok {
				nw, e := Pedviz3One(w, tree, child, node.Sex, focalSet)
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

type pedviz3Flags struct {
	HighlightPath string
}

func RunPedviz3() {
	var f pedviz3Flags
	flag.StringVar(&f.HighlightPath, "H", "", "Path to file containing line-separated focal IDs.")
	flag.Parse()
	focals, e := ReadLines(f.HighlightPath)
	if e != nil {
		log.Fatal(e)
	}
	ped, e := ParsePedFromReader(os.Stdin)
	if e != nil {
		log.Fatal(e)
	}
	_, e = Pedviz3(os.Stdout, BuildPedTree(ped...), focals)
	if e != nil {
		log.Fatal(e)
	}
}
