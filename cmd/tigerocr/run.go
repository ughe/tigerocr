package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/ughe/tigerocr/ocr"
)

func runService(image []byte, Service ocr.Client, dst string) (string, int64, error) {
	name := filepath.Base(dst)
	result, err := Service.Run(image)
	if err != nil {
		return "", 0, fmt.Errorf("%v:Run:%v", name, err)
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		return "", 0, fmt.Errorf("%v:Marshal:%v", name, err)
	}

	err = ioutil.WriteFile(dst, encoded, 0600)
	if err != nil {
		return "", 0, fmt.Errorf("%v:WriteFile:%v", name, err)
	}
	return name, result.Duration, nil
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

	ch := make(chan bool, len(services))
	for s, Service := range services {
		namepath := path.Join(wd, filepath.Base(filename)+"."+s+".json")
		go func(img []byte, Service ocr.Client, p string) {
			// Note: NOT Thread Safe. Services take different times
			name, duration, err := runService(img, Service, p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			} else {
				fmt.Printf("%s:%v\n", name, duration)
			}
			ch <- true
		}(buf, Service, namepath)
	}
	// Wait for each service to finish
	for i := 0; i < len(services); i++ {
		<-ch
	}
	// Sucess (even if sub-services error)
	return nil
}

// Return clients for each service. Map keys will appear in output files
func initServices(keys string, aws, azu, gcp bool) map[string]ocr.Client {
	m := make(map[string]ocr.Client, 3)
	if aws {
		m["aws"] = ocr.AWSClient{CredentialsPath: keys}
	}
	if azu {
		m["azu"] = ocr.AzureClient{CredentialsPath: keys}
	}
	if gcp {
		m["gcp"] = ocr.GCPClient{CredentialsPath: keys}
	}
	return m
}

// Executes OCR for each of the services on each filename
func runCommand(keys string, aws, azu, gcp bool, filenames []string) error {
	m := initServices(keys, aws, azu, gcp)

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
