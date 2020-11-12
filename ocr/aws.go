package ocr

import (
	"encoding/json"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
)

type AWSClient struct {
	CredentialsPath string
}

// Method required by ocr.Client
// Returns AWS document text detection Result
// Reference: https://docs.aws.amazon.com/textract/
func (c *AWSClient) Run(image []byte) (*Result, error) {
	const (
		service    = "AWS"
		keyName    = "credentials"
		configName = "config"
		maxSizeMB  = 5
	)

	credentialsFile := path.Join(c.CredentialsPath, keyName)
	configFile := path.Join(c.CredentialsPath, configName)

	three := new(int)
	*three = 3
	config := aws.Config{
		MaxRetries: three,
	}
	s, err := session.NewSessionWithOptions(
		session.Options{
			SharedConfigFiles: []string{credentialsFile, configFile},
			SharedConfigState: session.SharedConfigEnable,
		},
	)
	if err != nil {
		return nil, err
	}
	client := textract.New(s, &config)

	doc := textract.Document{
		Bytes: image,
	}
	ddti := textract.DetectDocumentTextInput{
		Document: &doc,
	}

	start := time.Now()
	result, err := client.DetectDocumentText(&ddti)
	milli := int64(time.Since(start) / time.Millisecond)
	if err != nil {
		return nil, err
	}

	version := *result.DetectDocumentTextModelVersion
	blocks := result.Blocks
	var lines []string
	for _, block := range blocks {
		if *block.BlockType == "LINE" && *block.Confidence >= 0.0 {
			lines = append(lines, *block.Text)
		}
	}
	fullText := strings.Join(lines[:], "\n")

	date := fmtTime(start.UTC())

	encoded, err := json.Marshal(result)
	return &Result{
		Service:  service,
		Version:  version,
		FullText: fullText,
		Duration: milli,
		Date:     date,
		Raw:      encoded,
	}, err
}

func geometryToBox(g *textract.Geometry, wi, hi int) (string, error) {
	b := g.BoundingBox
	w, h := float64(wi), float64(hi)
	return encodeBounds(int(*b.Left*w+.5), int(*b.Top*h+.5),
		int(*b.Width*w+.5), int(*b.Height*h+.5)), nil
}

func (c *AWSClient) RawToDetection(raw []byte, w, h int) (*Detection, error) {
	var response textract.DetectDocumentTextOutput
	err := json.Unmarshal(raw, &response)
	if err != nil {
		return nil, err
	}

	/*
		regions := make([]Region, 0, len(response.Blocks))
		for _, b := range response.Blocks {
			if *b.BlockType == "LINE" {
				lines := make([]Line, 0, len(r.Lines))
				for _, l := range r.Lines {
					words := make([]Word, 0, len(l.Words))
					for _, w := range words {
						words = append(words, Word{w.Confidence, w.Bounds, w.Text})
					}
					lines = append(lines, Line{l.Confidence, l.Bounds, words})
				}
				regions = append(regions, Region{r.Confidence, r.Bounds, lines})
			}
		}
		return &Detection{regions}, nil
	*/
	return nil, err // TODO AWS is more complicated (since we need to match relationship ids of words to lines)
}
