package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ughe/tigerocr/ocr"
)

var (
	blue   = color.RGBA{0, 0, 255, 255}
	red    = color.RGBA{255, 0, 0, 255}
	orange = color.RGBA{255, 165, 0, 255}
)

func annotate(img []byte, detection *ocr.Detection, b, l, w bool, dstFilename string) error {
	lid := strings.ToLower(detection.AlgoID)
	var col color.Color
	if strings.Contains(lid, "aws") {
		col = orange
	} else if strings.Contains(lid, "azu") {
		col = blue
	} else if strings.Contains(lid, "gcp") {
		col = red
	} else {
		col = color.Black
	}

	dstImg, err := detection.Annotate(img, col, b, l, w)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dstFilename, dstImg, 0600); err != nil {
		return err
	}
	fmt.Printf("[INFO] Annotated image: %v\n", dstFilename)
	return nil
}

// Annotates the block, lines, and/or words on the image given the coordinates
func annotateCommand(b, l, w bool, imageFilename, coordFilename string) error {
	buf, err := ioutil.ReadFile(imageFilename)
	if err != nil {
		return err
	}
	raw, err := ioutil.ReadFile(coordFilename)
	if err != nil {
		return err
	}

	detection, err := convertToBLW(buf, raw, coordFilename)

	blw := ""
	if b {
		blw += "b"
	}
	if l {
		blw += "l"
	}
	if w {
		blw += "w"
	}
	dstFilename := fmt.Sprintf("%v.%v.%v.jpg", strings.TrimSuffix(filepath.Base(imageFilename), filepath.Ext(imageFilename)), blw, strings.ToLower(detection.AlgoID))

	if len(detection.Blocks) == 0 {
		return fmt.Errorf("Failed to annotate. No blocks, lines, or words in: %s\n", coordFilename)
	}

	return annotate(buf, detection, b, l, w, dstFilename)
}
