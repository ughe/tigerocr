package bresenham

import (
	"archive/tar"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"testing"
)

const TEST_TAR = "bresenham_test.tar"

var (
	Red    = color.RGBA{255, 0, 0, 255}
	Orange = color.RGBA{255, 165, 0, 255}
	Yellow = color.RGBA{255, 255, 0, 255}
	Lime   = color.RGBA{0, 255, 0, 255}
	Blue   = color.RGBA{0, 0, 255, 255}
	Purple = color.RGBA{128, 0, 128, 255}
	colors = [7]color.Color{color.Black, Red, Orange, Yellow, Lime, Blue, Purple}
	dim    = 300
)

func blankImage(w, h int) draw.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			img.Set(i, j, color.White)
		}
	}
	return img
}

func readTAR(tarPath, header string) ([]byte, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == header {
			buf, err := ioutil.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			return buf, err
		}
	}
	return nil, fmt.Errorf("Header '%v' not found in tarfile: %v", header, tarPath)
}

// Append header to tarfile if header is not in tarfile already
func appendTAR(tarfile string, header string, buf []byte) error {
	_, err := os.Stat(tarfile)
	isNew := err != nil
	f, err := os.OpenFile(tarfile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if !isNew { // TODO experiment seeking past start
		_, err = f.Seek(-512*2, os.SEEK_END)
		if err != nil {
			return err
		}
	}
	tw := tar.NewWriter(f)
	defer tw.Close()
	hdr := &tar.Header{Name: header, Mode: 0600, Size: int64(len(buf))}
	if err := tw.WriteHeader(hdr); err != nil {
		return err // Corrupts the tarfile
	}
	if _, err := tw.Write(buf); err != nil {
		return err // Corrupts the tarfile
	}
	return nil
}

func isHeaderInTAR(tarfile string, header string) (bool, error) {
	f, err := os.Open(tarfile)
	if err != nil {
		return false, nil // Return does not exist (no err)
	}
	defer f.Close()
	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if hdr.Name == header {
			return true, nil
		}
	}
}

func checkPNG(t *testing.T, name string, img draw.Image) {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		log.Fatal(err)
	}
	exists, err := isHeaderInTAR(TEST_TAR, name)
	if err != nil {
		log.Fatal(err)
	}
	if exists {
		// Run unit test
		buf_answer, err := readTAR(TEST_TAR, name)
		if err != nil {
			log.Fatal(err)
		}
		if !bytes.Equal(buf.Bytes(), buf_answer) {
			debugName := "debug_" + name
			err = ioutil.WriteFile(debugName, buf.Bytes(), 0600)
			if err != nil {
				log.Fatal(err)
			}
			t.Fatalf("%v (in %v) does not match %v", name, TEST_TAR, debugName)
		}
	} else {
		// Pin current image result for future test
		err = appendTAR(TEST_TAR, name, buf.Bytes())
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("[INFO] Pinning test to: %v\n", name)
	}
}

func TestPoint(t *testing.T) {
	name := "TestPoint.png"
	img := blankImage(dim, dim)
	rows := 10
	for r := 0; r < rows; r++ {
		y := (dim / (rows + 1)) * (r + 1)
		for c := 0; c < len(colors); c++ {
			x := (dim / (len(colors) + 1)) * (c + 1)
			Point(img, image.Point{x, y}, colors[c], c)
		}
	}
	checkPNG(t, name, img)
}

func TestLineAngles(t *testing.T) {
	name := "TestLineAngles.png"
	img := blankImage(dim, dim)

	// Frame
	p := 30
	Rect(img, image.Point{p, p}, dim-2*p, dim-2*p, colors[0], 0)
	origin := image.Point{2 * p, dim - 2*p} // Bottom left
	w := dim - p*4

	h30 := int(math.Tan(math.Pi/6) * float64(w))   // 30 degrees
	p30 := image.Point{origin.X + w, origin.Y - h30}
	h45 := int(math.Tan(math.Pi/4) * float64(w))   // 45 degrees
	p45 := image.Point{origin.X + w, origin.Y - h45}
	h60 := int(math.Tan(math.Pi/3) * float64(w/2)) // 60 degrees
	p60 := image.Point{origin.X + w/2, origin.Y - h60}
	Line(img, origin, p30, colors[0], 1)
	Line(img, origin, p45, colors[0], 1)
	Line(img, origin, p60, colors[0], 1)
	Point(img, p30, colors[1], 3)
	Point(img, p45, colors[2], 3)
	Point(img, p60, colors[4], 3)
	Point(img, origin, colors[5], 3)

	checkPNG(t, name, img)
}

func TestLineTriangle(t *testing.T) {
	name := "TestLineTriangle.png"
	img := blankImage(dim, dim)
	p := 30
	p0 := image.Point{dim/2, p}
	p1 := image.Point{p, dim-p}
	p2 := image.Point{dim-p, dim-p}
	Line(img, p0, p1, colors[0], 1)
	Line(img, p1, p2, colors[0], 1)
	Line(img, p2, p0, colors[0], 1)
	Point(img, p1, colors[2], 3)
	Point(img, p2, colors[4], 3)
	Point(img, p0, colors[1], 3)
	checkPNG(t, name, img)
}

func TestRect(t *testing.T) {
	name := "TestRect.png"
	img := blankImage(dim, dim)
	for i := 0; i < len(colors); i++ {
		xy := (dim / 2 / (len(colors) + 1)) * (i + 1)
		wh := dim - 2*xy
		Rect(img, image.Point{xy, xy}, wh, wh, colors[i], i)
	}
	checkPNG(t, name, img)
}

func TestRectBricks(t *testing.T) {
	name := "TestRectBricks.png"
	img := blankImage(dim, dim)
	p := 15
	cols := 7
	h := (dim-p) / (len(colors)*2)
	w := (dim-p) / cols
	for r := 0; r < len(colors)*2; r++ {
		for c := 0; c < cols; c++ {
			Rect(img, image.Point{p+c*w, p+r*h}, w - p, h - p, colors[r / 2], 2)
		}
	}
	checkPNG(t, name, img)
}

func TestPoly(t *testing.T) {
	name := "TestPoly.png"
	img := blankImage(dim, dim)
	for i := 0; i < len(colors); i++ { // Draws a diamond (rotated square)
		offset := (dim / 2 / (len(colors) + 1)) * (i + 1)
		p0 := image.Point{offset, dim / 2}
		p1 := image.Point{dim / 2, offset}
		p2 := image.Point{dim - offset, dim / 2}
		p3 := image.Point{dim / 2, dim - offset}
		Poly(img, p0, p1, p2, p3, colors[i], i)
	}
	checkPNG(t, name, img)
}

/*
func check(t *testing.T, as string, bs string, exp int) {
	a := []byte(as);
	b := []byte(bs);
	dists := levenshtein(a, b)
	dist := dists[len(a)][len(b)]
	if exp != dist {
		t.Fatalf("Expected: %v. Received: %v. Lev '%v' '%v'\n%v",
			exp, dist, as, bs, printTable(as, bs, dists))
	}
	dist = Levenshtein(b, a)
	if exp != dist {
		t.Fatalf("Expected: %v. Received: %v. Lev '%v' '%v'\n%v",
			exp, dist, bs, as, printTable(bs, as, dists))
	}
}
*/
