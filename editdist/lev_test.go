package editdist

import (
	"testing"
)

func  check(t *testing.T, as string, bs string, exp int) {
	a := []byte(as);
	b := []byte(bs);
	dist := Levenshtein(a, b)
	if exp != dist {
		t.Fatalf("Expected: %v. Actual: %v. Lev '%v' '%v'", exp, dist, as, bs)
	}
	dist = Levenshtein(b, a)
	if exp != dist {
		t.Fatalf("Expected: %v. Actual: %v. Lev '%v' '%v'", exp, dist, bs, as)
	}
}

func TestLevenshtein(t *testing.T) {
	check(t, "", "", 0);
	check(t, "a", "a", 0);
	check(t, "ab", "abb", 1);
	check(t, "potatoe", "tomatoe", 2);
	check(t, "bonanza", "gonzaga", 4);
	check(t, "alfalfa", "allfalf", 2);
	check(t, "1234567", "ABCDEFG", 7);
}
