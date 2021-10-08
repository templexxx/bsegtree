package bsegtree

type Tree interface {
	// Push new interval [from, to] to stack
	Push(from, to []byte)
	// PushArray push new intervals [from, to] to stack
	PushArray(from, to [][]byte)
	// Build builds segment tree out of interval stack
	Build()
	// Query interval
	Query(from, to []byte) []IntervalBytes
	// Clear the interval stack
	Clear()
}
