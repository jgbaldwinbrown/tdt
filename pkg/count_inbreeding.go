package tdt

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"maps"

	"github.com/gammazero/deque"
)

type Ancestor struct {
	ID string
	Distance int64
}

type Direction int

const (
	DirUp Direction = iota
	DirDown Direction = iota
)

func WalkDown(seen map[string]Node, entry Node, target Node, tree map[string]Node, top string) []Ancestor {
	if entry.IndividualID == target.IndividualID {
		if len(seen) > 1 {
			return []Ancestor{Ancestor{ID: top, Distance: int64(len(seen))+1}}
		}
		return []Ancestor{}
	}

	seen2 := maps.Clone(seen)
	seen2[entry.IndividualID] = entry
	out := []Ancestor{}
	for childid, _ := range entry.ChildIDs {
		if cnode, cok := tree[childid]; cok {
			if _, isseen := seen[childid]; !isseen {
				out = append(out, WalkDown(seen2, cnode, target, tree, top)...)
			}
		}
	}
	return out
}

func FindLoopsCore(seen map[string]Node, entry Node, target Node, tree map[string]Node) []Ancestor {
	seen2 := maps.Clone(seen)
	seen2[entry.IndividualID] = entry
	out := []Ancestor{}
	for _, pid := range []string{entry.PaternalID, entry.MaternalID} {
		if pnode, pok := tree[pid]; pok {
			if _, found := seen[pnode.IndividualID]; !found {
				out = append(out, FindLoopsCore(seen2, pnode, target, tree)...)
			}
		}
	}

	for childid, _ := range entry.ChildIDs {
		if cnode, cok := tree[childid]; cok {
			if _, isseen := seen[childid]; !isseen {
				out = append(out, WalkDown(seen2, cnode, target, tree, entry.IndividualID)...)
			}
		}
	}

	return out
}

func FindLoops(entry Node, tree map[string]Node) []Ancestor {
	out := []Ancestor{}
	if p, pok := tree[entry.PaternalID]; pok {
		out = append(out, FindLoopsCore(map[string]Node{}, p, entry, tree)...)
	}
	// if p, pok := tree[entry.MaternalID]; pok {
	// 	out = append(out, FindLoopsCore(map[string]Node{}, p, entry, tree)...)
	// }
	return out
}

func appendLoopsOld(dst []Ancestor, entry PedEntry, tree map[string]Node) []Ancestor {
	seen := make(map[string]Ancestor)
	looped := make(map[string]Ancestor)
	d := new(deque.Deque[Ancestor])
	d.PushBack(Ancestor{ID: entry.PaternalID, Distance: 1})
	d.PushBack(Ancestor{ID: entry.MaternalID, Distance: 1})
	for d.Len() > 0 {
		rel := d.PopFront()
		node, ok := tree[rel.ID]
		if !ok {
			continue
		}

		seenrel, seenok := seen[rel.ID]
		_, loopedok := looped[rel.ID]
		if seenok && !loopedok {
			fmt.Printf("seen: %#v; looped: %#v\n", seen, looped)
			outrel := Ancestor{ID: rel.ID, Distance: rel.Distance+seenrel.Distance+1}
			dst = append(dst, outrel)
			looped[outrel.ID] = outrel
		} else if !seenok {
			seen[rel.ID] = rel
		}
		d.PushBack(Ancestor{ID: node.PaternalID, Distance: rel.Distance+1})
		d.PushBack(Ancestor{ID: node.MaternalID, Distance: rel.Distance+1})
	}
	return dst
}

func InbreedingCoefficientCore(loops []Ancestor, coeffs map[string]float64, tree map[string]Node) float64 {
	sum := 0.0
	for _, loop := range loops {
		if _, ok := tree[loop.ID]; !ok {
			continue
		}
		InbreedingCoefficient(loop.ID, coeffs, tree)
		sum += math.Pow(0.5, float64(loop.Distance)-1) * (1 + coeffs[loop.ID])
	}
	return sum
}

func InbreedingCoefficient(ID string, coeffs map[string]float64, tree map[string]Node) {
	if _, ok := coeffs[ID]; ok {
		return
	}
	node, ok := tree[ID]
	if !ok {
		return
	}
	loops := FindLoops(node, tree)
	// log.Printf("%#v\n", loops)
	coeff := InbreedingCoefficientCore(loops, coeffs, tree)
	coeffs[ID] = coeff
}

func InbreedingCoefficients(tree map[string]Node) map[string]float64 {
	coeffs := make(map[string]float64, len(tree))
	for id, _ := range tree {
		InbreedingCoefficient(id, coeffs, tree)
	}
	return coeffs
}

func PrintPedEntryHanging(w io.Writer, p PedEntry) error {
	_, e := fmt.Fprintf(w,
		"%v\t%v\t%v\t%v\t%v\t%v",
		p.FamilyID,
		p.IndividualID,
		p.PaternalID,
		p.MaternalID,
		p.Sex,
		p.Phenotype,
	)
	return e
}

func FullInbreedingCoefficients() {
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
	coeffs := InbreedingCoefficients(tree)
	for _, node := range tree {
		coeff, ok := coeffs[node.IndividualID]
		if !ok {
			continue
		}
		if e := PrintPedEntryHanging(w, node.PedEntry); e != nil {
			log.Fatal(e)
		}
		if _, e := fmt.Fprintf(w, "\t%v\n", coeff); e != nil {
			log.Fatal(e)
		}
	}
}
