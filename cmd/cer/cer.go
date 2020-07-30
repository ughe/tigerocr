package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ughe/tigerocr/editdist"
)

func cer(dist int, blen int) float64 {
	if dist == 0 {
		return 0.0 // Perfect match
	} else if blen == 0 {
		return 1.0 // 100% error if should be empty and not
	} else {
		return float64(dist) / float64(blen)
	}
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: cer first.txt second.txt")
	}
	bufa, err := ioutil.ReadFile(os.Args[1]);
	if err != nil {
		log.Fatal(err)
	}
	bufb, err := ioutil.ReadFile(os.Args[2]);
	if err != nil {
		log.Fatal(err)
	}
	dist := editdist.Levenshtein(bufa, bufb);
	e := cer(dist, len(bufb));
	fmt.Printf("%v\n", e);
}
