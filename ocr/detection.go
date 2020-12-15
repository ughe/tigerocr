package ocr

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"strconv"
	"strings"

	"github.com/ughe/tigerocr/bresenham"
)

type Detection struct {
	AlgoID string  `json:"algo"`
	Date   string  `json:"date"`
	Millis uint32  `json:"millis"`
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Confidence float32 `json:"conf"`
	Bounds     string  `json:"xywh"`
	Lines      []Line  `json:"lines"`
}

type Line struct {
	Confidence float32 `json:"conf"`
	Bounds     string  `json:"xywh"`
	Words      []Word  `json:"words"`
}

type Word struct {
	Confidence float32 `json:"conf"`
	Bounds     string  `json:"xywh"`
	Text       string  `json:"text"`
}

type Bounds struct {
	X int
	Y int
	W int
	H int
}

func (d *Detection) Plaintext() string {
	var blocks []string
	for _, b := range d.Blocks {
		var lines []string
		for _, l := range b.Lines {
			var words []string
			for _, w := range l.Words {
				words = append(words, w.Text)
			}
			lines = append(lines, strings.Join(words[:], " "))
		}
		blocks = append(blocks, strings.Join(lines[:], "\n"))
	}
	fullText := strings.Join(blocks[:], "\n")
	return fullText
}

func (d *Detection) CountBLW() (int, int, int) {
	nb, nl, nw := 0, 0, 0
	for _, b := range d.Blocks {
		nb++
		for _, l := range b.Lines {
			nl++
			nw += len(l.Words)
		}
	}
	return nb, nl, nw
}

func (d *Detection) Annotate(src []byte, c color.Color, ab, al, aw bool) ([]byte, error) {
	m, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	img := image.NewRGBA(m.Bounds())
	draw.Draw(img, img.Bounds(), m, image.ZP, draw.Src)
	for _, block := range d.Blocks {
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

func isAlphaNumeric(r byte) bool {
	return (r >= '0' && r <= '9') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z')
}

func sanitizeString(str string) string {
	san := make([]byte, 0, len(str))
	for i, _ := range str {
		if isAlphaNumeric(str[i]) || str[i] == '-' {
			san = append(san, str[i])
		} else {
			san = append(san, '_')
		}
	}
	return strings.ToLower(string(san))
}

func encodeBounds(x, y, w, h int) string {
	return fmt.Sprintf("%d,%d,%d,%d", x, y, w, h)
}

func decodeBounds(bounds string) (int, int, int, int, error) {
	s := strings.SplitN(bounds, ",", 4)
	if len(s) != 4 {
		return 0, 0, 0, 0, errors.New(fmt.Sprintf(
			"Expected 4 fields. Found %d", len(s)))
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

// Returns true if two intervals intersect
func intersects(left0, right0, left1, right1 int) bool {
	// Invariant: left0 <= right0 && left1 <= right1
	if left0 <= left1 {
		return right0 >= left1
	} else {
		return left0 <= right1
	}
}

// Returns true if the rectangles overlap. Rectangles have (x,y) and (width, height)
func overlaps(b0, b1 Bounds) bool {
	return (intersects(b0.X, b0.X+b0.W, b1.X, b1.X+b1.W) &&
		intersects(b0.Y, b0.Y+b0.H, b1.Y, b1.Y+b1.H))
}

func intersectionLen(left0, right0, left1, right1 int) int {
	if left0 <= left1 {
		if right0 <= left1 {
			return 0
		} else if right0 <= right1 {
			return right0 - left1
		} else {
			return right1 - left1
		}
	} else { // left0 > left1
		if left0 >= right1 {
			return 0
		} else if right0 <= right1 {
			return right0 - left0
		} else {
			return right1 - left0
		}
	}
}

// Returns the area of the overlap
func intersectionArea(b0, b1 Bounds) int {
	return (intersectionLen(b0.X, b0.X+b0.W, b1.X, b1.X+b1.W) *
		intersectionLen(b0.Y, b0.Y+b0.H, b1.Y, b1.Y+b1.H))
}
