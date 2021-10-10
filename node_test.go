package bsegtree

import (
	"fmt"
	"testing"
)

// {369 70017 825170}
// {457 824392 883250}
func TestNode_Disjoint(t *testing.T) {

	node := new(node)
	node.from = 824392
	node.to = 883250

	fmt.Println(node.Disjoint(824723, 825021))
}
