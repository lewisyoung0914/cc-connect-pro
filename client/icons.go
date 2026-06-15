package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

var defaultIconBytes []byte
var accentIconBytes []byte

func init() {
	defaultIconBytes = generateIcon(color.RGBA{R: 0x86, G: 0x86, B: 0x8B, A: 0xFF})
	accentIconBytes = generateIcon(color.RGBA{R: 0x58, G: 0x56, B: 0xD6, A: 0xFF})
}

// generateIcon creates a 16x16 PNG icon depicting two dots connected by a line,
// on a transparent background, using the given foreground color.
func generateIcon(fg color.RGBA) []byte {
	const size = 16
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Transparent background is the zero value for RGBA, so no fill needed.

	// Draw a line from top-left dot center (3,3) to bottom-right dot center (12,12)
	drawLine(img, 3, 3, 12, 12, fg)

	// Draw two dots (filled circles of radius 2)
	drawFilledCircle(img, 3, 3, 2, fg)
	drawFilledCircle(img, 12, 12, 2, fg)

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// drawLine draws a 1-pixel-wide line from (x0,y0) to (x1,y1) using Bresenham's algorithm.
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := step(x0, x1)
	sy := step(y0, y1)
	err := dx - dy

	for {
		setPixel(img, x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// drawFilledCircle draws a filled circle centered at (cx,cy) with radius r.
func drawFilledCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for y := cy - r; y <= cy + r; y++ {
		for x := cx - r; x <= cx + r; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx + dy*dy <= r*r {
				setPixel(img, x, y, c)
			}
		}
	}
}

func setPixel(img *image.RGBA, x, y int, c color.RGBA) {
	if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
		img.SetRGBA(x, y, c)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func step(from, to int) int {
	if from < to {
		return 1
	}
	return -1
}
