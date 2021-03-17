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

// https://eastus.dev.cognitive.microsoft.com/docs/services/computer-vision-v3-1-ga/
type azureReadResponse struct {
	Error               azureError             `json:"error,omitempty"`
	Status              string                 `json:"status"`
	CreatedDateTime     string                 `json:"createdDateTime"`
	LastUpdatedDateTime string                 `json:"lastUpdatedDateTime"`
	AnalyzeResult       azureReadAnalyzeResult `json:"analyzeResult"`
}

type azureError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	statusNotStarted = "notStarted"
	statusRunning    = "running"
	statusFailed     = "failed"
	statusSucceeded  = "succeeded"
)

type azureReadAnalyzeResult struct {
	Version     string            `json:"version"`
	ReadResults []azureReadResult `json:"readResults"`
}

type azureReadResult struct {
	Page   int             `json:"page"`
	Lang   string          `json:"language"`
	Angle  float64         `json:"angle"`
	Width  int             `json:"width"`
	Height int             `json:"height"`
	Lines  []azureReadLine `json:"lines"`
}

type azureReadLine struct {
	Bounds [8]int          `json:"boundingBox"`
	Text   string          `json:"text"`
	Words  []azureReadWord `json:"words"`
}

type azureReadWord struct {
	Bounds [8]int  `json:"boundingBox"`
	Text   string  `json:"text"`
	Conf   float64 `json:"confidence"`
}

type AzureReadClient struct {
	CredentialsPath string
}

// Method required by ocr.Client
// Returns Azure READ API document text detection Result
func (c AzureReadClient) Run(image []byte) (*Result, error) {
	const (
		service     = "AzureRead"
		keyName     = "azure.json"
		uriVersion  = "vision/v3.1/read/analyze" // NOTE: May be different from version Azure returns in JSON
		httpTimeout = time.Second * 15
	)

	credentialsPath := path.Join(c.CredentialsPath, keyName)
	credentials, err := loadCredentials(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot read credentials: %s (%v)", service, credentialsPath, err)
	}

	base := credentials.Endpoint + uriVersion
	// params := "?readingOrder=natural" // Applies only in v3.2-preview.* (not used currently)
	url := base

	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequest("POST", url, bytes.NewReader(image))
	if err != nil {
		return nil, fmt.Errorf("%s: configuration error: %v", service, err)
	}
	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("Ocp-Apim-Subscription-Key", credentials.Key)

	start := time.Now()
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: OCR request failed - %v", service, err)
	}
	defer response.Body.Close()
	if response.StatusCode != 202 {
		return nil, fmt.Errorf("%s: received status code %v (expected 202)", service, response.StatusCode)
	}
	oploc := response.Header.Get("Operation-Location")
	if oploc == "" {
		return nil, fmt.Errorf("%s: empty Operation-Location (no results URL given)", service)
	}

	req2, err := http.NewRequest("GET", oploc, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: configuration error for result url: %s (%v)", service, oploc, err)
	}
	req2.Header.Add("Ocp-Apim-Subscription-Key", credentials.Key)

	// Query results location every second until results are ready
	var milli int64
	MAX_RESULT_TIMEOUTS := 15
	var result azureReadResponse
	for i := 0; i < MAX_RESULT_TIMEOUTS; i++ {
		time.Sleep(1 * time.Second)
		response, err = client.Do(req2)
		milli = int64(time.Since(start) / time.Millisecond)
		if err != nil {
			return nil, fmt.Errorf("%s: OCR (result) request failed - %v", service, err)
		}

		responseJson, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("%s: cannot read http response: %v", service, err)
		}
		err = json.Unmarshal(responseJson, &result)
		if err != nil {
			return nil, fmt.Errorf("%s: cannot unmarshal json response: %v", service, err)
		}
		if result.Error.Code != "" {
			return nil, fmt.Errorf("%s: request failed - code %s: %s - at url: %s", service, result.Error.Code, result.Error.Message, oploc)
		}
		if result.Status == statusFailed {
			return nil, fmt.Errorf("%s: The operation has failed at url: %s", service, oploc)
		}
		if result.Status == statusSucceeded {
			break
		}
	}
	if result.Status != statusSucceeded {
		return nil, fmt.Errorf("%s: timed-out waiting for a result. Last status was: %s for url: %s", result.Status, oploc)
	}

	// Parse the associated times from the result to get operation duration
	// Anecdotally, this is rounded to the nearest second. Therefore, we can just use our own millis (accurate within 1 second)
	/*
		layout := "2006-01-02T15:04:05Z" // i.e. for "2019-10-03T14:32:04Z"
		begin, err := time.Parse(layout, result.CreatedDateTime)
		if err != nil {
			return nil, fmt.Errorf("%s: cannot parse date: %v (%v)", service, result.CreatedDateTime, err)
		}
		end, err := time.Parse(layout, result.LastUpdatedDateTime)
		if err != nil {
			return nil, fmt.Errorf("%s: cannot parse date: %v (%v)", service, result.LastUpdatedDateTime, err)
		}
		duration := int64(end.Sub(begin) / time.Millisecond)
		if duration > milli { // Expect duration < milli
			fmt.Fprintf(os.Stderr, "[WARNING] %s: remote duration %v is longer than local duration %v (milliseconds)\n", service, milli, duration)
		}
	*/

	var lines []string
	for i := 0; i < len(result.AnalyzeResult.ReadResults); i++ {
		region := (result.AnalyzeResult.ReadResults)[i]
		for j := 0; j < len(region.Lines); j++ {
			lines = append(lines, region.Lines[j].Text)
		}
	}
	fullText := strings.Join(lines[:], "\n")

	date := fmtTime(start.UTC())

	encoded, err := json.Marshal(result)
	return &Result{
		Service:  service,
		Version:  result.AnalyzeResult.Version,
		FullText: fullText,
		Duration: milli,
		Date:     date,
		Raw:      encoded,
	}, err
}

// Convert four (X,Y) vertices into single (X,Y) with (W,H)
// Similar to GCP's polyToBox function
func boundsToBox(bb [8]int) string {
	minx, miny := bb[0], bb[1]
	maxx, maxy := minx, miny
	for i := 1; i < 4; i++ {
		x, y := bb[2*i], bb[2*i+1]
		if x < minx {
			minx = x
		}
		if y < miny {
			miny = y
		}
		if x > maxx {
			maxx = x
		}
		if y > maxy {
			maxy = y
		}
	}
	return encodeRawBounds(int(minx), int(miny), int(maxx-minx), int(maxy-miny))
}

func (_ AzureReadClient) ResultToDetection(result *Result, _, _ int) (*Detection, error) {
	var response azureReadResponse
	err := json.Unmarshal(result.Raw, &response)
	if err != nil {
		return nil, err
	}

	blocks := make([]Block, 0, len(response.AnalyzeResult.ReadResults))
	for _, r := range response.AnalyzeResult.ReadResults {
		lines := make([]Line, 0, len(r.Lines))
		for _, l := range r.Lines {
			words := make([]Word, 0, len(l.Words))
			for _, w := range l.Words {
				words = append(words, Word{boundsToBox(w.Bounds), w.Text})
			}
			lines = append(lines, Line{boundsToBox(l.Bounds), words})
		}
		blocks = append(blocks, Block{encodeRawBounds(0, 0, r.Width, r.Height), lines})
	}
	algoID := sanitizeString(result.Service + "-" + result.Version)
	millis := uint32(result.Duration)
	return &Detection{algoID, result.Date, millis, blocks}, nil
}
