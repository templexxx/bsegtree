// Copyright 2012 Thomas Oberndörfer. All rights reserved.
// Copyright 2021 Temple3x. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsegtree

import (
	"bytes"
	"sort"
)

const (
	intervalCacheStartSize = 32 * 1024
)

type BSTree struct {
	count int // Number of intervals
	root  *node
	// Interval stack
	intervalCache []byte
	base          []Interval
	// Min value of all intervals
	min [2]int
	// Max value of all intervals
	max [2]int
}

type node struct {
	// A segment is a interval represented by the node
	segment     Segment
	left, right *node
	// All intervals that overlap with segment
	overlap []*Interval
}

func (n *node) Segment() Segment {
	return n.segment
}

func (n *node) Left() *node {
	return n.left
}

func (n *node) Right() *node {
	return n.right
}

// Overlap transforms []*Interval to []Interval
func (n *node) Overlap() []Interval {
	if n.overlap == nil {
		return nil
	}
	interval := make([]Interval, len(n.overlap))
	for i, pintrvl := range n.overlap {
		interval[i] = *pintrvl
	}
	return interval
}

type Interval struct {
	Id int // unique
	Segment
}

// IntervalBytes used as return value with origin from to in bytes.
type IntervalBytes struct {
	Id   int
	From []byte
	To   []byte
}

type Segment struct {
	From [2]int // Origin bytes is in interval cache, here only keep the offset & size.
	To   [2]int // Same as From.
}

// Represents overlapping intervals of a segment
type SegmentOverlap struct {
	Segment  Segment
	Interval []Interval
}

// Relations of two intervals
const (
	SUBSET = iota
	DISJOINT
	INTERSECT_OR_SUPERSET
)

// NewTree returns a Tree interface with underlying segment tree implementation
func NewTree() Tree {
	t := new(BSTree)
	t.Clear()
	return t
}

// Push new interval to stack
func (t *BSTree) Push(from, to []byte) {

	fromOff := len(t.intervalCache)
	t.intervalCache = append(t.intervalCache, from...)
	toOff := len(t.intervalCache)
	t.intervalCache = append(t.intervalCache, to...)

	t.base = append(t.base, Interval{t.count, Segment{[2]int{fromOff, len(from)}, [2]int{toOff, len(to)}}})
	t.count++
}

func (t *BSTree) PushArray(from, to [][]byte) {
	for i := 0; i < len(from); i++ {
		t.Push(from[i], to[i])
	}
}

// Clear the interval stack
func (t *BSTree) Clear() {
	t.count = 0
	t.root = nil
	t.base = make([]Interval, 0, 100)

	t.intervalCache = make([]byte, 0, intervalCacheStartSize)

	t.min = [2]int{}
	t.max = [2]int{}
}

func (t *BSTree) Build() {
	if len(t.base) == 0 {
		panic("No intervals in stack to build tree. Push intervals first")
	}
	var endpoint [][2]int
	endpoint, t.min, t.max = Endpoints(t, t.base)
	// Create tree nodes from interval endpoints
	t.root = t.insertNodes(endpoint)
	for i := range t.base {
		insertInterval(t.intervalCache, t.root, &t.base[i])
	}
}

// Endpoints returns a slice with all endpoints (sorted, unique)
func Endpoints(t *BSTree, base []Interval) (result [][2]int, min, max [2]int) {
	baseLen := len(base)
	positions := make([][2]int, baseLen*2)
	for i, interval := range base {
		positions[i] = interval.From
		positions[i+baseLen] = interval.To
	}
	result = Dedup(endpoints{positions: positions, cache: t.intervalCache})
	min = result[0]
	max = result[len(result)-1]
	return
}

type endpoints struct {
	positions [][2]int
	cache     []byte
}

func (e endpoints) Len() int {
	return len(e.positions)
}

func (e endpoints) Less(i, j int) bool {
	vi := e.cache[e.positions[i][0] : e.positions[i][0]+e.positions[i][1]]
	vj := e.cache[e.positions[j][0] : e.positions[j][0]+e.positions[j][1]]

	return bytes.Compare(vi, vj) == -1
}

func (e endpoints) Swap(i, j int) {
	e.positions[i], e.positions[j] = e.positions[j], e.positions[i]
}

