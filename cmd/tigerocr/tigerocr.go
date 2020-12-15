package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/ughe/tigerocr/ocr"
)

func abs(n int) int {
	if n >= 0 {
		return n
	} else {
		return -n
	}
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
