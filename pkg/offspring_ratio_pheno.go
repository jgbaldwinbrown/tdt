package tdt

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func OffspringRatioPhenos(r io.Reader, w io.Writer, f OffspringRatioPhenosFlags) error {
	ped, e := ParsePedFromReader(r)
	if e != nil {
		log.Fatal(e)
	}
	tree := BuildPedTree(ped...)
	ped = CanonicalTree(tree)
	for _, ent := range ped {
		node, ok := tree[ent.IndividualID]
		if !ok {
			return fmt.Errorf("OffspringRatioPhenos: entry %v ID %v missing from tree %v", ent, ent.IndividualID, tree)
		}
		maleKids := 0
		for childID, _ := range node.ChildIDs {
			childNode, ok := tree[childID]
			if !ok {
				return fmt.Errorf("OffspringRatioPhenos: childID %v missing from tree %v", childID, tree)
			}
			if childNode.Sex == 1 {
				maleKids++
			}
		}
		if !f.Counts {
			ent.Phenotype = 0.5
			if len(node.ChildIDs) > 0 && node.Sex == 1 {
				if !f.Female {
					ent.Phenotype = float64(maleKids) / float64(len(node.ChildIDs))
				} else {
					ent.Phenotype = (float64(len(node.ChildIDs)) - float64(maleKids)) / float64(len(node.ChildIDs))
				}
			}
		} else {
			ent.Phenotype = 0
			ent.Extra = append(ent.Extra, 0)
			if len(node.ChildIDs) > 0 && node.Sex == 1 {
				ent.Extra[0] = len(node.ChildIDs)
				if !f.Female {
					ent.Phenotype = float64(maleKids)
				} else {
					ent.Phenotype = float64(len(node.ChildIDs) - maleKids)
				}
			}
		}
					
		if e := WritePedEntry(w, ent); e != nil {
			return e
		}
	}
	return nil
}

type OffspringRatioPhenosFlags struct {
	Female bool
	Counts bool
}

func FullOffspringRatioPhenos() {
	var f OffspringRatioPhenosFlags
	flag.BoolVar(&f.Female, "f", false, "Set phenotypes to female ratio")
	flag.BoolVar(&f.Counts, "c", false, "Output counts, not ratio")
	flag.Parse()

	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer func() {
		if e := w.Flush(); e != nil {
			log.Fatal(e)
		}
	}()
	if e := OffspringRatioPhenos(r, w, f); e != nil {
		log.Fatal(e)
	}
}
