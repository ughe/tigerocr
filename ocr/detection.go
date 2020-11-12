package ocr

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Detection struct {
	Regions []Region `json:"regions"`
}

type Region struct {
	Confidence float32 `json:"conf"`
	Bounds     string  `json:"boundingBox"`
	Lines      []Line  `json:"lines"`
}

type Line struct {
	Confidence float32 `json:"conf"`
	Bounds     string  `json:"boundingBox"`
	Words      []Word  `json:"words"`
}

type Word struct {
	Confidence float32 `json:"conf"`
	Bounds     string  `json:"boundingBox"`
	Text       string  `json:"text"`
}

func encodeBounds(x, y, w, h int) string {
	return fmt.Sprintf("%d,%d,%d,%d", x, y, w, h)
}

func decodeBounds(bounds string) (int, int, int, int, error) {
	s := strings.SplitN(bounds, ",", 4)
	if len(s) != 4 {
		return 0, 0, 0, 0, errors.New(fmt.Sprintf(
			"Expected 4 fields. Found %d", len(s)))
	}
	x0, err := strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	y0, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	x1, err := strconv.ParseInt(s[2], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	y1, err := strconv.ParseInt(s[3], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return int(x0), int(y0), int(x1), int(y1), nil
}
