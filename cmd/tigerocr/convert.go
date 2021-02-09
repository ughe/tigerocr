package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/ughe/tigerocr/ocr"
)

func convert(imgFilename, jsnFilename, basename string) (string, error) {
	img, err := ioutil.ReadFile(imgFilename)
	if err != nil {
		return "", err
	}
	raw, err := ioutil.ReadFile(jsnFilename)
	if err != nil {
		return "", err
	}

	var result ocr.Result
	err = json.Unmarshal(raw, &result)
	if err != nil {
		return "", err
	}

	dstFilename := path.Join(basename, strings.TrimSuffix(filepath.Base(imgFilename), filepath.Ext(imgFilename))+"."+strings.ToLower(result.Service)[:3]+".blw")

	detection, err := convertToBLW(img, &result)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(detection)
	if err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(dstFilename, encoded, 0600); err != nil {
		return "", err
	}
	return dstFilename, nil
}

func convertCommand(imgFilename, jsnFilename string) error {
	dstFilename, err := convert(imgFilename, jsnFilename, "")
	if err != nil {
		return err
	}
	fmt.Printf("[INFO] Converted json to blw: %v\n", dstFilename)
	return nil
}
