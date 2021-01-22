package kdtree

import (
	"testing"
)

func TestInvariants(t *testing.T) {
	if K != 2 {
		t.Fatalf("Tests expect kdtree of dim k=2")
	}
	// Check that T can be compared
	var a, b T
	if a < b {
	}
}

func TestInsert(t *testing.T) {
	var root *Node
	root = nil
	root = root.Insert([K]int{0, 0})
	if root == nil {
		t.Fatalf("Expected ret to be root node address")
	}

	vals := [][K]int{
		[K]int{1, 2},
		[K]int{3, 4},
		[K]int{5, 6},
		[K]int{7, 8},
		[K]int{9, 10},
	}

	for i := 0; i < len(vals); i++ {
		ret := root.Insert(vals[i])
		if ret != nil {
			t.Fatalf("i = %d, vals[i] = %v, ret = %v", i, vals[i], ret)
		}
	}

	ret := root.Insert([K]int{0, 0})
	if ret != root {
		t.Fatalf("Expected ret == root")
	}

	ret = root.Insert(vals[0])
	if ret == nil {
		t.Fatalf("Expected ret != nil")
	}
}
