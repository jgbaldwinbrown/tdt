package tdt

import (
	"testing"
	"reflect"
)

func examplePed() []PedEntry {
	return []PedEntry {
		PedEntry{"1", "1", "0", "0", 1, 1},
		PedEntry{"1", "2", "1", "0", 1, 1},
		PedEntry{"1", "3", "1", "2", 1, 1},
		PedEntry{"1", "4", "1", "0", 1, 1},
		PedEntry{"1", "5", "0", "0", 1, 1},
	}
}

func TestFindEmbeddingFamily(t *testing.T) {
	tree := BuildPedTree(examplePed()...)
	fam := FindEmbeddingFamily(tree, []string{"3"})
	expect := map[string]struct{}{}
	for _, s := range []string{"1", "2", "3", "4"} {
		expect[s] = struct{}{}
	}
	if !reflect.DeepEqual(fam, expect) {
		t.Errorf("fam %v != expect %v", fam, expect)
	}
}
