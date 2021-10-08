package bsegtree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestDedup(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	s := make([][]byte, 2048)
	for i := 0; i < 1024; i++ {
		s[i] = make([]byte, 8)
		binary.LittleEndian.PutUint64(s[i], uint64(i))
	}

	exp := make([][]byte, 1024)
	for i := range exp {
		exp[i] = make([]byte, 8)
		copy(exp[i], s[i])
	}
	sort.Sort(bytess(exp))

	// double s
	for i := 1024; i < 2048; i++ {
		s[i] = make([]byte, 8)
		copy(s[i], s[i-1024])
	}

	rand.Shuffle(2048, func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})
	pos := make([][2]int, 2048)
	for i := range pos {
		pos[i] = [2]int{i*8, 8}
	}
	cache := make([]byte, 2048*8)
	for i := range s {
		copy(cache[i*8:i*8+8], s[i])
	}

	newPos := Dedup(endpoints{
		positions: pos,
		cache:     cache,
	})
	
	if len(newPos) != 1024 {
		t.Fatal("dedup failed: redundant elements still exited")
	}

	for i := range newPos {
		v := cache[newPos[i][0]:newPos[i][0]+8]
		if !bytes.Equal(v, exp[i]) {
			t.Fatal("dedup failed: wrong order")
		}
	}
}

type bytess [][]byte

func (e bytess) Len() int {
	return len(e)
}

func (e bytess) Less(i, j int) bool {

	return bytes.Compare(e[i], e[j]) == -1
}

func (e bytess) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func dedup(sl []int) []int {
	sort.Ints(sl)

	cnt := len(sl)
	cntDup := 0
	for i := 1; i < cnt; i++ {

		if sl[i] == sl[i-1] {
			cntDup++
		} else {
			sl[i-cntDup] = sl[i]
		}
	}

	return sl[:cnt-cntDup]
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
