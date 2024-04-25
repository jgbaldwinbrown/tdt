package tdt

func HasFather(p PedEntry, tree map[string]Node) bool {
	_, ok := tree[p.PaternalID]
	if p.IndividualID == "0" || p.IndividualID == "999999" {
		return false
	}
	return ok
}

func FindFocals(ped ...PedEntry) (orphanFocal []PedEntry, nonOrphanFocal []PedEntry) {
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
