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

func (n *Node) regionSearch(target, b bounds, r *[]*Node, d int) {
	// R1
	if n.inRegion(target) {
		*r = append(*r, n)
	}
	// R2
	bl := bounds{b[0], b[1]}
	bh := bounds{b[0], b[1]}
	bl[1][d] = n.val[d]
	bh[0][d] = n.val[d]
	// R3
	if n.lo != nil && intersects(target, bl) {
		n.lo.regionSearch(target, bl, r, (d+1)%K)
	}
	// R4
	if n.hi != nil && intersects(target, bh) {
		n.hi.regionSearch(target, bh, r, (d+1)%K)
	}
}

func (n *Node) RegionSearch(b bounds) *[]*Node {
	if n == nil {
		return nil
	}
	var everywhere bounds
	for i := 0; i < K; i++ {
		everywhere[0][i] = minT()
		everywhere[1][i] = maxT()
	}
	results := make([]*Node, 0)
	n.regionSearch(b, everywhere, &results, 0)
	return &results
}

func (n *Node) distance(val [K]T) T {
	var acc T
	for d := 0; d < K; d++ {
		if val[d] > n.val[d] {
			acc += val[d] - n.val[d]
		} else {
			acc += n.val[d] - val[d]
		}
	}
	return acc
}

// Node n has dim d. Returns node with minimum val in dim j
func (n *Node) findMin(d, j int) *Node {
	if d == j {
	} else {
		// Search both subtrees
	}
}

func (n *Node) Delete() *Node {
	// D1
	if n.hi == nil && n.lo == nil {
		return nil
	}
	var child **node
	if n.hi == nil {
		// lo
		// Q <- J-minimum node in p.hi
		// QFAT <- father of Q
		// QSON <- f(QFAT) = Q (hi or lo)
	} else if n.lo == nil {
		// hi
	} else {
		// randomly chose
		rand.Intn(2)
	}

	q := n
	for j := 0; n != nil; j = (j + 1) % K {
		// D2
		if n.hi == nil {
			// D4
		} else {
			// D3
		}
		// D5
		// D6
	}
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
