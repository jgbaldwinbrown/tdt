package tdt

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/jgbaldwinbrown/zfile"
)

type lineageSize struct {
	Name string
	Fathers float64
	ChildrenPerFather float64
}

func meanSize(fams ...Family) float64 {
	count := 0.0
	sum := 0.0
	for _, fam := range fams {
		if math.IsNaN(fam.MaleF1) || math.IsNaN(fam.FemaleF1) {
			continue
		}
		sum += fam.MaleF1
		sum += fam.FemaleF1
		count++
	}
	return sum/count
}

// Run all-Y TDT test on the command line
func FullKidCount() {
	pedPath := flag.String("i", "", "path to input .ped file")
	outPath := flag.String("o", "", "path to write output")
	flag.Parse()
	if *pedPath == "" {
		log.Fatal(fmt.Errorf("missing -i"))
	}
	if *outPath == "" {
		log.Fatal(fmt.Errorf("missing -o"))
	}

	r, e := zfile.Open(*pedPath)
	Must(e)
	defer r.Close()
	peds, e := ParsePedSafe(r)
	Must(e)

	w, e := zfile.Create(*outPath)
	Must(e)
	defer func() {
		Must(w.Close())
	}()
	enc := json.NewEncoder(w)

	orphanFocal, _ := FindFocals(peds...)

	for _, f := range orphanFocal {
		fams := BuildFamiliesY(f.IndividualID, peds...)
		meansize := meanSize(fams...)
		
		err := enc.Encode(lineageSize{
			Name: f.IndividualID,
			Fathers: float64(len(fams)),
			ChildrenPerFather: meansize,
		})
		Must(err)
	}
}
