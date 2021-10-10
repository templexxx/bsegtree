// Copyright 2012 Thomas Oberndörfer. All rights reserved.
// Copyright 2021 Temple3x. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsegtree

type BSTree struct {
	count int // Number of intervals
	root  *node
	// interval stack
	base []interval
	// Min value of all intervals
	min uint64
	// Max value of all intervals
	max uint64
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

// Push new interval [from, to] to stack
// This new interval will be added after Build.
func (t *BSTree) Push(from, to []byte) {

	fa := AbbreviatedKey(from)
	ft := AbbreviatedKey(to)

	t.base = append(t.base, interval{t.count, fa, ft})
	t.count++
}

// PushArray push new intervals [from, to] to stack.
// These new intervals will be added after Build.
func (t *BSTree) PushArray(from, to [][]byte) {
	for i := 0; i < len(from); i++ {
		t.Push(from[i], to[i])
	}
}

// Build builds segment tree out of interval stack
func (t *BSTree) Build() {

	if len(t.base) == 0 {
		panic("No intervals in stack to build tree. Push intervals first")
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
		panic("Can't run query on empty tree. Call Build() first")
	}

	var result []int

	result = make([]int, 0, 1)

	fa, ta := AbbreviatedKey(from), AbbreviatedKey(to)

	querySingle(t.root, fa, ta, &result)

	// on small result-set, we check for duplicates without allocation.
	// https://github.com/toberndo/go-stree/pull/5/files
	if len(result) < 2 || (len(result) == 2 && result[0] != result[1]) || (len(result) == 3 && result[0] != result[1] && result[0] != result[2] && result[1] != result[2]) {
		return result
	}

	// on larger sets, use a map.
	m := make(map[int]struct{})
	for _, id := range result {
		m[id] = struct{}{}
	}
	if len(m) == len(result) {
		return result
	}
	result = result[:0]
	for id := range m {
		result = append(result, id)
	}

	return result
}

// querySingle traverse tree in search of overlaps
func querySingle(node *node, from, to uint64, result *[]int) {

	if !node.Disjoint(from, to) {

		for _, interval := range node.overlap {

			*result = append(*result, interval.id)
		}
		if node.right != nil {
			querySingle(node.right, from, to, result)
		}
		if node.left != nil {
			querySingle(node.left, from, to, result)
		}
	}
}

func (t *BSTree) QueryPoint(p []byte) []int {

	if t.root == nil {
		panic("Can't run query on empty tree. Call Build() first")
	}

	pa := AbbreviatedKey(p)
	result := make([]int, 0, 1)

	queryPoint(t.root, pa, &result)

	// on small result-set, we check for duplicates without allocation.
	// https://github.com/toberndo/go-stree/pull/5/files
	if len(result) < 2 || (len(result) == 2 && result[0] != result[1]) || (len(result) == 3 && result[0] != result[1] && result[0] != result[2] && result[1] != result[2]) {
		return result
	}

	// on larger sets, use a map.
	m := make(map[int]struct{})
	for _, id := range result {
		m[id] = struct{}{}
	}
	if len(m) == len(result) {
		return result
	}
	result = result[:0]
	for id := range m {
		result = append(result, id)
	}

	return result
}

func queryPoint(node *node, p uint64, result *[]int) {

	if node.from <= p && node.to >= p {
		for _, interval := range node.overlap {
			*result = append(*result, interval.id)
		}
		if node.left != nil {
			queryPoint(node.left, p, result)
		}
		if node.right != nil {
			queryPoint(node.right, p, result)
		}
	}
}

// Clear reset Tree.
func (t *BSTree) Clear() {
	t.count = 0
	t.root = nil
	t.base = make([]interval, 0, 1024)

	t.min = 0
	t.max = 0
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
