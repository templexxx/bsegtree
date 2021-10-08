package bsegtree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"
)

func TestNextBytes(t *testing.T) {

	// exited := make(map[uint64]struct{})

	start := uint64(0)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, start)
	next := make([]byte, 8)
	for i := start; i < 300; i++ {
		nextBytesBuf(b, next)
		copy(b, next)
		fmt.Println(next, string(next))
	}
	fmt.Println(binary.LittleEndian.Uint64(next))
}

func TestTreeEqualSerial(t *testing.T) {
	tree := NewTree()
	serial := NewSerial()

	minB, maxB := make([]byte, 8), make([]byte, 8)
	for i := 0; i < 1024; i++ {
		min := rand.Int63n(1000000)
		max := rand.Int63n(1000000)
		binary.LittleEndian.PutUint64(minB, uint64(min))
		binary.LittleEndian.PutUint64(maxB, uint64(max))

		if bytes.Compare(minB, maxB) == 1 {
			minB, maxB = maxB, minB
		}
		tree.Push(minB, maxB)
		serial.Push(minB, maxB)
	}
	tree.Build()

	binary.LittleEndian.PutUint64(minB, 0)
	binary.LittleEndian.PutUint64(maxB, 983039)
	treeresult := tree.Query(minB, maxB)
	// fmt.Println(treeresult)
	serialresult := serial.Query(minB, maxB)
	// fmt.Println(serialresult)
	treemap := make(map[int]IntervalBytes)
	fail := false
	if len(treeresult) != len(serialresult) {
		fail = true
		fmt.Printf("unequal result length")
		goto Fail
	}
	for _, value := range treeresult {
		treemap[value.Id] = value
	}
	for _, intrvl := range serialresult {
		f := treemap[intrvl.Id].From
		to := treemap[intrvl.Id].To
		if !bytes.Equal(f, intrvl.From) || !bytes.Equal(to, intrvl.To) {
			fail = true
			fmt.Printf("result interval mismatch")
			break
		}
	}
Fail:
	if fail {
		t.Errorf("Result not equal")
	}
}
