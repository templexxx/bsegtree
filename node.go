package bsegtree

import (
	"sort"
)

type node struct {
	from uint64
	to   uint64

	left, right *node

	overlap []interval
}

func (n *node) CompareTo(other interval) int {

	if other.from > n.to || other.to < n.from {
		return DISJOINT
	}

	if other.from <= n.from && other.to >= n.to {
		return SUBSET
	}

	return INTERSECT_OR_SUPERSET
}

func (n *node) Disjoint(from, to uint64) bool {

	if from > n.to || to < n.from {
		return true
	}
	return false
}

type interval struct {
	id   int // unique
	from uint64
	to   uint64
}

// Disjoint returns true if Segment does not overlap with interval
func (p interval) Disjoint(from, to uint64) bool {

	if from > p.to || to < p.from {
		return true
	}
	return false
}

// Endpoints returns a slice with all endpoints (sorted, unique)
func Endpoints(base []interval) (result []uint64, min, max uint64) {
	baseLen := len(base)
	endpoints := make([]uint64, baseLen*2)
	for i, interval := range base {
		endpoints[i] = interval.from
		endpoints[i+baseLen] = interval.to
	}
	result = Dedup(endpoints)
	min = result[0]
	max = result[len(result)-1]
	return
}

func elementaryIntervals(endpoints []uint64) [][2]uint64 {
	if len(endpoints) == 1 {
		return [][2]uint64{{endpoints[0], endpoints[0]}}
	}

	intervals := make([][2]uint64, len(endpoints)*2-1)

	for i := 0; i < len(endpoints); i++ {
		intervals[i*2] = [2]uint64{endpoints[i], endpoints[i]}
		if i < len(endpoints)-1 {
			intervals[i*2+1] = [2]uint64{endpoints[i], endpoints[i+1]}
		}
	}
	return intervals
}

type endpoints []uint64

func (e endpoints) Len() int {
	return len(e)
}

func (e endpoints) Less(i, j int) bool {

	return e[i] < e[j]
}

func (e endpoints) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

// Dedup removes duplicates from a given slice
func Dedup(e []uint64) []uint64 {

	sort.Sort(endpoints(e))

	cnt := len(e)
	cntDup := 0
	for i := 1; i < cnt; i++ {

		if e[i] == e[i-1] {
			cntDup++
		} else {
			e[i-cntDup] = e[i]
		}
	}

	return e[:cnt-cntDup]
}

// Inserts interval into given tree structure
func (n *node) insertInterval(i interval) {

	if n.CompareTo(i) == SUBSET {
		// interval of node is a subset of the specified interval or equal
		if n.overlap == nil {
			n.overlap = make([]interval, 0, 2)
		}
		n.overlap = append(n.overlap, i)

	} else {
		if n.left != nil && n.left.CompareTo(i) != DISJOINT {
			n.left.insertInterval(i)
		}
		if n.right != nil && n.right.CompareTo(i) != DISJOINT {
			n.right.insertInterval(i)
		}
	}

	// switch node.CompareTo(n) {
	// case SUBSET:
	// 	// interval of node is a subset of the specified interval or equal
	// 	if node.overlap == nil {
	// 		node.overlap = make([]interval, 0, 2)
	// 	}
	//
	// 	node.overlap = append(node.overlap, n)
	//
	// 	if n.id == 369 || n.id == 457 {
	//
	// 		fmt.Println(node, n, "fuck2")
	// 	}
	//
	// case INTERSECT_OR_SUPERSET:
	// 	// interval of node is a superset, have to look in both children
	// 	if node.left != nil {
	// 		insertInterval(node.left, n)
	// 	}
	// 	if node.right != nil {
	// 		insertInterval(node.right, n)
	// 	}
	// case DISJOINT:
	// 	// nothing to do
	// }
}
