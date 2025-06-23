package tdt

import (
	"testing"
	"fmt"
	"slices"
	"strings"

	"github.com/jgbaldwinbrown/iterh"
)

const exampleWarp = `1	1	0	0	1	1		0.3	0.5	0.2	banana
1	2	0	0	2	0		0.3	0.5	0.2	banana
1	3	1	2	1	1		0.3	0.5	0.2	banana
1	4	1	2	1	1		0.3	0.4	0.2	banana
1	5	0	0	2	0		0.3	0.2	0.2	banana
1	6	4	5	1	1		0.3	0.1	0.2	banana
1	1001	1002	1003	1	1		0.3	0.5	0.2	banana`

func TestRelativeClusters(t *testing.T) {
	warp, err := iterh.CollectWithError(ParsePed(strings.NewReader(exampleWarp), false))
	if err != nil {
		t.Error(err)
	}
	tree := BuildPedTree(slices.Collect(ToPedEntries(slices.Values(warp)))...)
	goods := slices.Collect(ToIndividualID(FilterPed(slices.Values(warp), 0.45)))
	clusters := RelativeClusters(tree, 1, goods...)
	fmt.Println("clusters:", clusters)
	fmt.Println("tree:", tree)
	fmt.Println("goods:", goods)
}
