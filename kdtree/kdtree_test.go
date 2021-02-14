package kdtree

import (
	"math/rand"
	"testing"
)

var uniqueVals = [][K]T{
	[K]T{1, 2},
	[K]T{3, 4},
	[K]T{5, 6},
	[K]T{7, 8},
	[K]T{9, 10},
}

func valEquals(a, b [K]T) bool {
	for i := 0; i < K; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestInvariants(t *testing.T) {
	if K != 2 {
		t.Fatalf("Tests expect kdtree of dim k=2")
	}
	// Check that T can be compared and maxT > minT
	var a, b T
	a, b = minT(), maxT()
	if a >= b {
		t.Fatalf("Expected minT() < maxT()")
	}
}

func TestInsert(t *testing.T) {
	// 1. Check root inserts into empty k-d tree
	var n, root *Node
	root = nil
	rootVal := [K]T{0, 0}
	root = root.Insert(rootVal)
	if root == nil {
		t.Fatalf("Expected ret to be root node address")
	}
	if !valEquals(root.val, rootVal) {
		t.Fatalf("Expected ret.val to equal rootVal")
	}

	// 2. Check that duplicate inserts yield the same element
	n = root.Insert(rootVal)
	if n == nil {
		t.Fatalf("Expected insert to not add a new element")
	}
	if !valEquals(n.val, rootVal) {
		t.Fatal("Expected duplicate inserts to yield same element")
	}

	// 3. Check that a bunch of unique values insert successfully
	for i := 0; i < len(uniqueVals); i++ {
		ret := root.Insert(uniqueVals[i])
		if ret != nil {
			t.Fatalf("Unexpected insert failure for uniqueVals[%d]", i)
		}
	}

	// 4. Try a couple of times to reinsert the same values
	nDups := 3 // Try to insert duplicate value 3 times
	for i := 0; i < nDups; i++ {
		ret := root.Insert(uniqueVals[i])
		if ret == nil {
			t.Fatalf("Expected insert to return uniqueVals[%d]", i)
		}
		if !valEquals(ret.val, uniqueVals[i]) {
			t.Fatalf("Expected inserted value to be uniqueVals[%d]", i)
		}
	}

	// 5. Try re-inserting the root one more time. Equivalent value, new memory
	newRootVal := [K]T{0, 0}
	if !valEquals(newRootVal, rootVal) {
		t.Fatalf("Internal error. Expected newRootVal to equal rootVal")
	}
	ret := root.Insert([K]T{0, 0})
	if ret != root {
		t.Fatalf("Expected root value to already exist")
	}
}

func TestSearch(t *testing.T) {
	var root *Node
	root = nil

	// 1. Check that searching an empty k-d tree yields no result
	for i := 0; i < len(uniqueVals); i++ {
		ret := root.Search(uniqueVals[i])
		if ret != nil {
			t.Fatalf("Expected searching empty tree should yield nil")
		}
	}

	// 2. Insert all uniqueVals
	root = root.Insert(uniqueVals[0])
	if root == nil {
		t.Fatalf("Expected first insert to return the root node")
	}
	for i := 1; i < len(uniqueVals); i++ {
		ret := root.Insert(uniqueVals[i])
		if ret != nil {
			t.Fatalf("Expected insert to return nil")
		}
	}

	// 3. Search for all uniqueVals
	for i := 0; i < len(uniqueVals); i++ {
		ret := root.Search(uniqueVals[i])
		if ret == nil {
			t.Fatalf("Expected search to return node")
		}
		if !valEquals(ret.val, uniqueVals[i]) {
			t.Fatalf("Search returned node with wrong value")
		}
	}
}

func TestInRegion(t *testing.T) {
	// 1. Check point is inRegion
	origin := [K]T{10, 10}
	var root *Node = nil
	root = root.Insert(origin)
	if root == nil {
		t.Fatalf("Expected first insert to return the root")
	}
	if !root.inRegion([2][K]T{
		origin,
		origin,
	}) {
		t.Fatalf("Expected origin to be in the region of the origin")
	}

	// 2. Check not in region low and hi
	if root.inRegion([2][K]T{
		[K]T{5, 5},
		[K]T{7, 7},
	}) {
		t.Fatalf("Expected origin to not be in the given region")
	}
	if root.inRegion([2][K]T{
		[K]T{12, 15},
		[K]T{15, 12},
	}) {
		t.Fatalf("Expected origin to not be in the given region")
	}

	// 3. Check partial match fails
	if root.inRegion([2][K]T{
		[K]T{0, 11},
		[K]T{100, 100},
	}) {
		t.Fatalf("Expected origin to not be in the similar region")
	}

	// 4. Check intersection is in region
	if !root.inRegion([2][K]T{
		[K]T{0, 9},
		[K]T{100, 11},
	}) {
		t.Fatalf("Expected origin to be in the given region")
	}
}

func TestIntersects(t *testing.T) {
	// 1. Check that a point intersects
	pointB := [2][K]T{
		[K]T{0, 0}, // Lo
		[K]T{0, 0}, // Hi
	}
	if !intersects(pointB, pointB) {
		t.Fatalf("Expected intersection of point with itself")
	}

	// 2. Check that disjoint sets do not intersect
	a2 := [2][K]T{
		[K]T{0, 10},
		[K]T{4, 14},
	}
	b2 := [2][K]T{
		[K]T{5, 15},
		[K]T{9, 19},
	}
	if intersects(a2, b2) {
		t.Fatalf("Unexpected intersection of a2 and b2")
	}
	if intersects(b2, a2) {
		t.Fatalf("Unexpected intersection of b2 and a2")
	}

	// 3. Check one-sided intersection
	a3 := [2][K]T{
		[K]T{0, 0},
		[K]T{3, 3},
	}
	b3 := [2][K]T{
		[K]T{1, 1},
		[K]T{4, 4},
	}
	if !intersects(a3, b3) {
		t.Fatalf("Expected intersection of a3 and b3")
	}
	if !intersects(b3, a3) {
		t.Fatalf("Expected intersection of b3 and a3")
	}

	// 4. Check subset intersection
	a4 := [2][K]T{
		[K]T{2, 2},
		[K]T{8, 8},
	}
	b4 := [2][K]T{
		[K]T{0, 0},
		[K]T{10, 10},
	}
	if !intersects(a4, b4) {
		t.Fatalf("Expected intersection of a4 and b4")
	}
	if !intersects(b4, a4) {
		t.Fatalf("Expected intersection of b4 and a4")
	}

	// 5. Check partial intersection fails
	a5 := [2][K]T{
		[K]T{0, 10},
		[K]T{3, 14},
	}
	b5 := [2][K]T{
		[K]T{1, 15},
		[K]T{4, 19},
	}
	if intersects(a5, b5) {
		t.Fatalf("Unexpected intersection of a5 and b5")
	}
	if intersects(b5, a5) {
		t.Fatalf("Unexpected intersection of b5 and a5")
	}
}

func TestRegionSearch(t *testing.T) {
	// Randomly insert points from 1 to 100 and check subsets
	const N = 100
	var points [N][K]T
	for i := 0; i < N; i++ {
		var v [K]T
		for j := 0; j < K; j++ {
			v[j] = T(i)
		}
		points[i] = v
	}
	rand.Seed(1)
	rand.Shuffle(len(points), func(i, j int) {
		points[i], points[j] = points[j], points[i]
	})
	var root *Node
	root = root.Insert(points[0])
	if root == nil {
		t.Fatalf("Expected first insertion to return the root")
	}
	for i := 1; i < N; i++ {
		ret := root.Insert(points[i])
		if ret != nil {
			t.Fatalf("Expected insertion to return nil for point %d", i)
		}
	}

	// 1. Check each point is returned given bounds for itself
	for i := 0; i < N; i++ {
		var res *[]*Node
		res = root.RegionSearch([2][K]T{
			points[i],
			points[i],
		})
		if res == nil {
			t.Fatalf("Expected RegionSearch to return a result at %d", i)
		}
		if len(*res) != 1 {
			t.Fatalf("Expected RegionSearch to have len(*res) == 1. Got: %d", len(*res))
		}
		if !valEquals((*res)[0].val, points[i]) {
			t.Fatalf("Expected RegionSearch to return the right node")
		}
	}

	// 2. Check that entire bounds returns all nodes

	// 3. Check that top half returns top half

	// 4. Check that bottom half returns bottom half

	// 5. Check that non intersecting bounds return nil
}
