package tdt

// Check if p has a father in tree
func HasFather(p PedEntry, tree map[string]Node) bool {
	_, ok := tree[p.PaternalID]
	if p.IndividualID == "0" || p.IndividualID == "999999" {
		return false
	}
	return ok
}

// Deprecated
func findFocalsInconsistent(ped ...PedEntry) (orphanFocal []PedEntry, nonOrphanFocal []PedEntry) {
	tree := BuildPedTree(ped...)
	for _, p := range ped {
		if p.Sex != 1 {
			continue
		}
		if !HasFather(p, tree) {
			orphanFocal = append(orphanFocal, p)
		} else {
			nonOrphanFocal = append(nonOrphanFocal, p)
		}
	}
	return orphanFocal, nonOrphanFocal
}

// From the entries in ped, find all individuals that are male. If they have a father, put them in nonOrphanFocal otherwise, put them in orphanFocal
func FindFocals(ped ...PedEntry) (orphanFocal []PedEntry, nonOrphanFocal []PedEntry) {
	tree := BuildPedTree(ped...)
	for _, node := range tree {
		p := node.PedEntry
		if p.Sex != 1 {
			continue
		}
		if !HasFather(p, tree) {
			orphanFocal = append(orphanFocal, p)
		} else {
			nonOrphanFocal = append(nonOrphanFocal, p)
		}
	}
	return orphanFocal, nonOrphanFocal
}

// Remove duplicates from the pedigree
func UniqPed(ped ...PedEntry) []PedEntry {
	tree := BuildPedTree(ped...)
	out := make([]PedEntry, 0, len(tree))
	for _, node := range tree {
		out = append(out, node.PedEntry)
	}
	return out
}
