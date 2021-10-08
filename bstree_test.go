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

	cmpQueryWithSerial(t, tree, serial, minB, maxB, 1024, true)

	for i := 0; i < 1024; i++ {
		min := rand.Int63n(1000000)
		max := rand.Int63n(1000000)
		binary.LittleEndian.PutUint64(minB, uint64(min))
		binary.LittleEndian.PutUint64(maxB, uint64(max))

		if bytes.Compare(minB, maxB) == 1 {
			minB, maxB = maxB, minB
		}

		cmpQueryWithSerial(t, tree, serial, minB, maxB, 0, false)
	}
}

func cmpQueryWithSerial(t *testing.T, tree, serial Tree, from ,to []byte, expN int, checkExpN bool) {

	treeresult := tree.Query(from, to)

	if checkExpN {

		if len(treeresult) != expN {	// Test full ranges here.
			t.Fatalf("result count mismatched, exp: %d, got: %d",expN, len(treeresult))
		}
	}

	serialresult := serial.Query(from, to)
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

func TestMinimalTree(t *testing.T) {
	tree := NewTree()
	tree.Push([]byte("3"), []byte("7"))
	tree.Build()
	fail := false
	result := tree.Query([]byte("1"), []byte("2"))
	if len(result) != 0 {
		fail = true
	}
	result = tree.Query([]byte("2"),[]byte("3"))
	if len(result) != 1 {
		fail = true
	}
	if fail {
		t.Errorf("fail query minimal tree")
	}
}

func TestMinimalTree2(t *testing.T) {
	tree := NewTree()
	tree.Push([]byte("1"), []byte("1"))
	tree.Build()
	if result := tree.Query([]byte("1"), []byte("1")); len(result) != 1 {
		t.Errorf("fail query minimal tree for (1, 1)")
	}
	if result := tree.Query([]byte("1"), []byte("2")); len(result) != 1 {
		t.Errorf("fail query minimal tree for (1, 2)")
	}
	if result := tree.Query([]byte("2"), []byte("3")); len(result) != 0 {
		t.Errorf("fail query minimal tree for (2, 3)")
	}
}

func TestNormalTree(t *testing.T) {
	tree := NewTree()
	tree.Push([]byte("1"), []byte("1"))
	tree.Push([]byte("2"), []byte("3"))
	tree.Push([]byte("5"), []byte("7"))
	tree.Push([]byte("4"), []byte("6"))
	tree.Push([]byte("6"), []byte("9"))
	tree.Build()
	if result := tree.Query([]byte("3"), []byte("5")); len(result) != 3 {
		t.Errorf("fail query multiple tree for (3, 5)")
	}
	qvalid := map[string]int{
		"0": 0,
		"1": 1,
		"2": 1,
		"3": 1,
		"4": 1,
		"5": 2,
		"6": 3,
		"7": 2,
		"8": 1,
		"9": 1,
	}

	for k, v := range qvalid {
		if result := tree.Query([]byte(k), []byte(k)); len(result) != v {
			t.Errorf("fail query multiple tree for (%s, %s), exp: %d, got: %d", k, k, v, len(result))
		}
	}
}

func BenchmarkBuildSmallTree(b *testing.B) {

	tree := NewTree()
	tree.Push([]byte{1}, []byte{1})
	tree.Push([]byte{2}, []byte{3})
	tree.Push([]byte{5}, []byte{7})
	tree.Push([]byte{4}, []byte{6})
	tree.Push([]byte{6}, []byte{9})
	tree.Push([]byte{9}, []byte{14})
	tree.Push([]byte{10}, []byte{13})
	tree.Push([]byte{11}, []byte{11})

	for i := 0; i < b.N; i++ {

		tree.Build()
	}
}

func BenchmarkBuildMidTree(b *testing.B) {

	tree := NewTree()

	f, to := make([]byte, 8), make([]byte, 8)
	for i := 0; i < 1024; i++ {
		binary.BigEndian.PutUint64(f, uint64(i))
		binary.BigEndian.PutUint64(to, uint64(i+1024))
		if bytes.Compare(f, to) > 1{
			f, to = to, f
		}
		tree.Push(f, to)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		tree.Build()
	}
}

var tree Tree
var ser Tree

// func init() {
// 	tree = NewTree()
// 	ser = NewSerial()
// 	for j := 0; j < 100000; j++ {
// 		min := rand.Int()
// 		max := rand.Int()
// 		if min > max {
// 			min, max = max, min
// 		}
// 		tree.Push(min, max)
// 		ser.Push(min, max)
// 	}
// 	tree.BuildTree()
// }
//
// func BenchmarkBuildTree1000(b *testing.B) {
// 	tree := NewTree()
// 	buildTree(b, tree, 1000)
// }
//
// func BenchmarkBuildTree100000(b *testing.B) {
// 	tree := NewTree()
// 	buildTree(b, tree, 100000)
// }
//
// func BenchmarkQueryTree(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		tree.Query(0, 100000)
// 	}
// }
//
// func BenchmarkQuerySerial(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		ser.Query(0, 100000)
// 	}
// }
//
// func BenchmarkQueryTreeMax(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		tree.Query(0, math.MaxInt32)
// 	}
// }
//
// func BenchmarkQuerySerialMax(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		ser.Query(0, math.MaxInt32)
// 	}
// }
//
// func BenchmarkQueryTreeArray(b *testing.B) {
// 	from := []int{0, 1000000, 2000000, 3000000, 4000000, 5000000, 6000000, 7000000, 8000000, 9000000}
// 	to := []int{10, 1000010, 2000010, 3000010, 4000010, 5000010, 6000010, 7000010, 8000010, 9000010}
// 	for i := 0; i < b.N; i++ {
// 		tree.QueryArray(from, to)
// 	}
// }
//
// func BenchmarkQuerySerialArray(b *testing.B) {
// 	from := []int{0, 1000000, 2000000, 3000000, 4000000, 5000000, 6000000, 7000000, 8000000, 9000000}
// 	to := []int{10, 1000010, 2000010, 3000010, 4000010, 5000010, 6000010, 7000010, 8000010, 9000010}
// 	for i := 0; i < b.N; i++ {
// 		ser.QueryArray(from, to)
// 	}
// }
//
// func buildTree(b *testing.B, tree Tree, count int) {
// 	for i := 0; i < b.N; i++ {
// 		b.StopTimer()
// 		tree.Clear()
// 		pushRandom(tree, count)
// 		b.StartTimer()
// 		tree.BuildTree()
// 	}
// }
//
// func pushRandom(tree Tree, count int) {
// 	for j := 0; j < count; j++ {
// 		min := rand.Int()
// 		max := rand.Int()
// 		if min > max {
// 			min, max = max, min
// 		}
// 		tree.Push(min, max)
// 	}
// }
//
// func BenchmarkEndpoints100000(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		b.StopTimer()
// 		tree := NewTree().(*stree)
// 		pushRandom(tree, 100000)
// 		b.StartTimer()
// 		Endpoints(tree.base)
// 	}
// }
//
// func BenchmarkInsertNodes100000(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		b.StopTimer()
// 		tree := NewTree().(*stree)
// 		pushRandom(tree, 100000)
// 		var endpoint []int
// 		endpoint, tree.min, tree.max = Endpoints(tree.base)
// 		//fmt.Println(len(endpoint))
// 		b.StartTimer()
// 		tree.root = tree.insertNodes(endpoint)
// 	}
// }
//
// func BenchmarkInsertIntervals100000(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		b.StopTimer()
// 		tree := NewTree().(*stree)
// 		pushRandom(tree, 100000)
// 		var endpoint []int
// 		endpoint, tree.min, tree.max = Endpoints(tree.base)
// 		tree.root = tree.insertNodes(endpoint)
// 		b.StartTimer()
// 		for i := range tree.base {
// 			insertInterval(tree.root, &tree.base[i])
// 		}
// 	}
// }