package bresenham

import (
	"image"
	"image/color"
	"image/draw"
)

// Bool to Int
func btoi(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

// Absolute value
func abs(n int) int {
	if n >= 0 {
		return n
	} else {
		return -n
	}
}

func Point(img draw.Image, p image.Point, c color.Color, size int) {
	for i := -size; i <= size; i++ {
		for j := -size; j <= size; j++ {
			img.Set(p.X+i, p.Y+j, c)
		}
	}
}

func Line(img draw.Image, p0, p1 image.Point, c color.Color, weight int) {
	x0, y0, x1, y1 := p0.X, p0.Y, p1.X, p1.Y
	if x1 < x0 { // Ensure x0 <= x1
		x0, y0, x1, y1 = x1, y1, x0, y0
	}
	w, h := abs(x1-x0), abs(y1-y0)
	if h == 0 { // Horizontal line special case
		for x := x0; x <= x1; x++ {
			Point(img, image.Point{x, y0}, c, weight)
		}
		return
	} else if w == 0 { // Vertical line special case
		if y1 < y0 { // Ensure y0 <= y1
			x0, y0, x1, y1 = x1, y1, x0, y0
		}
		for y := y0; y <= y1; y++ {
			Point(img, image.Point{x0, y}, c, weight)
		}
		return
	}
	// Bresenham's algorithm directly from Wikipedia:
	// https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm
	dx, dy := w, -h
	sx, sy := 1, 1
	if x1 < x0 {
		sx = -1
	}
	if y1 < y0 {
		sy = -1
	}
	x, y := x0, y0
	acc := dx + dy
	for {
		Point(img, image.Point{x, y}, c, weight)
		if x == x1 && y == y1 {
			break
		}
		acc2 := 2*acc
		if acc2 >= dy {
			acc += dy
			x += sx
		}
		if acc2 <= dx {
			acc += dx
			y += sy
		}
	}
}

func Rect(img draw.Image, p image.Point, w, h int, c color.Color, weight int) {
	Line(img, p, image.Point{p.X + w, p.Y}, c, weight)
	Line(img, p, image.Point{p.X, p.Y + h}, c, weight)
	Line(img, image.Point{p.X + w, p.Y + h}, image.Point{p.X + w, p.Y}, c, weight)
	Line(img, image.Point{p.X + w, p.Y + h}, image.Point{p.X, p.Y + h}, c, weight)
}

func Poly(img draw.Image, p0, p1, p2, p3 image.Point, c color.Color, weight int) {
	Line(img, p0, p1, c, weight)
	Line(img, p1, p2, c, weight)
	Line(img, p2, p3, c, weight)
	Line(img, p3, p0, c, weight)
}
