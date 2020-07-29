package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ughe/tigerocr/editdist"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: ./lev first.txt second.txt")
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
	fmt.Printf("%v\n", dist);
}
