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

func (t *serial) Print() {
	panic("Print() not supported for serial data structure")
}

func (t *serial) Tree2Array() []SegmentOverlap {
	panic("Tree2Array() not supported for serial data structure")
}

// Query interval by looping through the interval stack
func (t *serial) Query(from, to []byte) []IntervalBytes {
	result := make([]IntervalBytes, 0, 10)
	for _, intrvl := range t.base {
		if !intrvl.Segment.Disjoint(t.intervalCache, from, to) {
			ib := IntervalBytes{
				Id:   intrvl.Id,
				From: t.intervalCache[intrvl.From[0] : intrvl.From[0]+intrvl.From[1]],
				To:   t.intervalCache[intrvl.To[0] : intrvl.To[0]+intrvl.To[1]],
			}
			result = append(result, ib)
		}
	}
	return result
}
