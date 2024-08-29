package tdt

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jgbaldwinbrown/iterh"
	"github.com/jgbaldwinbrown/kolm/pkg"
	"io"
	"iter"
	"log"
	"os"
)

func ReadResultsJson(r io.Reader) iter.Seq2[TDTResult, error] {
	return func(y func(TDTResult, error) bool) {
		dec := json.NewDecoder(r)
		var err error
		for err != io.EOF {
			var j TDTResultJson
			err = dec.Decode(j)
			if err != nil && !y(TDTResult{}, err) {
				return
			}
			t := FromJson(j)
			if !y(t, err) {
				return
			}
		}
	}
}

func FullKolm() {
	ress, e := iterh.CollectWithError(ReadResultsJson(bufio.NewReader(os.Stdin)))
	if e != nil {
		log.Fatal(e)
	}
	chis := func(y func(float64) bool) {
		for _, res := range ress {
			if !y(res.Chisq) {
				return
			}
		}
	}

	k, e := kolm.KolmogorovSmirnovChi2(chis)
	if e != nil {
		log.Fatal(e)
	}
	fmt.Printf("%#v\n", k)
}
