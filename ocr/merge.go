package ocr

import (
	"fmt"
)

func Merge(blws []Detection) (*Detection, error) {
	fmt.Println("[INFO] hello, world")
	// 1. Flatten
	dets := make([][]bWord, len(blws))
	for _, blw := range blws {
		det, err := blw.Flatten()
		if err != nil {
			return nil, err
		}
		dets = append(dets, det)
	}
	return nil, nil
}
