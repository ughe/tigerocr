package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"path"
	"time"

	vision "cloud.google.com/go/vision/apiv1"
	"google.golang.org/api/option"
)

type GCPClient struct {
	CredentialsPath string
}

// Method required by ocr.Client
// Returns GCP document text detection Result
// Reference: https://cloud.google.com/vision/docs/apis
func (c GCPClient) Run(file []byte) (*Result, error) {
	const (
		service = "GCP"
		version = "v1"
		keyName = "gcp.json"
	)

	credentialsFile := path.Join(c.CredentialsPath, keyName)
	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(
		ctx,
		option.WithCredentialsFile(credentialsFile),
	)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	fileReader := bytes.NewReader(file)
	image, err := vision.NewImageFromReader(fileReader)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	annotation, err := client.DetectDocumentText(ctx, image, nil)
	milli := int64(time.Since(start) / time.Millisecond)
	if err != nil {
		return nil, err
	}

	// Extract full text
	fullText := ""
	if annotation != nil {
		fullText = annotation.Text
	}

	date := fmtTime(start.UTC())

	encoded, err := json.Marshal(annotation)
	return &Result{
		Service:  service,
		Version:  version,
		FullText: fullText,
		Duration: milli,
		Date:     date,
		Raw:      encoded,
	}, err
}
