package main

import (
	"fmt"
	"io/ioutil"

	"github.com/ughe/tigerocr/editdist"
)

func editdistCommand(srcFilename, dstFilename string, cer bool) error {
	bufa, err := ioutil.ReadFile(srcFilename)
	if err != nil {
		return err
	}
	bufb, err := ioutil.ReadFile(dstFilename)
	if err != nil {
		return err
	}
	dist := editdist.Levenshtein(bufa, bufb)
	if cer {
		fmt.Printf("%.5f\n", editdist.CER(dist, len(bufb)))
	} else {
		fmt.Printf("%d\n", dist)
	}
	return nil
}
