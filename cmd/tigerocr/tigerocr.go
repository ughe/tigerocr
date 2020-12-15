package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/ughe/tigerocr/editdist"
	"github.com/ughe/tigerocr/ocr"
)

func runService(image []byte, Service ocr.Client, dst string) (string, error) {
	name := filepath.Base(dst)
	result, err := Service.Run(image)
	if err != nil {
		return "", fmt.Errorf("%v:Run:%v", name, err)
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("%v:Marshal:%v", name, err)
	}

	err = ioutil.WriteFile(dst, encoded, 0600)
	if err != nil {
		return "", fmt.Errorf("%v:WriteFile:%v", name, err)
	}
	return fmt.Sprintf("%s:%v", name, result.Duration), nil
}
func runOCR(filename string, services map[string]ocr.Client) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	namepath := path.Join(wd, filepath.Base(filename))
	if err != nil {
		return err
	}

	ch := make(chan bool, len(services))
	for service, Service := range services {
		go func(service string, Service ocr.Client, p string, img []byte) {
			duration, err := runService(img, Service, p+"."+service+".json")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			} else {
				fmt.Println(duration)
			}
			ch <- true
		}(service, Service, namepath, buf)
	}
	// Wait for each service to finish
	for i := 0; i < len(services); i++ {
		<-ch
	}
	// Sucess (even if sub-services error)
	return nil
}

// Executes OCR for each of the services on each filename
func runCommand(keys string, aws, azu, gcp bool, filenames []string) error {
	m := make(map[string]ocr.Client, 3)
	if aws {
		m["aws"] = ocr.AWSClient{CredentialsPath: keys}
	}
	if azu {
		m["azure"] = ocr.AzureClient{CredentialsPath: keys}
	}
	if gcp {
		m["gcp"] = ocr.GCPClient{CredentialsPath: keys}
	}

	errs := make([]error, 0)
	for _, filename := range filenames {
		if err := runOCR(filename, m); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		allErr := errs[0]
		for err := range errs[1:] {
			allErr = fmt.Errorf("%v\n\n%v", allErr, err)
		}
		return allErr
	}
	return nil
}

var (
	blue   = color.RGBA{0, 0, 255, 255}
	red    = color.RGBA{255, 0, 0, 255}
	orange = color.RGBA{255, 165, 0, 255}
)

func annotate(img []byte, detection *ocr.Detection, b, l, w bool, dstFilename string) error {
	lid := strings.ToLower(detection.AlgoID)
	var col color.Color
	if strings.Contains(lid, "aws") {
		col = orange
	} else if strings.Contains(lid, "azu") {
		col = blue
	} else if strings.Contains(lid, "gcp") {
		col = red
	} else {
		col = color.Black
	}

	dstImg, err := ocr.Annotate(img, detection, col, b, l, w)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dstFilename, dstImg, 0600); err != nil {
		return err
	}
	fmt.Printf("[INFO] Annotated image: %v\n", dstFilename)
	return nil
}
func abs(n int) int {
	if n >= 0 {
		return n
	} else {
		return -n
	}
}

// Annotates the block, lines, and/or words on the image given the coordinates
func annotateCommand(b, l, w bool, imageFilename, coordFilename string) error {
	buf, err := ioutil.ReadFile(imageFilename)
	if err != nil {
		return err
	}

	var detection *ocr.Detection
	raw, err := ioutil.ReadFile(coordFilename)
	if err != nil {
		return err
	}
	switch filepath.Ext(coordFilename) {
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
		detection, err = convertToBLW(buf, &result)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Expected *.json or *.blw coordinate file instead of: %v", filepath.Ext(coordFilename))
	}

	blw := ""
	if b {
		blw += "b"
	}
	if l {
		blw += "l"
	}
	if w {
		blw += "w"
	}
	dstFilename := fmt.Sprintf("%v.%v.%v.jpg", strings.TrimSuffix(filepath.Base(imageFilename), filepath.Ext(imageFilename)), blw, strings.ToLower(detection.AlgoID))

	if len(detection.Blocks) == 0 {
		return fmt.Errorf("Failed to annotate. No blocks, lines, or words in: %s\n", coordFilename)
	}

	return annotate(buf, detection, b, l, w, dstFilename)
}

func editdistCommand(srcFilename, dstFilename string, cer bool) error {
	bufa, err := ioutil.ReadFile(srcFilename)
	if err != nil {
		return err
	}
	bufb, err := ioutil.ReadFile(dstFilename)
	if err != nil {
		return err
	}
	dist := editdist.Levenshtein(bufa, bufb)
	if cer {
		fmt.Printf("%.5f\n", editdist.CER(dist, len(bufb)))
	} else {
		fmt.Printf("%d\n", dist)
	}
	return nil
}

