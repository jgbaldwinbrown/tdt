package tdt

import (
	"io"
	"os"
	"fmt"
)

func ToGraphViz(w io.Writer, ps ...PedEntry) (n int, err error) {
	red := `"#cc6666"`
	blue := `"#6666cc"`
	nwritten, e := fmt.Fprintf(w, "digraph full {\n")
	n += nwritten
	if e != nil {
		return n, e
	}

	for _, p := range ps {
		if p.PaternalID != 0 {
			nwritten, e := fmt.Fprintf(w, "p%v -> p%v\np%v [style=filled; fillcolor=%v]\n", p.PaternalID, p.IndividualID, p.PaternalID, blue)
			n += nwritten
			if e != nil {
				return n, e
			}
		}
		if p.MaternalID != 0 {
			nwritten, e := fmt.Fprintf(w, "p%v -> p%v\np%v [style=filled; fillcolor=%v]\n", p.MaternalID, p.IndividualID, p.MaternalID, red)
			n += nwritten
			if e != nil {
				return n, e
			}
		}
		if p.Sex == 1 {
			nwritten, e := fmt.Fprintf(w, "p%v [style=filled; fillcolor=%v]\n", p.IndividualID, blue)
			n += nwritten
			if e != nil {
				return n, e
			}
		} else if p.Sex == 2 {
			nwritten, e := fmt.Fprintf(w, "p%v [style=filled; fillcolor=%v]\n", p.IndividualID, red)
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

func FullToGraphViz() {
	ps, e := ParsePed(os.Stdin)
	if e != nil {
		panic(e)
	}
	_, e = ToGraphViz(os.Stdout, ps...)
	if e != nil {
		panic(e)
	}
}
