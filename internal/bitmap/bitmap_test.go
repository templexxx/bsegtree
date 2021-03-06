package bitmap

import (
	"testing"
)

func TestDataOrCopy(t *testing.T) {
	data := make([]byte, 5)
	data[0] = 1
	d2 := dataOrCopy(data, false)
	d2[0] = 2
	if data[0] != 2 {
		t.Fatal("wrong data")
	}
	d3 := dataOrCopy(data, true)
	d3[0] = 3
	if data[0] != 2 {
		t.Fatal("wrong data")
	}
}

func TestBitwiseOperations(t *testing.T) {
	data := byte(0)
	data = SetBit(data, 0, true)
	if data != 1 {
		t.Fatal("wrong data")
	}
	if GetBit(data, 0) != true {
		t.Fatal("wrong getbit")
	}
	data = SetBit(data, 0, false)
	if data != 0 {
		t.Fatal("wrong data")
	}
	if GetBit(data, 0) != false {
		t.Fatal("wrong getbit")
	}
}

func TestBitmap(t *testing.T) {
	bm := New(50)
	bm.Set(30, true)
	if bm.Get(30) != true {
		t.Fatal("wrong GET")
	}
	if bm.Len() != 56 {
		t.Fatal("wrong length")
	}
	data := bm.Data(true)
	if Get(data, 30) != true {
		t.Fatal("wrong data copy")
	}
}

func TestNewSlice(t *testing.T) {
	bm := NewSlice(7)
	if len(bm) != 1 {
		t.Fatal("wrong length")
	}
	bm = NewSlice(10)
	if len(bm) != 2 {
		t.Fatal("wrong length")
	}
}

func TestLen(t *testing.T) {
	bitmap := []byte{0, 0, 0}
	if Len(bitmap) != 24 {
		t.Fatal("wrong length")
	}
}

func TestGetSet(t *testing.T) {
	bm := NewSlice(50)
	for i := 0; i < Len(bm); i++ {
		if Get(bm, i) != false {
			t.Fatal("wrong return value")
		}
		Set(bm, i, true)
	}

	for i := Len(bm) - 1; i >= 0; i-- {
		if Get(bm, i) != true {
			t.Fatal("wrong return value")
		}
		Set(bm, i, false)
	}
	for i := 0; i < Len(bm); i++ {
		if Get(bm, i) != false {
			t.Fatal("wrong return value")
		}
	}
}

func BenchmarkFuncs(b *testing.B) {
	bm := New(1000 * 1000)
	index := 0
	for i := 0; i < b.N; i++ {
		index++
		if index == 1000 {
			index = 0
		}
		Set(bm, index, !Get(bm, index))
	}
}
