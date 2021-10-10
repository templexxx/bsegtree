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
	panic("Build() not supported for serial data structure")
}

// Query interval by looping through the interval stack
func (t *serial) Query(from, to []byte) []int {

	fa, ta := AbbreviatedKey(from), AbbreviatedKey(to)

	result := make([]int, 0, 10)
	for _, intrvl := range t.base {
		if !intrvl.Disjoint(fa, ta) {
			result = append(result, intrvl.id)
		}
	}
	return result
}
