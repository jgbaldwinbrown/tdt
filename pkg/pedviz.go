package tdt

import (
	"sort"
	"flag"
	"io"
	"os"
	"fmt"
)

func Red() string {
	return `"#cc8888"`
}

func Blue() string {
	return `"#aaaaee"`
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
		if p.PaternalID != "0" {
			nwritten, e := fmt.Fprintf(w, "p%v -> p%v\np%v [style=filled; fillcolor=%v]\n", p.PaternalID, p.IndividualID, p.PaternalID, Blue())
			n += nwritten
			if e != nil {
				return n, e
			}
		}
		if p.MaternalID != "0" {
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
	ID string
	Cluster []PedEntry
}

func (c NamedCluster) MaleFrac() float64 {
	return MaleFrac(c.Cluster)
}

func MaleFrac(ps []PedEntry) float64 {
	nmales := 0
	printit := false
	for _, pe := range ps {
		if pe.Sex == 1 {
			nmales++
		}
		if pe.IndividualID == "16" {
			printit = true
		}
	}
	if printit {
		fmt.Fprintln(os.Stderr, "16 family:", ps)
	}
	return float64(nmales) / float64(len(ps))
}

func SortClusters(clusters map[string][]PedEntry) []NamedCluster {
	out := make([]NamedCluster, 0, len(clusters))
	for cid, cps := range clusters {
		out = append(out, NamedCluster{cid, cps})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ID >= "10000" && out[j].ID < "10000" {
			return true
		}
		if out[j].ID >= "10000" && out[i].ID < "10000" {
			return false
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func ClusterYs(focalID string, tree map[string]Node, ps ...PedEntry) (unclustered []PedEntry, clusters []NamedCluster) {
	mclusters := make(map[string][]PedEntry)
	for _, p := range ps {
		if HasY(tree[p.PaternalID].PedEntry, focalID, tree) {
			mclusters[p.PaternalID] = append(mclusters[p.PaternalID], p)
		} else {
			unclustered = append(unclustered, p)
		}
	}
	return unclustered, SortClusters(mclusters)
}

func ShouldPrint(p PedEntry, focalID string, tree map[string]Node, opts GraphVizOpts) bool {
	if !opts.StripUninf {
		return true
	}
	return DadHasY(p, focalID, tree) || HasY(p, focalID, tree) || (p.PaternalID == "0" && p.MaternalID == "0")
}

func ShouldPrintAParent(p PedEntry, focalID string, tree map[string]Node, opts GraphVizOpts) bool {
	if !opts.StripUninf {
		return true
	}
	if p.PaternalID != "0" {
		if d, ok := tree[p.PaternalID]; ok {
			if ShouldPrint(d.PedEntry, focalID, tree, opts) {
				return true
			}
		}
	}

	if p.MaternalID != "0" {
		if m, ok := tree[p.MaternalID]; ok {
			if ShouldPrint(m.PedEntry, focalID, tree, opts) {
				return true
			}
		}
	}

	return false
}

func PedEntryToGraphVizY(w io.Writer, focalID string, tree map[string]Node, p PedEntry, opts GraphVizOpts, prevparent *Set[string], prevparentpair *Set[Int64Pair]) (n int, err error) {
	if p.PaternalID != "0" {
		extra := " [color = \"#888888\"]"
		// extra := " [style=dotted]"
		if HasY(p, focalID, tree) {
			extra = ""
		}
		nwritten, e := fmt.Fprintf(w, "p%v -> p%v%v\np%v [style=filled; fillcolor=%v]\n", p.PaternalID, p.IndividualID, extra, p.PaternalID, Blue())
		n += nwritten
		if e != nil {
			return n, e
		}
	}
	if p.MaternalID != "0" {
		nwritten, e := fmt.Fprintf(w, "p%v -> p%v [color = \"#888888\"]\np%v [style=filled; fillcolor=%v]\n", p.MaternalID, p.IndividualID, p.MaternalID, Red())
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

type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	s := new(Set[T])
	s.m = make(map[T]struct{})
	return s
}

func (s *Set[T]) Add(val T) {
	s.m[val] = struct{}{}
}

func (s *Set[T]) Contains(val T) bool {
	_, ok := s.m[val]
	return ok
}

func PedEntryToGraphVizYShape(w io.Writer, focalID string, tree map[string]Node, p PedEntry, opts GraphVizOpts, prevparent *Set[string], prevparentpair *Set[Int64Pair]) (n int, err error) {
	labeltxt := `, label = ""`
	if opts.LabelNumber {
		labeltxt = ""
	}

	printit := ShouldPrint(p, focalID, tree, opts)
	printparent := ShouldPrintAParent(p, focalID, tree, opts)
	fmt.Fprintf(os.Stderr, "printit: %v; printparent: %v\n", printit, printparent)

	if (!printit) {
		if ((prevparent.Contains(p.MaternalID) || p.MaternalID == "0") &&
			(prevparent.Contains(p.PaternalID) || p.PaternalID == "0")) {
			return n, err
		}
	}


	if p.PaternalID != "0" && printparent && p.MaternalID != "0" {
		if !prevparentpair.Contains(Int64Pair{p.PaternalID, p.MaternalID}) {
			if printit {
				nwritten, e := fmt.Fprintf(w, "p%v ->px%vmx%v\np%v [style = filled%v%v]\n", p.PaternalID, p.PaternalID, p.MaternalID, p.PaternalID, MaleAes(), labeltxt)
				n += nwritten
				if e != nil {
					return n, e
				}
				nwritten, e = fmt.Fprintf(w, "p%v ->px%vmx%v\np%v [style = filled%v%v]\n", p.MaternalID, p.PaternalID, p.MaternalID, p.MaternalID, FemAes(), labeltxt)
				n += nwritten
				if e != nil {
					return n, e
				}

				nwritten, e = fmt.Fprintf(w, "px%vmx%v [shape = point, width = 0.06, height = 0.06]\n", p.PaternalID, p.MaternalID)
				n += nwritten
				if e != nil {
					return n, e
				}
			} else {
				nwritten, e := fmt.Fprintf(w, "p%v [style = \"filled\"%v]\n", p.PaternalID, labeltxt)
				n += nwritten
				if e != nil {
					return n, e
				}
				nwritten, e = fmt.Fprintf(w, "p%v [style = \"filled\"%v]\n", p.MaternalID, labeltxt)
				n += nwritten
				if e != nil {
					return n, e
				}
			}
		}

		// extra := " [style=dotted]"
		extra := " [color = \"#888888\"]"
		if HasY(p, focalID, tree) {
			extra = ""
		}
		// nwritten, e := fmt.Fprintf(w, "p%v -> p%v%v\np%v [style=filled%v]\n", p.PaternalID, myid, extra, p.PaternalID, MaleAes())

		if printit {
			nwritten, e := fmt.Fprintf(w, "px%vmx%v -> p%v%v\n", p.PaternalID, p.MaternalID, p.IndividualID, extra)
			n += nwritten
			if e != nil {
				return n, e
			}
		}
		prevparentpair.Add(Int64Pair{p.PaternalID, p.MaternalID})
	} else if p.PaternalID != "0" && printparent {
		if !prevparent.Contains(p.PaternalID) {
			if printit {
				nwritten, e := fmt.Fprintf(w, "p%v ->px%v\np%v [style = filled%v%v]\n", p.PaternalID, p.PaternalID, p.PaternalID, MaleAes(), labeltxt)
				n += nwritten
				if e != nil {
					return n, e
				}

				nwritten, e = fmt.Fprintf(w, "px%v [shape = point, width = 0.06, height = 0.06]\n", p.PaternalID)
				n += nwritten
				if e != nil {
					return n, e
				}
			} else {
				nwritten, e := fmt.Fprintf(w, "p%v [style = \"filled\"%v]\n", p.PaternalID, labeltxt)
				n += nwritten
				if e != nil {
					return n, e
				}
			}
		}

		// extra := " [style=dotted]"
		extra := " [color = \"#888888\"]"
		if HasY(p, focalID, tree) {
			extra = ""
		}
		// nwritten, e := fmt.Fprintf(w, "p%v -> p%v%v\np%v [style=filled%v]\n", p.PaternalID, myid, extra, p.PaternalID, MaleAes())

		if printit {
			nwritten, e := fmt.Fprintf(w, "px%v -> p%v%v\n", p.PaternalID, p.IndividualID, extra)
			n += nwritten
			if e != nil {
				return n, e
			}
		}
		prevparent.Add(p.PaternalID)
	} else if p.MaternalID != "0" && printparent {
		if !prevparent.Contains(p.MaternalID) {

			if printit {
				nwritten, e := fmt.Fprintf(w, "p%v ->px%v\np%v [style = filled%v%v]\n", p.MaternalID, p.MaternalID, p.MaternalID, FemAes(), labeltxt)
				n += nwritten
				if e != nil {
					return n, e
				}
				nwritten, e = fmt.Fprintf(w, "px%v [shape = point, width = 0.06, height = 0.06]\n", p.MaternalID)
				n += nwritten
				if e != nil {
					return n, e
				}
			} else {
				nwritten, e := fmt.Fprintf(w, "p%v [style = \"filled,dashed\"%v]\n", p.MaternalID, labeltxt)
				n += nwritten
				if e != nil {
					return n, e
				}
			}
		}

		// extra := " [style=dotted]"
		extra := " [color = \"#888888\"]"
		// nwritten, e := fmt.Fprintf(w, "p%v -> p%v%v\np%v [style=filled%v]\n", p.MaternalID, myid, extra, p.MaternalID, FemAes())

		if printit {
			nwritten, e := fmt.Fprintf(w, "px%v -> p%v%v\n", p.MaternalID, p.IndividualID, extra)
			n += nwritten
			if e != nil {
				return n, e
			}
		}
		prevparent.Add(p.MaternalID)
	}
	if p.Sex == 1 && printit {
		nwritten, e := fmt.Fprintf(w, "p%v [style=filled%v%v]\n", p.IndividualID, MaleAes(), labeltxt)
		n += nwritten
		if e != nil {
			return n, e
		}
	} else if p.Sex == 2 && printit {
		nwritten, e := fmt.Fprintf(w, "p%v [style=filled%v%v]\n", p.IndividualID, FemAes(), labeltxt)
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

type PedEncodeFunc func(w io.Writer, focalID string, tree map[string]Node, p PedEntry, opts GraphVizOpts, prevparent *Set[string], prevparentpair *Set[Int64Pair]) (n int, err error)

func GetStylePedFunc(style string) PedEncodeFunc {
	switch style {
	case "YShape": return PedEntryToGraphVizYShape
	case "Y": return PedEntryToGraphVizY
	default: return PedEntryToGraphVizY
	}
}

func Percentify(f float64) string {
	return fmt.Sprintf("%.0f%%", f * 100.0)
}

type Int64Pair struct {
	I0 string
	I1 string
}

func ToGraphVizY(w io.Writer, opts GraphVizOpts, ps ...PedEntry) (n int, err error) {
	prevparent := NewSet[string]()
	prevparentpair := NewSet[Int64Pair]()
	f := opts.FocalID
	tree := BuildPedTree(ps...)

	pedfunc := GetStylePedFunc(opts.Style)

	unclustered, clusters := ClusterYs(f, tree, ps...)

	nwritten, e := fmt.Fprintf(w, `digraph full {
fontsize = 24;
graph [ranksep="0.5"];
overlap = true;
splines = true;
node [width = 0.1, height = 0.5, margin = 0.03];
edge [dir = none];
`)
	// graph [pad="0.5", nodesep="1", ranksep="2"];

	// nwritten, e := fmt.Fprintf(w, "digraph full {\n")
	n += nwritten
	if e != nil { return n, e }

	for _, p := range unclustered {
		if ShouldPrint(p, f, tree, opts) {
			nwritten, e := fmt.Fprintf(w, "p%v\n", p.IndividualID)
			n += nwritten
			if e != nil { return n, e }
		}
	}

	for _, cps := range clusters {
		cid := cps.ID
		malefrac := cps.MaleFrac()

		nwritten, e := fmt.Fprintf(w, `subgraph cluster_%v {
label = "%v"
labeljust = "l"
labelloc = "l"
graph [color = "#888888"]
`, cid, Percentify(malefrac))
		n += nwritten
		if e != nil { return n, e }

		for _, p := range cps.Cluster {
			nwritten, e := pedfunc(w, f, tree, p, opts, prevparent, prevparentpair)
			n += nwritten
			if e != nil { return n, e }
		}

		nwritten, e = fmt.Fprintf(w, "}\n")
		n += nwritten
		if e != nil { return n, e }
	}

	for _, p := range unclustered {
		nwritten, e := pedfunc(w, f, tree, p, opts, prevparent, prevparentpair)
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
	FocalID string
	StripUninf bool
	LabelNumber bool
}

func GetOpts() GraphVizOpts {
	g := GraphVizOpts{}
	var f string
	flag.StringVar(&g.Style, "s", "", "Style for printing (try Y or X)")
	flag.StringVar(&f, "f", "-1", "Focal ID")
	flag.BoolVar(&g.StripUninf, "strip", false, "Strip offspring that do not contribute to Y test")
	flag.BoolVar(&g.LabelNumber, "l", false, "Add number labels to nodes")
	flag.Parse()
	g.FocalID = f
	return g
}

func FullToGraphViz() {
	opts := GetOpts()
	ps, e := ParsePedFromReader(os.Stdin)
	if e != nil {
		panic(e)
	}
	_, e = ToGraphViz(os.Stdout, opts, ps...)
	if e != nil {
		panic(e)
	}
}
