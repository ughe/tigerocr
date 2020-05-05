package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"

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

	err = ioutil.WriteFile(dst, encoded, 0660)
	if err != nil {
		return "", fmt.Errorf("%v:WriteFile:%v", name, err)
	}
	return fmt.Sprintf("%s:%v", name, result.Duration), nil
}

func runOCR(filename string, services map[string]ocr.Client, errchan chan error) {
	file, err := os.Open(filename)
	if err != nil {
		errchan <- err
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		errchan <- err
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		errchan <- err
		return
	}
	namepath := path.Join(wd, filepath.Base(filename))
	if err != nil {
		errchan <- err
		return
	}

	log.Printf("Running: %v\n", filename)
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
	// Wait for each service
	for i := 0; i < len(services); i++ {
		<-ch
	}
	errchan <- nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: tigerocr image.jpg")
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to read user's directory: %v", err)
	}
	homekeys := path.Join(usr.HomeDir, ".aws")

	keys := flag.String("keys", homekeys, "Path to credentials directory")

	aws_ref := "https://docs.aws.amazon.com/textract/latest/dg/setup-awscli-sdk.html"
	aws_help := "Key files: credentials config\nMore info: " + aws_ref
	awso := flag.Bool("aws", false, "Run AWS Textract OCR. "+aws_help)
	azu_ref := "https://docs.microsoft.com/azure/cognitive-services/cognitive-services-apis-create-account"
	azu_ins := "\nNote: Create a json file with 'subscription_key' and 'endpoint' items"
	azu_help := "Key file: azure.json\nMore info: " + azu_ref + azu_ins
	azuo := flag.Bool("azure", false, "Run Azure CognitiveServices OCR. "+azu_help)
	gcp_ref := "https://cloud.google.com/vision/docs/before-you-begin"
	gcp_help := "Key file: gcp.json\nMore info: " + gcp_ref
	gcpo := flag.Bool("gcp", false, "Run GCP Vision OCR. "+gcp_help)

	flag.Parse()

	m := make(map[string]ocr.Client, 3)
	if *awso {
		m["aws"] = ocr.AWSClient{*keys}
	}
	if *azuo {
		m["azure"] = ocr.AzureClient{*keys}
	}
	if *gcpo {
		m["gcp"] = ocr.GCPClient{*keys}
	}

	ch := make(chan error, flag.NArg())
	for _, filename := range flag.Args() {
		go runOCR(filename, m, ch)
	}
	// Wait for every file to finish
	for i := 0; i < flag.NArg(); i++ {
		err = <-ch
		if err != nil {
			log.Println(err)
		}
	}
}
