package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ughe/tigerocr/ocr"
	"github.com/ughe/tigerocr/util"
)

var (
	blue   = color.RGBA{0, 0, 255, 255}
	red    = color.RGBA{255, 0, 0, 255}
	orange = color.RGBA{255, 165, 0, 255}
)

func annotate(result *ocr.Result, src []byte, c ocr.Client, col color.Color, b, l, w bool, dstName string) {
	m, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		log.Fatalf("annotate:image.Decode: %v", err)
	}
	mb := m.Bounds()
	width, height := util.Abs(mb.Max.X-mb.Min.X), util.Abs(mb.Max.Y-mb.Min.Y)

	detection, err := c.ResultToDetection(result, width, height)
	if err != nil {
		log.Fatalf("ResultToDetection:%v: %v", result.Service, err)
	}
	dst, err := ocr.Annotate(src, detection, col, b, l, w)
	if err != nil {
		log.Fatalf("Annotate:%v: %v", result.Service, err)
	}
	util.Write(dst, dstName)
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
	img, err := util.Read(flag.Arg(0))
	if err != nil {
		log.Fatalf("Failed to read image: %v", err)
	}
	raw, err := util.Read(flag.Arg(1))
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
		annotate(&result, img, ocr.AWSClient{""}, orange, b, l, w, dstName)
	case "Azure":
		annotate(&result, img, ocr.AzureClient{""}, blue, b, l, w, dstName)
	case "GCP":
		annotate(&result, img, ocr.GCPClient{""}, red, b, l, w, dstName)
	default:
		log.Fatalf("Service %v is not {AWS, Azure, GCP}", result.Service)
	}
}
