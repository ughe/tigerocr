package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ughe/explorer"
)

const DPI = "300"
const FMT = "png"
const QUA = "00" // imagemagick.org/script/command-line-options.php#quality

// Count pages in PDF using GhostScript executable
func countPages(gs, pdf string) (int, error) {
	cmd := exec.Command(gs, "-q", "-dNOSAFER", "-dNODISPLAY", "-c",
		"("+pdf+") (r) file runpdfbegin pdfpagecount = quit")
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return 0, fmt.Errorf("GhostScript: %s", string(exitError.Stderr))
		}
		return 0, err
	}
	i, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, err
	}
	return i, nil
}

// Convert PDF to Image using ImageMagick executable
func convertPDF(dstDir, pdf string, pageCount int) error {
	pdfName := strings.TrimSuffix(filepath.Base(pdf), filepath.Ext(pdf))
	nDigits := strconv.Itoa(int(math.Ceil(math.Log10(float64(pageCount)))))
	out := path.Join(dstDir, pdfName+"-%0"+nDigits+"d."+FMT)
	_, err := exec.Command("magick", "convert", "-density", DPI,
		"-alpha", "off", "-quality", QUA, pdf, out).Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("ImageMagick: %s", string(exitError.Stderr))
		}
		return err
	}
	return nil
}

func exploreCommand(pdf string) error {
	// Check pdf file exists
	if _, err := os.Stat(pdf); err != nil {
		return err
	}

	// Check for ImageMagick
	if _, err := exec.LookPath("magick"); err != nil {
		return fmt.Errorf("Missing magick (imagemagick.org)")
	}

	// Check for GhostScript
	var gs string
	if path, err := exec.LookPath("gs"); err == nil {
		gs = path
	} else if path, err := exec.LookPath("gswin32c"); err == nil {
		gs = path
	} else if path, err := exec.LookPath("gswin64c"); err == nil {
		gs = path
	} else {
		return fmt.Errorf("Missing gs (ghostscript.com)")
	}

	// Get number of pages
	pc, err := countPages(gs, pdf)
	if err != nil {
		return err
	}

	// Convert PDF to images
	fmt.Printf("[INFO] Converting PDF to %d images (%s) ...\n", pc, FMT)
	imgsDir := path.Join("explorer", "data", "imgs")
	os.MkdirAll(imgsDir, 0755)
	err = convertPDF(imgsDir, pdf, pc)
	if err != nil {
		return err
	}

	// TODO: If PDF has text associated, save it page-by-page
	// TODO: Copy Explorer Info
	// TODO: Execute OCR
	// TODO: Run Levenshtein
	// TODO: Create config.csv and results.csv

	// Copy explorer static files
	indexDst := path.Join("explorer", "index.html")
	if err := ioutil.WriteFile(indexDst, explorer.Index, 0644); err != nil {
		return err
	}
	styleDst := path.Join("explorer", "style.css")
	if err := ioutil.WriteFile(styleDst, explorer.Style, 0644); err != nil {
		return err
	}
	os.MkdirAll(path.Join("explorer", "js"), 0755) // Create js dir
	mainDst := path.Join("explorer", "js", "main.js")
	if err := ioutil.WriteFile(mainDst, explorer.Main, 0644); err != nil {
		return err
	}
	gridDst := path.Join("explorer", "js", "grid.js")
	if err := ioutil.WriteFile(gridDst, explorer.Grid, 0644); err != nil {
		return err
	}
	fmt.Printf("[DONE] Run: cd explorer && tigerocr serve\n")

	return nil
}
