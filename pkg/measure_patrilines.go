package tdt

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
)

func FindPaternalTop(id string, tree map[string]Node) Node {
	n := tree[id]
	for !IsOrphan(n.PaternalID) {
		n = tree[n.PaternalID]
	}
	return n
}

func EntryToPedEntry(ent Entry) PedEntry {
	sex, e := strconv.ParseInt(ent.Sex, 0, 64)
	if e != nil {
		sex = 0
	}
	pheno, e := strconv.ParseInt(ent.Phenotype, 0, 64)
	if e != nil {
		pheno = 0
	}
	return PedEntry{
		FamilyID: ent.FamilyID,
		IndividualID: ent.IndividualID,
		PaternalID: ent.FatherID,
		MaternalID: ent.MotherID,
		Sex: sex,
		Phenotype: pheno,
	}
}

func EntriesToPed(ents []Entry) []PedEntry {
	out := make([]PedEntry, 0, len(ents))
	for _, ent := range ents {
		out = append(out, EntryToPedEntry(ent))
	}
	return out
}

func MeasurePatriline(ents []Entry, likelihoodThresh float64) ([]TDTResult, bool) {
	outlier := GetBiggestOutlier(slices.Values(ents))
	if outlier.Posterior < likelihoodThresh {
		return nil, false
	}

	ped := EntriesToPed(ents)
	tree := BuildPedTree(ped...)
	focalNode := FindPaternalTop(outlier.IndividualID, tree)
	focal := focalNode.IndividualID

	out := make([]TDTResult, 0, 4)

	res := TDTTest(BuildFamiliesFemaleX(focal, ped...)...)
	res.Name = "FemaleX"
	out = append(out, res)

	res = TDTTest(BuildFamiliesFemDescentFemaleX(focal, ped...)...)
	res.Name = "FemDescentFemaleX"
	out = append(out, res)

	res = TDTTest(BuildFamiliesY(focal, ped...)...)
	res.Name = "Y"
	out = append(out, res)

	res = TDTTest(BuildFamiliesAuto(focal, ped...)...)
	res.Name = "Auto"
	out = append(out, res)

	return out, true
}

func AverageResults(results []TDTResult) TDTResult {
	var out TDTResult
	for _, r := range results {
		out.Name = r.Name
		out.Totals.MaleF1 += r.Totals.MaleF1
		out.Totals.FemaleF1 += r.Totals.FemaleF1
		out.Nfamilies += r.Nfamilies
		out.MaleProportion += r.MaleProportion
		out.MeanMalesPerFam += r.MeanMalesPerFam
		out.MeanFemalesPerFam += r.MeanFemalesPerFam
		out.MeanChildrenPerFam += r.MeanChildrenPerFam
		out.Chisq += r.Chisq
		out.P += r.P
		out.Orphan = r.Orphan
	}

	l := float64(len(results))
	out.Totals.MaleF1 /= l
	out.Totals.FemaleF1 /= l
	out.Nfamilies /= l
	out.MaleProportion /= l
	out.MeanMalesPerFam /= l
	out.MeanFemalesPerFam /= l
	out.MeanChildrenPerFam /= l
	out.Chisq /= l
	out.P /= l

	return out
}

type MeasurePatrilinesFlags struct {
	LikelihoodThreshold float64
}

func MeasurePatrilinesFull() {
	var f MeasurePatrilinesFlags
	flag.Float64Var(&f.LikelihoodThreshold, "l", -1.0, "Minimum threshold for highest likelihood to include in results")
	flag.Parse()

	paths, e := readReaderLines(os.Stdin)
	if e != nil {
		log.Fatal(e)
	}
	allResults := make([][]TDTResult, 0, len(paths))
	for _, path := range paths {
		entsit := ParsePedPath(path, true)
		ents := []Entry{}
		for ent, e := range entsit {
			if e != nil {
				log.Fatal(e)
			}
			ents = append(ents, ent)
		}
		
		res, aboveThresh := MeasurePatriline(ents, f.LikelihoodThreshold)
		if aboveThresh {
			allResults = append(allResults, res)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	for _, set := range allResults {
		for _, res := range set {
			if res.Name == "Y" {
				e := enc.Encode(res)
				if e != nil {
					log.Fatal(e)
				}
			}
		}
	}
}

func AveragePatrilinesFull() {
	dec := json.NewDecoder(os.Stdin)
	results := []TDTResult{}
	for {
		var r TDTResult
		if e := dec.Decode(&r); e != nil {
			if e == io.EOF {
				break
			} else {
				log.Fatal(e)
			}
		}
		results = append(results, r)
	}
	avg := AverageResults(results)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	e := enc.Encode(avg)
	if e != nil {
		log.Fatal(e)
	}
}
