package ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"
)

type azureClientCredentials struct {
	Key      string `json:"subscription_key"`
	Endpoint string `json:"endpoint"`
}

type azureVisionResponse struct {
	StatusCode  string        `json:"code,omitempty"`
	StatusMsg   string        `json:"message,omitempty"`
	Language    string        `json:"language"`
	Oreintation string        `json:"orientation"`
	Regions     []azureRegion `json:"regions"`
}

type azureRegion struct {
	Bounds string      `json:"boundingBox"`
	Lines  []azureLine `json:"lines"`
}

type azureLine struct {
	Bounds string      `json:"boundingBox"`
	Words  []azureWord `json:"words"`
}

type azureWord struct {
	Bounds string `json:"boundingBox"`
	Text   string `json:"text"`
}

type AzureClient struct {
	CredentialsPath string
}

// Method required by ocr.Client
// Returns Azure document text detection Result
// Example: https://docs.microsoft.com/en-us/azure/cognitive-services/computer-vision/quickstarts/go-print-text
func (c AzureClient) Run(image []byte) (*Result, error) {
	const (
		service     = "Azure"
		keyName     = "azure.json"
		uriVersion  = "vision/v2.1/ocr"
		httpTimeout = time.Second * 15
	)

	credentialsFile := path.Join(c.CredentialsPath, keyName)
	f, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		return nil, err
	}
	credentials := &azureClientCredentials{}
	err = json.Unmarshal(f, credentials)
	if err != nil {
		return nil, err
	}
	if credentials.Endpoint == "" || credentials.Key == "" {
		err = fmt.Errorf("No 'subscription_key' or 'endpoint' in " +
			credentialsFile)
		return nil, err
	}

	base := credentials.Endpoint + uriVersion
	params := "?language=unk&detectOrientation=false"
	url := base + params

	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequest("POST", url, bytes.NewReader(image))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("Ocp-Apim-Subscription-Key", credentials.Key)

	start := time.Now()
	response, err := client.Do(req)
	milli := int64(time.Since(start) / time.Millisecond)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseJson, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	result := azureVisionResponse{}
	err = json.Unmarshal(responseJson, &result)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != "" {
		err = fmt.Errorf("%v: %v", result.StatusCode, result.StatusMsg)
		return nil, err
	}

	var lines []string
	for i := 0; i < len(result.Regions); i++ {
		region := (result.Regions)[i]
		for j := 0; j < len(region.Lines); j++ {
			var words []string
			line := (region.Lines)[j]
			for k := 0; k < len(line.Words); k++ {
				word := (line.Words)[k]
				words = append(words, word.Text)
			}
			lines = append(lines, strings.Join(words[:], " "))
		}
	}
	fullText := strings.Join(lines[:], "\n")

	date := fmtTime(start.UTC())

	encoded, err := json.Marshal(result)
	return &Result{
		Service:  service,
		Version:  uriVersion,
		FullText: fullText,
		Duration: milli,
		Date:     date,
		Raw:      encoded,
	}, err
}
