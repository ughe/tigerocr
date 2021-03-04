package main

import (
	"fmt"
	"io/ioutil"
)

func extractCommand(filename string, stat, algoid, speed, date, text bool) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	// Extraction never uses image width or height. So okay to use bogus
	detection, err := convertToBLW(nil, raw, filename)
	if err != nil {
		return err
	}

	// Extract the fields
	if stat {
		// Human readable output
		b, l, w := detection.CountBLW()
		fmt.Printf("algoid: %s\n", detection.AlgoID)
		fmt.Printf("millis: %d\n", detection.Millis)
		fmt.Printf("date:   %s\n", detection.Date)
		fmt.Printf("blocks: %d\n", b)
		fmt.Printf("lines:  %d\n", l)
		fmt.Printf("words:  %d\n", w)
	} else if algoid {
		fmt.Printf("%s\n", detection.AlgoID)
	} else if speed {
		fmt.Printf("%d\n", detection.Millis)
	} else if date {
		fmt.Printf("%s\n", detection.Date)
	} else if text {
		fmt.Printf("%s\n", detection.Plaintext())
	} else {
		return fmt.Errorf("Error: no flags specified")
	}
	return nil
}
