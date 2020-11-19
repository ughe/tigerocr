package ocr

import (
	"testing"
)

func assert(t *testing.T, cond bool, err string) {
	if !cond {
		t.Fatalf("[FAILED] Test name: %v", err)
	}
}

func TestIntersects(t *testing.T) {
	assert(t, intersects(0, 0, 0, 0), "point")
	assert(t, intersects(-10, 10, 0, 0), "point in interval")
	assert(t, !intersects(-10, 10, 100, 100), "point outside interval")
	l, r := 0, 10
	assert(t, intersects(l, r, l, r), "identity")
	assert(t, intersects(l, r, l+1, r-1), "subset")
	assert(t, intersects(l+1, r-1, l, r), "subset flipped")
	assert(t, intersects(l, r, l-1, l+1), "left side intersect")
	assert(t, intersects(l, r, r-1, r+1), "right side intersect")
	assert(t, !intersects(l, r, l-2, l-1), "disjoint to left")
	assert(t, !intersects(l, r, r+1, r+2), "disjoint to right")
}

func TestIntersectionLen(t *testing.T) {
	assert(t, intersectionLen(0, 0, 0, 0) == 0, "point empty")
	assert(t, intersectionLen(-10, 10, 0, 0) == 0, "point in interval is empty")
	assert(t, intersectionLen(-10, 10, 100, 100) == 0, "point outside interval is empty")
	l, r := 0, 10
	assert(t, intersectionLen(l, r, l, r) == r-l, "identity *")
	assert(t, intersectionLen(l, r, l+1, r-1) == r-l-2, "subset *")
	assert(t, intersectionLen(l+1, r-1, l, r) == r-l-2, "subset flipped *")
	assert(t, intersectionLen(l, r, l-1, l+1) == 1, "left side intersect *")
	assert(t, intersectionLen(l, r, r-1, r+1) == 1, "right side intersect *")
	assert(t, intersectionLen(l, r, l-2, l-1) == 0, "disjoint to left *")
	assert(t, intersectionLen(l, r, r+1, r+2) == 0, "disjoint to right *")
}
