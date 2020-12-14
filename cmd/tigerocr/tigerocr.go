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
		var c ocr.Client
		switch result.Service {
		case "AWS":
			c = ocr.AWSClient{""}
		case "Azure":
			c = ocr.AzureClient{""}
		case "GCP":
			c = ocr.GCPClient{""}
		default:
			return fmt.Errorf("Datafile %v has invalid Service tag. Service %v is not {AWS, Azure, GCP}", coordFilename, result.Service)
		}
		img, _, err := image.Decode(bytes.NewReader(buf))
		if err != nil {
			return err
		}
		mb := img.Bounds()
		width, height := abs(mb.Max.X-mb.Min.X), abs(mb.Max.Y-mb.Min.Y)
		detection, err = c.ResultToDetection(&result, width, height)
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

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s <command> [arguments]\n\nThe commands are:\n\n\t%v\n\t%v\n\n",
			os.Args[0],
			"run     \t execute ocr on selected providers",
			"annotate\t draw bounding boxes of words on the original image",
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
		if err := runCommand(*keys, *awso, *azuo, *gcpo, runSet.Args()); err != nil {
			log.Fatal(err)
		}
	case "annotate":
		annotateSet.Parse(os.Args[2:])
		if annotateSet.NArg() != 2 {
			// log.Fatalf("usage: tigerocr-annotate image.jpg result.json")
			annotateSet.Usage()
			os.Exit(1)
		}
		imageFilename := annotateSet.Arg(0)
		coordFilename := annotateSet.Arg(1)
		if err := annotateCommand(*bo, *lo, *wo, imageFilename, coordFilename); err != nil {
			log.Fatal(err)
		}
	default:
		flag.Usage()
		os.Exit(1)
	}

}
