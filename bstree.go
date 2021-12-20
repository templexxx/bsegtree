// Copyright 2012 Thomas Oberndörfer. All rights reserved.
// Copyright 2021 Temple3x. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsegtree

import (
	"github.com/templexxx/bsegtree/internal/bitmap"
)

type BSTree struct {
	count int // Number of intervals
	root  *node
	// interval stack
	base []Interval
	// Min value of all intervals
	min uint64
	// Max value of all intervals
	max uint64

	// sum of To - from in intervals.
	totalDeltas uint64

	// (max - min) / totalDeltas
	disjointPoint float64
}

func (t *BSTree) GetAll() []Interval {
	return t.base
}

// Relations of two intervals
const (
	SUBSET = iota
	DISJOINT
	INTERSECT_OR_SUPERSET
)

// New creates a Tree with segment tree implementation.
func New() Tree {
	t := new(BSTree)
	t.Clear()
	return t
}

// Push new interval [from, To] To stack
// This new interval will be added after Build.
func (t *BSTree) Push(from, to []byte) {

	fa := AbbreviatedKey(from)
	ta := AbbreviatedKey(to)

	t.base = append(t.base, Interval{t.count, fa, ta})
	t.count++

	if ta > t.max {
		t.max = ta
	}
	if fa < t.min {
		t.min = fa
	}

	t.totalDeltas += ta - fa

	if t.totalDeltas != 0 && t.max-t.min != 0 {
		t.disjointPoint = float64(t.max-t.min) / float64(t.totalDeltas)
	}
}

// PushArray push new intervals [from, To] To stack.
// These new intervals will be added after Build.
func (t *BSTree) PushArray(from, to [][]byte) {
	for i := 0; i < len(from); i++ {
		t.Push(from[i], to[i])
	}
}

// Build builds segment tree out of interval stack
func (t *BSTree) Build() {

	if len(t.base) == 0 {
		panic("No intervals in stack To build tree. Push intervals first")
	}
	var endpoint []uint64
	endpoint, t.min, t.max = Endpoints(t.base)
	leaves := elementaryIntervals(endpoint)
	// Create tree nodes from interval endpoints
	t.root = t.insertNodes(leaves)
	for i := range t.base {
		t.root.insertInterval(t.base[i])
	}
}

// Query interval, return interval id.
func (t *BSTree) Query(from, to []byte) []int {

	if t.root == nil {
		return nil
	}

	fa, ta := AbbreviatedKey(from), AbbreviatedKey(to)

	if ta > t.max {
		ta = t.max
	}
	if fa < t.min {
		fa = t.min
	}

	cnt := t.estimateIntervals(fa, ta)

	if (cnt >= 48 && t.count <= 1024) || t.count <= 48 { // If true, serial will be faster.
		result := make([]int, 0, cnt)
		for _, i := range t.base {
			if !i.Disjoint(fa, ta) {
				result = append(result, i.ID)
			}
		}
		return result
	}

	result := make([]int, 0, cnt)

	var bm bitmap.Bitmap
	if cnt != 1 { // There is no need to check repeated result when there will be only 1 interval.
		bm = bitmap.New(t.count)
	}

	var bmp *bitmap.Bitmap
	if bm != nil {
		bmp = &bm
	}

	querySingle(t.root, fa, ta, &result, bmp)

	if cnt == 1 {
		if len(result) <= 1 {
			return result
		}
		// on small result-set, we check for duplicates without allocation.
		// https://github.com/toberndo/go-stree/pull/5/files
		if (len(result) == 2 && result[0] != result[1]) || (len(result) == 3 && result[0] != result[1] && result[0] != result[2] && result[1] != result[2]) {
			return result
		}
		bm = bitmap.New(t.count)
		for _, id := range result {
			bm.Set(id, true)
		}
		result = result[:0]
		for i := 0; i < t.count; i++ {
			if bm.Get(i) {
				result = append(result, i)
			}
		}
	}

	return result
}

// querySingle traverse tree in search of overlaps
func querySingle(node *node, from, to uint64, result *[]int, bm *bitmap.Bitmap) {

	if !node.Disjoint(from, to) {

		for _, i := range node.overlap {
			if bm != nil {
				if !bm.Get(i.ID) {
					*result = append(*result, i.ID)
					bm.Set(i.ID, true)
				}
			} else {
				*result = append(*result, i.ID)
			}
		}
		if node.right != nil {
			querySingle(node.right, from, to, result, bm)
		}
		if node.left != nil {
			querySingle(node.left, from, to, result, bm)
		}
	}
}

func (t *BSTree) QueryPoint(p []byte) []int {

	return t.Query(p, p)
}

// Clear reset Tree.
func (t *BSTree) Clear() {
	t.count = 0
	t.root = nil
	t.base = t.base[:0]

	t.min = 0
	t.max = 0

	t.totalDeltas = 0
	t.disjointPoint = 0
}

func (t *BSTree) Clone() Tree {

	nt := &BSTree{
		count:         t.count,
		root:          nil,
		base:          make([]Interval, 0, 1024),
		min:           t.min,
		max:           t.max,
		totalDeltas:   t.totalDeltas,
		disjointPoint: t.disjointPoint,
	}

	for _, i := range t.base {
		nt.base = append(nt.base, Interval{
			ID:   i.ID,
			From: i.From,
			To:   i.To,
		})
	}
	return nt
}

// insertNodes builds tree structure from given endpoints
func (t *BSTree) insertNodes(ls [][2]uint64) *node {
	var n *node
	if len(ls) == 1 {
		n = &node{from: ls[0][0], to: ls[0][1]}
		n.left = nil
		n.right = nil
	} else {
		n = &node{from: ls[0][0], to: ls[len(ls)-1][1]}

		center := len(ls) / 2
		n.left = t.insertNodes(ls[:center])
		n.right = t.insertNodes(ls[center:])
	}
	return n
}

// estimateIntervals estimates possible intervals count will be returned by Query/QueryPoint.
// We assume the dealt of each interval is smooth. I hope so :D
func (t *BSTree) estimateIntervals(from, to uint64) int {

	if t.max == t.min {
		return 1
	}

	delta := float64(to - from)

	var cnt int
	if delta == 0 && t.disjointPoint != 0 {

		cnt = int(round(1/t.disjointPoint, 0))

	} else {
		cnt = int((delta*float64(t.count))/float64(t.max-t.min)) + 1 // +1 for potential cross intervals and point query.
	}

	if cnt < 1 {
		return 1
	}
	if cnt > t.count {
		return t.count
	}
	return cnt

}
