package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ughe/explorer"
	"github.com/ughe/tigerocr/ocr"
)

const DPI = "300"
const FMT = "png"
const QUA = "00" // imagemagick.org/script/command-line-options.php#quality

const FILE_PERM = 0444
const DIR_PERM = 0755

// Count pages in PDF using GhostScript executable
func countPages(gs, pdf string) (int, error) {
	cmd := exec.Command(gs, "-q", "-dNOSAFER", "-dNODISPLAY", "-c",
		"("+pdf+") (r) file runpdfbegin pdfpagecount = quit")
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return 0, fmt.Errorf("[ERROR] GhostScript pdfpagecount: %s", string(exitError.Stderr))
		}
		return 0, err
	}
	i, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, err
	}
	return i, nil
}

// Extract text from PDFs. No functionality for telling if it is useful
func extractText(gs, pdf, dstPath string, pageCount int) error {
	pdfName := strings.TrimSuffix(filepath.Base(pdf), filepath.Ext(pdf))
	nDigits := strconv.Itoa(int(math.Ceil(math.Log10(float64(pageCount)))))
	ptr := pdfName + "-%0" + nDigits + "d"
	for i := 0; i < pageCount; i++ {
		is := strconv.Itoa(i + 1) // gs uses page numbers from 1 instead of 0
		dst := path.Join(dstPath, fmt.Sprintf(ptr, i)+".txt")
		cmd := exec.Command(gs, "-q", "-sDEVICE=txtwrite", "-dBATCH", "-dNOPAUSE",
			"-dFirstPage="+is, "-dLastPage="+is, "-sOutputFile="+dst, pdf)
		out, err := cmd.Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				return fmt.Errorf("[ERROR] GhostScript txtwrite: %s", string(exitError.Stderr))
			}
			return err
		}
		stdout := strings.TrimSpace(string(out))
		if stdout != "" {
			return fmt.Errorf("[WARNING] GhostScript unexpected stdout: %s", stdout)
		}
		err = os.Chmod(dst, FILE_PERM)
		if err != nil {
			return err
		}
	}
	return nil
}

// Convert PDF to Image using ImageMagick executable. Returns list of pointers
func convertPDF(magick[]string, dstDir, pdf string, pageCount int) ([]string, error) {
	pdfName := strings.TrimSuffix(filepath.Base(pdf), filepath.Ext(pdf))
	nDigits := strconv.Itoa(int(math.Ceil(math.Log10(float64(pageCount)))))
	ptr := pdfName + "-%0" + nDigits + "d"
	out := path.Join(dstDir, ptr+"."+FMT)
	magick = append(magick, "-density", DPI, "-alpha", "off", "-quality", QUA, pdf, out)
	_, err := exec.Command(magick[0], magick[1:]...).Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("ImageMagick: %s", string(exitError.Stderr))
		}
		return nil, err
	}
	// Create pointers (matching magick's output format)
	ptrs := make([]string, pageCount)
	for i := 0; i < pageCount; i++ {
		ptrs[i] = fmt.Sprintf(ptr, i)
		if err := os.Chmod(path.Join(dstDir, ptrs[i]+"."+FMT), FILE_PERM); err != nil {
			return nil, err
		}
	}
	return ptrs, nil
}

