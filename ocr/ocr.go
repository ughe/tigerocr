package ocr

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"time"

	"github.com/ughe/tigerocr/bresenham"
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
	// Converts lossless Result format to unified Detection format
	ResultToDetection(result *Result, width, height int) (*Detection, error)
}

func fmtTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 MST")
}

func Annotate(src []byte, response *Detection, c color.Color, ab, al, aw bool) ([]byte, error) {
	m, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	img := image.NewRGBA(m.Bounds())
	draw.Draw(img, img.Bounds(), m, image.ZP, draw.Src)
	for _, block := range response.Blocks {
		x, y, w, h, err := decodeBounds(block.Bounds)
		if err != nil {
			return nil, err
		}
		if ab {
			bresenham.Rect(img, image.Point{x, y}, w, h, c, 1)
		}
		for _, line := range block.Lines {
			x, y, w, h, err = decodeBounds(line.Bounds)
			if err != nil {
				return nil, err
			}
			if al {
				bresenham.Rect(img, image.Point{x, y}, w, h, c, 1)
			}
			for _, word := range line.Words {
				x, y, w, h, err = decodeBounds(word.Bounds)
				if err != nil {
					return nil, err
				}
				if aw {
					bresenham.Rect(img, image.Point{x, y}, w, h, c, 1)
				}
			}
		}
	}
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, nil)
	return buf.Bytes(), nil
}
