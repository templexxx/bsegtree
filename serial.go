// Copyright 2012 Thomas Oberndörfer. All rights reserved.
// Copyright 2021 Temple3x. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsegtree

// serial is a structure that allows to query intervals
// with a sequential algorithm
type serial struct {
	BSTree
}

// NewSerial returns a Tree interface with underlying serial algorithm
func NewSerial() Tree {
	t := new(serial)
	t.Clear()
	return t
}

func (t *serial) Build() {
	return
}

// Query interval by looping through the interval stack
func (t *serial) Query(from, to []byte) []int {

	fa, ta := AbbreviatedKey(from), AbbreviatedKey(to)

	result := make([]int, 0, t.estimateIntervals(fa, ta))
	for _, i := range t.base {
		if !i.Disjoint(fa, ta) {
			result = append(result, i.ID)
		}
	}
	return result
}

func (t *serial) QueryPoint(p []byte) []int {

	pa := AbbreviatedKey(p)

	result := make([]int, 0, t.estimateIntervals(pa, pa))
	for _, i := range t.base {
		if i.From <= pa && i.To >= pa {
			result = append(result, i.ID)
		}
	}
	return result
}
