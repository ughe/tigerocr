package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	"github.com/ughe/tigerocr/editdist"
	"github.com/ughe/tigerocr/ocr"
)

const DPI = 300
const FMT = "png"
const QUA = "00" // imagemagick.org/script/command-line-options.php#quality

const FILE_PERM = 0444
const DIR_PERM = 0755

// Count pages in PDF using GhostScript executable
func countPages(gs, pdfPath string) (int, error) {
	cmd := exec.Command(gs, "-q", "-dNOSAFER", "-dNODISPLAY", "-c",
		"("+pdfPath+") (r) file runpdfbegin pdfpagecount = quit")
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
func extractText(gs, pdfPath, dstPath string, pageCount int) error {
	pdfName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	nDigits := strconv.Itoa(int(math.Ceil(math.Log10(float64(pageCount)))))
	ptr := pdfName + "-%0" + nDigits + "d"
	for i := 0; i < pageCount; i++ {
		is := strconv.Itoa(i + 1) // gs uses page numbers from 1 instead of 0
		dst := path.Join(dstPath, fmt.Sprintf(ptr, i)+".txt")
		cmd := exec.Command(gs, "-q", "-sDEVICE=txtwrite", "-dBATCH", "-dNOPAUSE",
			"-dFirstPage="+is, "-dLastPage="+is, "-sOutputFile="+dst, pdfPath)
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
func convertPDF(magick []string, dstDir, pdfPath string, pageCount int) ([]string, error) {
	pdfName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	nDigits := strconv.Itoa(int(math.Ceil(math.Log10(float64(pageCount)))))
	ptr := pdfName + "-%0" + nDigits + "d"
	out := path.Join(dstDir, ptr+"."+FMT)
	magick = append(magick, "-density", fmt.Sprintf("%d", DPI), "-alpha", "off", "-quality", QUA, pdfPath, out)
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

// Executes OCR. Returns map from providers to map from pointer to seconds
func execOCR(ptrs []string, services map[string]ocr.Client, artDir, imgsDir, ocrDir string) (map[string]map[string]string, error) {
	os.MkdirAll(artDir, DIR_PERM)
	os.MkdirAll(ocrDir, DIR_PERM)

	fout, err := os.OpenFile(path.Join(artDir, "ocr-logs.txt"), os.O_CREATE|os.O_WRONLY, FILE_PERM)
	if err != nil {
		return nil, err
	}
	ferr, err := os.OpenFile(path.Join(artDir, "ocr-errs.txt"), os.O_CREATE|os.O_WRONLY, FILE_PERM)
	if err != nil {
		return nil, err
	}
	stdout := log.New(fout, "", 0)
	stderr := log.New(ferr, "", 0)

	// Run each ptr, in order, on each service, in alphabetical order
	for _, ptr := range ptrs {
		imgPath := path.Join(imgsDir, ptr+"."+FMT)
		err := runOCR(imgPath, ocrDir, stdout, stderr, services)
		if err != nil {
			return nil, err
		}
	}

	if err := fout.Close(); err != nil {
		return nil, err
	}
	if err := ferr.Close(); err != nil {
		return nil, err
	}

	// Read the logs
	results := make(map[string]map[string]string)
	for s, _ := range services {
		results[s] = make(map[string]string)
	}

	fout, err = os.Open(path.Join(artDir, "ocr-logs.txt"))
	if err != nil {
		return nil, err
	}
	defer fout.Close()
	scanner := bufio.NewScanner(fout)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			return nil, fmt.Errorf("Expected 1 ':' Found: %s", line)
		}
		file, millisStr := fields[0], fields[1]
		parts := strings.Split(file, ".")
		if len(parts) != 3 {
			return nil, fmt.Errorf("Expected 2 '.' Found: %s", file)
		}
		ptr, s := parts[0], parts[1]
		millis, err := strconv.Atoi(millisStr)
		if err != nil {
			return nil, err
		}
		// Format %.02f without any trailing zeros
		const MILLIS_PER_SEC = 1000.0
		secs := fmt.Sprintf("%.02f", float64(millis)/MILLIS_PER_SEC)
		if secs[len(secs)-3:] == ".00" {
			secs = secs[:len(secs)-3]
		} else if secs[len(secs)-1:] == "0" {
			secs = secs[:len(secs)-1]
		}
		if len(secs) == 0 {
			secs = "0"
		}
		results[s][ptr] = secs
	}
	return results, nil
}

// Creates the explorer website
func createExplorer(ptrs []string, metrics map[string][]string, metricLimits, metricOrder []string, txtDirs, pdfPath, baseDir string) error {
	os.MkdirAll(path.Join(baseDir, "js"), DIR_PERM)
	os.MkdirAll(path.Join(baseDir, "data"), DIR_PERM)

	// Write Explorer website static files
	indexDst := path.Join(baseDir, "index.html")
	if err := ioutil.WriteFile(indexDst, explorer.Index, FILE_PERM); err != nil {
		return err
	}
	styleDst := path.Join(baseDir, "style.css")
	if err := ioutil.WriteFile(styleDst, explorer.Style, FILE_PERM); err != nil {
		return err
	}
	mainDst := path.Join(baseDir, "js", "main.js")
	if err := ioutil.WriteFile(mainDst, explorer.Main, FILE_PERM); err != nil {
		return err
	}
	gridDst := path.Join(baseDir, "js", "grid.js")
	if err := ioutil.WriteFile(gridDst, explorer.Grid, FILE_PERM); err != nil {
		return err
	}

	// Create config.csv
	configDst := path.Join(baseDir, "data", "config.csv")
	pdfName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	config := fmt.Sprintf("title,%s Explorer\nimgs-fmt,%s\ntxts-dirs,%s\n[]links,Data;data/\n[]range,CER;0;1\n%s", pdfName, FMT, txtDirs, strings.Join(metricLimits, "\n"))
	if err := ioutil.WriteFile(configDst, []byte(config), FILE_PERM); err != nil {
		return err
	}

	// Create results.csv
	resultsDst := path.Join(baseDir, "data", "results.csv")
	results := make([]string, 0)
	results = append(results, "ptr,"+strings.Join(ptrs, ","))
	for _, k := range metricOrder {
		results = append(results, k+","+strings.Join(metrics[k], ","))
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

func exploreCommand(keys string, aws, azu, azuR, gcp bool, pdfPath string) error {
	// Check pdf file exists
	if _, err := os.Stat(pdfPath); err != nil {
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
	services := initServices(keys, aws, azu, azuR, gcp)
	// Sort the services alphabetically
	providers := make([]string, 0, len(services))
	for s, _ := range services {
		providers = append(providers, s)
	}
	sort.Strings(providers)

	// Check if explorer exists
	pdfName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	baseDir := fmt.Sprintf("explorer-%s", pdfName)
	_, err := os.Stat(baseDir)
	if err == nil || !os.IsNotExist(err) {
		return fmt.Errorf("Please remove directory: ./%s", baseDir)
	}

	// Get number of pages
	pc, err := countPages(gs, pdfPath)
	if err != nil {
		return err
	}

	// Convert PDF to images
	fmt.Printf("[INFO] PDF to %s (Total: %d) ... \t\t", strings.ToUpper(FMT), pc)
	start := time.Now()
	imgsDir := path.Join(baseDir, "data", "imgs")
	os.MkdirAll(imgsDir, DIR_PERM)
	ptrs, err := convertPDF(magick, imgsDir, pdfPath, pc)
	if err != nil {
		return err
	}
	secs := int(time.Since(start) / time.Second)
	fmt.Printf("%d secs\n", secs)

	// Extract text from PDF
	txtsDir := path.Join(baseDir, "data", "txts")
	dstPath := path.Join(txtsDir, "pdf")
	os.MkdirAll(dstPath, DIR_PERM)
	fmt.Printf("[INFO] PDF to TXT (Total: %d) ... \t\t", pc)
	start = time.Now()
	err = extractText(gs, pdfPath, dstPath, pc)
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
	_, err = fmt.Scanln(&resp)
	if err != nil {
		resp = "" // Treat any err as an empty response
	}
	resp = strings.ToLower(strings.TrimSpace(resp))
	if resp != "y" && resp != "yes" {
		fmt.Printf("[DONE] OCR cancelled. Created directory: %s\n", baseDir)
		return nil
	}

	fmt.Printf("[INFO] Executing OCR (Total: %d) ... \t\t", len(ptrs)*len(services))
	start = time.Now()
	artDir := path.Join(baseDir, "data", "artifacts")
	ocrDir := path.Join(artDir, "json")

	results, err := execOCR(ptrs, services, artDir, imgsDir, ocrDir)
	if err != nil {
		return err
	}
	// Calculate OCR stats
	tPassed, tFailed := 0, 0
	nPassed, nFailed := make(map[string]int), make(map[string]int)
	sFailed := ""
	for _, s := range providers {
		nPassed[s] = len(results[s])
		tPassed += nPassed[s]
		nFailed[s] = len(ptrs) - nPassed[s]
		tFailed += nFailed[s]
		if nFailed[s] > 0 { // Record errors if any
			sFailed += fmt.Sprintf("%s: %d, ", s, nFailed[s])
		}
	}
	if len(sFailed) >= 2 {
		sFailed = ". Errs: " + sFailed[:len(sFailed)-2] // Remove last ", "
	}
	secs = int(time.Since(start) / time.Second)
	fmt.Printf("%d secs%s\n", secs, sFailed)

	// Convert to BLW
	fmt.Printf("[INFO] JSON to BLW (Total: %d) ... \t\t", tPassed)
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
	txtDirs := "PDF"
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
	for _, p := range sptrs {
		unified = comm(unified, p)
	}
	for _, s := range providers {
		res = append(res, fmt.Sprintf("%s: %d", s, len(sptrs[s])))
	}
	sort.Strings(unified)

	// Create metrics
	metricOrder := make([]string, 0) // metric keys in order
	metrics := make(map[string][]string)
	metricLimits := make([]string, 0)

	// Levenshtein distance for each pair of providers
	fmt.Printf("[INFO] Running Levenshtein Distance ... \t")
	start = time.Now()
	firstLoop := true
	minl, maxl := 0, 0
	if len(providers) > 1 {
		for i := 0; i < len(providers); i++ {
			if len(providers) == 2 && i == 1 {
				break // Don't compare twice if only 2
			}
			// Run between provider i and i+1
			s1, s2 := strings.ToUpper(providers[i]), strings.ToUpper(providers[(i+1)%len(providers)])
			levs := make([]string, 0, len(unified))
			for _, ptr := range unified {
				bufa, err := ioutil.ReadFile(path.Join(txtsDir, s1, ptr+".txt"))
				if err != nil {
					return err
				}
				bufb, err := ioutil.ReadFile(path.Join(txtsDir, s2, ptr+".txt"))
				if err != nil {
					return err
				}
				dist := editdist.Levenshtein(bufa, bufb)
				levs = append(levs, strconv.Itoa(dist))
				if firstLoop {
					minl, maxl = dist, dist
					firstLoop = false
				} else if dist < minl {
					minl = dist
				} else if dist > maxl {
					maxl = dist
				}
			}
			name := fmt.Sprintf("%s vs %s", s1, s2)
			metrics[name] = levs
			metricOrder = append(metricOrder, name)
		}
	}
	// Compare PDF to first provider
	levs := make([]string, 0, len(unified))
	s1 := strings.ToUpper(providers[0])
	s2 := "PDF"
	for _, ptr := range unified {
		bufa, err := ioutil.ReadFile(path.Join(txtsDir, s1, ptr+".txt"))
		if err != nil {
			return err
		}
		bufb, err := ioutil.ReadFile(path.Join(txtsDir, s2, ptr+".txt"))
		if err != nil {
			return err
		}
		dist := editdist.Levenshtein(bufa, bufb)
		levs = append(levs, strconv.Itoa(dist))
		if firstLoop {
			minl, maxl = dist, dist
			firstLoop = false
		} else if dist < minl {
			minl = dist
		} else if dist > maxl {
			maxl = dist
		}
	}
	name := fmt.Sprintf("%s vs %s", s1, s2)
	metrics[name] = levs
	metricOrder = append(metricOrder, name)

	metricLimits = append(metricLimits, fmt.Sprintf(" vs ;%d;%d", minl, maxl))
	secs = int(time.Since(start) / time.Second)
	fmt.Printf("%d secs\n", secs) // Finished Levenshtein

	// Time
	mint, maxt := math.Inf(1), math.Inf(-1)
	for _, s := range providers {
		name := fmt.Sprintf("%s Seconds", strings.ToUpper(s))
		fields := make([]string, 0, len(unified))
		for _, ptr := range unified {
			f, err := strconv.ParseFloat(results[s][ptr], 64)
			if err != nil {
				return err
			}
			if f < mint {
				mint = f
			}
			if f > maxt {
				maxt = f
			}
			fields = append(fields, results[s][ptr])
		}
		metrics[name] = fields
		metricOrder = append(metricOrder, name)
	}
	metricLimits = append(metricLimits, fmt.Sprintf("Seconds;%.2f;%2.f", mint, maxt))

	// Word Count
	//fmt.Printf("[INFO] Running Word Count ... \t\t\t")
	//start = time.Now()
	// Count PDF words
	firstLoop = true
	minwc, maxwc := 0, 0
	wc := make([]string, 0, len(unified))
	for _, ptr := range unified {
		buf, err := ioutil.ReadFile(path.Join(txtsDir, "PDF", ptr+".txt"))
		if err != nil {
			return err
		}
		count := len(strings.Fields(string(buf)))
		if firstLoop { // Set initial min, max in first loop
			minwc, maxwc = count, count
			firstLoop = false
		} else if count < minwc {
			minwc = count
		} else if count > maxwc {
			maxwc = count
		}
		wc = append(wc, strconv.Itoa(count))
	}
	name = "PDF Word Count"
	metrics[name] = wc
	metricOrder = append(metricOrder, name)
	/*
		// Count words for each provider
		for _, s := range providers {
			wc := make([]string, 0, len(unified))
			for _, ptr := range unified {
				buf, err := ioutil.ReadFile(path.Join(txtsDir, strings.ToUpper(s), ptr+".txt"))
				if err != nil {
					return err
				}
				count := len(strings.Fields(string(buf)))
				if count < minwc {
					minwc = count
				}
				if count > maxwc {
					maxwc = count
				}
				wc = append(wc, strconv.Itoa(count))
			}
			name := fmt.Sprintf("%s Word Count", strings.ToUpper(s))
			metrics[name] = wc
			metricOrder = append(metricOrder, name)
		}
	*/
	metricLimits = append(metricLimits, fmt.Sprintf("Word Count;%d;%d", minwc, maxwc))
	//secs = int(time.Since(start) / time.Second)
	//fmt.Printf("%d secs\n", secs)

	// Create explorer
	fmt.Printf("[INFO] Creating Explorer ... \t\t\t")
	err = createExplorer(unified, metrics, metricLimits, metricOrder, txtDirs, pdfPath, baseDir)
	fmt.Printf("done\n")

	fmt.Printf("[INFO] Comparable Ptrs: %d (out of %d). %s\n", len(unified), len(ptrs), strings.Join(res, ", "))
	fmt.Printf("[DONE] Run: tigerocr serve ./%s\n", baseDir)

	return nil
}
