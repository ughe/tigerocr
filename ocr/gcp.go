package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"

	vision "cloud.google.com/go/vision/apiv1"
	"google.golang.org/api/option"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
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

func polyToBox(poly *pb.BoundingPoly) (string, error) {
	if len(poly.Vertices) != 4 {
		return "", fmt.Errorf("Found %d != 4 vertices", len(poly.Vertices))
	}
	minx, miny := poly.Vertices[0].X, poly.Vertices[0].Y
	maxx, maxy := minx, miny
	for _, v := range poly.Vertices {
		if v.X < minx {
			minx = v.X
		}
		if v.Y < miny {
			miny = v.Y
		}
		if v.X > maxx {
			maxx = v.X
		}
		if v.Y > maxy {
			maxy = v.Y
		}
	}
	return encodeBounds(int(minx), int(miny), int(maxx-minx), int(maxy-miny)), nil
}

func (_ GCPClient) RawToDetection(raw []byte) (*Detection, error) {
	var response pb.TextAnnotation
	err := json.Unmarshal(raw, &response)
	if err != nil {
		return nil, err
	}

	regions := make([]Region, 0, 3)
	for _, p := range response.Pages {
		for _, r := range p.Blocks {
			lines := make([]Line, 0, len(r.Paragraphs))
			for _, l := range r.Paragraphs {
				words := make([]Word, 0, len(l.Words))
				for _, w := range l.Words {
					bounds, err := polyToBox(w.BoundingBox)
					if err != nil {
						return nil, err
					}
					words = append(words, Word{w.Confidence, bounds, w.String()})
				}
				bounds, err := polyToBox(l.BoundingBox)
				if err != nil {
					return nil, err
				}
				lines = append(lines, Line{l.Confidence, bounds, words})
			}
			bounds, err := polyToBox(r.BoundingBox)
			if err != nil {
				return nil, err
			}
			regions = append(regions, Region{r.Confidence, bounds, lines})
		}
	}
	return &Detection{regions}, nil
}
