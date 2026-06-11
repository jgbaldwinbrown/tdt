package tdt

import (
	"flag"
	"fmt"
	"github.com/jgbaldwinbrown/csvh"
	"io"
	"log"
	"math/rand"
	"os"
)

// type PedEntry struct {
// 	FamilyID string
// 	IndividualID string
// 	PaternalID string
// 	MaternalID string
// 	Sex int64
// 	Phenotype int64
// }

// Randomly re-sort all ped entries
func ShufPedSex(ps []PedEntry, r *rand.Rand) {
	r.Shuffle(len(ps), func(i, j int) {
		ps[i].Sex, ps[j].Sex = ps[j].Sex, ps[i].Sex
	})
}

// aruments; ShufPhenos indicates to shuffle the phenotypes instead of the sexes
type ShufPedSexFlags struct {
	Inpath     string
	Outpre     string
	Reps       int
	Seed       int
	ShufPhenos bool
}

// Parse ped file (again?)
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
	return ParsePedFromReader(r)
}

// Write out ped entry
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

// Write out whole ped file
func WritePed(w io.Writer, ps []PedEntry) error {
	for _, p := range ps {
		if e := WritePedEntry(w, p); e != nil {
			return e
		}
	}
	return nil
}

// Write ped file to path
func WritePedPath(path string, ps []PedEntry) (err error) {
	w, e := csvh.CreateMaybeGz(path)
	if e != nil {
		return e
	}
	defer func() { csvh.DeferE(&err, w.Close()) }()

	return WritePed(w, ps)
}

type HistEntry struct {
	MaleCount float64
	FemaleCount float64
	TotalCount float64
	OtherCount float64
}

func (h HistEntry) MaleFrac() float64 {
	return h.MaleCount / (h.MaleCount+h.FemaleCount)
}

type Hist struct {
	Hist map[string]HistEntry
	Order []string
}

func (h *Hist) AddPedEntry(p PedEntry) {
	if _, ok := h.Hist[p.IndividualID]; !ok {
		h.Hist[p.IndividualID] = HistEntry{}
		h.Order = append(h.Order, p.IndividualID)
	}
	ent := h.Hist[p.IndividualID]
	if p.Phenotype == 1 {
		ent.MaleCount++
	} else if p.Phenotype == 2 {
		ent.FemaleCount++
	} else {
		ent.OtherCount++
	}
	ent.TotalCount++
	h.Hist[p.IndividualID] = ent
}

func (h *Hist) AddPedPath(path string) error {
	ps, e := ParsePedPathMaybe(path)
	if e != nil {
		return e
	}
	for _, ent := range ps {
		h.AddPedEntry(ent)
	}
	return nil
}

func (h *Hist) WriteTo(w io.Writer) error {
	if _, e := fmt.Fprintf(w, "ID\tmale_frac\tmale_count\tfemale_count\tother_count\ttotal_count\n"); e != nil {
		return e
	}
	for _, id := range h.Order {
		ent := h.Hist[id]
		_, e := fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n",
			id,
			ent.MaleFrac(),
			ent.MaleCount,
			ent.FemaleCount,
			ent.OtherCount,
			ent.TotalCount,
		)
		if e != nil {
			return e
		}
	}
	return nil
}

func HitHist(paths []string) (*Hist, error) {
	h := new(Hist)
	h.Hist = make(map[string]HistEntry)
	for _, path := range paths {
		if e := h.AddPedPath(path); e != nil {
			return h, e
		}
	}
	return h, nil
}

// Run the whole pedigree shuffling program on the command line
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

func FullShufHist() {
	paths, e := readReaderLines(os.Stdin)
	if e != nil {
		log.Fatal(e)
	}
	hist, e := HitHist(paths)
	if e != nil {
		log.Fatal(e)
	}
	if e := hist.WriteTo(os.Stdout); e != nil {
		log.Fatal(e)
	}
}
