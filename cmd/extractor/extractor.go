package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ughe/tigerocr/ocr"
)

func fieldToString(result *ocr.Result, key string) (string, error) {
	switch key {
	case "service":
		return result.Service, nil
	case "version":
		return result.Version, nil
	case "text":
		return result.FullText, nil
	case "milliseconds":
		return fmt.Sprintf("%d", result.Duration), nil
	case "date":
		return result.Date, nil
	case "raw":
		return string(result.Raw), nil
	}
	err := fmt.Errorf("fieldToString: %v not in struct ocr.Result", key)
	return "", err
}

func main() {
	// Keys must be valid json entries into the OCR output format
	fields := map[string]string{
		"service":      "Extracts the provider service",
		"version":      "Extracts the version",
		"text":         "Extracts full text",
		"milliseconds": "Extracts duration in milliseconds",
		"date":         "Extracts the date",
		"raw":          "Extracts raw data as json with x y coordinates",
	}

	// Flags
	flags := make(map[string]*bool)
	for k, v := range fields {
		flags[k] = flag.Bool(k, false, v)
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Read in the result file
	name := flag.Arg(0)
	f, err := os.Open(name)
	if err != nil {
		log.Fatalf("Failed to open file '%v': %v\n", name, err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("Failed to read json file '%v': %v\n", name, err)
	}

	// Parse the result file
	var result ocr.Result
	err = json.Unmarshal(buf, &result)
	if err != nil {
		log.Fatalf("Failed to parse json: %v\n", err)
	}

	// Extract the provided field
	for k, b := range flags {
		if *b {
			out, err := fieldToString(&result, k)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(out)
			return
		}
	}

	// No flag selected
	flag.Usage()
}
