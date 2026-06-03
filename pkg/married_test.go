package tdt

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

const exampleMarriedPed = `1	1	0	0	1	2		0.3	0.5	0.2	banana
1	2	0	0	2	1		0.3	0.5	0.2	banana
1	3	1	2	1	2		0.3	0.5	0.2	banana
1	4	1	2	1	2		0.3	0.4	0.2	banana
1	5	0	0	2	1		0.3	0.2	0.2	banana
1	6	4	5	1	2		0.3	0.1	0.2	banana
1	1001	0	0	1	2		0.3	0.5	0.2	banana
1	1004	0	0	2	1		0.3	0.5	0.2	banana
1	1007	0	0	2	1		0.3	0.5	0.2	banana`

func TestFlipSingles(t *testing.T) {
	ped, err := ParsePedFromReader(strings.NewReader(exampleMarriedPed))

	if err != nil {
		t.Error(err)
	}

	tree := BuildPedTree(ped...)
	clusters, cluster_order := ClusterParents(ped)
	nclustered := 0
	for _, i := range cluster_order {
		nclustered += len(clusters[i])
	}
	t.Logf("num individuals %v\n", len(tree))
	t.Logf("nclustered %v\n", nclustered)
	t.Logf("clusters: %v\n", clusters)
	t.Logf("cluster_order: %v\n", cluster_order)

	rng := rand.New(rand.NewSource(2))
	out := FlipClustersSexes(clusters, cluster_order, ped, tree, rng, true)

	var b strings.Builder
	fmt.Fprintln(&b, "")
	WritePed(&b, CanonicalTree(out))
	t.Log(b.String())
}
