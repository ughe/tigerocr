package main

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ughe/tigerocr/ocr"
)

func abs(n int) int {
	if n >= 0 {
		return n
	} else {
		return -n
	}
}

func convertToBLW(img []byte, result *ocr.Result, c ocr.Client, dstFilename string) error {
	m, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return err
	}
	mb := m.Bounds()
	width, height := abs(mb.Max.X-mb.Min.X), abs(mb.Max.Y-mb.Min.Y)

	detection, err := c.ResultToDetection(result, width, height)
	if err != nil {
		return err
	}

	encoded, err := json.Marshal(detection)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dstFilename, encoded, 0600); err != nil {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: ocr-json-to-blw image.jpg result.json")
	}

	imgName := os.Args[1]
	img, err := ioutil.ReadFile(imgName)
	if err != nil {
		log.Fatalf("Failed to read image: %v", err)
	}
	raw, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		log.Fatalf("Failed to read json: %v", err)
	}

	var result ocr.Result
	err = json.Unmarshal(raw, &result)
	if err != nil {
		log.Fatalf("Failed to unmarshal json into Result: %v", err)
	}

	dst := strings.TrimSuffix(filepath.Base(imgName), filepath.Ext(imgName)) + "." + strings.ToLower(result.Service)[:3] + ".blw"

	switch result.Service {
	case "AWS":
		err = convertToBLW(img, &result, ocr.AWSClient{""}, dst)
	case "Azure":
		err = convertToBLW(img, &result, ocr.AzureClient{""}, dst)
	case "GCP":
		err = convertToBLW(img, &result, ocr.GCPClient{""}, dst)
	default:
		log.Fatalf("Service %v is not {AWS, Azure, GCP}", result.Service)
	}
	if err != nil {
		log.Fatal(err)
	}
}
