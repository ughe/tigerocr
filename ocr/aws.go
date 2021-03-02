package ocr

import (
	"encoding/json"
	"fmt"
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
		if *block.BlockType == "LINE" {
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

func geometryToBox(g *textract.Geometry, wi, hi int) string {
	b := g.BoundingBox
	w, h := float64(wi), float64(hi)
	return encodeRawBounds(int(*b.Left*w+.5), int(*b.Top*h+.5),
		int(*b.Width*w+.5), int(*b.Height*h+.5))
}

func relsToIds(rels []*textract.Relationship) ([]*string, error) {
	for _, rel := range rels {
		// Invariant: len(r.Relationships) <= 2 because Type is {CHILD, VALUE}
		switch *rel.Type {
		case "CHILD":
			return rel.Ids, nil
		case "VALUE":
			continue
		default:
			return nil, fmt.Errorf("Invalid format type: %v", *rel.Type)
		}
	}
	return nil, nil // No error on empty
}

func (_ AWSClient) ResultToDetection(result *Result, width, height int) (*Detection, error) {
	var response textract.DetectDocumentTextOutput
	err := json.Unmarshal(result.Raw, &response)
	if err != nil {
		return nil, err
	}

	mpages := make([]textract.Block, 0)
	blks := make(map[string]*textract.Block)

	for _, b := range response.Blocks {
		switch *b.BlockType {
		case "PAGE":
			mpages = append(mpages, *b)
		case "LINE":
			blks[*b.Id] = b
		case "WORD":
			blks[*b.Id] = b
		default:
			return nil, fmt.Errorf("Invalid BlockType: %v", *b.BlockType)
		}
	}

	blocks := make([]Block, 0, len(mpages))
	for _, r := range mpages {
		ids, err := relsToIds(r.Relationships)
		if err != nil {
			return nil, err
		}
		if ids == nil {
			continue
		}
		lines := make([]Line, 0, len(ids))
		for _, id := range ids {
			l, ok := blks[*id]
			if !ok {
				return nil, fmt.Errorf("Line %v not found", *id)
			}
			ids, err := relsToIds(l.Relationships)
			if err != nil {
				return nil, err
			}
			if ids == nil {
				continue
			}
			words := make([]Word, 0, len(ids))
			for _, id := range ids {
				w, ok := blks[*id]
				if !ok {
					return nil, fmt.Errorf("Word %v not found", *id)
				}
				words = append(words, Word{geometryToBox(w.Geometry, width, height), *w.Text})
			}
			lines = append(lines, Line{geometryToBox(l.Geometry, width, height), words})
		}
		blocks = append(blocks, Block{geometryToBox(r.Geometry, width, height), lines})
	}

	algoID := sanitizeString(result.Service[:3] + "-" + result.Version)
	millis := uint32(result.Duration)
	return &Detection{algoID, result.Date, millis, blocks}, nil
}
