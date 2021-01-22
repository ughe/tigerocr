package kdtree

// https://dl.acm.org/doi/10.1145/361002.361007

const K = 2 // Number of dimensions

type T = int // All dimensions have the same type

type Node struct {
	val    [K]T
	lo, hi *Node
}

// Returns nil on successful insert. If element already exists, returns
// address of existing node. If root is nil, returns address of new root
func (n *Node) Insert(val [K]T) *Node {
	// I1
	if n == nil {
		return &Node{val, nil, nil}
	}
	for d := 0; ; d = (d + 1) % K {
		// I2
		eq := true
		for i := 0; i < K; i++ {
			eq = eq && (val[i] == n.val[i])
		}
		if eq {
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

// Returns the node containing val or nil if it doesn't exist
func (n *Node) Search(val [K]T) *Node {
	if n == nil {
		return nil
	}
	for d := 0; n != nil; d = (d + 1) % K {
		eq := true
		for i := 0; i < K; i++ {
			eq = eq && (val[i] == n.val[i])
		}
		if eq {
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
func (n *Node) inRegion(b [2][K]T) bool {
	for d := 0; d < K; d++ {
		if n.val[d] < b[0][d] || n.val[d] > b[1][d] {
			return false
		}
	}
	return true
}

// Returns true if the bounds overlap in every dimension
func (b0 [2][K]T) intersects(b1 [2][K]T) bool {
	for d := 0; d < K; d++ {
		if b1[0] > b0[1] || b1[1] < b0[0] {
			return false
		}
	}
	return true
}

func (n *Node) regionSearch(target, b [2][K]T, r *[]Node, d int) {
	// R1
	if n.inRegion(target) {
		*r = append(*r, n)
	}
	// R2
	bl = [2][K]T{b[0], b[1]}
	bh = [2][K]T{b[0], b[1]}
	bl[1][d] = n.val[d]
	bh[0][d] = n.val[d]
	// R3
	if n.lo != nil && target.intersects(bl) {
		n.lo.regionSearch(bl, r, (d+1)%K)
	}
	// R4
	if n.hi != nil && target.intersects(bh) {
		n.hi.regionSearch(bh, r, (d+1)%K)
	}
}

func (n *Node) RegionSearch(bounds [2][K]T) *[]Node {
	if n == nil {
		return nil
	}
	results := make([]Node, 0)
	regionSearch(bounds, &results, 0)
	return &results
}

func (n *Node) Delete() *Node {
	// D1
	if n.hi == nil && n.lo == nil {
		return nil
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

func (n *Node) optimize(int j) *Node {
	if n == nil {
		return nil
	}

	// Flatten
}

func (n *Node) Optimize() *Node {
	return n.optimize(0)
}

func (n *Node) RegionSearch() {
}
