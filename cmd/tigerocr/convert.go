package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
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

	detection, err := convertToBLW(img, raw, jsnFilename)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(detection)
	if err != nil {
		return "", err
	}

	dstFilename := path.Join(basename, strings.TrimSuffix(filepath.Base(imgFilename), filepath.Ext(imgFilename))+"."+strings.ToLower(detection.AlgoID)[:3]+".blw")

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

func stripExt(ls []os.FileInfo, subExt string) ([]string, string) {
	files := make([]string, 0, len(ls))
	for _, l := range ls {
		if !l.IsDir() {
			files = append(files, l.Name())
		}
	}
	ext := filepath.Ext(files[len(files)/2]) // Choose middle ext
	names := make([]string, 0, len(files))
	for _, file := range files {
		if file[len(file)-len(ext):] == ext {
			name := file[:len(file)-len(ext)]
			if subExt != "" {
				name = strings.TrimRight(name, "."+subExt)
			}
			names = append(names, name)
		}
	}
	return names, ext
}

func convertCommandPDF(imgDir, blwDir, providerPrefix string) error {
	imgDirListing, err := ioutil.ReadDir(imgDir)
	if err != nil {
		return err
	}
	blwDirListing, err := ioutil.ReadDir(blwDir)
	if err != nil {
		return err
	}
	// Ensure same list of names and extensions
	imgNames, imgExt := stripExt(imgDirListing, "")
	if imgExt != ".png" && imgExt != ".jpg" && imgExt != ".jpeg" {
		return fmt.Errorf("Expected (png|jpg|jpeg). Found: %s", imgExt)
	}
	blwNames, blwExt := stripExt(blwDirListing, providerPrefix)
	if blwExt != ".json" && blwExt != ".blw" {
		return fmt.Errorf("Expected (blw|json). Found: %s", blwExt)
	}
	if len(imgNames) != len(blwNames) {
		fmt.Printf("[INFO] %s: %d\n[INFO] %s: %d\n", imgDir, len(imgNames), blwDir, len(blwNames))
	}
	pointers := comm(imgNames, blwNames)
	if len(pointers) == 0 {
		return fmt.Errorf("No matching image and blw files in: %s and %s", imgDir, blwDir)
	}
	fmt.Printf("[INFO] Number of pages: %d\n", len(pointers))

	dstFilename := pointers[0] + ".pdf"
	pdf := gofpdf.New("P", "pt", "Letter", "")
	fontSize := 12.0
	pdf.SetFont("Courier", "", fontSize)
	hideText := false // gs cannot read text that is hidden
	hiddenLayer := pdf.AddLayer("OCR", !hideText)
	for _, ptr := range pointers {
		// Read the image
		buf, err := ioutil.ReadFile(path.Join(imgDir, ptr+imgExt))
		if err != nil {
			return err
		}
		width, height, err := imgToWidthHeight(buf)
		if err != nil {
			return err
		}
		w, h := float64(width), float64(height)
		// Read the BLW or JSON file
		blwName := ptr + blwExt
		if providerPrefix != "" {
			blwName = fmt.Sprintf("%s.%s%s", ptr, providerPrefix, blwExt)
		}
		raw, err := ioutil.ReadFile(path.Join(blwDir, blwName))
		if err != nil {
			return err
		}
		detection, err := convertToBLW(buf, raw, blwName)
		if err != nil {
			return err
		}
		// Create a PDF page
		pdf.AddPageFormat("P", gofpdf.SizeType{Wd: w, Ht: h})
		opt := gofpdf.ImageOptions{
			ImageType:             imgExt[1:],
			ReadDpi:               false,
			AllowNegativePosition: true,
		}
		// Add the image to the PDF
		pdf.RegisterImageOptionsReader(ptr, opt, bytes.NewReader(buf)).SetDpi(float64(DPI))
		// Add the text in behind the image (instead of in invisible layer)
		pdf.BeginLayer(hiddenLayer)
		fmt.Println("[INFO] Drawing text on a page")
		for _, b := range detection.Blocks {
			for _, l := range b.Lines {
				for _, w := range l.Words {
					bnds, err := ocr.DecodeBounds(w.Bounds)
					if err != nil {
						return err
					}
					ww, wh := float64(bnds.W), float64(bnds.H) // Word W,H
					size := fitFontSize(w.Text, ww)
					if size != 0 && size != fontSize {
						fontSize = size // Update last font size to new
						pdf.SetFontSize(fontSize)
					}
					// If font is smaller than height, shift up by diff
					_, h := pdf.GetFontSize()
					diff := 0.0
					if wh > h {
						diff = (wh - h)
					}
					// Set font if the size has changed
					pdf.Text(float64(bnds.X), float64(bnds.Y)+wh-diff, w.Text+" ")
				}
			}
		}
		pdf.EndLayer()
		// Image goes on top of the text
		pdf.ImageOptions(ptr, 0, 0, w, h, false, opt, 0, "")
	}
	if err := pdf.OutputFileAndClose(dstFilename); err != nil {
		return err
	}

	fmt.Printf("[INFO] Converted imgs and blw to pdf: %v\n", dstFilename)
	return nil
}

// Fits font (courier) size to width of string s
func fitFontSize(s string, width float64) float64 {
	if len(s) == 0 {
		return 0
	}
	// Create dummy PDF to call GetStringWidth on font size without editing
	pdf := gofpdf.New("P", "pt", "Letter", "")
	pdf.SetFont("Courier", "", 12)
	// Estimate the correct font size
	fs := float64(int(width / float64(len(s)) * 1.6))
	pdf.SetFontSize(fs)
	// Scale font down if too big
	for pdf.GetStringWidth(s) > width && fs > 1 {
		fs = fs - 1
		pdf.SetFontSize(fs)
	}
	// Scale font up if too small
	for pdf.GetStringWidth(s) < width {
		fs = fs + 1
		pdf.SetFontSize(fs)
	}
	return fs
}
