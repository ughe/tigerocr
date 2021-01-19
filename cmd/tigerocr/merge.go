package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ughe/tigerocr/ocr"
)

func mergeCommand(blwFilenames []string) error {
	blws := make([]ocr.Detection, len(blwFilenames))
	for i, f := range blwFilenames {
		if filepath.Ext(f) != ".blw" {
			return fmt.Errorf("Expected *.blw format instead of: %v", filepath.Ext(f))
		}
		raw, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &blws[i]); err != nil {
			return err
		}
	}

	blw, err := ocr.Merge(blws)
	if err != nil {
		return err
	}
	buf, err := json.Marshal(*blw)
	if err != nil {
		return err
	}
	dstFilename := fmt.Sprintf("%v.m%d.blw", strings.TrimSuffix(filepath.Base(blwFilenames[0]), ".blw"), len(blws))
	err = ioutil.WriteFile(dstFilename, buf, 0644)
	return err
}
