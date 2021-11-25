package bsegtree

import (
	"math"
	"sort"
)

type node struct {
	from uint64
	to   uint64

	left, right *node

	overlap []Interval
}

func (n *node) CompareTo(other Interval) int {

	if other.From > n.to || other.To < n.from {
		return DISJOINT
	}

	if other.From <= n.from && other.To >= n.to {
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

type Interval struct {
	ID   int // unique
	From uint64
	To   uint64
}

// Disjoint returns true if Segment does not overlap with interval
func (p Interval) Disjoint(from, to uint64) bool {

	if from > p.To || to < p.From {
		return true
	}
	return false
}

// Endpoints returns a slice with all endpoints (sorted, unique)
func Endpoints(base []Interval) (result []uint64, min, max uint64) {
	baseLen := len(base)
	points := make([]uint64, baseLen*2)
	for i, interval := range base {
		points[i] = interval.From
		points[i+baseLen] = interval.To
	}
	result = Dedup(points)
	min = result[0]
	max = result[len(result)-1]
	return
}

// Creates a slice of elementary intervals from a slice of (sorted) endpoints
// Input: [p1, p2, ..., pn]
// Output: [{p1 : p1}, {p1 : p2}, {p2 : p2},... , {pn : pn}
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
func (n *node) insertInterval(i Interval) {

	if n.CompareTo(i) == SUBSET {
		// interval of node is a subset of the specified interval or equal
		if n.overlap == nil {
			n.overlap = make([]Interval, 0, 2)
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
}

// round rounds a float64 and cuts it by n.
// n: decimal places.
// e.g.
// f = 1.001, n = 2, return 1.00
func round(f float64, n int) float64 {
	pow10n := math.Pow10(n)
	return math.Trunc(f*pow10n+0.5) / pow10n
}