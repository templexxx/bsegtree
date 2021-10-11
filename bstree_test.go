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

	s := make([]uint64, 2048+512)
	for i := 0; i < 1024; i++ {
		s[i] = uint64(i)
	}

	exp := make([]uint64, 1024)
	for i := range exp {
		exp[i] = s[i]
	}
	sort.Sort(endpoints(exp))

	// double s
	for i := 1024; i < 2048; i++ {
		s[i] = s[i-1024]
	}
	// tripe 1/2
	for i := 2048; i < 2048+512; i++ {
		s[i] = s[i-2048]
	}

	rand.Shuffle(2048+512, func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})

	act := Dedup(s)

	if len(act) != 1024 {
		t.Fatal("dedup failed: redundant elements still exited")
	}

	for i, v := range act {
		if v != exp[i] {
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

func TestTreeEstimateIntervals(t *testing.T) {

	for i := 0; i < 2048; i += 2 {
		cnt := tree.(*BSTree).estimateIntervals(0, uint64(i))
		if cnt != i/2 +1 {
			t.Fatalf("estimate intervals wrong, exp: %d, got: %d for to: %d", i/2+1, cnt, i)
		}
	}
}

// Test segment tree result with serial query:
// both of [from, to] for every interval is 8 bytes.
func TestTreeEqualSerialSameLenInterval(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	for j := 0; j < 16; j++ {
		tree := New()
		serial := NewSerial()

		from, to := make([]byte, 8), make([]byte, 8)
		min, max := make([]byte, 8), make([]byte, 8)
		hasMin, hasMax := false, false
		for i := 0; i < 1024; i++ {
			fn := rand.Int63n(1000000)
			tn := rand.Int63n(1000000)
			binary.BigEndian.PutUint64(from, uint64(fn))
			binary.BigEndian.PutUint64(to, uint64(tn))

			if bytes.Compare(from, to) == 1 {
				from, to = to, from
			}
			tree.Push(from, to)
			serial.Push(from, to)

			if !hasMin {
				copy(min, from)
				hasMin = true
			} else {
				if bytes.Compare(min, from) == 1 {
					copy(min, from)
				}
			}
			if !hasMax {
				copy(max, to)
				hasMax = true
			} else {
				if bytes.Compare(max, to) == -1 {
					copy(max, to)
				}
			}
		}
		tree.Build()

		cmpQueryWithSerial(t, tree, serial, min, max, 1024, true, false)

		for i := 0; i < 1024; i++ {
			min := rand.Int63n(1000000)
			max := rand.Int63n(1000000)
			binary.BigEndian.PutUint64(from, uint64(min))
			binary.BigEndian.PutUint64(to, uint64(max))

			if bytes.Compare(from, to) == 1 {
				from, to = to, from
			}

			cmpQueryWithSerial(t, tree, serial, from, to, 0, false, false)
			cmpQueryWithSerial(t, tree, serial, from, nil, 0, false, true)
		}
	}
}

// Test segment tree result with serial query:
// [from, to] for every interval is 1-10 bytes.
func TestTreeEqualSerialInterval(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	for j := 0; j < 16; j++ {
		tree := New()
		serial := NewSerial()

		from, to := make([]byte, 10), make([]byte, 10)
		min, max := make([]byte, 10), make([]byte, 10)
		minN, maxN := 0, 0
		hasMin, hasMax := false, false
		for i := 0; i < 1024; i++ {

			from, to = from[:10], to[:10]

			fn, tn := rand.Intn(11), rand.Intn(11)
			if fn == 0 {
				fn = 1
			}
			if tn == 0 {
				tn = 2
			}
			from, to = from[:fn], to[:tn]
			rand.Read(from)
			rand.Read(to)

			if bytes.Compare(from, to) == 1 {
				from, to = to, from
			}
			tree.Push(from, to)
			serial.Push(from, to)

			if !hasMin {
				copy(min, from)
				hasMin = true
				minN = len(from)
			} else {
				if bytes.Compare(min, from) == 1 {
					copy(min, from)
					minN = len(from)
				}
			}
			if !hasMax {
				copy(max, to)
				hasMax = true
				maxN = len(to)
			} else {
				if bytes.Compare(max, to) == -1 {
					copy(max, to)
					maxN = len(to)
				}
			}
		}
		tree.Build()

		cmpQueryWithSerial(t, tree, serial, min[:minN], max[:maxN], 1024, true, false)

		for i := 0; i < 1024; i++ {
			from, to = from[:10], to[:10]

			fn, tn := rand.Intn(11), rand.Intn(11)
			if fn == 0 {
				fn = 1
			}
			if tn == 0 {
				tn = 2
			}
			from, to = from[:fn], to[:tn]
			rand.Read(from)
			rand.Read(to)

			if bytes.Compare(from, to) == 1 {
				from, to = to, from
			}

			cmpQueryWithSerial(t, tree, serial, from, to, 0, false, false)
			cmpQueryWithSerial(t, tree, serial, from, nil, 0, false, true)
		}
	}
}

func cmpQueryWithSerial(t *testing.T, tree, serial Tree, from, to []byte, expN int, checkExpN, isPoint bool) {

	var treeresult []int
	if !isPoint {
		treeresult = tree.Query(from, to)
	} else {
		treeresult = tree.QueryPoint(from)
	}

	if checkExpN {

		if len(treeresult) != expN { // Test full ranges here.
			t.Fatalf("result count mismatched, exp: %d, got: %d", expN, len(treeresult))
		}
	}

	var serialresult []int
	if !isPoint {
		serialresult = serial.Query(from, to)
	} else {
		serialresult = serial.QueryPoint(from)
	}

	sort.Ints(treeresult)
	sort.Ints(serialresult)

	if len(treeresult) != len(serialresult) {

		t.Fatalf("wrong result length, exp: %d, got: %d for from: %d, to: %d",
			len(serialresult), len(treeresult), AbbreviatedKey(from), AbbreviatedKey(to))
	}

	for i, act := range treeresult {
		if serialresult[i] != act {
			t.Fatalf("wrong interval id, exp: %d, got: %d", serialresult[i], act)
		}
	}
}

// This testing is designed for reproducing wrong result issue.
// bstree_test.go:172: wrong result length, exp: 297, got: 295 for from: 824723, to: 825021
// missing:
// {369 70017 825170}
// {457 824392 883250}
func TestMissingResult(t *testing.T) {

	tree := New()

	tree.(*BSTree).base = []interval{{id: 0, from: 0xb2fde, to: 0xd2d66}, {id: 1, from: 0x272c7, to: 0x2d27a}, {id: 2, from: 0x4150a, to: 0x7b67d}, {id: 3, from: 0xce70c, to: 0xeacc6}, {id: 4, from: 0x27ba, to: 0x1cc92}, {id: 5, from: 0x664da, to: 0xe354b}, {id: 6, from: 0x20f16, to: 0x6262c}, {id: 7, from: 0x67391, to: 0x75c50}, {id: 8, from: 0xd657, to: 0x5b068}, {id: 9, from: 0xd193e, to: 0xd4459}, {id: 10, from: 0x5ecaf, to: 0xceb67}, {id: 11, from: 0x4c5be, to: 0xc9918}, {id: 12, from: 0x463e7, to: 0xd99bc}, {id: 13, from: 0xac7be, to: 0xd2a47}, {id: 14, from: 0x1c516, to: 0xadf0e}, {id: 15, from: 0x12308, to: 0xbe5cf}, {id: 16, from: 0x5c0b3, to: 0xf2574}, {id: 17, from: 0xdac, to: 0x75e26}, {id: 18, from: 0x189cb, to: 0x473cb}, {id: 19, from: 0x2e490, to: 0xd19ff}, {id: 20, from: 0x63516, to: 0xebafc}, {id: 21, from: 0x188a3, to: 0xb3032}, {id: 22, from: 0x9ebe9, to: 0xab0e7}, {id: 23, from: 0xde0d, to: 0x85c51}, {id: 24, from: 0x69825, to: 0x77076}, {id: 25, from: 0x4a54f, to: 0x60333}, {id: 26, from: 0x6e78f, to: 0xa4aa5}, {id: 27, from: 0x4538d, to: 0x57cd3}, {id: 28, from: 0x8ddc2, to: 0xd7a13}, {id: 29, from: 0xb0026, to: 0xc2cf0}, {id: 30, from: 0x3f4c5, to: 0x6a13a}, {id: 31, from: 0x2a431, to: 0xeb78c}, {id: 32, from: 0xceca6, to: 0xf08f9}, {id: 33, from: 0x69a81, to: 0xb34a4}, {id: 34, from: 0xda614, to: 0xf36ea}, {id: 35, from: 0x54908, to: 0xd5d69}, {id: 36, from: 0x3165b, to: 0xc2de1}, {id: 37, from: 0x965f7, to: 0x9ec14}, {id: 38, from: 0x2586a, to: 0x34764}, {id: 39, from: 0xa1d2f, to: 0xa7cce}, {id: 40, from: 0xaa641, to: 0xb3fa5}, {id: 41, from: 0x730c8, to: 0xd89c8}, {id: 42, from: 0xe56b, to: 0x4285c}, {id: 43, from: 0x5146, to: 0x4631a}, {id: 44, from: 0x2069d, to: 0xe32b4}, {id: 45, from: 0x32c12, to: 0x4e710}, {id: 46, from: 0x1b931, to: 0x1ee70}, {id: 47, from: 0xdc080, to: 0xe2dd8}, {id: 48, from: 0x1c0ce, to: 0xb485c}, {id: 49, from: 0x1b8fc, to: 0xe1504}, {id: 50, from: 0xe6075, to: 0xeadd3}, {id: 51, from: 0xb70aa, to: 0xcf087}, {id: 52, from: 0xdee64, to: 0xee5db}, {id: 53, from: 0x2896d, to: 0xb03eb}, {id: 54, from: 0x8a0f, to: 0x81280}, {id: 55, from: 0x214fa, to: 0x68f52}, {id: 56, from: 0x8fa16, to: 0xba72f}, {id: 57, from: 0x836c6, to: 0x93300}, {id: 58, from: 0x34320, to: 0xb1643}, {id: 59, from: 0x2e8ff, to: 0x58bb4}, {id: 60, from: 0x13b54, to: 0xb5d77}, {id: 61, from: 0x7b849, to: 0xd61ab}, {id: 62, from: 0x89772, to: 0xdadf8}, {id: 63, from: 0x12dfa, to: 0x5d588}, {id: 64, from: 0x5d4ed, to: 0xdeca1}, {id: 65, from: 0x39bec, to: 0x726ef}, {id: 66, from: 0x2b0ad, to: 0x52987}, {id: 67, from: 0x3fc7a, to: 0x46be3}, {id: 68, from: 0xdac43, to: 0xe1dfb}, {id: 69, from: 0x33bd, to: 0xcf73}, {id: 70, from: 0x2c978, to: 0xcfab0}, {id: 71, from: 0xabad5, to: 0xf0104}, {id: 72, from: 0xfe53, to: 0x2be8a}, {id: 73, from: 0xcc4b6, to: 0xd0829}, {id: 74, from: 0xd65b2, to: 0xe5b8b}, {id: 75, from: 0x9f1c, to: 0xd5237}, {id: 76, from: 0x2c0ce, to: 0xbe1d8}, {id: 77, from: 0x73eaf, to: 0xb6a08}, {id: 78, from: 0xe6cf9, to: 0xf1141}, {id: 79, from: 0x131c7, to: 0x9137f}, {id: 80, from: 0x5e037, to: 0x8571d}, {id: 81, from: 0xdd3c4, to: 0xdf34e}, {id: 82, from: 0x9406a, to: 0xc45cd}, {id: 83, from: 0xce3fa, to: 0xefe92}, {id: 84, from: 0x47b7c, to: 0xd74b7}, {id: 85, from: 0x16b1c, to: 0xd2064}, {id: 86, from: 0x1ca1a, to: 0x43879}, {id: 87, from: 0x3caa1, to: 0xa5ea6}, {id: 88, from: 0xae9d2, to: 0xcd157}, {id: 89, from: 0x214a0, to: 0x6c928}, {id: 90, from: 0x46246, to: 0x49f54}, {id: 91, from: 0x7ea7c, to: 0xcf143}, {id: 92, from: 0xc914b, to: 0xcb014}, {id: 93, from: 0x480c0, to: 0x5e127}, {id: 94, from: 0x286a5, to: 0x629d6}, {id: 95, from: 0x821bf, to: 0xcfee2}, {id: 96, from: 0x5884a, to: 0xebf23}, {id: 97, from: 0x5c115, to: 0xbdcdc}, {id: 98, from: 0x32aa3, to: 0x47d11}, {id: 99, from: 0x921c1, to: 0x95de2}, {id: 100, from: 0x1ea6d, to: 0x2277d}, {id: 101, from: 0x4ec6f, to: 0x56d43}, {id: 102, from: 0x31255, to: 0x8c890}, {id: 103, from: 0x5d5b2, to: 0xca6f7}, {id: 104, from: 0xf4ec, to: 0x6f8ee}, {id: 105, from: 0x4436a, to: 0xe4499}, {id: 106, from: 0x634de, to: 0x985eb}, {id: 107, from: 0x87dcf, to: 0xae060}, {id: 108, from: 0x294ef, to: 0xc6736}, {id: 109, from: 0x71219, to: 0xb4442}, {id: 110, from: 0x1a282, to: 0x45846}, {id: 111, from: 0x5fff2, to: 0x632dd}, {id: 112, from: 0x1181b, to: 0xc7c23}, {id: 113, from: 0x738c7, to: 0x9d28a}, {id: 114, from: 0x82af2, to: 0xbf7cd}, {id: 115, from: 0x7d005, to: 0xe80d7}, {id: 116, from: 0xbd9ab, to: 0xd50d8}, {id: 117, from: 0x4f74d, to: 0x85f4e}, {id: 118, from: 0x27edf, to: 0xb595f}, {id: 119, from: 0x35256, to: 0x8f9da}, {id: 120, from: 0x70c0d, to: 0xd02fc}, {id: 121, from: 0x2596, to: 0xefce2}, {id: 122, from: 0x43978, to: 0xbeef3}, {id: 123, from: 0x4967f, to: 0xc7935}, {id: 124, from: 0x32b7, to: 0x8d43b}, {id: 125, from: 0x26591, to: 0xa89e4}, {id: 126, from: 0x2794f, to: 0x94b02}, {id: 127, from: 0x9bc4, to: 0x32b34}, {id: 128, from: 0x16842, to: 0x797a7}, {id: 129, from: 0x1b6d2, to: 0x4210a}, {id: 130, from: 0x92844, to: 0xb8c6d}, {id: 131, from: 0xaa664, to: 0xf3b26}, {id: 132, from: 0x74cad, to: 0xd28a6}, {id: 133, from: 0xb22f6, to: 0xb8abf}, {id: 134, from: 0x3c1fd, to: 0xeb4bc}, {id: 135, from: 0x8e717, to: 0x95b02}, {id: 136, from: 0x310fb, to: 0xaa55e}, {id: 137, from: 0x18673, to: 0x43070}, {id: 138, from: 0x209d, to: 0xdd159}, {id: 139, from: 0x3bc4f, to: 0x7038f}, {id: 140, from: 0x21c37, to: 0x7839b}, {id: 141, from: 0x1212e, to: 0x1f9ee}, {id: 142, from: 0x765ec, to: 0x8fad0}, {id: 143, from: 0x5e097, to: 0xef4c9}, {id: 144, from: 0x5206a, to: 0xd0241}, {id: 145, from: 0x3e975, to: 0xb0dbe}, {id: 146, from: 0x147cc, to: 0x9fcdd}, {id: 147, from: 0x91ec7, to: 0x9db52}, {id: 148, from: 0x37a9a, to: 0x4e8b9}, {id: 149, from: 0x56f5, to: 0x96cb3}, {id: 150, from: 0xbb246, to: 0xc6b4f}, {id: 151, from: 0x51757, to: 0xbfcd3}, {id: 152, from: 0x7427d, to: 0xc9858}, {id: 153, from: 0x5929, to: 0x75fbe}, {id: 154, from: 0x4e807, to: 0x5a936}, {id: 155, from: 0x3a82b, to: 0xed88b}, {id: 156, from: 0x6bc05, to: 0xac165}, {id: 157, from: 0x1e6b1, to: 0xa58cf}, {id: 158, from: 0x6cc36, to: 0xb9021}, {id: 159, from: 0x12bf8, to: 0x5cbb8}, {id: 160, from: 0x2529a, to: 0x7905a}, {id: 161, from: 0x6de0f, to: 0x77705}, {id: 162, from: 0x15515, to: 0x267fa}, {id: 163, from: 0x8476b, to: 0xbc4ba}, {id: 164, from: 0xd8de, to: 0xf370e}, {id: 165, from: 0x7b509, to: 0xe69fb}, {id: 166, from: 0x2652c, to: 0x9514a}, {id: 167, from: 0x7f96a, to: 0xa8e6c}, {id: 168, from: 0x116d, to: 0xba06a}, {id: 169, from: 0x112c1, to: 0x25e8c}, {id: 170, from: 0x24c9e, to: 0x50f45}, {id: 171, from: 0x3a404, to: 0xad147}, {id: 172, from: 0x36f2c, to: 0x81995}, {id: 173, from: 0x4c20e, to: 0x593d2}, {id: 174, from: 0xb21ea, to: 0xeabd8}, {id: 175, from: 0x981eb, to: 0x9a982}, {id: 176, from: 0xe4d3, to: 0xdc4ed}, {id: 177, from: 0x53ed, to: 0x2de3b}, {id: 178, from: 0x33e62, to: 0xd885c}, {id: 179, from: 0xc862a, to: 0xeb325}, {id: 180, from: 0xa6381, to: 0xf224b}, {id: 181, from: 0x667, to: 0x3bdda}, {id: 182, from: 0x27a26, to: 0x7ffe8}, {id: 183, from: 0x5613d, to: 0xc07bb}, {id: 184, from: 0x432d5, to: 0x7270e}, {id: 185, from: 0x85767, to: 0x98a85}, {id: 186, from: 0x3ddaa, to: 0x9a88d}, {id: 187, from: 0x333b8, to: 0x79778}, {id: 188, from: 0x104ca, to: 0x66dde}, {id: 189, from: 0x3cd07, to: 0xbeefb}, {id: 190, from: 0x13fb7, to: 0x3f1d9}, {id: 191, from: 0x27236, to: 0xee818}, {id: 192, from: 0x5c5f2, to: 0xaad4c}, {id: 193, from: 0xa8f04, to: 0xb3425}, {id: 194, from: 0x1c3a2, to: 0xb4eb2}, {id: 195, from: 0x6b13b, to: 0x841b9}, {id: 196, from: 0x31e02, to: 0xd01a6}, {id: 197, from: 0x34f9c, to: 0xd6338}, {id: 198, from: 0x1a3f7, to: 0xebfc5}, {id: 199, from: 0x47ac2, to: 0xcf34f}, {id: 200, from: 0x40450, to: 0x94608}, {id: 201, from: 0x3fcd4, to: 0x418a9}, {id: 202, from: 0x1c5a7, to: 0xaa955}, {id: 203, from: 0xbfd, to: 0x8c4e8}, {id: 204, from: 0x49f60, to: 0x8daf8}, {id: 205, from: 0x19779, to: 0xdf02f}, {id: 206, from: 0x7acee, to: 0xda674}, {id: 207, from: 0x81ea6, to: 0xd43f3}, {id: 208, from: 0xb3caf, to: 0xdf9f4}, {id: 209, from: 0x4ded4, to: 0x5fb13}, {id: 210, from: 0x343e6, to: 0xe47f2}, {id: 211, from: 0x31097, to: 0x9bd1e}, {id: 212, from: 0x44f9b, to: 0xe8c10}, {id: 213, from: 0xa27e, to: 0x9bfa7}, {id: 214, from: 0x97530, to: 0xf34c3}, {id: 215, from: 0x78442, to: 0x8593d}, {id: 216, from: 0x5391f, to: 0x7c460}, {id: 217, from: 0xc3556, to: 0xd91a5}, {id: 218, from: 0x4cf60, to: 0x68f06}, {id: 219, from: 0x54d0e, to: 0x75574}, {id: 220, from: 0x925f3, to: 0x97186}, {id: 221, from: 0x32fba, to: 0x38662}, {id: 222, from: 0x8f11e, to: 0xda048}, {id: 223, from: 0xe0b7a, to: 0xe51b2}, {id: 224, from: 0x4b07f, to: 0xcf343}, {id: 225, from: 0x9ac48, to: 0xa556d}, {id: 226, from: 0x508d3, to: 0x84163}, {id: 227, from: 0x88dc5, to: 0xe28fa}, {id: 228, from: 0x5acf3, to: 0xca25c}, {id: 229, from: 0x42c40, to: 0x68f84}, {id: 230, from: 0x21410, to: 0xaaa1a}, {id: 231, from: 0x22de7, to: 0xac8c3}, {id: 232, from: 0x5c112, to: 0xb35ab}, {id: 233, from: 0x76dcb, to: 0xe00c2}, {id: 234, from: 0x1f926, to: 0xedf5f}, {id: 235, from: 0x31b09, to: 0xaaffb}, {id: 236, from: 0x5ccd3, to: 0xa48ea}, {id: 237, from: 0x98cfc, to: 0xdf386}, {id: 238, from: 0x4842e, to: 0x4c1aa}, {id: 239, from: 0x316a7, to: 0x73d70}, {id: 240, from: 0x34b11, to: 0xcd472}, {id: 241, from: 0x9fc77, to: 0xadaf2}, {id: 242, from: 0x54b29, to: 0xef260}, {id: 243, from: 0x21c6d, to: 0x715ef}, {id: 244, from: 0x80e59, to: 0xb9569}, {id: 245, from: 0x50698, to: 0x9fa2d}, {id: 246, from: 0x506a7, to: 0xf26d1}, {id: 247, from: 0x7f481, to: 0xb6c39}, {id: 248, from: 0x733b5, to: 0x85b9e}, {id: 249, from: 0x621fc, to: 0x804e4}, {id: 250, from: 0xb2226, to: 0xb6926}, {id: 251, from: 0x2ee80, to: 0x56138}, {id: 252, from: 0x10137, to: 0xe509e}, {id: 253, from: 0x63ac9, to: 0xd4401}, {id: 254, from: 0x17577, to: 0xba230}, {id: 255, from: 0xc8dfb, to: 0xe223c}, {id: 256, from: 0x7bc25, to: 0x8440a}, {id: 257, from: 0x18a43, to: 0xd54ac}, {id: 258, from: 0x1da47, to: 0x9320a}, {id: 259, from: 0xa15c9, to: 0xf2338}, {id: 260, from: 0x5417f, to: 0x882b8}, {id: 261, from: 0x615b2, to: 0x98088}, {id: 262, from: 0x6696b, to: 0x8d50a}, {id: 263, from: 0x293cf, to: 0x85bfb}, {id: 264, from: 0x4d8a9, to: 0xe74ae}, {id: 265, from: 0xa97d4, to: 0xdd5d8}, {id: 266, from: 0xd05fb, to: 0xd3811}, {id: 267, from: 0x77eee, to: 0xdc7d1}, {id: 268, from: 0x549ef, to: 0x92848}, {id: 269, from: 0x251b7, to: 0xb750b}, {id: 270, from: 0xbbd2b, to: 0xbfbb9}, {id: 271, from: 0x76b2, to: 0x7754e}, {id: 272, from: 0x85773, to: 0xb21e3}, {id: 273, from: 0x85c47, to: 0xeae60}, {id: 274, from: 0xe7657, to: 0xed321}, {id: 275, from: 0x2c555, to: 0xb1727}, {id: 276, from: 0x67d18, to: 0x69ca3}, {id: 277, from: 0x2fd74, to: 0xcca19}, {id: 278, from: 0x68319, to: 0x7c60a}, {id: 279, from: 0xec23e, to: 0xecaf7}, {id: 280, from: 0xd2dbe, to: 0xdda3c}, {id: 281, from: 0xa25c7, to: 0xc0227}, {id: 282, from: 0x3c552, to: 0xaae77}, {id: 283, from: 0x20b1d, to: 0x924a9}, {id: 284, from: 0x5f7cc, to: 0xd2963}, {id: 285, from: 0x695a8, to: 0xc20eb}, {id: 286, from: 0x5d725, to: 0xa70ed}, {id: 287, from: 0xfecc, to: 0x3b33f}, {id: 288, from: 0xd27d, to: 0x8e558}, {id: 289, from: 0x29611, to: 0xb5e51}, {id: 290, from: 0x44d6, to: 0x41a0e}, {id: 291, from: 0x1533, to: 0xaa8c5}, {id: 292, from: 0x43303, to: 0xcd544}, {id: 293, from: 0x29eab, to: 0x3f218}, {id: 294, from: 0x9144, to: 0x91ea6}, {id: 295, from: 0x9471c, to: 0xd9624}, {id: 296, from: 0xd0d41, to: 0xef784}, {id: 297, from: 0x2c2ed, to: 0xae87f}, {id: 298, from: 0x49edd, to: 0xa8c4f}, {id: 299, from: 0x4bf69, to: 0xf37f1}, {id: 300, from: 0x47fb4, to: 0x6a502}, {id: 301, from: 0x1c2bd, to: 0xa39e1}, {id: 302, from: 0x27052, to: 0xbb99d}, {id: 303, from: 0x4fc1a, to: 0x752c8}, {id: 304, from: 0x23cb1, to: 0x87992}, {id: 305, from: 0x2e3e7, to: 0xa1509}, {id: 306, from: 0x35d2a, to: 0x527cb}, {id: 307, from: 0x479cb, to: 0x8dc56}, {id: 308, from: 0xe64c1, to: 0xe7e6f}, {id: 309, from: 0x5cdf7, to: 0x6d124}, {id: 310, from: 0x948e9, to: 0xbb9fe}, {id: 311, from: 0x6a7ba, to: 0xdc402}, {id: 312, from: 0x3836a, to: 0xdfe11}, {id: 313, from: 0xd3f54, to: 0xf32c6}, {id: 314, from: 0xc36aa, to: 0xe4af4}, {id: 315, from: 0x51391, to: 0xb6fe0}, {id: 316, from: 0x13421, to: 0xe1389}, {id: 317, from: 0xbb748, to: 0xe5135}, {id: 318, from: 0x39d2e, to: 0x582f0}, {id: 319, from: 0x80f37, to: 0x8b031}, {id: 320, from: 0x2862e, to: 0x2aeb4}, {id: 321, from: 0x44081, to: 0xdfe50}, {id: 322, from: 0x1b569, to: 0xd4cf2}, {id: 323, from: 0x2ac74, to: 0xc8d1d}, {id: 324, from: 0x1f60e, to: 0xb6241}, {id: 325, from: 0x5e7c, to: 0x1e270}, {id: 326, from: 0x83ba6, to: 0xac962}, {id: 327, from: 0x99c08, to: 0x9a9bd}, {id: 328, from: 0x37c1b, to: 0xd4e6c}, {id: 329, from: 0x64534, to: 0xf34e5}, {id: 330, from: 0x6172f, to: 0x90329}, {id: 331, from: 0x18945, to: 0x6798c}, {id: 332, from: 0x247a8, to: 0x3b87d}, {id: 333, from: 0x9ec56, to: 0x9f0d6}, {id: 334, from: 0x11943, to: 0x215bd}, {id: 335, from: 0x221b6, to: 0xcba08}, {id: 336, from: 0x5386c, to: 0x590e8}, {id: 337, from: 0x260a8, to: 0xe0a91}, {id: 338, from: 0x40d53, to: 0x70eaf}, {id: 339, from: 0xbdb43, to: 0xc303e}, {id: 340, from: 0x635e, to: 0x22ed7}, {id: 341, from: 0x3efe3, to: 0x6f08c}, {id: 342, from: 0x1f553, to: 0x287dd}, {id: 343, from: 0xc13b4, to: 0xc4cc2}, {id: 344, from: 0x6afc, to: 0x41c0b}, {id: 345, from: 0x696f7, to: 0x92fff}, {id: 346, from: 0x368cc, to: 0xd8f2f}, {id: 347, from: 0x700ec, to: 0xbc3bb}, {id: 348, from: 0x430e0, to: 0x56ccb}, {id: 349, from: 0xae76d, to: 0xf19d7}, {id: 350, from: 0x249db, to: 0x5c76f}, {id: 351, from: 0x38a5e, to: 0x7f370}, {id: 352, from: 0xa02e4, to: 0xa1a0b}, {id: 353, from: 0xd95a, to: 0x6f21b}, {id: 354, from: 0x5e967, to: 0xf0ca9}, {id: 355, from: 0x62930, to: 0xa64cf}, {id: 356, from: 0x3f932, to: 0xbfaf3}, {id: 357, from: 0x8535e, to: 0xeadfa}, {id: 358, from: 0x77c31, to: 0xb0618}, {id: 359, from: 0x8591a, to: 0xa7069}, {id: 360, from: 0x80988, to: 0xde37f}, {id: 361, from: 0x118e8, to: 0x5e7c9}, {id: 362, from: 0x1f697, to: 0x62970}, {id: 363, from: 0x7f1d2, to: 0x9c919}, {id: 364, from: 0x51865, to: 0xa4431}, {id: 365, from: 0x3cc5e, to: 0x603b8}, {id: 366, from: 0x194, to: 0x35076}, {id: 367, from: 0x15b70, to: 0xb5558}, {id: 368, from: 0xd848, to: 0x2fb17}, {id: 369, from: 0x11181, to: 0xc9752}, {id: 370, from: 0x16f2f, to: 0x9d0c9}, {id: 371, from: 0x1f6c4, to: 0x70559}, {id: 372, from: 0x919fe, to: 0x96c50}, {id: 373, from: 0x17abe, to: 0xb185d}, {id: 374, from: 0x3a5ff, to: 0x7456c}, {id: 375, from: 0x5d531, to: 0x7aeb2}, {id: 376, from: 0x43947, to: 0xba783}, {id: 377, from: 0x8fe64, to: 0xa9a5b}, {id: 378, from: 0x12bd7, to: 0xc8182}, {id: 379, from: 0x5ea45, to: 0xdfe45}, {id: 380, from: 0x6fd98, to: 0xd427d}, {id: 381, from: 0x7a1e1, to: 0xd34ee}, {id: 382, from: 0x243a6, to: 0x79d11}, {id: 383, from: 0x11658, to: 0x575f0}, {id: 384, from: 0x1095e, to: 0xa58e2}, {id: 385, from: 0x32589, to: 0x5e4c3}, {id: 386, from: 0x56ece, to: 0x90ae0}, {id: 387, from: 0xade78, to: 0xd4874}, {id: 388, from: 0x26c53, to: 0x48cb7}, {id: 389, from: 0x5a5fa, to: 0xc2ec2}, {id: 390, from: 0x804e3, to: 0xa2a56}, {id: 391, from: 0x8abd8, to: 0xd358e}, {id: 392, from: 0x9d0c5, to: 0xd88d2}, {id: 393, from: 0x2b1c0, to: 0x336b2}, {id: 394, from: 0x54606, to: 0x95849}, {id: 395, from: 0x2f50a, to: 0xc910c}, {id: 396, from: 0x28d5f, to: 0xad094}, {id: 397, from: 0x6908, to: 0xd2dff}, {id: 398, from: 0x52049, to: 0xc072d}, {id: 399, from: 0x4260d, to: 0xde445}, {id: 400, from: 0x4df05, to: 0x57c72}, {id: 401, from: 0x310a9, to: 0x4f383}, {id: 402, from: 0x1ebf7, to: 0xabf73}, {id: 403, from: 0x418ce, to: 0x6e57c}, {id: 404, from: 0xcd51b, to: 0xe958b}, {id: 405, from: 0xa365e, to: 0xe8607}, {id: 406, from: 0x28f2b, to: 0x2b0e0}, {id: 407, from: 0x86a3a, to: 0x921e2}, {id: 408, from: 0x759cc, to: 0xadb60}, {id: 409, from: 0x230d, to: 0x8f20b}, {id: 410, from: 0x2baa5, to: 0x80eb0}, {id: 411, from: 0xa1477, to: 0xdac16}, {id: 412, from: 0x3560b, to: 0xdaebc}, {id: 413, from: 0x5eada, to: 0xe6c9b}, {id: 414, from: 0x755ba, to: 0xdc5fc}, {id: 415, from: 0xbfc21, to: 0xcb5f6}, {id: 416, from: 0x40c92, to: 0x93c83}, {id: 417, from: 0x7be08, to: 0x85699}, {id: 418, from: 0x66fa9, to: 0x6db54}, {id: 419, from: 0x33c3c, to: 0x73341}, {id: 420, from: 0x9b231, to: 0xde4c8}, {id: 421, from: 0x91487, to: 0xe5ad9}, {id: 422, from: 0x8380a, to: 0xdc343}, {id: 423, from: 0xe2b0c, to: 0xed0ff}, {id: 424, from: 0xf813, to: 0x22b48}, {id: 425, from: 0x3c149, to: 0x89ffa}, {id: 426, from: 0x2ae84, to: 0x8b08d}, {id: 427, from: 0x47ea, to: 0xcbe0b}, {id: 428, from: 0x7a680, to: 0xf3ad5}, {id: 429, from: 0x2e667, to: 0x9d727}, {id: 430, from: 0xbe9ad, to: 0xc49da}, {id: 431, from: 0x498c8, to: 0x5eb8b}, {id: 432, from: 0x1ec0b, to: 0xaaf5e}, {id: 433, from: 0x278da, to: 0x4d7d6}, {id: 434, from: 0x80383, to: 0x83a43}, {id: 435, from: 0x2a5cf, to: 0x95c33}, {id: 436, from: 0x606a, to: 0x7425e}, {id: 437, from: 0x65941, to: 0x8f07a}, {id: 438, from: 0x55e74, to: 0x9ac5d}, {id: 439, from: 0xb78d5, to: 0xbcb54}, {id: 440, from: 0x2b5e6, to: 0xcaa02}, {id: 441, from: 0x67044, to: 0x72ee4}, {id: 442, from: 0x8e41a, to: 0xc0fe2}, {id: 443, from: 0xa01c2, to: 0xa60b9}, {id: 444, from: 0x11c6f, to: 0x11db1}, {id: 445, from: 0x5d5d, to: 0xd33e3}, {id: 446, from: 0x3f574, to: 0xb6c87}, {id: 447, from: 0x35300, to: 0x7d37b}, {id: 448, from: 0x70dec, to: 0x76cdf}, {id: 449, from: 0x9ed63, to: 0xd354f}, {id: 450, from: 0x61331, to: 0xe8720}, {id: 451, from: 0x8e406, to: 0xeea7a}, {id: 452, from: 0xaa0d4, to: 0xd680a}, {id: 453, from: 0x293fe, to: 0x8e057}, {id: 454, from: 0xa9a47, to: 0xac13b}, {id: 455, from: 0x3c381, to: 0x5654f}, {id: 456, from: 0xb06e, to: 0x64b17}, {id: 457, from: 0xc9448, to: 0xd7a32}, {id: 458, from: 0x51e9d, to: 0x7d106}, {id: 459, from: 0x14bb4, to: 0x406e2}, {id: 460, from: 0x785d, to: 0xd6dd9}, {id: 461, from: 0x72ab2, to: 0xd48f2}, {id: 462, from: 0xbf55a, to: 0xd7291}, {id: 463, from: 0x1ea22, to: 0xe075e}, {id: 464, from: 0x58331, to: 0xbfc57}, {id: 465, from: 0x25a9f, to: 0xa5a2a}, {id: 466, from: 0x57414, to: 0x79a49}, {id: 467, from: 0x17e13, to: 0x60569}, {id: 468, from: 0x2e038, to: 0xabe2a}, {id: 469, from: 0x76720, to: 0x9fcb1}, {id: 470, from: 0xc5602, to: 0xd7ae7}, {id: 471, from: 0x1eba0, to: 0xba7af}, {id: 472, from: 0x7ef74, to: 0xccb75}, {id: 473, from: 0x480af, to: 0x64b2d}, {id: 474, from: 0xadbad, to: 0xb1d17}, {id: 475, from: 0x50793, to: 0x60976}, {id: 476, from: 0x373ef, to: 0x5256c}, {id: 477, from: 0x7c2a9, to: 0x9c5e4}, {id: 478, from: 0x33b4d, to: 0xee2fe}, {id: 479, from: 0x5adf, to: 0x15b15}, {id: 480, from: 0x46ff8, to: 0xe0221}, {id: 481, from: 0x34937, to: 0x4dd13}, {id: 482, from: 0x97e1, to: 0xad319}, {id: 483, from: 0x64571, to: 0x6e0f7}, {id: 484, from: 0x8def2, to: 0xa50bc}, {id: 485, from: 0x7eeba, to: 0xd6193}, {id: 486, from: 0x7d77f, to: 0xe946c}, {id: 487, from: 0x343fe, to: 0xe5196}, {id: 488, from: 0x46250, to: 0x7bd33}, {id: 489, from: 0x683cb, to: 0xe3b85}, {id: 490, from: 0x2fe21, to: 0xa81e7}, {id: 491, from: 0xc08e, to: 0x1aa5f}, {id: 492, from: 0xb31f9, to: 0xed559}, {id: 493, from: 0x8a7ae, to: 0xc4537}, {id: 494, from: 0xab71, to: 0xa8dc9}, {id: 495, from: 0x1a1ce, to: 0x3c426}, {id: 496, from: 0x3fb6d, to: 0x45fdb}, {id: 497, from: 0x7ab6b, to: 0xc9c03}, {id: 498, from: 0xc679a, to: 0xe6b29}, {id: 499, from: 0xc1392, to: 0xe7b94}, {id: 500, from: 0xaf43f, to: 0xd93cd}, {id: 501, from: 0x6a51d, to: 0x8dc15}, {id: 502, from: 0x10185, to: 0x5813d}, {id: 503, from: 0x55920, to: 0x7e816}, {id: 504, from: 0x840dc, to: 0x8cb80}, {id: 505, from: 0x2d728, to: 0x713c0}, {id: 506, from: 0xa3417, to: 0xb6a27}, {id: 507, from: 0xba4b9, to: 0xd4886}, {id: 508, from: 0xad1bc, to: 0xe1d37}, {id: 509, from: 0x3830b, to: 0x81a3f}, {id: 510, from: 0x34ab2, to: 0x8b2df}, {id: 511, from: 0x66666, to: 0xbc4a5}, {id: 512, from: 0x4afd7, to: 0x53b60}, {id: 513, from: 0xa345a, to: 0xd0bcd}, {id: 514, from: 0x6fc60, to: 0xc424a}, {id: 515, from: 0x53415, to: 0x7a82e}, {id: 516, from: 0x67760, to: 0xb2ba4}, {id: 517, from: 0xb0024, to: 0xb0bd6}, {id: 518, from: 0xad39e, to: 0xf2153}, {id: 519, from: 0xa9ce3, to: 0xc4678}, {id: 520, from: 0x2c483, to: 0xaaf0f}, {id: 521, from: 0xab7b8, to: 0xe1fe3}, {id: 522, from: 0xedc2, to: 0x9b48f}, {id: 523, from: 0x2816e, to: 0x8797e}, {id: 524, from: 0x7bc71, to: 0xa97e7}, {id: 525, from: 0x5f2b5, to: 0x94960}, {id: 526, from: 0x40402, to: 0xa39dd}, {id: 527, from: 0x939c8, to: 0xd5c57}, {id: 528, from: 0x7f2da, to: 0xae691}, {id: 529, from: 0x4e3f2, to: 0xee3d5}, {id: 530, from: 0x2a3ce, to: 0xb2653}, {id: 531, from: 0x5526, to: 0x42cc0}, {id: 532, from: 0x6677f, to: 0xefcd4}, {id: 533, from: 0x91ce3, to: 0xc0199}, {id: 534, from: 0x5e8d8, to: 0x78db0}, {id: 535, from: 0x38538, to: 0x74e30}, {id: 536, from: 0x7ced7, to: 0xcea3d}, {id: 537, from: 0x39d87, to: 0x6eca9}, {id: 538, from: 0x2adc8, to: 0x50700}, {id: 539, from: 0xdbcf, to: 0x62d2e}, {id: 540, from: 0xb8661, to: 0xe61b8}, {id: 541, from: 0x32188, to: 0xc4c07}, {id: 542, from: 0x3c7f5, to: 0x5c787}, {id: 543, from: 0x9673, to: 0xd385b}, {id: 544, from: 0x67412, to: 0x9682c}, {id: 545, from: 0x33eed, to: 0x57fc3}, {id: 546, from: 0x4d712, to: 0xe3d34}, {id: 547, from: 0x34e1f, to: 0x9d360}, {id: 548, from: 0x9e3ef, to: 0xd5038}, {id: 549, from: 0x6ded4, to: 0xaf181}, {id: 550, from: 0x27aac, to: 0x6eb73}, {id: 551, from: 0x93b14, to: 0x9e3fd}, {id: 552, from: 0x75956, to: 0xa2e8e}, {id: 553, from: 0x56ae7, to: 0x750ae}, {id: 554, from: 0x660c, to: 0x4367f}, {id: 555, from: 0x7a933, to: 0xcc9e9}, {id: 556, from: 0x1009, to: 0x9ad5f}, {id: 557, from: 0xaaba, to: 0xe2689}, {id: 558, from: 0xa83c6, to: 0xdf5be}, {id: 559, from: 0x74ae4, to: 0xe7abe}, {id: 560, from: 0x3db4, to: 0xa9896}, {id: 561, from: 0x28fc4, to: 0x5ee16}, {id: 562, from: 0x28f7c, to: 0x90750}, {id: 563, from: 0x38416, to: 0x96a43}, {id: 564, from: 0x1e5b2, to: 0x8f529}, {id: 565, from: 0x9791d, to: 0xf2731}, {id: 566, from: 0xfe52, to: 0x86f28}, {id: 567, from: 0xcdd31, to: 0xedda9}, {id: 568, from: 0x51df9, to: 0x942f6}, {id: 569, from: 0xad5f, to: 0x2cda7}, {id: 570, from: 0x3c95, to: 0x6bfd4}, {id: 571, from: 0x4031, to: 0x6adfb}, {id: 572, from: 0xb5b7e, to: 0xcffd5}, {id: 573, from: 0x67e7a, to: 0xf4135}, {id: 574, from: 0x1df6b, to: 0x56b65}, {id: 575, from: 0x84c47, to: 0x940e6}, {id: 576, from: 0x22206, to: 0x75878}, {id: 577, from: 0x6bfc5, to: 0x92fc9}, {id: 578, from: 0x5a45b, to: 0x7e4c9}, {id: 579, from: 0x526dc, to: 0xf152f}, {id: 580, from: 0x2114a, to: 0x4a85a}, {id: 581, from: 0x419e3, to: 0xb8f27}, {id: 582, from: 0x6e132, to: 0x7e07c}, {id: 583, from: 0x1eafc, to: 0xe60fe}, {id: 584, from: 0x5bef9, to: 0x73993}, {id: 585, from: 0x52421, to: 0x5bc26}, {id: 586, from: 0x8c55a, to: 0xbd549}, {id: 587, from: 0x6cb29, to: 0xb8a0e}, {id: 588, from: 0x4bb72, to: 0x90bc6}, {id: 589, from: 0xaeb3f, to: 0xd04eb}, {id: 590, from: 0x29dfa, to: 0xeb77c}, {id: 591, from: 0x495fe, to: 0xf04d3}, {id: 592, from: 0x383f, to: 0xe4aa7}, {id: 593, from: 0x25200, to: 0xc4ce8}, {id: 594, from: 0x7e639, to: 0xcd91e}, {id: 595, from: 0x8522, to: 0xe20f8}, {id: 596, from: 0x58971, to: 0xba5a5}, {id: 597, from: 0x682c8, to: 0x72558}, {id: 598, from: 0xb27b6, to: 0xf3235}, {id: 599, from: 0x5fa36, to: 0xf3e1f}, {id: 600, from: 0x2e8c4, to: 0x50b26}, {id: 601, from: 0x32287, to: 0x49530}, {id: 602, from: 0x1106e, to: 0xc59b2}, {id: 603, from: 0x569c5, to: 0xdf0dd}, {id: 604, from: 0xb8ed5, to: 0xc80e7}, {id: 605, from: 0x6740e, to: 0xd1a19}, {id: 606, from: 0x20fdc, to: 0x9d772}, {id: 607, from: 0x58b6f, to: 0x84ac1}, {id: 608, from: 0x3d8dc, to: 0x79411}, {id: 609, from: 0x62dfd, to: 0xedeb8}, {id: 610, from: 0x1338, to: 0x7ddd9}, {id: 611, from: 0x4b72d, to: 0x52bf7}, {id: 612, from: 0x7af82, to: 0x941e6}, {id: 613, from: 0x2010, to: 0x43a67}, {id: 614, from: 0x3cc8d, to: 0x9e307}, {id: 615, from: 0x18e40, to: 0x3de49}, {id: 616, from: 0x382a6, to: 0xdf2a7}, {id: 617, from: 0x26d0c, to: 0x3a961}, {id: 618, from: 0x3e5c7, to: 0xc07c1}, {id: 619, from: 0x4102a, to: 0xae552}, {id: 620, from: 0x6c5c4, to: 0x91e00}, {id: 621, from: 0x7ede7, to: 0xd71aa}, {id: 622, from: 0x41036, to: 0x52841}, {id: 623, from: 0x50144, to: 0x8025d}, {id: 624, from: 0x95e72, to: 0xbfdb4}, {id: 625, from: 0x48e7d, to: 0x68a20}, {id: 626, from: 0xaeae, to: 0xb3884}, {id: 627, from: 0x5aa95, to: 0x70c2d}, {id: 628, from: 0x35d55, to: 0xc22cf}, {id: 629, from: 0x42ed0, to: 0xd1469}, {id: 630, from: 0x53e9d, to: 0xa4779}, {id: 631, from: 0x4f546, to: 0xef03a}, {id: 632, from: 0x45280, to: 0x74041}, {id: 633, from: 0x65cec, to: 0xdac61}, {id: 634, from: 0x1ec1a, to: 0x21a5f}, {id: 635, from: 0x3f58f, to: 0x5a275}, {id: 636, from: 0x4ecbc, to: 0x90a15}, {id: 637, from: 0xbae23, to: 0xc269a}, {id: 638, from: 0x9f558, to: 0xb0eb3}, {id: 639, from: 0x1a27e, to: 0x1e50d}, {id: 640, from: 0x793d1, to: 0xdfd4c}, {id: 641, from: 0xe0d44, to: 0xe8cd3}, {id: 642, from: 0x85825, to: 0x9e5d7}, {id: 643, from: 0xa38da, to: 0xdabe4}, {id: 644, from: 0x28008, to: 0xdd390}, {id: 645, from: 0x2cada, to: 0xd7b52}, {id: 646, from: 0x5534, to: 0x976a8}, {id: 647, from: 0x9e70, to: 0x8f013}, {id: 648, from: 0x6d162, to: 0xbd257}, {id: 649, from: 0x20956, to: 0x88149}, {id: 650, from: 0x400f6, to: 0x49e19}, {id: 651, from: 0x6e99b, to: 0xd14f9}, {id: 652, from: 0x5884d, to: 0x964f5}, {id: 653, from: 0x19105, to: 0xb033b}, {id: 654, from: 0x6f08e, to: 0xd49ff}, {id: 655, from: 0x36971, to: 0x574ae}, {id: 656, from: 0x5caaa, to: 0xa72dc}, {id: 657, from: 0x39ddd, to: 0xa8221}, {id: 658, from: 0x9b7f, to: 0x832ca}, {id: 659, from: 0x9e660, to: 0xebe51}, {id: 660, from: 0x53231, to: 0x714c5}, {id: 661, from: 0x19866, to: 0xa93eb}, {id: 662, from: 0x869c, to: 0x878f8}, {id: 663, from: 0x20110, to: 0x56806}, {id: 664, from: 0x2716, to: 0x18c1a}, {id: 665, from: 0x90bc2, to: 0xb0d62}, {id: 666, from: 0x503ec, to: 0x741a5}, {id: 667, from: 0xc6ce7, to: 0xe70f1}, {id: 668, from: 0x8d94b, to: 0xa33dd}, {id: 669, from: 0x96d9, to: 0x92a0c}, {id: 670, from: 0x1fbf0, to: 0xc032d}, {id: 671, from: 0x8e920, to: 0xb9238}, {id: 672, from: 0x6804f, to: 0x8e058}, {id: 673, from: 0x87876, to: 0xa3f4f}, {id: 674, from: 0x4e6f2, to: 0x9b321}, {id: 675, from: 0x4ec, to: 0x6fdc4}, {id: 676, from: 0xa481f, to: 0xda395}, {id: 677, from: 0x7baf3, to: 0xd011f}, {id: 678, from: 0x6c787, to: 0xee267}, {id: 679, from: 0xd095e, to: 0xf0878}, {id: 680, from: 0x4cfba, to: 0x82c07}, {id: 681, from: 0x3036, to: 0x2bf05}, {id: 682, from: 0x46ac4, to: 0x850f9}, {id: 683, from: 0x14e5c, to: 0x39317}, {id: 684, from: 0x24390, to: 0x4fa78}, {id: 685, from: 0xe7be5, to: 0xeaea6}, {id: 686, from: 0x39104, to: 0xd3527}, {id: 687, from: 0xc6163, to: 0xe591b}, {id: 688, from: 0x16799, to: 0x7bd76}, {id: 689, from: 0xad9b5, to: 0xd5417}, {id: 690, from: 0x33e77, to: 0x4850e}, {id: 691, from: 0x1d800, to: 0x21460}, {id: 692, from: 0xa05a1, to: 0xa1586}, {id: 693, from: 0x5b603, to: 0xc1b5d}, {id: 694, from: 0x90e9f, to: 0xacb16}, {id: 695, from: 0x87714, to: 0xcc150}, {id: 696, from: 0x1d969, to: 0x1f8c8}, {id: 697, from: 0x41f, to: 0x65b34}, {id: 698, from: 0x7d5ca, to: 0xbfc5a}, {id: 699, from: 0x34a08, to: 0xe54aa}, {id: 700, from: 0xa4ca8, to: 0xbf7a2}, {id: 701, from: 0x24eba, to: 0x63e0a}, {id: 702, from: 0x5053a, to: 0x8901f}, {id: 703, from: 0x8839b, to: 0xa7ced}, {id: 704, from: 0x5e5b6, to: 0xa117c}, {id: 705, from: 0x29956, to: 0xc539f}, {id: 706, from: 0xa8dd1, to: 0xd8d1a}, {id: 707, from: 0x13ecc, to: 0xe4462}, {id: 708, from: 0x64f9c, to: 0x96d06}, {id: 709, from: 0x2afd5, to: 0xc439c}, {id: 710, from: 0x2b6c7, to: 0x90da5}, {id: 711, from: 0x4da06, to: 0xa49a5}, {id: 712, from: 0x63d3, to: 0xae745}, {id: 713, from: 0xbea6e, to: 0xe7fd1}, {id: 714, from: 0x7692, to: 0x2f7f5}, {id: 715, from: 0x328fa, to: 0x9ba5f}, {id: 716, from: 0x39358, to: 0xa9334}, {id: 717, from: 0x4a765, to: 0xedb5d}, {id: 718, from: 0x8bff2, to: 0xd7ea3}, {id: 719, from: 0x32b50, to: 0x39b13}, {id: 720, from: 0x1a681, to: 0xb277f}, {id: 721, from: 0x3b9ac, to: 0xc977a}, {id: 722, from: 0x2b16, to: 0xeff8a}, {id: 723, from: 0x7320a, to: 0xa2c96}, {id: 724, from: 0x4a6da, to: 0xb901f}, {id: 725, from: 0x1f979, to: 0x74cc9}, {id: 726, from: 0x26fbb, to: 0xef1ae}, {id: 727, from: 0x4f286, to: 0x5599b}, {id: 728, from: 0x58e11, to: 0xb1421}, {id: 729, from: 0x3f988, to: 0xefb6c}, {id: 730, from: 0xa54e9, to: 0xb028d}, {id: 731, from: 0x82ba2, to: 0xb99e3}, {id: 732, from: 0xe198d, to: 0xe8c0c}, {id: 733, from: 0x1e58, to: 0x695f5}, {id: 734, from: 0x7e3e, to: 0x798b2}, {id: 735, from: 0x16754, to: 0x5db80}, {id: 736, from: 0x1e7ea, to: 0x9e2fc}, {id: 737, from: 0x5619e, to: 0xccf73}, {id: 738, from: 0xdd861, to: 0xe91ca}, {id: 739, from: 0xd5a3, to: 0x32901}, {id: 740, from: 0x27ac7, to: 0x61196}, {id: 741, from: 0x62f21, to: 0x78436}, {id: 742, from: 0x9cc46, to: 0xf3e42}, {id: 743, from: 0x1b231, to: 0x84155}, {id: 744, from: 0x5b283, to: 0xec2f5}, {id: 745, from: 0x2fb9f, to: 0xa4920}, {id: 746, from: 0x66812, to: 0xbec52}, {id: 747, from: 0x9409e, to: 0xd80ba}, {id: 748, from: 0x64866, to: 0xb789f}, {id: 749, from: 0x41584, to: 0x73389}, {id: 750, from: 0x3d06e, to: 0x71f8d}, {id: 751, from: 0x5e002, to: 0x89c2c}, {id: 752, from: 0x175b5, to: 0xeb6e8}, {id: 753, from: 0x51dac, to: 0xc098a}, {id: 754, from: 0x9bc92, to: 0xcc9cd}, {id: 755, from: 0x5a1b4, to: 0xf1bbb}, {id: 756, from: 0x9afa6, to: 0xbb1fe}, {id: 757, from: 0x3cf9b, to: 0x6d8a2}, {id: 758, from: 0xc3571, to: 0xc5782}, {id: 759, from: 0x97585, to: 0xc1e8b}, {id: 760, from: 0xb0224, to: 0xb4c5b}, {id: 761, from: 0xcbb7c, to: 0xe5b26}, {id: 762, from: 0x1fa62, to: 0x9ad1a}, {id: 763, from: 0x741ee, to: 0xf3f66}, {id: 764, from: 0x7848a, to: 0xa2981}, {id: 765, from: 0xa8807, to: 0xeb07a}, {id: 766, from: 0x44daa, to: 0x76f7b}, {id: 767, from: 0x5b1b, to: 0x1af5d}, {id: 768, from: 0x5c3eb, to: 0x89ceb}, {id: 769, from: 0x6046a, to: 0xa18df}, {id: 770, from: 0x7f183, to: 0xbd100}, {id: 771, from: 0x11e5f, to: 0x70432}, {id: 772, from: 0x6d5c7, to: 0xee96d}, {id: 773, from: 0x12ef5, to: 0x8c1e5}, {id: 774, from: 0x9c687, to: 0xa5f41}, {id: 775, from: 0x2da02, to: 0xcfa53}, {id: 776, from: 0x50639, to: 0xcfdb2}, {id: 777, from: 0x9a6b1, to: 0xc1480}, {id: 778, from: 0x418cd, to: 0xa67aa}, {id: 779, from: 0x18938, to: 0x1d649}, {id: 780, from: 0x653f9, to: 0x6fad0}, {id: 781, from: 0x6f6c, to: 0x60627}, {id: 782, from: 0x9d187, to: 0xd1c10}, {id: 783, from: 0x1b049, to: 0xb1bb4}, {id: 784, from: 0x87a28, to: 0xd8659}, {id: 785, from: 0x16d2a, to: 0x2a137}, {id: 786, from: 0x77de2, to: 0xae38a}, {id: 787, from: 0x563d7, to: 0xe4dad}, {id: 788, from: 0x57d2f, to: 0x6b1ab}, {id: 789, from: 0x7043a, to: 0x71360}, {id: 790, from: 0x82d5, to: 0xe331}, {id: 791, from: 0xa348c, to: 0xb26a4}, {id: 792, from: 0x52027, to: 0x9d6fe}, {id: 793, from: 0x24944, to: 0x4c3ef}, {id: 794, from: 0x9dd29, to: 0xd8042}, {id: 795, from: 0x15c0d, to: 0xe516a}, {id: 796, from: 0x2035f, to: 0x5d619}, {id: 797, from: 0x417bc, to: 0xb3f5c}, {id: 798, from: 0xa1ac6, to: 0xc262f}, {id: 799, from: 0x44ac2, to: 0xd33a1}, {id: 800, from: 0x9683d, to: 0xcc319}, {id: 801, from: 0x99cdc, to: 0xca345}, {id: 802, from: 0x1d53e, to: 0xda20b}, {id: 803, from: 0x81ee6, to: 0x9e6b9}, {id: 804, from: 0x3c28, to: 0xebfe}, {id: 805, from: 0x2d4c9, to: 0x3cdb2}, {id: 806, from: 0x5041e, to: 0x814c3}, {id: 807, from: 0x27260, to: 0xac0d0}, {id: 808, from: 0x490d4, to: 0xd940f}, {id: 809, from: 0x9d2f6, to: 0xbb780}, {id: 810, from: 0x2ed38, to: 0x36a4d}, {id: 811, from: 0x28d40, to: 0x6159a}, {id: 812, from: 0x3201e, to: 0xe2223}, {id: 813, from: 0x10991, to: 0x55020}, {id: 814, from: 0x49752, to: 0x6b5ac}, {id: 815, from: 0x50198, to: 0x6068d}, {id: 816, from: 0x276f5, to: 0x32d94}, {id: 817, from: 0x9c00, to: 0x5f594}, {id: 818, from: 0xc12f2, to: 0xd2725}, {id: 819, from: 0x1ae8e, to: 0xd7637}, {id: 820, from: 0xa380, to: 0xc40e}, {id: 821, from: 0xa4004, to: 0xd43b4}, {id: 822, from: 0x28478, to: 0xbfb8a}, {id: 823, from: 0x400fc, to: 0xb39f2}, {id: 824, from: 0x1db5e, to: 0x332ee}, {id: 825, from: 0x453c, to: 0xb3226}, {id: 826, from: 0x7067c, to: 0xf3337}, {id: 827, from: 0xbb7f5, to: 0xc6fe7}, {id: 828, from: 0x431bb, to: 0xd3c91}, {id: 829, from: 0x40cc1, to: 0xd25c5}, {id: 830, from: 0x671e7, to: 0x712fa}, {id: 831, from: 0x78725, to: 0xc50ed}, {id: 832, from: 0x8ddaa, to: 0xa291b}, {id: 833, from: 0x19a5d, to: 0x23f47}, {id: 834, from: 0x2a7ed, to: 0xc028a}, {id: 835, from: 0x7b00, to: 0x9b187}, {id: 836, from: 0x238be, to: 0x93e05}, {id: 837, from: 0x70c01, to: 0x83792}, {id: 838, from: 0x7cea, to: 0x1f53a}, {id: 839, from: 0x30984, to: 0xc1557}, {id: 840, from: 0x62b98, to: 0x9833d}, {id: 841, from: 0x32ae7, to: 0x527a5}, {id: 842, from: 0x47808, to: 0x7a056}, {id: 843, from: 0xb69a0, to: 0xd63da}, {id: 844, from: 0x9e28d, to: 0xcc52d}, {id: 845, from: 0x43511, to: 0xcdce4}, {id: 846, from: 0x703c0, to: 0xbe31c}, {id: 847, from: 0x34d3a, to: 0x7db51}, {id: 848, from: 0x94f56, to: 0xb4e18}, {id: 849, from: 0x1a445, to: 0x24fb2}, {id: 850, from: 0x4eb00, to: 0x9b1d4}, {id: 851, from: 0x7c7ce, to: 0xdb5b2}, {id: 852, from: 0x4a2b2, to: 0xa722c}, {id: 853, from: 0x72211, to: 0x9ae35}, {id: 854, from: 0x68490, to: 0x82060}, {id: 855, from: 0x2b8c6, to: 0xe6029}, {id: 856, from: 0x15026, to: 0xd59de}, {id: 857, from: 0x7f02b, to: 0x8f72f}, {id: 858, from: 0xb0b3, to: 0xcfd86}, {id: 859, from: 0x19d29, to: 0x7d449}, {id: 860, from: 0x55194, to: 0xda85f}, {id: 861, from: 0x418b7, to: 0xa7468}, {id: 862, from: 0x28e4f, to: 0xe2a2c}, {id: 863, from: 0x3b49c, to: 0x9ba9f}, {id: 864, from: 0x9030f, to: 0xb7c27}, {id: 865, from: 0x1cd65, to: 0xd00d8}, {id: 866, from: 0x248a8, to: 0x2b023}, {id: 867, from: 0x47cec, to: 0x8b88c}, {id: 868, from: 0xa1154, to: 0xbbc2a}, {id: 869, from: 0x4886a, to: 0xe70fd}, {id: 870, from: 0x32e9d, to: 0x7341b}, {id: 871, from: 0x184ef, to: 0xaf979}, {id: 872, from: 0x60109, to: 0xe160a}, {id: 873, from: 0xda23, to: 0x909f3}, {id: 874, from: 0x2fb7b, to: 0x7b8c7}, {id: 875, from: 0x8afdc, to: 0xe847f}, {id: 876, from: 0x279bc, to: 0xdd578}, {id: 877, from: 0xb9a86, to: 0xc9010}, {id: 878, from: 0xca8a8, to: 0xccf61}, {id: 879, from: 0x2d64d, to: 0xc7669}, {id: 880, from: 0x30bcc, to: 0xdcf1b}, {id: 881, from: 0xab3b, to: 0x6407a}, {id: 882, from: 0xbbf41, to: 0xe4c28}, {id: 883, from: 0x2e6d4, to: 0xb8eee}, {id: 884, from: 0xdce09, to: 0xebda8}, {id: 885, from: 0x33d85, to: 0x7c586}, {id: 886, from: 0x2ef2c, to: 0xaf6b6}, {id: 887, from: 0xc7649, to: 0xf0abb}, {id: 888, from: 0x3ff3f, to: 0x530ce}, {id: 889, from: 0x58ca4, to: 0xa0596}, {id: 890, from: 0x51c75, to: 0x97020}, {id: 891, from: 0x7be76, to: 0xc5101}, {id: 892, from: 0x558a5, to: 0xa5d26}, {id: 893, from: 0x6b097, to: 0x974a9}, {id: 894, from: 0x2184, to: 0xce25c}, {id: 895, from: 0x1ab40, to: 0xbe9ac}, {id: 896, from: 0x85703, to: 0xc56ca}, {id: 897, from: 0x17da4, to: 0xdbe1a}, {id: 898, from: 0xa6adf, to: 0xd9dfc}, {id: 899, from: 0xd91a, to: 0xf0e8b}, {id: 900, from: 0xf4df, to: 0x230d2}, {id: 901, from: 0x57859, to: 0x9d93d}, {id: 902, from: 0xa47e0, to: 0xea06f}, {id: 903, from: 0x76b5a, to: 0x8d201}, {id: 904, from: 0x803e1, to: 0xd321b}, {id: 905, from: 0x8e672, to: 0xb294d}, {id: 906, from: 0x56445, to: 0x6879c}, {id: 907, from: 0x350c7, to: 0x6a020}, {id: 908, from: 0x1c8c6, to: 0x81960}, {id: 909, from: 0x986fb, to: 0xe7227}, {id: 910, from: 0xa870, to: 0xe3d81}, {id: 911, from: 0x54c24, to: 0x63ac8}, {id: 912, from: 0x3e96f, to: 0x6b2de}, {id: 913, from: 0x4a2b1, to: 0x4b019}, {id: 914, from: 0x34857, to: 0xd25a9}, {id: 915, from: 0x938ba, to: 0x944c4}, {id: 916, from: 0x71ae0, to: 0xb182f}, {id: 917, from: 0x43bf9, to: 0xdfa20}, {id: 918, from: 0x2d3d0, to: 0x412ae}, {id: 919, from: 0x1d950, to: 0x89eaf}, {id: 920, from: 0x9cc1b, to: 0xe52f4}, {id: 921, from: 0x1ed7a, to: 0x8beb1}, {id: 922, from: 0x2d278, to: 0xef9cc}, {id: 923, from: 0x8296d, to: 0x9f25c}, {id: 924, from: 0xc299b, to: 0xe9f92}, {id: 925, from: 0x71859, to: 0xc3ba1}, {id: 926, from: 0xb1470, to: 0xc514e}, {id: 927, from: 0x58419, to: 0x7bc7e}, {id: 928, from: 0x11ede, to: 0xb4dd8}, {id: 929, from: 0x8de5d, to: 0x9aea4}, {id: 930, from: 0x75457, to: 0xd81c8}, {id: 931, from: 0x512d6, to: 0xd7edc}, {id: 932, from: 0x5d83f, to: 0x71402}, {id: 933, from: 0x99def, to: 0xcc219}, {id: 934, from: 0x1d935, to: 0x9b055}, {id: 935, from: 0xd3480, to: 0xef6c3}, {id: 936, from: 0x243e1, to: 0x7d25e}, {id: 937, from: 0x74985, to: 0xe8f5f}, {id: 938, from: 0x40240, to: 0xbb73b}, {id: 939, from: 0x5cacd, to: 0xed375}, {id: 940, from: 0xc3a1, to: 0x1e6bb}, {id: 941, from: 0x1ec8a, to: 0xc62c2}, {id: 942, from: 0x6e92b, to: 0xde37c}, {id: 943, from: 0x97bdc, to: 0x9b741}, {id: 944, from: 0x35bd8, to: 0xa0631}, {id: 945, from: 0x10822, to: 0x9c6d7}, {id: 946, from: 0x1cca7, to: 0x2e8db}, {id: 947, from: 0x4a6ea, to: 0x4c29c}, {id: 948, from: 0x5167d, to: 0x87908}, {id: 949, from: 0x4c4f0, to: 0xec6c8}, {id: 950, from: 0x4ca85, to: 0xd8765}, {id: 951, from: 0xabfdc, to: 0xcd7ef}, {id: 952, from: 0x191af, to: 0x4804b}, {id: 953, from: 0xa93ad, to: 0xf1a9d}, {id: 954, from: 0x49cf6, to: 0x4ce80}, {id: 955, from: 0xd516b, to: 0xe7f43}, {id: 956, from: 0x3bc44, to: 0xba6bb}, {id: 957, from: 0x25513, to: 0xeb733}, {id: 958, from: 0x1ee52, to: 0x5ede9}, {id: 959, from: 0x22df8, to: 0x29d58}, {id: 960, from: 0x52e1c, to: 0x8e4fb}, {id: 961, from: 0xbe3c6, to: 0xdb50f}, {id: 962, from: 0xeaf48, to: 0xf3965}, {id: 963, from: 0x513a6, to: 0x5165c}, {id: 964, from: 0x4426c, to: 0x88e3c}, {id: 965, from: 0x5ca21, to: 0xc85b9}, {id: 966, from: 0xa480, to: 0xdac61}, {id: 967, from: 0xbdeb7, to: 0xc7803}, {id: 968, from: 0x2a72c, to: 0xc7eff}, {id: 969, from: 0x5f43d, to: 0x69e49}, {id: 970, from: 0x33e54, to: 0x980ec}, {id: 971, from: 0xcc00b, to: 0xe9cfb}, {id: 972, from: 0x80ee4, to: 0xb62ca}, {id: 973, from: 0x738ae, to: 0x9d4de}, {id: 974, from: 0xade5a, to: 0xd9f42}, {id: 975, from: 0x35366, to: 0xc0f22}, {id: 976, from: 0x135c, to: 0x49467}, {id: 977, from: 0x62533, to: 0xd7a57}, {id: 978, from: 0xcde08, to: 0xe9e1a}, {id: 979, from: 0x66f7f, to: 0x7e8b4}, {id: 980, from: 0xbc52f, to: 0xdefcc}, {id: 981, from: 0x25b40, to: 0x4622b}, {id: 982, from: 0x3d538, to: 0xba954}, {id: 983, from: 0x4cade, to: 0xb88b4}, {id: 984, from: 0xc518, to: 0xeea0e}, {id: 985, from: 0x6ba76, to: 0xc0e5b}, {id: 986, from: 0xbe9d7, to: 0xd60fe}, {id: 987, from: 0x36dea, to: 0xe6bf0}, {id: 988, from: 0x40f18, to: 0xa0b5e}, {id: 989, from: 0xb0cdc, to: 0xd1791}, {id: 990, from: 0x3a3b3, to: 0xab315}, {id: 991, from: 0xaa2fb, to: 0xde1fc}, {id: 992, from: 0x8582, to: 0xed775}, {id: 993, from: 0x535b, to: 0x1a5bd}, {id: 994, from: 0xe9a62, to: 0xed586}, {id: 995, from: 0x6134f, to: 0x630d3}, {id: 996, from: 0x787e2, to: 0xba2ca}, {id: 997, from: 0x22ab1, to: 0x7ac87}, {id: 998, from: 0xe7123, to: 0xee3cd}, {id: 999, from: 0xeffc8, to: 0xf14fc}, {id: 1000, from: 0x1c03, to: 0x5ad64}, {id: 1001, from: 0xd0eef, to: 0xeec06}, {id: 1002, from: 0x8fe70, to: 0xd26b2}, {id: 1003, from: 0x1c28d, to: 0x49e0e}, {id: 1004, from: 0x898e, to: 0x3c2f4}, {id: 1005, from: 0x2b7f4, to: 0x79fb7}, {id: 1006, from: 0xd1889, to: 0xdab8a}, {id: 1007, from: 0x7576, to: 0x76aa5}, {id: 1008, from: 0x93475, to: 0xe9329}, {id: 1009, from: 0x290bd, to: 0x89a5f}, {id: 1010, from: 0x45584, to: 0x6bae9}, {id: 1011, from: 0x2daa2, to: 0x43a47}, {id: 1012, from: 0x591db, to: 0xad6d4}, {id: 1013, from: 0x65b75, to: 0x89551}, {id: 1014, from: 0x7d6c6, to: 0xe514d}, {id: 1015, from: 0x2406e, to: 0x6a5e0}, {id: 1016, from: 0x1362c, to: 0x81b6e}, {id: 1017, from: 0xca42f, to: 0xef63a}, {id: 1018, from: 0x48d8d, to: 0xc3007}, {id: 1019, from: 0x34bc6, to: 0xee83b}, {id: 1020, from: 0x422cd, to: 0xd7a08}, {id: 1021, from: 0x96da9, to: 0xd6726}, {id: 1022, from: 0x2bb49, to: 0xbfd77}, {id: 1023, from: 0x9863c, to: 0xbc748}}
	tree.(*BSTree).count = 1024
	tree.Build()

	from, to := make([]byte, 8), make([]byte, 8)
	binary.BigEndian.PutUint64(from, 824723)
	binary.BigEndian.PutUint64(to, 825021)

	st := NewSerial()
	st.(*serial).base = tree.(*BSTree).base

	cmpQueryWithSerial(t, tree, st, from, to, 0, false, false)
}

func TestMinimalTree(t *testing.T) {
	tree := New()
	tree.Push([]byte("3"), []byte("7"))
	tree.Build()
	fail := false
	result := tree.Query([]byte("1"), []byte("2"))
	if len(result) != 0 {
		fail = true
	}
	result = tree.Query([]byte("2"), []byte("3"))
	if len(result) != 1 {
		fail = true
	}
	if fail {
		t.Errorf("fail query minimal tree")
	}
}

func TestMinimalTree2(t *testing.T) {
	tree := New()
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
	tree := New()
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

	tree := New()
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

// Re-build a tree with 1024 ranges (no same element)
func BenchmarkBuildMidTree(b *testing.B) {

	tree := New()

	ftos := make([][]byte, 2048)
	for i := 0; i < 2048; i++ {
		ftos[i] = make([]byte, 8)
		binary.BigEndian.PutUint64(ftos[i], uint64(i))
	}
	sort.Sort(bytess(ftos))

	for i := 0; i < 2048; i += 2 {
		tree.Push(ftos[i], ftos[i+1])
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		tree.Build()
	}
}

var tree Tree
var ser Tree

func init() {
	tree = New()
	ser = NewSerial()

	from, to := make([]byte, 8), make([]byte, 8)
	for j := 0; j < 2048; j += 2 {
		binary.BigEndian.PutUint64(from, uint64(j))
		binary.BigEndian.PutUint64(to, uint64(j+1))

		tree.Push(from, to)
		ser.Push(from, to)
	}
	tree.Build()
}

func BenchmarkQueryFullTree(b *testing.B) {

	from, to := make([]byte, 8), make([]byte, 8)
	binary.BigEndian.PutUint64(from, 0)
	binary.BigEndian.PutUint64(to, 2048)

	for i := 0; i < b.N; i++ {
		_ = tree.Query(from, to)
	}
}


func BenchmarkQueryFullTreeSerial(b *testing.B) {

	from, to := make([]byte, 8), make([]byte, 8)
	binary.BigEndian.PutUint64(from, 0)
	binary.BigEndian.PutUint64(to, 2048)

	for i := 0; i < b.N; i++ {
		_ = ser.Query(from, to)
	}
}

func BenchmarkQueryPartTree(b *testing.B) {

	for i := 1; i <= 1024; i *= 4 {
		b.Run(fmt.Sprintf("%d result", i), func(b *testing.B) {
			benchmarkQueryPart(b, tree, i)
		})
	}
}

func BenchmarkQueryPartTreeSerial(b *testing.B) {

	for i := 1; i <= 1024; i *= 4 {
		b.Run(fmt.Sprintf("%d result", i), func(b *testing.B) {
			benchmarkQueryPart(b, ser, i)
		})
	}
}

func benchmarkQueryPart(b *testing.B, t Tree, c int) {

	from, to := make([]byte, 8), make([]byte, 8)

	binary.BigEndian.PutUint64(from, 0)
	binary.BigEndian.PutUint64(to, uint64((c-1)*2))


	for i := 0; i < b.N; i++ {
		_ = t.Query(from, to)
	}
}

func BenchmarkQueryPoint(b *testing.B) {

	from := make([]byte, 8)
	binary.BigEndian.PutUint64(from, uint64(rand.Intn(2048)))

	for i := 0; i < b.N; i++ {
		_ = tree.QueryPoint(from)
	}
}


func BenchmarkQueryPointSerial(b *testing.B) {

	from := make([]byte, 8)
	binary.BigEndian.PutUint64(from, uint64(rand.Intn(2048)))

	for i := 0; i < b.N; i++ {
		_ = ser.QueryPoint(from)
	}
}

func BenchmarkQueryPointSerialCapacity(b *testing.B) {

	t := NewSerial()
	p := make([]byte, 8)
	binary.BigEndian.PutUint64(p, 1)

	for i := 4; i <= 1024; i *= 4 {

		b.Run(fmt.Sprintf("%d", i), func(b *testing.B) {
			benchmarkQueryPointCapacity(b, t, p, i)
		})
	}

}

func BenchmarkQueryPointCapacity(b *testing.B) {

	t := New()
	p := make([]byte, 8)
	binary.BigEndian.PutUint64(p, 1)

	for i := 4; i <= 1024; i *= 4 {


		b.Run(fmt.Sprintf("%d", i), func(b *testing.B) {

			benchmarkQueryPointCapacity(b, t, p, i)
		})

	}
}

func benchmarkQueryPointCapacity(b *testing.B, t Tree, point []byte, c int) {

	from, to := make([]byte, 8), make([]byte, 8)
	for j := 0; j < c*2; j += 2 {
		binary.BigEndian.PutUint64(from, uint64(j))
		binary.BigEndian.PutUint64(to, uint64(j+1))

		t.Push(from, to)
	}
	t.Build()
	defer t.Clear()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = t.QueryPoint(point)
	}
}