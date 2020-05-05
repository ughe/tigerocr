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
func (c AWSClient) Run(image []byte) (*Result, error) {
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
