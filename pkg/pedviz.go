package tdt

import (
	"sort"
	"flag"
	"io"
	"os"
	"fmt"
)

func Red() string {
	return `"#cc6666"`
}

func Blue() string {
	return `"#6666cc"`
}

func FemShape() string {
	return `circle`
}

func MaleShape() string {
	return `square`
}

func FemAes() string {
	return fmt.Sprintf(`; fillcolor=%v; shape=%v`, Red(), FemShape())
}

func MaleAes() string {
	return fmt.Sprintf(`; fillcolor=%v; shape=%v`, Blue(), MaleShape())
}

func ToGraphVizSimple(w io.Writer, ps ...PedEntry) (n int, err error) {
	nwritten, e := fmt.Fprintf(w, "digraph full {\n")
	n += nwritten
	if e != nil {
		return n, e
	}

	for _, p := range ps {
		if p.PaternalID != 0 {
			nwritten, e := fmt.Fprintf(w, "p%v -> p%v\np%v [style=filled; fillcolor=%v]\n", p.PaternalID, p.IndividualID, p.PaternalID, Blue())
			n += nwritten
			if e != nil {
				return n, e
			}
		}
		if p.MaternalID != 0 {
			nwritten, e := fmt.Fprintf(w, "p%v -> p%v\np%v [style=filled; fillcolor=%v]\n", p.MaternalID, p.IndividualID, p.MaternalID, Red())
			n += nwritten
			if e != nil {
				return n, e
			}
		}
		if p.Sex == 1 {
			nwritten, e := fmt.Fprintf(w, "p%v [style=filled; fillcolor=%v]\n", p.IndividualID, Blue())
			n += nwritten
			if e != nil {
				return n, e
			}
		} else if p.Sex == 2 {
			nwritten, e := fmt.Fprintf(w, "p%v [style=filled; fillcolor=%v]\n", p.IndividualID, Red())
			n += nwritten
			if e != nil {
				return n, e
			}
		} else if p.Sex == 0 {
		} else {
			fmt.Fprintf(os.Stderr, "weird sex: %v; %v\n", p.Sex, p)
		}
	}
	nwritten, e = fmt.Fprintf(w, "}\n")
	n += nwritten
	if e != nil {
		return n, e
	}

	return n, nil
}

type NamedCluster struct {
	ID int64
	Cluster []PedEntry
}

func (c NamedCluster) MaleFrac() float64 {
	nmales := 0
	printit := false
	for _, pe := range c.Cluster {
		if pe.Sex == 1 {
			nmales++
		}
		if pe.IndividualID == 16 {
			printit = true
		}
	}
	if printit {
		fmt.Fprintln(os.Stderr, "16 family:", c)
	}
	return float64(nmales) / float64(len(c.Cluster))
}

