package kdtree

// k-d tree implementation adopted from Jon Bentley's 1975 paper:
// Multidimensional Binary Search Trees Used for Associative Searching
// https://dl.acm.org/doi/10.1145/361002.361007

const K = 2 // Number of dimensions

type T = uint // dimension type. minT and maxT functions MUST match T

func minT() T {
	return 0
}
func maxT() T {
	return ^T(0)
}

type bounds [2][K]T

type Node struct {
	val    [K]T
	lo, hi *Node
}

// Inserts value starting from the root node of the kd-tree. Returns nil
// on success. If root is nil or val already exists, returns the node
func (root *Node) Insert(val [K]T) *Node {
	// I1
	if root == nil {
		return &Node{val, nil, nil}
	}
	n := root
	for d := 0; ; d = (d + 1) % K {
		// I2
		match := true
		for i := 0; i < K; i++ {
			match = match && (val[i] == n.val[i])
		}
		if match {
			return n
		}
		var child **Node
		if val[d] > n.val[d] {
			child = &n.hi
		} else {
			child = &n.lo
		}
		if *child != nil {
			// I3
			n = *child
		} else {
			// I4
			*child = &Node{val, nil, nil}
			return nil
		}
	}
}

// Returns address of node containing val or nil if val does not exist
func (n *Node) Search(val [K]T) *Node {
	if n == nil {
		return nil
	}
	for d := 0; n != nil; d = (d + 1) % K {
		match := true
		for i := 0; i < K; i++ {
			match = match && (val[i] == n.val[i])
		}
		if match {
			return n
		}
		if val[d] > n.val[d] {
			n = n.hi
		} else {
			n = n.lo
		}
	}
	return nil
}

// Return true if the current node's value is within the given bounds
func (n *Node) inRegion(b bounds) bool {
	for d := 0; d < K; d++ {
		if n.val[d] < b[0][d] || n.val[d] > b[1][d] {
			return false
		}
	}
	return true
}

// Returns true if the bounds overlap in every dimension
func intersects(b0, b1 bounds) bool {
	for d := 0; d < K; d++ {
		if b1[0][d] > b0[1][d] || b1[1][d] < b0[0][d] {
			return false
		}
	}
	return true
}

func (n *Node) regionSearch(target, b bounds, r []*Node, d int) []*Node {
	// R1
	if n.inRegion(target) {
		r = append(r, n)
	}
	// R2
	bl := bounds{b[0], b[1]}
	bh := bounds{b[0], b[1]}
	bl[1][d] = n.val[d]
	bh[0][d] = n.val[d]
	// R3
	if n.lo != nil && intersects(target, bl) {
		r = n.lo.regionSearch(target, bl, r, (d+1)%K)
	}
	// R4
	if n.hi != nil && intersects(target, bh) {
		r = n.hi.regionSearch(target, bh, r, (d+1)%K)
	}
	return r
}

// Return a list of *Node's that are within the given bounds
func (n *Node) RegionSearch(b bounds) []*Node {
	if n == nil {
		return nil
	}
	var everywhere bounds
	for i := 0; i < K; i++ {
		everywhere[0][i] = minT()
		everywhere[1][i] = maxT()
	}
	results := make([]*Node, 0)
	return n.regionSearch(b, everywhere, results, 0)
}

// Return the absolute value distance between the node and the given value
func (n *Node) distance(val [K]T) T {
	var acc T
	for d := 0; d < K; d++ {
		if val[d] >= n.val[d] {
			acc += val[d] - n.val[d]
		} else {
			acc += n.val[d] - val[d]
		}
	}
	return acc
}

// Returns the node with the smallest value at index j. Requires d,
// which is the current discriminator level for n
func jmin(n **Node, j, d int) **Node {
	if j == d {
		// Smallest values must be in LO
		if (*n).lo == nil {
			return n
		}
		l := jmin(&(*n).lo, j, (d+1)%K)
		if (*n).val[j] <= (*l).val[j] {
			return n
		} else {
			return l
		}
	} else {
		if (*n).lo == nil && (*n).hi == nil {
			return n
		}
		var l, h **Node
		if (*n).lo != nil {
			l = jmin(&(*n).lo, j, (d+1)%K)
		}
		if (*n).hi != nil {
			h = jmin(&(*n).hi, j, (d+1)%K)
		}

		if l != nil && *l != nil &&
			(h != nil && *h != nil && (*l).val[j] <= (*h).val[j]) &&
			(n != nil && *n != nil && (*l).val[j] <= (*n).val[j]) {
			return l
		} else if h != nil && *h != nil &&
			(l != nil && *l != nil && (*h).val[j] <= (*l).val[j]) &&
			(n != nil && *n != nil && (*h).val[j] <= (*h).val[j]) {
			return n
		} else { // we know: (*n).val[j] <= (*l).val[j] && (*n).val[j] <= (*h).val[j]
			return n
		}
	}
}

// Returns the node with the largest value at index j. Requires d,
// which is the current discriminator level for n
func jmax(n **Node, j, d int) **Node {
	if j == d {
		// Largest values must be in HI
		if (*n).hi == nil {
			return n
		}
		h := jmax(&(*n).hi, j, (d+1)%K)
		if (*n).val[j] >= (*h).val[j] {
			return n
		} else {
			return h
		}
	} else {
		if (*n).lo == nil && (*n).hi == nil {
			return n
		}
		var l, h **Node
		if (*n).lo != nil {
			l = jmax(&(*n).lo, j, (d+1)%K)
		}
		if (*n).hi != nil {
			h = jmax(&(*n).hi, j, (d+1)%K)
		}

		if l != nil && *l != nil &&
			(h != nil && *h != nil && (*l).val[j] >= (*h).val[j]) &&
			(n != nil && *n != nil && (*l).val[j] >= (*n).val[j]) {
			return l
		} else if h != nil && *h != nil &&
			(l != nil && *l != nil && (*h).val[j] >= (*l).val[j]) &&
			(n != nil && *n != nil && (*h).val[j] >= (*h).val[j]) {
			return h
		} else { // we know: (*n).val[j] >= (*l).val[j] && (*n).val[j] >= (*h).val[j]
			return n
		}
	}
}

// Returns a new tree with the old root deleted. Requires j, the root's
// discriminator, which increments every level of the tree and is mod K
// Repeated deletes will unbalance tree since hi is removed before lo
func (P *Node) Delete(j int) *Node {
	if P == nil {
		return nil
	}

	var child **Node
	var Q Node // Q will take P's place as the new root

	// D1
	if P.hi == nil && P.lo == nil {
		return nil
	}
	if P.hi != nil {
		// D3
		child = jmin(&P.hi, j, (j+1)%K)
	} else {
		// D4
		child = jmax(&P.lo, j, (j+1)%K)
	}

	// D5
	Q = **child
	*child = (*child).Delete((j + 1) % K)

	// D6
	Q.hi = P.hi
	Q.lo = P.lo
	return &Q
}

/*
func (n *Node) optimize(j int) *Node {
	if n == nil {
		return nil
	}

	// Flatten
}

func (n *Node) Optimize() *Node {
	return n.optimize(0)
}
*/