// Run OCR in series
func runSyncOCR(ptrs []string, services map[string]ocr.Client, artDir, imgsDir, ocrDir string) (int, map[string]map[string]string, error) {
	errs := make(map[string]map[string]string) // service name -> ptr -> err msg
	logs := make(map[string]map[string]int64)  // service name -> ptr -> time
	for s, _ := range services {
		errs[s] = make(map[string]string)
		logs[s] = make(map[string]int64, len(ptrs))
	}

	serviceKeys := make([]string, 0, len(services))
	for k, _ := range services {
		serviceKeys = append(serviceKeys, k)
	}
	sort.Strings(serviceKeys)
	sort.Strings(ptrs)

	// Run each ptr, in order, on each service, in alphabetical order
	os.MkdirAll(ocrDir, DIR_PERM)
	for _, ptr := range ptrs {
		imgPath := path.Join(imgsDir, ptr+"."+FMT)
		f, err := os.Open(imgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERR] Failed to open image: %s\n", imgPath)
			continue
		}
		defer f.Close()
		buf, err := ioutil.ReadAll(bufio.NewReader(f))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERR] Failed to read image: %s\n", imgPath)
			continue
		}
		for _, s := range serviceKeys {
			jsonPath := path.Join(ocrDir, ptr+"."+s+".json")
			_, millis, err := runService(buf, services[s], jsonPath)
			if err != nil {
				errs[s][ptr] = fmt.Sprintf("%v", err)
			} else {
				logs[s][ptr] = millis
			}
		}
	}
	// Finished OCR

	// Dump logs to explorer/data/artifacts/{ocr-errs.txt,ocr-logs.txt}
	nErrs := 0
	ocrerrs := make([]string, 0)
	ocrlogs := make([]string, 0, len(ptrs)*len(services)) // output log
	results := make(map[string]map[string]string)         // service name -> ptr -> time
	for s, _ := range services {
		results[s] = make(map[string]string, len(ptrs))
	}

	// Need to replicate order in which OCR was run
	for _, ptr := range ptrs {
		for _, s := range serviceKeys {
			if millis, ok := logs[s][ptr]; ok {
				// Append duration to logs and results
				const MILLIS_PER_SEC = 1000.0
				// Format %.02f without any trailing zeros
				secs := fmt.Sprintf("%.02f", float64(millis)/MILLIS_PER_SEC)
				if secs[len(secs)-3:] == ".00" {
					secs = secs[:len(secs)-3]
				} else if secs[len(secs)-1:] == "0" {
					secs = secs[:len(secs)-1]
				}
				if len(secs) == 0 {
					secs = "0"
				}
				ocrlogs = append(ocrlogs, s+","+ptr+","+secs)
				results[s][ptr] = secs
			} else if err, ok := errs[s][ptr]; ok {
				// Append err to logs
				nErrs++
				safeErr := strings.Replace(strings.Replace(err, "\n", "", -1), ",", "", -1)
				ocrerrs = append(ocrerrs, s+","+ptr+","+safeErr)
			} else {
				return 0, nil, fmt.Errorf("Internal Error: a log is missing")
			}
		}
	}

	// Write logs
	if err := ioutil.WriteFile(path.Join(artDir, "ocr-errs.txt"), []byte(strings.Join(ocrerrs, "\n")), FILE_PERM); err != nil {
		return 0, nil, err
	}
	if err := ioutil.WriteFile(path.Join(artDir, "ocr-logs.txt"), []byte(strings.Join(ocrlogs, "\n")), FILE_PERM); err != nil {
		return 0, nil, err
	}

	return nErrs, results, nil
}

