package tdt

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

func ParsePedSafe(r io.Reader) ([]PedEntry, error) {
	s := bufio.NewScanner(r)
	var ps []PedEntry
	for i := 0; s.Scan(); i++ {
		if s.Err() != nil {
			return nil, s.Err()
		}
		if ShouldSkipPedLine(s.Text()) {
			continue
		}

		p, e := ParsePedEntry(s.Text())
		if e != nil {
			log.Println(fmt.Errorf("ParsePedSafe: line %v: ped %v: %w", i, p, e))
		}
		ps = append(ps, p)
	}
	return ps, nil
}

