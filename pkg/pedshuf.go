package tdt

import (
	"io"
	"os"
	"github.com/jgbaldwinbrown/csvh"
	"fmt"
	"log"
	"math/rand"
	"flag"
)

// type PedEntry struct {
// 	FamilyID string
// 	IndividualID string
// 	PaternalID string
// 	MaternalID string
// 	Sex int64
// 	Phenotype int64
// }

func ShufPedSex(ps []PedEntry, r *rand.Rand) {
	r.Shuffle(len(ps), func(i, j int) {
		ps[i].Sex, ps[j].Sex = ps[j].Sex, ps[i].Sex
	})
}

type ShufPedSexFlags struct {
	Inpath string
	Outpre string
	Reps int
	Seed int
	ShufPhenos bool
}

func ParsePedPathMaybe(path string) ([]PedEntry, error) {
	var r io.Reader = os.Stdin
	if path != "" {
		f, e := csvh.OpenMaybeGz(path)
		if e != nil {
			return nil, e
		}
		defer f.Close()
		r = f
	}
	return ParsePed(r)
}

func WritePedEntry(w io.Writer, p PedEntry) error {
	_, e := fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n",
			p.FamilyID,
			p.IndividualID,
			p.PaternalID,
			p.MaternalID,
			p.Sex,
			p.Phenotype,
	)
	return e
}

func WritePed(w io.Writer, ps []PedEntry) error {
	for _, p := range ps {
		if e := WritePedEntry(w, p); e != nil {
			return e
		}
	}
	return nil
}

func WritePedPath(path string, ps []PedEntry) (err error) {
	w, e := csvh.CreateMaybeGz(path)
	if e != nil {
		return e
	}
	defer func() { csvh.DeferE(&err, w.Close()) }()

	return WritePed(w, ps)
}

func FullShufPedSex() {
	var f ShufPedSexFlags
	flag.StringVar(&f.Inpath, "i", "", "input .ped path (default stdin)")
	flag.StringVar(&f.Outpre, "o", "shuf_ped_sex_out", "output prefix")
	flag.IntVar(&f.Reps, "r", 1, "shuffle replicates")
	flag.IntVar(&f.Seed, "s", 0, "random seed")
	flag.BoolVar(&f.ShufPhenos, "p", false, "huffle phenotype instead of sex")
	flag.Parse()

	ps, e := ParsePedPathMaybe(f.Inpath)
	if e != nil {
		log.Fatal(e)
	}
	ps = UniqPed(ps...)

	r := rand.New(rand.NewSource(int64(f.Seed)))

	for i := 0; i < f.Reps; i++ {
		if f.ShufPhenos {
			ShufPedPheno(ps, r)
		} else {
			ShufPedSex(ps, r)
		}
		opath := fmt.Sprintf("%v_%v.ped.gz", f.Outpre, i)
		if e := WritePedPath(opath, ps); e != nil {
			log.Fatal(e)
		}
	}
}
