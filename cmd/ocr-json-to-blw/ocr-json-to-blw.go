package main

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ughe/tigerocr/ocr"
	"github.com/ughe/tigerocr/util"
)

func convertToBLW(img []byte, result *ocr.Result, c ocr.Client, dst string) {
	m, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		log.Fatalf("Failed to Decode the image: %v", err)
	}
	mb := m.Bounds()
	width, height := util.Abs(mb.Max.X-mb.Min.X), util.Abs(mb.Max.Y-mb.Min.Y)

	detection, err := c.ResultToDetection(result, width, height)
	if err != nil {
		log.Fatalf("Failed to convert raw json to detection: %v", err)
	}

	encoded, err := json.Marshal(detection)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	util.Write(encoded, dst)
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: ocr-json-to-blw image.jpg result.json")
	}

	imgName := os.Args[1]
	img, err := util.Read(imgName)
	if err != nil {
		log.Fatalf("Failed to read image: %v", err)
	}
	raw, err := util.Read(os.Args[2])
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
		convertToBLW(img, &result, ocr.AWSClient{""}, dst)
	case "Azure":
		convertToBLW(img, &result, ocr.AzureClient{""}, dst)
	case "GCP":
		convertToBLW(img, &result, ocr.GCPClient{""}, dst)
	default:
		log.Fatalf("Service %v is not {AWS, Azure, GCP}", result.Service)
	}
}