// Creates the explorer website
func createExplorer(ptrs []string, metrics map[string][]string, txtDirs, pdf string) error {
	os.MkdirAll(path.Join("explorer", "js"), DIR_PERM)
	os.MkdirAll(path.Join("explorer", "data"), DIR_PERM)

	// Write Explorer website static files
	indexDst := path.Join("explorer", "index.html")
	if err := ioutil.WriteFile(indexDst, explorer.Index, FILE_PERM); err != nil {
		return err
	}
	styleDst := path.Join("explorer", "style.css")
	if err := ioutil.WriteFile(styleDst, explorer.Style, FILE_PERM); err != nil {
		return err
	}
	mainDst := path.Join("explorer", "js", "main.js")
	if err := ioutil.WriteFile(mainDst, explorer.Main, FILE_PERM); err != nil {
		return err
	}
	gridDst := path.Join("explorer", "js", "grid.js")
	if err := ioutil.WriteFile(gridDst, explorer.Grid, FILE_PERM); err != nil {
		return err
	}

	// Create config.csv
	configDst := path.Join("explorer", "data", "config.csv")
	pdfName := strings.TrimSuffix(filepath.Base(pdf), filepath.Ext(pdf))
	config := fmt.Sprintf("title,%s Explorer\nimgs-fmt,%s\ntxts-dirs,%s\n[]links,Data;data/\n[]range,CER;0;1\n[]range,Seconds;0;12\n", pdfName, FMT, txtDirs)
	if err := ioutil.WriteFile(configDst, []byte(config), FILE_PERM); err != nil {
		return err
	}

	// Create results.csv
	resultsDst := path.Join("explorer", "data", "results.csv")
	results := make([]string, 0)
	results = append(results, "ptr,"+strings.Join(ptrs, ","))
	for k, v := range metrics {
		results = append(results, k+","+strings.Join(v, ","))
	}
	resultBytes := []byte(strings.Join(results, "\n"))
	if err := ioutil.WriteFile(resultsDst, resultBytes, FILE_PERM); err != nil {
		return err
	}

	return nil
}

// Union of two []string, similar to unix `comm -12 a b`
func comm(a, b []string) []string {
	sort.Strings(a)
	sort.Strings(b)

	c := make([]string, 0)

	for i, j := 0, 0; i < len(a) && j < len(b); {
		if a[i] == b[j] {
			c = append(c, a[i])
			i++
			j++
		} else if a[i] < b[j] {
			i++
		} else { // a[i] > b[j]
			j++
		}
	}
	return c
}

