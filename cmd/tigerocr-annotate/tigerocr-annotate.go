package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"

	"github.com/ughe/tigerocr/ocr"
)

var (
	blue   = color.RGBA{0, 0, 255, 255}
	red    = color.RGBA{255, 0, 0, 255}
	orange = color.RGBA{255, 165, 0, 255}
)

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

func annotate(service string, src []byte, jsn []byte, c ocr.Client, col color.Color, b, l, w bool) {
	detection, err := c.RawToDetection(jsn)
	if err != nil {
		log.Fatalf("Failed to convert %v raw json to detection: %v", service, err)
	}
	dst, err := ocr.Annotate(src, detection, col, b, l, w)
	if err != nil {
		log.Fatalf("Failed to annotate %v img: %v", service, err)
	}
	write(dst, fmt.Sprintf("annotated_%v.jpg", service))
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: tigerocr-annotate image.jpg result.json")
	}

	img, err := read(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to read image: %v", err)
	}
	raw, err := read(os.Args[2])
	if err != nil {
		log.Fatalf("Failed to read json: %v", err)
	}

	var result ocr.Result
	err = json.Unmarshal(raw, &result)
	if err != nil {
		log.Fatalf("Failed to unmarshal json into Result: %v", err)
	}

	b, l, w := false, false, true // Words only

	switch result.Service {
	case "AWS":
		annotate(result.Service, img, result.Raw,
			ocr.AWSClient{""}, orange, b, l, w)
	case "Azure":
		annotate(result.Service, img, result.Raw,
			ocr.AzureClient{""}, blue, b, l, w)
	case "GCP":
		annotate(result.Service, img, result.Raw,
			ocr.GCPClient{""}, red, b, l, w)
	default:
		log.Fatalf("Service %v is not {AWS, Azure, GCP}", result.Service)
	}
}
