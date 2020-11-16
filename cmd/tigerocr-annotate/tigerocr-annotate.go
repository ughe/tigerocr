package main

import (
	"bufio"
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

func read(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func write(buf []byte, fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	n := 0
	for n < len(buf) {
		m, err := f.Write(buf[n:])
		if err != nil {
			return err
		}
		n += m
	}
	return nil
}

func annotate(service string, imgName string, src []byte, jsn []byte, c ocr.Client, col color.Color, b, l, w bool) {
	m, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		log.Fatalf("Failed to Decode the image: %v", err)
	}
	mb := m.Bounds()
	width, height := abs(mb.Max.X-mb.Min.X), abs(mb.Max.Y-mb.Min.Y)

	detection, err := c.RawToDetection(jsn, width, height)
	if err != nil {
		log.Fatalf("Failed to convert %v raw json to detection: %v", service, err)
	}
	dst, err := ocr.Annotate(src, detection, col, b, l, w)
	if err != nil {
		log.Fatalf("Failed to annotate %v img: %v", service, err)
	}
	name := strings.TrimSuffix(filepath.Base(imgName), filepath.Ext(imgName))
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
	// Write to CWD
	write(dst, fmt.Sprintf("%v.%v.%v.jpg", name, blw, strings.ToLower(service)))
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
	img, err := read(flag.Arg(0))
	if err != nil {
		log.Fatalf("Failed to read image: %v", err)
	}
	raw, err := read(flag.Arg(1))
	if err != nil {
		log.Fatalf("Failed to read json: %v", err)
	}

	var result ocr.Result
	err = json.Unmarshal(raw, &result)
	if err != nil {
		log.Fatalf("Failed to unmarshal json into Result: %v", err)
	}

	switch result.Service {
	case "AWS":
		annotate(result.Service, imgName, img, result.Raw,
			ocr.AWSClient{""}, orange, b, l, w)
	case "Azure":
		annotate(result.Service, imgName, img, result.Raw,
			ocr.AzureClient{""}, blue, b, l, w)
	case "GCP":
		annotate(result.Service, imgName, img, result.Raw,
			ocr.GCPClient{""}, red, b, l, w)
	default:
		log.Fatalf("Service %v is not {AWS, Azure, GCP}", result.Service)
	}
}
