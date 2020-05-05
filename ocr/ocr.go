package ocr

import (
	"time"
)

type Result struct {
	Service  string `json:"service"`
	Version  string `json:"version"`
	FullText string `json:"text"`
	Duration int64  `json:"milliseconds"`
	Date     string `json:"date"`
	Raw      []byte `json:"raw"`
}

type Client interface {
	Run(image []byte) (*Result, error)
}

func fmtTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 MST")
}
