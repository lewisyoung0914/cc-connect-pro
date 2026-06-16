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

// generateIcon creates a 32x32 PNG icon depicting a connection symbol:
// two filled circles connected by a thick line, on a transparent background.
// 32x32 is the recommended size for Windows system tray icons.
func generateIcon(fg color.RGBA) []byte {
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Draw a thick line from top-left dot center (7,7) to bottom-right dot center (24,24)
	drawThickLine(img, 7, 7, 24, 24, 2, fg)

	// Draw two filled circles of radius 5 (large enough to be clearly visible)
	drawFilledCircle(img, 7, 7, 5, fg)
	drawFilledCircle(img, 24, 24, 5, fg)

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// drawThickLine draws a line of the given thickness from (x0,y0) to (x1,y1).
func drawThickLine(img *image.RGBA, x0, y0, x1, y1, thickness int, c color.RGBA) {
	// Draw the center line
	drawLine(img, x0, y0, x1, y1, c)
	// Draw offset lines for thickness
	for offset := 1; offset < thickness; offset++ {
		drawLine(img, x0+offset, y0, x1+offset, y1, c)
		drawLine(img, x0, y0+offset, x1, y1+offset, c)
		drawLine(img, x0-offset, y0, x1-offset, y1, c)
		drawLine(img, x0, y0-offset, x1, y1-offset, c)
	}
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