func exploreCommand(keys string, aws, azu, gcp bool, pdf string) error {
	// Check pdf file exists
	if _, err := os.Stat(pdf); err != nil {
		return err
	}

	// Check for ImageMagick
	var magick []string
	if _, err := exec.LookPath("magick"); err == nil {
		magick = []string{"magick", "convert"}
	} else if _, err := exec.LookPath("convert"); err == nil {
		magick = []string{"convert"}
	} else {
		return fmt.Errorf("Missing magick convert (imagemagick.org)")
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

	// Set up OCR Clients
	services := initServices(keys, aws, azu, gcp)
	// Sort the services alphabetically
	providers := make([]string, 0, len(services))
	for s, _ := range services {
		providers = append(providers, s)
	}
	sort.Strings(providers)

	// Get number of pages
	pc, err := countPages(gs, pdf)
	if err != nil {
		return err
	}

	// Convert PDF to images
	fmt.Printf("[INFO] PDF to %s (Total: %d) ... \t\t", FMT, pc)
	start := time.Now()
	imgsDir := path.Join("explorer", "data", "imgs")
	os.MkdirAll(imgsDir, DIR_PERM)
	ptrs, err := convertPDF(magick, imgsDir, pdf, pc)
	if err != nil {
		return err
	}
	secs := int(time.Since(start) / time.Second)
	fmt.Printf("%d secs\n", secs)

	// Extract text from PDF
	txtsDir := path.Join("explorer", "data", "txts")
	dstPath := path.Join(txtsDir, "pdf")
	os.MkdirAll(dstPath, DIR_PERM)
	fmt.Printf("[INFO] PDF to TXT (Total: %d) ... \t\t", pc)
	start = time.Now()
	err = extractText(gs, pdf, dstPath, pc)
	if err != nil {
		return err
	}
	secs = int(time.Since(start) / time.Second)
	fmt.Printf("%d secs\n", secs)

	// Execute OCR
	estCost := 0.0015
	nCalls := len(ptrs) * len(services)
	fmt.Printf("[ATTN] Estimate: $%.2f (%d ops). Run? [N/y]: \t", float64(nCalls)*estCost, nCalls)
	var resp string
	_, err = fmt.Scanf("%s\n", &resp)
	if err != nil {
		return err
	}
	resp = strings.ToLower(strings.TrimSpace(resp))
	if resp != "y" && resp != "yes" {
		fmt.Printf("[DONE] OCR cancelled. Created explorer dir\n")
		return nil
	}

	fmt.Printf("[INFO] Executing OCR (Total: %d) ... \t\t", len(ptrs)*len(services))
	start = time.Now()
	artDir := path.Join("explorer", "data", "artifacts")
	ocrDir := path.Join(artDir, "json")
	nFailed, results, err := runSyncOCR(ptrs, services, artDir, imgsDir, ocrDir)
	nSuccess := nCalls - nFailed // Should always equal number of json files created
	if err != nil {
		return err
	}
	secs = int(time.Since(start) / time.Second)
	fmt.Printf("%d secs. Total errors: %d\n", secs, nFailed)

	// Convert to BLW
	fmt.Printf("[INFO] JSON to BLW (Total: %d) ... \t\t", nSuccess)
	start = time.Now()
	blwDir := path.Join(artDir, "blw")
	os.MkdirAll(blwDir, DIR_PERM)
	for s, ptr_ := range results {
		for ptr, _ := range ptr_ {
			img := path.Join(imgsDir, ptr+"."+FMT)
			jsn := path.Join(ocrDir, ptr+"."+s+"."+"json")
			if _, err := convert(img, jsn, blwDir); err != nil {
				return err
			}
		}
	}
	secs = int(time.Since(start) / time.Second)
	fmt.Printf("%d secs\n", secs)

	// Extract TXT
	txtDirs := "pdf"
	fmt.Printf("[INFO] BLW to TXT ... \t\t\t\t")
	start = time.Now()
	for s, ptr_ := range results {
		S := strings.ToUpper(s)
		txtDirs += ";" + S
		sDir := path.Join(txtsDir, S)
		os.MkdirAll(sDir, DIR_PERM)
		for ptr, _ := range ptr_ {
			blwFile := path.Join(blwDir, ptr+"."+s+".blw")
			var detection ocr.Detection
			raw, err := ioutil.ReadFile(blwFile)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(raw, &detection); err != nil {
				return err
			}
			if err := ioutil.WriteFile(path.Join(sDir, ptr+".txt"), []byte(detection.Plaintext()), FILE_PERM); err != nil {
				return err
			}
		}
	}
	secs = int(time.Since(start) / time.Second)
	fmt.Printf("%d secs\n", secs)

	// Determine final list of common pointers
	sptrs := make(map[string][]string)
	var anyKey string
	for s, ptr_ := range results {
		anyKey = s
		sptrs[s] = make([]string, 0)
		for ptr, _ := range ptr_ {
			sptrs[s] = append(sptrs[s], ptr)
		}
	}
	unified := make([]string, len(sptrs[anyKey]))
	copy(unified, sptrs[anyKey])
	var res []string
	for s, p := range sptrs {
		unified = comm(unified, p)
		res = append(res, fmt.Sprintf("%s: %d", s, len(p)))
	}
	sort.Strings(unified)

	// TODO: Run Levenshtein

	// Create metrics
	fmt.Printf("[INFO] Creating Metrics ... \t\t\t")
	start = time.Now()
	metrics := make(map[string][]string)
	// Time
	for _, s := range providers {
		name := fmt.Sprintf("%s Seconds", strings.ToUpper(s))
		fields := make([]string, 0, len(unified))
		for _, ptr := range unified {
			fields = append(fields, results[s][ptr])
		}
		metrics[name] = fields
	}
	// TODO: CER, Character Count
	secs = int(time.Since(start) / time.Second)
	fmt.Printf("%d secs\n", secs)

	// Create explorer
	fmt.Printf("[INFO] Creating Explorer ... \t\t\t")
	err = createExplorer(unified, metrics, txtDirs, pdf)
	fmt.Printf("done\n")

	fmt.Printf("[INFO] Comparable Ptrs: %d (out of %d). %s\n", len(unified), len(ptrs), strings.Join(res, ","))
	fmt.Printf("[DONE] Run: tigerocr serve ./explorer\n")

	return nil
}
