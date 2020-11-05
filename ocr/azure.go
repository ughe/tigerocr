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

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func btoi(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func drawPoint(img draw.Image, x, y int) {
	black := color.RGBA{0, 0, 0, 255}
	img.Set(x+0, y+0, black)
	img.Set(x+0, y+1, black)
	img.Set(x+0, y-1, black)
	img.Set(x+1, y+0, black)
	img.Set(x+1, y+1, black)
	img.Set(x+1, y-1, black)
	img.Set(x-1, y+0, black)
	img.Set(x-1, y+1, black)
	img.Set(x-1, y-1, black)
}

func drawLine(img draw.Image, x0, y0, x1, y1 int) {
	fmt.Printf("   |> Drawing line: (%d, %d), (%d, %d)\n", x0, y0, x1, y1)
	if x1 < x0 { // Ensure x0 <= x1
		x0, y0, x1, y1 = x1, y1, x0, y0
	}
	w, h := abs(x1 - x0), abs(y1 - y0)
	if (h == 0) { // Horizontal line special case
		for x := x0; x <= x1; x++ {
			drawPoint(img, x, y0)
		}
		return
	} else if (w == 0) { // Vertical line special case
		if y1 < y0 { // Ensure y0 <= y1
			x0, y0, x1, y1 = x1, y1, x0, y0
		}
		for y := y0; y <= y1; y++ {
			drawPoint(img, x0, y)
		}
		return
	}
	// Use Bresenham's algorithm:
	// https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm
	// https://www.cl.cam.ac.uk/projects/raspberrypi/tutorials/os/screen02.html#lines
	dx := abs(x1-x0)
	sx := 1 - btoi(x1 < x0)*2
	dy := abs(y1-y0)
	sy := 1 - btoi(y1 < y0)*2
	err := 0
	if dx > dy {
		for x0 != x1 {
			drawPoint(img, x0, y0)
			err += dx
			if err*2 >= dy {
				y0 += sy
				err -= dy
			}
			x0 = x0 + sx;
		}
	} else {
		for y0 != y1 {
			drawPoint(img, x0, y0)
			err += dy
			if err*2 >= dx {
				x0 += sx
				err -= dx
			}
			y0 = y0 + sy;
		}
	}

	/*
	for x0 != x1 && y0 != y1 {
		img.Set(x0, y0, black)
		err2 := err*2
		if err2 >= dy {
			x0 += sx
			err += dy
		}
		if err2 <= dx {
			y0 += sy
			err += dx
		}
		i += 1
	}
	*/
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
		drawLine(img, x, y, x+w, y)
		drawLine(img, x, y, x, y+h)
		drawLine(img, x+w, y+h, x+w, y)
		drawLine(img, x+w, y+h, x, y+h)
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
