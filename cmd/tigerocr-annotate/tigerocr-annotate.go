package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/ughe/tigerocr/ocr"
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

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: tigerocr-annotate image.jpg azure.json")
	}

	img, err := read(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to read image: %v", err)
	}
	jsn, err := read(os.Args[2])
	if err != nil {
		log.Fatalf("Failed to read json: %v", err)
	}

	var response ocr.AzureVisionResponse
	json.Unmarshal(jsn, &response)
	result, err := ocr.Annotate(img, response)
	if err != nil {
		log.Fatalf("Failed to annotate img: %v", err)
	}
	write(result, "out.jpg")
}
