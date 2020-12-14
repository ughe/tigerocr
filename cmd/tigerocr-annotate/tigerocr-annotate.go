package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ughe/tigerocr/ocr"
)

var (
	blue   = color.RGBA{0, 0, 255, 255}
	red    = color.RGBA{255, 0, 0, 255}
	orange = color.RGBA{255, 165, 0, 255}
)

func abs(n int) int {
	if n >= 0 {
		return n
	} else {
		return -n
	}
}

func annotate(result *ocr.Result, src []byte, c ocr.Client, col color.Color, b, l, w bool, dstFilename string) error {
	m, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return err
	}
	mb := m.Bounds()
	width, height := abs(mb.Max.X-mb.Min.X), abs(mb.Max.Y-mb.Min.Y)

	detection, err := c.ResultToDetection(result, width, height)
	if err != nil {
		return err
	}
	dst, err := ocr.Annotate(src, detection, col, b, l, w)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dstFilename, dst, 0600); err != nil {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: tigerocr-annotate image.jpg result.json")
	}

	bo := flag.Bool("b", false, "Draw blocks (or regions)")
	lo := flag.Bool("l", false, "Draw lines")
	wo := flag.Bool("w", false, "Draw words [Default]")
	flag.Parse()

	b, l, w := false, false, false
	if *bo {
		b = true
	}
	if *lo {
		l = true
	}
	if *wo {
		w = true
	}
	if !b && !l && !w {
		w = true // draw words if all others are unspecified
	}

	imgName := flag.Arg(0)
	img, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("Failed to read image: %v", err)
	}
	raw, err := ioutil.ReadFile(flag.Arg(1))
	if err != nil {
		log.Fatalf("Failed to read json: %v", err)
	}

	var result ocr.Result
	err = json.Unmarshal(raw, &result)
	if err != nil {
		log.Fatalf("Failed to unmarshal json into Result: %v", err)
	}

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
	dstName := fmt.Sprintf("%v.%v.%v.jpg",
		strings.TrimSuffix(filepath.Base(imgName), filepath.Ext(imgName)),
		blw, strings.ToLower(result.Service))

	switch result.Service {
	case "AWS":
		err = annotate(&result, img, ocr.AWSClient{""}, orange, b, l, w, dstName)
	case "Azure":
		err = annotate(&result, img, ocr.AzureClient{""}, blue, b, l, w, dstName)
	case "GCP":
		err = annotate(&result, img, ocr.GCPClient{""}, red, b, l, w, dstName)
	default:
		log.Fatalf("Service %v is not {AWS, Azure, GCP}", result.Service)
	}
	if err != nil {
		log.Fatal(err)
	}
}
