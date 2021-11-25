// Copyright 2011 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package bsegtree

import "encoding/binary"

type Tree interface {
	// Push new interval [from, to] to stack
	// This new interval will be added after Build.
	Push(from, to []byte)
	// PushArray push new intervals [from, to] to stack.
	// These new intervals will be added after Build.
	PushArray(from, to [][]byte)
	// Build builds segment tree out of interval stack
	Build()
	// Query interval, return interval id.
	Query(from, to []byte) []int
	// QueryPoint queries a pont, return all intervals contains this point.
	QueryPoint(p []byte) []int
	// Clear reset Tree.
	Clear()

	// Clone this tree to a new one.
	// Build the new tree before query.
	Clone() Tree

	GetAll() []Interval
}

// AbbreviatedKey returns a fixed length prefix of a user key such that AbbreviatedKey(a)
// < AbbreviatedKey(b) iff a < b and AbbreviatedKey(a) > AbbreviatedKey(b) iff a > b. If
// AbbreviatedKey(a) == AbbreviatedKey(b) an additional comparison is required to
// determine if the two keys are actually equal.
//
// This helps optimize indexed batch comparisons for cache locality. If a Split
// function is specified, AbbreviatedKey usually returns the first eight bytes
// of the user key prefix in the order that gives the correct ordering.
//
// Copied from PebbleDB.
func AbbreviatedKey(key []byte) uint64 {
	if len(key) >= 8 {
		return binary.BigEndian.Uint64(key)
	}
	var v uint64
	for _, b := range key {
		v <<= 8
		v |= uint64(b)
	}
	return v << uint(8*(8-len(key)))
}
