package tdt

import (
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

func ToGraphVizY(w io.Writer, opts GraphVizOpts, ps ...PedEntry) (n int, err error) {
	f := opts.FocalID
	tree := BuildPedTree(ps...)

	nwritten, e := fmt.Fprintf(w, "digraph full {\n")
	n += nwritten
	if e != nil {
		return n, e
	}

	for _, p := range ps {
		if p.PaternalID != 0 {
			extra := " [style=dotted]"
			if HasY(p, f, tree) {
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
	}

	nwritten, e = fmt.Fprintf(w, "}\n")
	n += nwritten
	if e != nil {
		return n, e
	}

	return n, nil
}

func ToGraphViz(w io.Writer, opts GraphVizOpts, ps ...PedEntry) (n int, err error) {
	if opts.Style == "Y" {
		return ToGraphVizY(w, opts, ps...)
	}

	return ToGraphVizSimple(w, ps...)
}

type GraphVizOpts struct {
	Style string
	FocalID int64
}

func GetOpts() GraphVizOpts {
	g := GraphVizOpts{}
	var f int
	flag.StringVar(&g.Style, "s", "", "Style for printing (try Y)")
	flag.IntVar(&f, "f", -1, "Focal ID")
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
