package main

import (
	"bufio"
	"image"
	"image/color"
	"image/draw"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var drawer = font.Drawer{
	Src:  image.NewUniform(color.Black),
	Face: basicfont.Face7x13, // Built-in fixed size font
	// Dot sets the baseline location where text rendering starts
}
var img = Image{}
var mu = sync.Mutex{}
var uniform = image.NewUniform(color.Gray{0xFF})

func DrawText(text string, lineLimit int) *Image {
	mu.Lock()
	defer mu.Unlock()

	draw.Draw(&img, img.Bounds(), uniform, image.Point{}, draw.Src)

	var (
		x    = 0
		y    = 0
		line = 0
	)

	drawer.Dst = &Scaled2xImage{&img}
	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	metrics := drawer.Face.Metrics()
	lineHeight := metrics.Height

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		drawer.Dot.X = fixed.I(0)
		drawer.Dot.Y += lineHeight
		drawer.DrawString(scanner.Text())
		line += 1

		if line > lineLimit {
			break
		}
	}

	return &img
}