// Dedup removes duplicates from a given slice
func Dedup(sl endpoints) [][2]int {

	if !sort.IsSorted(sl) {
		sort.Sort(sl)
	}

	cnt := len(sl.positions)
	cntDup := 0
	for i := 1; i < cnt; i++ {

		vi := sl.cache[sl.positions[i][0] : sl.positions[i][0]+sl.positions[i][1]]
		vip := sl.cache[sl.positions[i-1][0] : sl.positions[i-1][0]+sl.positions[i-1][1]]
		if bytes.Equal(vi, vip) {
			cntDup++
		} else {
			sl.positions[i-cntDup] = sl.positions[i]
		}
	}

	return sl.positions[:cnt-cntDup]
}

// insertNodes builds tree structure from given endpoints
func (t *BSTree) insertNodes(endpoint [][2]int) *node {
	var n *node
	if len(endpoint) == 1 {
		n = &node{segment: Segment{endpoint[0], endpoint[0]}}
		n.left = nil
		n.right = nil
	} else {
		n = &node{segment: Segment{endpoint[0], endpoint[len(endpoint)-1]}}
		center := len(endpoint) / 2
		n.left = t.insertNodes(endpoint[:center])
		n.right = t.insertNodes(endpoint[center:])
	}
	return n
}

// CompareTo compares two Segments and returns: DISJOINT, SUBSET or INTERSECT_OR_SUPERSET
func (s *Segment) CompareTo(cache []byte, other *Segment) int {
	if compareTo(cache, other.From, s.To) == 1 || compareTo(cache, other.To, s.From) == -1 {
		return DISJOINT
	}

	fromC := compareTo(cache, other.From, s.From)
	toC := compareTo(cache, other.To, s.To)
	if fromC < 1 && toC >= 0 {
		return SUBSET
	}

	return INTERSECT_OR_SUPERSET
}

func compareTo(cache []byte, pos0, pos1 [2]int) int {

	v0 := cache[pos0[0] : pos0[0]+pos0[1]]
	v1 := cache[pos1[0] : pos1[0]+pos1[1]]

	return bytes.Compare(v0, v1)
}

// Disjoint returns true if Segment does not overlap with interval
func (s *Segment) Disjoint(cache []byte, from, to []byte) bool {

	sto := cache[s.To[0] : s.To[0]+s.To[1]]
	if bytes.Compare(from, sto) == 1 || bytes.Compare(to, cache[s.From[0]:s.From[0]+s.From[1]]) == -1 {
		return true
	}
	return false
}

// Inserts interval into given tree structure
func insertInterval(cache []byte, node *node, intrvl *Interval) {
	switch node.segment.CompareTo(cache, &intrvl.Segment) {
	case SUBSET:
		// interval of node is a subset of the specified interval or equal
		if node.overlap == nil {
			node.overlap = make([]*Interval, 0, 10)
		}
		node.overlap = append(node.overlap, intrvl)
	case INTERSECT_OR_SUPERSET:
		// interval of node is a superset, have to look in both children
		if node.left != nil {
			insertInterval(cache, node.left, intrvl)
		}
		if node.right != nil {
			insertInterval(cache, node.right, intrvl)
		}
	case DISJOINT:
		// nothing to do
	}
}

// Query interval
func (t *BSTree) Query(from, to []byte) []IntervalBytes {
	if t.root == nil {
		panic("Can't run query on empty tree. Call Build() first")
	}
	result := make(map[int]IntervalBytes)
	querySingle(t.intervalCache, t.root, from, to, &result)
	// transform map to slice
	sl := make([]IntervalBytes, 0, len(result))
	for _, intrvl := range result {
		sl = append(sl, intrvl)
	}
	return sl
}

// querySingle traverse tree in search of overlaps
func querySingle(cache []byte, node *node, from, to []byte, result *map[int]IntervalBytes) {
	if !node.segment.Disjoint(cache, from, to) {
		for _, pintrvl := range node.overlap {
			(*result)[pintrvl.Id] = IntervalBytes{
				Id:   pintrvl.Id,
				From: cache[pintrvl.From[0] : pintrvl.From[0]+pintrvl.From[1]],
				To:   cache[pintrvl.To[0] : pintrvl.To[0]+pintrvl.To[1]],
			}
		}
		if node.right != nil {
			querySingle(cache, node.right, from, to, result)
		}
		if node.left != nil {
			querySingle(cache, node.left, from, to, result)
		}
	}
}
