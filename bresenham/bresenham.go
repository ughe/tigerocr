package bresenham

import (
	"image"
	"image/color"
	"image/draw"
)

func Point(img draw.Image, p image.Point, c color.RGBA) {
	const size = 1

	for i := -size; i <= size; i++ {
		for j := -size; j <= size; j++ {
			img.Set(p.X+i, p.Y+j, black)
		}
	}
}

func Line(img draw.Image, p0, p1 image.Point) {
	x0, y0, x1, y1 := p0.X, p0.Y, p1.X, p1.Y
	if x1 < x0 { // Ensure x0 <= x1
		x0, y0, x1, y1 = x1, y1, x0, y0
	}
	w, h := abs(x1-x0), abs(y1-y0)
	if h == 0 { // Horizontal line special case
		for x := x0; x <= x1; x++ {
			Point(img, image.Point{x, y0})
		}
		return
	} else if w == 0 { // Vertical line special case
		if y1 < y0 { // Ensure y0 <= y1
			x0, y0, x1, y1 = x1, y1, x0, y0
		}
		for y := y0; y <= y1; y++ {
			Point(img, image.Point{x0, y})
		}
		return
	}
	// Use Bresenham's algorithm:
	// https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm
	// https://www.cl.cam.ac.uk/projects/raspberrypi/tutorials/os/screen02.html#lines
	dx := abs(x1 - x0)
	sx := 1 - btoi(x1 < x0)*2
	dy := abs(y1 - y0)
	sy := 1 - btoi(y1 < y0)*2
	err := 0
	if dx > dy {
		for x0 != x1 {
			Point(img, image.Point{x0, y0})
			err += dx
			if err*2 >= dy {
				y0 += sy
				err -= dy
			}
			x0 = x0 + sx
		}
	} else {
		for y0 != y1 {
			Point(img, image.Point{x0, y0})
			err += dy
			if err*2 >= dx {
				x0 += sx
				err -= dx
			}
			y0 = y0 + sy
		}
	}

	/*
		for x0 != x1 && y0 != y1 {
			img.Set(x0, y0, black)
			err2 := err*2
			if err2 >= dy {
				x0 += sx
				err += dy
			}
			if err2 <= dx {
				y0 += sy
				err += dx
			}
			i += 1
		}
	*/
}

func Rect(img draw.Image, p image.Point, w, h int) {
	Line(img, p, image.Point{p.X + w, p.Y})
	Line(img, p, image.Point{p.X, p.Y + h})
	Line(img, image.Point{p.X + w, p.Y + h}, image.Point{p.X + w, p.Y})
	Line(img, image.Point{p.X + w, p.Y + h}, image.Point{p.X, p.Y + h})
}