func convertToBLW(img []byte, result *ocr.Result) (*ocr.Detection, error) {
	var c ocr.Client
	switch result.Service {
	case "AWS":
		c = ocr.AWSClient{""}
	case "Azure":
		c = ocr.AzureClient{""}
	case "GCP":
		c = ocr.GCPClient{""}
	default:
		return nil, fmt.Errorf("Service %v is not {AWS, Azure, GCP}", result.Service)
	}
	m, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return nil, err
	}
	mb := m.Bounds()
	width, height := abs(mb.Max.X-mb.Min.X), abs(mb.Max.Y-mb.Min.Y)
	return c.ResultToDetection(result, width, height)
}

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

func main() {
	// run command
	runSet := flag.NewFlagSet("run", flag.ExitOnError)
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to read user's directory: %v", err)
	}
	keys := runSet.String("keys", path.Join(usr.HomeDir, ".aws"), "Path to credentials directory")
	aws_ref := "https://docs.aws.amazon.com/textract/latest/dg/setup-awscli-sdk.html"
	aws_help := "Key files: credentials config\nMore info: " + aws_ref
	awso := runSet.Bool("aws", false, "Run AWS Textract OCR. "+aws_help)
	azu_ref := "https://docs.microsoft.com/azure/cognitive-services/cognitive-services-apis-create-account"
	azu_ins := "\nNote: Create a json file with 'subscription_key' and 'endpoint' items"
	azu_help := "Key file: azure.json\nMore info: " + azu_ref + azu_ins
	azuo := runSet.Bool("azure", false, "Run Azure CognitiveServices OCR. "+azu_help)
	gcp_ref := "https://cloud.google.com/vision/docs/before-you-begin"
	gcp_help := "Key file: gcp.json\nMore info: " + gcp_ref
	gcpo := runSet.Bool("gcp", false, "Run GCP Vision OCR. "+gcp_help)
	runSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s %s [-keys=~/keydir/] [-aws] [-azure] [-gcp] image.jpg\n\n", os.Args[0], os.Args[1])
		runSet.PrintDefaults()
	}

	// annotate command
	annotateSet := flag.NewFlagSet("annotate", flag.ExitOnError)
	bo := annotateSet.Bool("b", false, "Annotate blocks on original image")
	lo := annotateSet.Bool("l", false, "Annotate lines on origianl image")
	wo := annotateSet.Bool("w", true, "Annotate words on original image")
	annotateSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s %s [-b] [-l] [-w=false] image.jpg ocr.blw\n\n", os.Args[0], os.Args[1])
		annotateSet.PrintDefaults()
	}

	// editdist command
	editdistSet := flag.NewFlagSet("editdist", flag.ExitOnError)
	cero := editdistSet.Bool("c", false, "Output character error rate instead of levenshtein dist")
	editdistSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s %s [-c] test.txt truth.txt\n\n", os.Args[0], os.Args[1])
		editdistSet.PrintDefaults()
	}

	// convert command
	convertSet := flag.NewFlagSet("convert", flag.ExitOnError)
	convertSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s %s img.jpg ocr.json\n\n", os.Args[0], os.Args[1])
		// convertSet.PrintDefaults() // No flags used
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s <command> [arguments]\n\nThe commands are:\n\n"+
			"\t%v\n\t%v\n\t%v\n\t%v\n"+"\n", os.Args[0],
			"run     \t execute ocr on selected providers",
			"annotate\t draw bounding boxes of words on the original image",
			"editdist\t calculate levenshtein distance of two plaintext files",
			"convert \t convert json ocr responses to unified blw files",
		)
		flag.PrintDefaults()
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runSet.Parse(os.Args[2:])
		if runSet.NArg() < 1 {
			runSet.Usage()
			os.Exit(1)
		}
		if !*awso && !*azuo && !*gcpo {
			fmt.Fprintf(os.Stderr, "Error: No service(s) selected.")
			runSet.Usage()
			os.Exit(1)
		}
		err = runCommand(*keys, *awso, *azuo, *gcpo, runSet.Args())
	case "annotate":
		annotateSet.Parse(os.Args[2:])
		if annotateSet.NArg() != 2 {
			annotateSet.Usage()
			os.Exit(1)
		}
		imageFilename := annotateSet.Arg(0)
		coordFilename := annotateSet.Arg(1)
		err = annotateCommand(*bo, *lo, *wo, imageFilename, coordFilename)
	case "editdist":
		editdistSet.Parse(os.Args[2:])
		if editdistSet.NArg() != 2 {
			editdistSet.Usage()
			os.Exit(1)
		}
		srcFilename := editdistSet.Arg(0)
		dstFilename := editdistSet.Arg(1)
		err = editdistCommand(srcFilename, dstFilename, *cero)
	case "convert":
		convertSet.Parse(os.Args[2:])
		if convertSet.NArg() != 2 {
			convertSet.Usage()
			os.Exit(1)
		}
		imgFilename := convertSet.Arg(0)
		jsnFilename := convertSet.Arg(1)
		err = convertCommand(imgFilename, jsnFilename)
	default:
		flag.Usage()
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err)
	}
}
