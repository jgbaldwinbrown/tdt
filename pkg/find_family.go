package tdt

import (
	"io"
	"fmt"
	"os"
	"bufio"
	"flag"
	"log"
	"slices"
	"maps"
)

func IsOrphan(id string) bool {
	return id == "0" || id == "999999"
}

func InMap[M ~map[K]V, K comparable, V any](m M, k K) bool {
	_, ok := m[k]
	return ok
}

func FindEmbeddingFamily(tree map[string]Node, focalIDs []string) map[string]struct{} {
	family := map[string]struct{}{}
	toCheck := map[string]struct{}{}
	top := map[string]struct{}{}
	for _, id := range focalIDs {
		family[id] = struct{}{}
		toCheck[id] = struct{}{}
	}
	for len(toCheck) > 0 {
		toCheckNow := slices.AppendSeq(make([]string, 0, len(toCheck)), maps.Keys(toCheck))
		for _, id := range toCheckNow {
			n := tree[id]
			if IsOrphan(n.PaternalID) || IsOrphan(n.MaternalID) {
				top[id] = struct{}{}
			}
			if !IsOrphan(n.PaternalID) && !InMap(family, n.PaternalID) {
				family[n.PaternalID] = struct{}{}
				toCheck[n.PaternalID] = struct{}{}
			}
			if !IsOrphan(n.MaternalID) && !InMap(family, n.MaternalID) {
				family[n.MaternalID] = struct{}{}
				toCheck[n.MaternalID] = struct{}{}
			}
			delete(toCheck, id)
		}
	}
	for id, _ := range top {
		toCheck[id] = struct{}{}
	}
	for len(toCheck) > 0 {
		toCheckNow := slices.AppendSeq(make([]string, 0, len(toCheck)), maps.Keys(toCheck))
		for _, id := range toCheckNow {
			n := tree[id]
			for child, _ := range n.ChildIDs {
				if !InMap(family, child) {
					family[child] = struct{}{}
					toCheck[child] = struct{}{}
				}
			}
			delete(toCheck, id)
		}
	}
	return family
}

type FindFamilyFlags struct {
	FocalIDPath string
}

// Read all lines of a file into a slice
func ReadLines(path string) (lines []string, err error) {
	r, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer func() {
		e := r.Close()
		if err == nil {
			err = e
		}
	}()
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e15)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	return lines, s.Err()
}

func PrintPedEntry(w io.Writer, p PedEntry) error {
	_, e := fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n", p.FamilyID, p.IndividualID, p.PaternalID, p.MaternalID, p.Sex, p.Phenotype)
	return e
}

func RunFindFamily() {
	var f FindFamilyFlags
	flag.StringVar(&f.FocalIDPath, "f", "", "Path to file containing line-separated focal IDs (required).")
	flag.Parse()
	if f.FocalIDPath == "" {
		log.Fatal("Missing -f flag")
	}
	focals, e := ReadLines(f.FocalIDPath)
	if e != nil {
		log.Fatal(e)
	}
	ped, e := ParsePedFromReader(os.Stdin)
	if e != nil {
		log.Fatal(e)
	}
	tree := BuildPedTree(ped...)
	fam := FindEmbeddingFamily(tree, focals)
	sfam := slices.AppendSeq(make([]string, 0, len(fam)), maps.Keys(fam))
	slices.Sort(sfam)
	for _, id := range sfam {
		if e := PrintPedEntry(os.Stdout, tree[id].PedEntry); e != nil {
			log.Fatal(e)
		}
	}
}
