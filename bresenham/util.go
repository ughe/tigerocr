package bresenham

// Bool to Int
func btoi(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

// Absolute value
func abs(n int) int {
	if n >= 0 {
		return n
	} else {
		return -n
	}
}
