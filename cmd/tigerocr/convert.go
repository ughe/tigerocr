package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ughe/tigerocr/ocr"
)

func convertCommand(imgFilename, jsnFilename string) error {
	img, err := ioutil.ReadFile(imgFilename)
	if err != nil {
		return err
	}
	raw, err := ioutil.ReadFile(jsnFilename)
	if err != nil {
		return err
	}

	var result ocr.Result
	err = json.Unmarshal(raw, &result)
	if err != nil {
		return err
	}

	dstFilename := strings.TrimSuffix(filepath.Base(imgFilename), filepath.Ext(imgFilename)) + "." + strings.ToLower(result.Service)[:3] + ".blw"

	detection, err := convertToBLW(img, &result)
	if err != nil {
		return err
	}

	encoded, err := json.Marshal(detection)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dstFilename, encoded, 0600); err != nil {
		return err
	} else {
		fmt.Printf("[INFO] Converted json to blw: %v\n", dstFilename)
		return nil
	}
}
