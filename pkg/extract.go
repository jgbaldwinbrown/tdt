package tdt

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func ExtractFamilyAuto(focalID string, ps ...PedEntry) []PedEntry {
	tree := BuildPedTree(ps...)
	famped := []PedEntry{}
	for _, p := range ps {
		if HasAuto(p, focalID, tree) {
			famped = append(famped, p)
		}
	}
	return famped
}

func PrintPed(w io.Writer, ped ...PedEntry) (n int, err error) {
	for _, p := range ped {
		nwrit, err := fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n",
			p.FamilyID, p.IndividualID,
			p.PaternalID, p.MaternalID,
			p.Sex, p.Phenotype,
		)
		n += nwrit
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

type ExtractFlags struct {
	FocalID string
}

func FullExtract() {
	var f ExtractFlags
	flag.StringVar(&f.FocalID, "f", "", "ID to extract")
	flag.Parse()

	ped, e := ParsePedFromReader(os.Stdin)
	if e != nil {
		log.Fatal(e)
	}
	eped := ExtractFamilyAuto(f.FocalID, ped...)

	_, e = PrintPed(os.Stdout, eped...)
	if e != nil {
		log.Fatal(e)
	}
}
