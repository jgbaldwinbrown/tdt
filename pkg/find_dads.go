package tdt

import (
	"flag"
	"fmt"
	"os"
	"log"
	"bufio"
)

// Run all-Y TDT test on the command line
func FullFindDads() {
	focalID := flag.String("f", "", "IndividualID for focal individual (default is to do TDT for all males)")
	flag.Parse()
	if *focalID == "" {
		log.Fatal(fmt.Errorf("missing -f"))
	}

	r := bufio.NewReader(os.Stdin)
	peds, e := ParsePedSafe(r)
	Must(e)

	tree := BuildPedTree(peds...)

	node := tree[*focalID]
	for {
		if e := WritePedEntry(os.Stdout, node.PedEntry); e != nil {
			log.Fatal(e)
		}
		if IsOrphan(node.PaternalID) {
			break
		}
		node = tree[node.PaternalID]
	}
}
