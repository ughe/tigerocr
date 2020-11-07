package ocr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ughe/tigerocr/bresenham"
)

type azureClientCredentials struct {
	Key      string `json:"subscription_key"`
	Endpoint string `json:"endpoint"`
}

type AzureVisionResponse struct {
	StatusCode  string        `json:"code,omitempty"`
	StatusMsg   string        `json:"message,omitempty"`
	Language    string        `json:"language"`
	Orientation string        `json:"orientation"`
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

func parseBounds(bounds string) (int, int, int, int, error) {
	s := strings.SplitN(bounds, ",", 4)
	if len(s) != 4 {
		return 0, 0, 0, 0, errors.New(fmt.Sprintf("Expected 4 coordinates. Found %d", len(s)))
	}
	x0, err := strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	y0, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	x1, err := strconv.ParseInt(s[2], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	y1, err := strconv.ParseInt(s[3], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return int(x0), int(y0), int(x1), int(y1), nil
}

func Annotate(src []byte, response AzureVisionResponse) ([]byte, error) {
	m, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	img := image.NewRGBA(m.Bounds())
	draw.Draw(img, img.Bounds(), m, image.ZP, draw.Src)
	for _, region := range response.Regions {
		x, y, w, h, err := parseBounds(region.Bounds)
		if err != nil {
			return nil, err
		}
		fmt.Printf("We found a region: (%d, %d) (%d, %d)\n", x, y, w, h)
		blue := color.RGBA{0, 0, 255, 255}
		bresenham.Rect(img, image.Point{x, y}, w, h, blue)
	}
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, nil)
	return buf.Bytes(), nil
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
	result := AzureVisionResponse{}
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
