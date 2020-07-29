package editdist

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
)

const TEST_FILENAME = "lev_test.csv"

func TestLevenshtein(t *testing.T) {
	f, err := os.Open(TEST_FILENAME)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error parsing %v: %v", TEST_FILENAME, err)
		}
		if len(record) != 3 {
			log.Fatalf("Expected string,string,int but got: %v", record)
		}
		dist, err := strconv.Atoi(record[2])
		if err != nil {
			log.Fatalf("%v is not an int. %v", record[2], err)
		}
		check(t, record[0], record[1], dist)
	}
}

func  check(t *testing.T, as string, bs string, exp int) {
	a := []byte(as);
	b := []byte(bs);
	dists := levenshtein(a, b)
	dist := dists[len(a)][len(b)]
	if exp != dist {
		t.Fatalf("Expected: %v. Received: %v. Lev '%v' '%v'\n%v",
			exp, dist, as, bs, printTable(as, bs, dists))
	}
	dist = Levenshtein(b, a)
	if exp != dist {
		t.Fatalf("Expected: %v. Received: %v. Lev '%v' '%v'\n%v",
			exp, dist, bs, as, printTable(bs, as, dists))
	}
}

func printTable(a string, b string, dist [][]int) string {
	var c strings.Builder
	// Header
	fmt.Fprint(&c, "            ")
	for j := 0; j < len(b); j++ {
		fmt.Fprintf(&c, "   %v", string(b[j]))
	}
	fmt.Fprintf(&c, "\n            ")
	for j := 0; j < len(b); j++ {
		fmt.Fprintf(&c, " ---")
	}
	// First row of numbers
	fmt.Fprint(&c, "\n        ")
	for j := 0; j < len(b)+1; j++ {
		fmt.Fprintf(&c, " %3d", dist[0][j])
	}
	// Other rows
	for i := 1; i < len(a)+1; i++ {
		fmt.Fprintf(&c, "\n   %v   |", string(a[i-1]))
		for j := 0; j < len(b)+1; j++ {
			fmt.Fprintf(&c, " %3d", dist[i][j])
		}
	}
	return c.String()
}
