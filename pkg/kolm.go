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

// Read the results of the TDT test as JSON
func ReadResultsJson(r io.Reader) iter.Seq2[TDTResult, error] {
	return func(y func(TDTResult, error) bool) {
		dec := json.NewDecoder(r)
		var err error
		for {
			var j TDTResultJson
			err = dec.Decode(&j)
			if err == io.EOF {
				return
			}
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

// Run the whole kolmogorov-smirnov test on a set of TDT results, comparing them to the chi-squared distribution
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
	fmt.Printf("%.20g\n", k.PValue)
}
