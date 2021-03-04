package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/ughe/tigerocr/ocr"
)

func abs(n int) int {
	if n >= 0 {
		return n
	} else {
		return -n
	}
}

func imgToWidthHeight(img []byte) (int, int, error) {
	m, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return 0, 0, err
	}
	mb := m.Bounds()
	width, height := abs(mb.Max.X-mb.Min.X), abs(mb.Max.Y-mb.Min.Y)
	return width, height, nil
}

// Converts a raw buffer to BLW. If format is ocr.Result, converts to ocr.Detection
func convertToBLW(img []byte, raw []byte, rawName string) (*ocr.Detection, error) {
	switch filepath.Ext(rawName) {
	case ".blw":
		var detection *ocr.Detection
		if err := json.Unmarshal(raw, &detection); err != nil {
			return nil, err
		}
		return detection, nil
	case ".json":
		// Dynamically convert json to blw format (entails overhead)
		var result ocr.Result
		if err := json.Unmarshal(raw, &result); err != nil {
			return nil, err
		}
		var c ocr.Client
		switch result.Service {
		case "AWS":
			c = ocr.AWSClient{CredentialsPath: ""}
		case "Azure":
			c = ocr.AzureClient{CredentialsPath: ""}
		case "GCP":
			c = ocr.GCPClient{CredentialsPath: ""}
		default:
			return nil, fmt.Errorf("Service %v is not {AWS, Azure, GCP}", result.Service)
		}
		if img == nil {
			bogus := new(bytes.Buffer)
			err := jpeg.Encode(bogus, image.NewRGBA(image.Rect(0, 0, 1, 1)), nil)
			if err != nil {
				return nil, err
			}
			img = bogus.Bytes()
		}
		width, height, err := imgToWidthHeight(img)
		if err != nil {
			return nil, err
		}
		return c.ResultToDetection(&result, width, height)
	default:
		return nil, fmt.Errorf("Expected *.json or *.blw coordinate file instead of: %v", filepath.Ext(rawName))
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

	// merge command
	mergeSet := flag.NewFlagSet("merge", flag.ExitOnError)
	mergeSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s %s fst.blw snd.blw ...\n\n", os.Args[0], os.Args[1])
		mergeSet.PrintDefaults()
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
	diro := convertSet.Bool("pdf", false, "Convert to PDF. Same arguments as directories, not files")
	filter := convertSet.String("select", "", "Select BLW prefix for PDF. i.e. -pdf -select=azu for *.azu.blw")
	convertSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s %s img.jpg ocr.json\nusage: %s %s -pdf imgs/ blw/\n\n", os.Args[0], os.Args[1], os.Args[0], os.Args[1])
		convertSet.PrintDefaults()
	}

	// extract command
	extractSet := flag.NewFlagSet("extract", flag.ExitOnError)
	stato := extractSet.Bool("stat", false, "Combined, human-readable summary of all metadata")
	algoido := extractSet.Bool("algoid", false, "Algorithm ID is composed of the service name and version")
	speedo := extractSet.Bool("speed", false, "Speed is the duration in milliseconds to run OCR")
	dateo := extractSet.Bool("date", false, "Date the OCR was run")
	texto := extractSet.Bool("text", false, "OCR transcription in plaintext")
	extractSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s %s [-stat] [-algoid] [-speed] [-date] [-text] ocr.blw\n\n", os.Args[0], os.Args[1])
		extractSet.PrintDefaults()
	}

	// explore command
	exploreSet := flag.NewFlagSet("explore", flag.ExitOnError)
	xkeys := exploreSet.String("keys", path.Join(usr.HomeDir, ".aws"), "Path to credentials directory")
	xawso := exploreSet.Bool("aws", false, "Run AWS Textract OCR. "+aws_help)
	xazuo := exploreSet.Bool("azure", false, "Run Azure CognitiveServices OCR. "+azu_help)
	xgcpo := exploreSet.Bool("gcp", false, "Run GCP Vision OCR. "+gcp_help)
	exploreSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s %s [-keys=~/keydir/] [-aws] [-azure] [-gcp] file.pdf\n\n", os.Args[0], os.Args[1])
		exploreSet.PrintDefaults()
	}

	// serve command
	serveSet := flag.NewFlagSet("serve", flag.ExitOnError)
	serveSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s %s ./explorer\n\n", os.Args[0], os.Args[1])
		// serveSet.PrintDefaults() // No flags used
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s <command> [arguments]\n\nThe commands are:\n\n"+
			strings.Repeat("\t%v\n", 8)+"\n", os.Args[0],
			"run     \t execute ocr on selected providers",
			"annotate\t draw bounding boxes of words on the original image",
			"merge   \t combine multiple blw files for boosted transcription",
			"editdist\t calculate levenshtein distance of two text files",
			"convert \t convert json ocr responses to unified blw format (*)",
			"extract \t extract metadata from a blw or json datafile",
			"explore \t execute pdf ocr and output results as a web explorer",
			"serve   \t serve current directory at "+addr,
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
			fmt.Fprintf(os.Stderr, "Error: No service(s) selected.\n")
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
	case "merge":
		mergeSet.Parse(os.Args[2:])
		if mergeSet.NArg() < 2 {
			mergeSet.Usage()
			os.Exit(1)
		}
		err = mergeCommand(mergeSet.Args())
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
		if *diro {
			// Accepts directories instead of filenames
			err = convertCommandPDF(imgFilename, jsnFilename, *filter)
		} else {
			err = convertCommand(imgFilename, jsnFilename)
		}
	case "extract":
		extractSet.Parse(os.Args[2:])
		if extractSet.NArg() != 1 {
			extractSet.Usage()
			os.Exit(1)
		}
		if !(*stato || *algoido || *speedo || *dateo || *texto) {
			fmt.Fprintf(os.Stderr, "Error: no flags. Please specify one flag.\n\n")
			extractSet.Usage()
			os.Exit(1)
		}
		// Check for more than one flag specified
		if (*stato && *algoido) || (*stato && *speedo) || (*stato && *dateo) || (*stato && *texto) || (*algoido && *speedo) || (*algoido && *dateo) || (*algoido && *texto) || (*speedo && *dateo) || (*speedo && *texto) || (*dateo && *texto) {
			fmt.Fprintf(os.Stderr, "Error: multiple flags. Please specify one flag.\n\n")
			extractSet.Usage()
			os.Exit(1)
		}
		dataFilename := extractSet.Arg(0)
		err = extractCommand(dataFilename, *stato, *algoido, *speedo, *dateo, *texto)
	case "explore":
		exploreSet.Parse(os.Args[2:])
		if exploreSet.NArg() != 1 {
			exploreSet.Usage()
			os.Exit(1)
		}
		if !*xawso && !*xazuo && !*xgcpo {
			fmt.Fprintf(os.Stderr, "Error: No service(s) selected.\n")
			exploreSet.Usage()
			os.Exit(1)
		}
		pdfName := exploreSet.Arg(0)
		err = exploreCommand(*xkeys, *xawso, *xazuo, *xgcpo, pdfName)
	case "serve":
		serveSet.Parse(os.Args[2:])
		if serveSet.NArg() > 1 {
			serveSet.Usage()
			os.Exit(1)
		}
		dirName := "."
		if serveSet.NArg() == 1 {
			dirName = serveSet.Arg(0)
		}
		err = serve(dirName)
	default:
		flag.Usage()
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("[ERROR] %s", err)
	}
}