func SortClusters(clusters map[int64][]PedEntry) []NamedCluster {
	out := make([]NamedCluster, 0, len(clusters))
	for cid, cps := range clusters {
		out = append(out, NamedCluster{cid, cps})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ID >= 10000 && out[j].ID < 10000 {
			return true
		}
		if out[j].ID >= 10000 && out[i].ID < 10000 {
			return false
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func ClusterYs(focalID int64, tree map[int64]Node, ps ...PedEntry) (unclustered []PedEntry, clusters []NamedCluster) {
	mclusters := make(map[int64][]PedEntry)
	for _, p := range ps {
		if HasY(tree[p.PaternalID].PedEntry, focalID, tree) {
			mclusters[p.PaternalID] = append(mclusters[p.PaternalID], p)
		} else {
			unclustered = append(unclustered, p)
		}
	}
	return unclustered, SortClusters(mclusters)
}

func ShouldPrint(p PedEntry, focalID int64, tree map[int64]Node, opts GraphVizOpts) bool {
	if !opts.StripUninf {
		return true
	}
	return DadHasY(p, focalID, tree) || HasY(p, focalID, tree) || (p.PaternalID == 0 && p.MaternalID == 0)
}

func ShouldPrintAParent(p PedEntry, focalID int64, tree map[int64]Node, opts GraphVizOpts) bool {
	if !opts.StripUninf {
		return true
	}
	if p.PaternalID != 0 {
		if d, ok := tree[p.PaternalID]; ok {
			if ShouldPrint(d.PedEntry, focalID, tree, opts) {
				return true
			}
		}
	}

	if p.MaternalID != 0 {
		if m, ok := tree[p.MaternalID]; ok {
			if ShouldPrint(m.PedEntry, focalID, tree, opts) {
				return true
			}
		}
	}

	return false
}

func PedEntryToGraphVizY(w io.Writer, focalID int64, tree map[int64]Node, p PedEntry, opts GraphVizOpts) (n int, err error) {
	if p.PaternalID != 0 {
		extra := " [style=dotted]"
		if HasY(p, focalID, tree) {
			extra = ""
		}
		nwritten, e := fmt.Fprintf(w, "p%v -> p%v%v\np%v [style=filled; fillcolor=%v]\n", p.PaternalID, p.IndividualID, extra, p.PaternalID, Blue())
		n += nwritten
		if e != nil {
			return n, e
		}
	}
	if p.MaternalID != 0 {
		nwritten, e := fmt.Fprintf(w, "p%v -> p%v [style=dotted]\np%v [style=filled; fillcolor=%v]\n", p.MaternalID, p.IndividualID, p.MaternalID, Red())
		n += nwritten
		if e != nil {
			return n, e
		}
	}
	if p.Sex == 1 {
		nwritten, e := fmt.Fprintf(w, "p%v [style=filled; fillcolor=%v]\n", p.IndividualID, Blue())
		n += nwritten
		if e != nil {
			return n, e
		}
	} else if p.Sex == 2 {
		nwritten, e := fmt.Fprintf(w, "p%v [style=filled; fillcolor=%v]\n", p.IndividualID, Red())
		n += nwritten
		if e != nil {
			return n, e
		}
	} else if p.Sex == 0 {
	} else {
		fmt.Fprintf(os.Stderr, "weird sex: %v; %v\n", p.Sex, p)
	}

	return n, nil
}

func PedEntryToGraphVizYShape(w io.Writer, focalID int64, tree map[int64]Node, p PedEntry, opts GraphVizOpts) (n int, err error) {
	printit := ShouldPrint(p, focalID, tree, opts)
	printparent := ShouldPrintAParent(p, focalID, tree, opts)
	fmt.Fprintf(os.Stderr, "printit: %v; printparent: %v\n", printit, printparent)


	if p.PaternalID != 0 && printparent {
		myid := fmt.Sprintf("%v", p.IndividualID)
		if !printit {
			myid = fmt.Sprintf("x%v", p.PaternalID)
		}

		extra := " [style=dotted]"
		if HasY(p, focalID, tree) {
			extra = ""
		}
		nwritten, e := fmt.Fprintf(w, "p%v -> p%v%v\np%v [style=filled%v]\n", p.PaternalID, myid, extra, p.PaternalID, MaleAes())
		n += nwritten
		if e != nil {
			return n, e
		}
	}
	if p.MaternalID != 0 && printparent {
		myid := fmt.Sprintf("%v", p.IndividualID)
		if !printit {
			myid = fmt.Sprintf("x%v", p.MaternalID)
		}

		nwritten, e := fmt.Fprintf(w, "p%v -> p%v [style=dotted]\np%v [style=filled%v]\n", p.MaternalID, myid, p.MaternalID, FemAes())
		n += nwritten
		if e != nil {
			return n, e
		}
	}
	if p.Sex == 1 && printit {
		nwritten, e := fmt.Fprintf(w, "p%v [style=filled%v]\n", p.IndividualID, MaleAes())
		n += nwritten
		if e != nil {
			return n, e
		}
	} else if p.Sex == 2 && printit {
		nwritten, e := fmt.Fprintf(w, "p%v [style=filled%v]\n", p.IndividualID, FemAes())
		n += nwritten
		if e != nil {
			return n, e
		}
	} else if printparent && !printit {
		if p.MaternalID != 0 {
			nwritten, e := fmt.Fprintf(w, "px%v [shape = plaintext; label = \"...\"]\n", p.MaternalID)
			n += nwritten
			if e != nil {
				return n, e
			}
		}
		if p.PaternalID != 0 {
			nwritten, e := fmt.Fprintf(w, "px%v [shape = plaintext; label = \"...\"]\n", p.PaternalID)
			n += nwritten
			if e != nil {
				return n, e
			}
		}
	} else if p.Sex == 0 {
	} else {
		fmt.Fprintf(os.Stderr, "weird sex: %v; %v\n", p.Sex, p)
	}

	return n, nil
}

type PedEncodeFunc func(w io.Writer, focalID int64, tree map[int64]Node, p PedEntry, opts GraphVizOpts) (n int, err error)

func GetStylePedFunc(style string) PedEncodeFunc {
	switch style {
	case "YShape": return PedEntryToGraphVizYShape
	case "Y": return PedEntryToGraphVizY
	default: return PedEntryToGraphVizY
	}
}

func Percentify(f float64) string {
	return fmt.Sprintf("%.2f%%", f * 100.0)
}

func ToGraphVizY(w io.Writer, opts GraphVizOpts, ps ...PedEntry) (n int, err error) {
	f := opts.FocalID
	tree := BuildPedTree(ps...)

	pedfunc := GetStylePedFunc(opts.Style)

	unclustered, clusters := ClusterYs(f, tree, ps...)

	nwritten, e := fmt.Fprintf(w, "digraph full {\n")
	n += nwritten
	if e != nil { return n, e }

	for _, p := range unclustered {
		nwritten, e := fmt.Fprintf(w, "p%v\n", p.IndividualID)
		n += nwritten
		if e != nil { return n, e }
	}

	for _, cps := range clusters {
		cid := cps.ID
		malefrac := cps.MaleFrac()

		nwritten, e := fmt.Fprintf(w, "subgraph cluster_%v {\nlabel = \"%v\"\n", cid, Percentify(malefrac))
		n += nwritten
		if e != nil { return n, e }

		for _, p := range cps.Cluster {
			nwritten, e := pedfunc(w, f, tree, p, opts)
			n += nwritten
			if e != nil { return n, e }
		}

		nwritten, e = fmt.Fprintf(w, "}\n")
		n += nwritten
		if e != nil { return n, e }
	}

	for _, p := range unclustered {
		nwritten, e := pedfunc(w, f, tree, p, opts)
		n += nwritten
		if e != nil { return n, e }
	}

	nwritten, e = fmt.Fprintf(w, "}\n")
	n += nwritten
	if e != nil { return n, e }

	return n, nil
}

func ToGraphViz(w io.Writer, opts GraphVizOpts, ps ...PedEntry) (n int, err error) {
	switch opts.Style {
	case "YShape":
		fallthrough
	case "Y":
		return ToGraphVizY(w, opts, ps...)
	case "X":
		return ToGraphVizX(w, opts, ps...)
	default:
		return ToGraphVizSimple(w, ps...)
	}
}

type GraphVizOpts struct {
	Style string
	FocalID int64
	StripUninf bool
}

func GetOpts() GraphVizOpts {
	g := GraphVizOpts{}
	var f int
	flag.StringVar(&g.Style, "s", "", "Style for printing (try Y or X)")
	flag.IntVar(&f, "f", -1, "Focal ID")
	flag.BoolVar(&g.StripUninf, "strip", false, "Strip offspring that do not contribute to Y test")
	flag.Parse()
	g.FocalID = int64(f)
	return g
}

func FullToGraphViz() {
	opts := GetOpts()
	ps, e := ParsePed(os.Stdin)
	if e != nil {
		panic(e)
	}
	_, e = ToGraphViz(os.Stdout, opts, ps...)
	if e != nil {
		panic(e)
	}
}
