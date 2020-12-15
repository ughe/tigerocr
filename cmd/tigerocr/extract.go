package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ughe/tigerocr/ocr"
)

func extractCommand(filename string, stat, algoid, speed, date, text bool) error {
	var detection *ocr.Detection
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	switch filepath.Ext(filename) {
	case ".blw":
		if err := json.Unmarshal(raw, &detection); err != nil {
			return err
		}
	case ".json":
		// Dynamically convert json to blw format (entails overhead)
		var result ocr.Result
		if err := json.Unmarshal(raw, &result); err != nil {
			return err
		}

		var c ocr.Client
		switch result.Service {
		case "AWS":
			c = ocr.AWSClient{""}
		case "Azure":
			c = ocr.AzureClient{""}
		case "GCP":
			c = ocr.GCPClient{""}
		default:
			return fmt.Errorf("Service %v is not {AWS, Azure, GCP}", result.Service)
		}
		// We can use bogus width, height of zero since we do not return
		// any information related to the physical coordinates and we do
		// not save this new detection instance
		detection, err = c.ResultToDetection(&result, 0, 0)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Expected *.json or *.blw coordinate file instead of: %v", filepath.Ext(filename))
	}

	// Extract the fields
	if stat {
		// Human readable output
		b, l, w := detection.CountBLW()
		fmt.Printf("algoid: %s\n", detection.AlgoID)
		fmt.Printf("millis: %d\n", detection.Millis)
		fmt.Printf("date:   %s\n", detection.Date)
		fmt.Printf("blocks: %d\n", b)
		fmt.Printf("lines:  %d\n", l)
		fmt.Printf("words:  %d\n", w)
	} else if algoid {
		fmt.Printf("%s\n", detection.AlgoID)
	} else if speed {
		fmt.Printf("%d\n", detection.Millis)
	} else if date {
		fmt.Printf("%s\n", detection.Date)
	} else if text {
		fmt.Printf("%s\n", detection.Plaintext())
	} else {
		return fmt.Errorf("Error: no flags specified")
	}
	return nil
}
