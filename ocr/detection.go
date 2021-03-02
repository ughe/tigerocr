package ocr

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math"
	"math/rand"
	"sort"
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
	Bounds     string  `json:"xywh"`
	Lines      []Line  `json:"lines"`
}

type Line struct {
	Bounds     string  `json:"xywh"`
	Words      []Word  `json:"words"`
}

type Word struct {
	Bounds     string  `json:"xywh"`
	Text       string  `json:"text"`
}

type Bounds struct {
	X int
	Y int
	W int
	H int
}

// Helper objects
type bBlock struct {
	c float32
	b Bounds
	l []bLine
}
type bLine struct {
	c float32
	b Bounds
	w []bWord
}
type bWord struct {
	c float32
	b Bounds
	t string
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
		x, y, w, h, err := decodeRawBounds(block.Bounds)
		if err != nil {
			return nil, err
		}
		if ab {
			bresenham.Rect(img, image.Point{x, y}, w, h, c, 1)
		}
		for _, line := range block.Lines {
			x, y, w, h, err = decodeRawBounds(line.Bounds)
			if err != nil {
				return nil, err
			}
			if al {
				bresenham.Rect(img, image.Point{x, y}, w, h, c, 1)
			}
			for _, word := range line.Words {
				x, y, w, h, err = decodeRawBounds(word.Bounds)
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

func encodeRawBounds(x, y, w, h int) string {
	return fmt.Sprintf("%d,%d,%d,%d", x, y, w, h)
}

func encodeBounds(b Bounds) string {
	return encodeRawBounds(b.X, b.Y, b.W, b.H)
}

func DecodeBounds(bounds string) (Bounds, error) {
	s := strings.SplitN(bounds, ",", 4)
	if len(s) != 4 {
		return Bounds{}, errors.New(fmt.Sprintf(
			"Expected 4 fields. Found %d", len(s)))
	}
	x0, err := strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		return Bounds{}, err
	}
	y0, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		return Bounds{}, err
	}
	x1, err := strconv.ParseInt(s[2], 10, 64)
	if err != nil {
		return Bounds{}, err
	}
	y1, err := strconv.ParseInt(s[3], 10, 64)
	if err != nil {
		return Bounds{}, err
	}
	return Bounds{int(x0), int(y0), int(x1), int(y1)}, nil
}

func decodeRawBounds(bounds string) (int, int, int, int, error) {
	b, err := DecodeBounds(bounds)
	return b.X, b.Y, b.W, b.H, err
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func halfway(a, b int) int {
	x0, x1 := min(a, b), max(a, b)
	return x0 + (x1-x0)/2
}

// Expects b0 to be above b1
func lineSeparator(b0, b1 Bounds) (*image.Point, *image.Point, error) {
	midy := halfway(b0.Y+b0.H, b1.Y)

	l := image.Point{min(b0.X, b1.X), midy}
	r := image.Point{max(b0.X+b0.W, b1.X+b1.W), midy}
	return &l, &r, nil
}

func (d *Detection) Flatten() ([]bWord, error) {
	var ws []bWord
	for _, b := range d.Blocks {
		for _, l := range b.Lines {
			for _, w := range l.Words {
				x0, y0, w0, h0, err := decodeRawBounds(w.Bounds)
				if err != nil {
					return nil, err
				}
				bounds := Bounds{x0, y0, w0, h0}
				ws = append(ws, bWord{w.Confidence, bounds, w.Text})
			}
		}
	}
	return ws, nil
}

func mean(xs []int) float64 {
	mu := 0.0
	for _, x := range xs {
		mu += float64(x)
	}
	return mu / float64(len(xs))
}

func stddev(xs []int) float64 {
	mu := mean(xs)
	sigma := 0.0
	for _, x := range xs {
		sigma += math.Pow(float64(x)-mu, 2)
	}
	return math.Sqrt(sigma / float64(len(xs)-1))
}

// Attempts to find line boundaries and draw them
func (d *Detection) AnnotateLineBoundaries(src []byte, c color.Color) ([]byte, error) {
	bwords, err := d.Flatten()
	if err != nil {
		return nil, err
	}

	// Sort by y
	sort.Slice(bwords[:], func(i, j int) bool {
		return bwords[i].b.Y < bwords[j].b.Y
	})

	// Successive difference (separation)
	sep := make([]int, 0, len(bwords)-1)
	for i := 1; i < len(bwords); i++ {
		// Compare baseline not the top of words
		sep = append(sep, (bwords[i].b.Y+bwords[i].b.H)-(bwords[i-1].b.Y+bwords[i-1].b.H))
	}

	mu := mean(sep)
	sigma := stddev(sep)
	fmt.Printf("[INFO] mean:   %v\n", mu)
	fmt.Printf("[INFO] stddev: %v\n", sigma)

	m, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	img := image.NewRGBA(m.Bounds())
	draw.Draw(img, img.Bounds(), m, image.ZP, draw.Src)

	// Find number of separations greater than one stddev
	n := 0
	for i, s := range sep {
		if float64(s) > mu+sigma {
			n++
			p0, p1, err := lineSeparator(bwords[i].b, bwords[i+1].b)
			if err != nil {
				return nil, err
			}
			r := uint8(rand.Intn(256))
			g := uint8(rand.Intn(256))
			b := uint8(rand.Intn(256))
			col := color.RGBA{r, g, b, 255}
			bresenham.Rect(img, image.Point{bwords[i].b.X, bwords[i].b.Y}, bwords[i].b.W, bwords[i].b.H, col, 1)
			bresenham.Rect(img, image.Point{bwords[i+1].b.X, bwords[i+1].b.Y}, bwords[i+1].b.W, bwords[i+1].b.H, col, 1)
			bresenham.Line(img, *p0, *p1, col, 1)

			fmt.Printf("[INFO] %d `%s` `%s` | %d - %d = %d\n", s, bwords[i].t, bwords[i+1].t, bwords[i+1].b.Y+bwords[i+1].b.H, bwords[i].b.Y+bwords[i].b.H, s)
		}
	}
	fmt.Printf("[INFO] Total lines detected: %d\n", n)

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, nil)
	return buf.Bytes(), nil
}
