package main

// Immutable three-level component storage. A million values fit in 64 branches
// of 64 leaves of 256 scalar parts. Only constructors mutate unpublished nodes;
// every later edit copies one root, one branch and one leaf. Snapshots share
// immutable nodes and copy only the source descriptor/revision.
const maxStoredParts = 64 * 64 * 256

type partLeaf [256]part
type partBranch [64]*partLeaf
type partRoot [64]*partBranch

type partStore struct {
	root  *partRoot
	count int
}

func buildPartStore(count int, value func(int) part) partStore {
	if count < 0 || count > maxStoredParts {
		panic("component storage size outside bound")
	}
	s := partStore{root: new(partRoot), count: count}
	for i := 0; i < count; i++ {
		a, b := i>>14, (i>>8)&63
		if s.root[a] == nil {
			s.root[a] = new(partBranch)
		}
		if s.root[a][b] == nil {
			s.root[a][b] = new(partLeaf)
		}
		s.root[a][b][i&255] = value(i)
	}
	return s
}
func storeParts(values []part) partStore {
	return buildPartStore(len(values), func(i int) part { return values[i] })
}
func (s partStore) at(i int) part {
	if i < 0 || i >= s.count {
		panic("component index outside storage")
	}
	return s.root[i>>14][(i>>8)&63][i&255]
}
func (s partStore) with(i int, value part) partStore {
	if i < 0 || i >= s.count {
		panic("component index outside storage")
	}
	root := *s.root
	a, b := i>>14, (i>>8)&63
	branch := *root[a]
	leaf := *branch[b]
	leaf[i&255] = value
	branch[b] = &leaf
	root[a] = &branch
	return partStore{root: &root, count: s.count}
}

// Export only on a worker or in a fixture: this enumerates the full document.
func (s partStore) values() []part {
	out := make([]part, s.count)
	for i := range out {
		out[i] = s.at(i)
	}
	return out
}
