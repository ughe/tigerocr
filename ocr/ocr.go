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
	RawToDetection(raw []byte) (*Detection, error)
}

func fmtTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 MST")
}

func Annotate(src []byte, response *Detection) ([]byte, error) {
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
