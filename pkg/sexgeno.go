package tdt

import (
	"fmt"
	"io"
	"github.com/jgbaldwinbrown/csvh"
	"math/rand"
)

func SexPheno(p *PedEntry) {
	p.Phenotype = p.Sex
}

func ShufPedPheno(ps []PedEntry, r *rand.Rand) {
	r.Shuffle(len(ps), func(i, j int) {
		ps[i].Phenotype, ps[j].Phenotype = ps[j].Phenotype, ps[i].Phenotype
	})
}

func SexGeno(sex int64) string {
	if sex == 1 {
		return "Aa"
	}
	return "AA"
}

func WriteSexGenos(w io.Writer, ped ...PedEntry) error {
	if _, e := fmt.Fprintf(w, "fam\tind\tgeno\n"); e != nil {
		return e
	}

	for _, p := range ped {
		if _, e := fmt.Fprintf(w, "%v\t%v\t%v\n", p.FamilyID, p.IndividualID, SexGeno(p.Sex)); e != nil {
			return e
		}
	}
	return nil
}

func WriteSexGenosPath(path string, ped ...PedEntry) (err error) {
	w, e := csvh.CreateMaybeGz(path)
	if e != nil {
		return e
	}
	defer func() { csvh.DeferE(&err, w.Close()) }()
	return WriteSexGenos(w, ped...)
}
