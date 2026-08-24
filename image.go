package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/peter-mueller/esppaper/crowpanel579"
)

type Image struct {
	Stride [(crowpanel579.Height * crowpanel579.Width) / 8]byte
}

func (i Image) Bounds() image.Rectangle {
	return image.Rect(0, 0, crowpanel579.Width, crowpanel579.Height)
}
func (i Image) pos(x, y int) (stride int, bitMask byte, ok bool) {
	pos := (y * crowpanel579.Width) + x
	stride = pos >> 3
	bitOffset := pos % 8

	ok = x < crowpanel579.Width && y < crowpanel579.Height
	return stride, (1 << (7 - bitOffset)), ok
}
func (i *Image) At(x, y int) color.Color {
	return i.GrayAt(x, y)
}

func (i *Image) GrayAt(x, y int) color.Gray {
	stride, bitMask, ok := i.pos(x, y)
	if !ok {
		return color.Gray{0x00}
	}

	bit := i.Stride[stride] & bitMask
	if bit != 0 {
		return color.Gray{0xFF}
	}
	return color.Gray{0x00}
}

func (i *Image) Set(x, y int, c color.Color) {
	stride, bitMask, ok := i.pos(x, y)
	if !ok {
		return
	}
	r, g, b, _ := c.RGBA()

	whiteR, whiteG, whiteB, _ := color.Gray{0xFF}.RGBA()
	isWhite := r == whiteR && g == whiteG && b == whiteB
	if isWhite {
		i.Stride[stride] |= bitMask
	} else {
		i.Stride[stride] &= ^bitMask
	}
}
func (i *Image) ColorModel() color.Model {
	return color.GrayModel
}

// Scaled2xImage wraps a target draw.Image and scales all Set/SetRGBA operations 2x.
type Scaled2xImage struct {
	Dst draw.Image
}

func (s *Scaled2xImage) ColorModel() color.Model {
	return s.Dst.ColorModel()
}

func (s *Scaled2xImage) Bounds() image.Rectangle {
	b := s.Dst.Bounds()
	return image.Rect(b.Min.X/2, b.Min.Y/2, b.Max.X/2, b.Max.Y/2)
}

func (s *Scaled2xImage) At(x, y int) color.Color {
	return s.Dst.At(x*2, y*2)
}

func (s *Scaled2xImage) Set(x, y int, c color.Color) {
	startX, startY := x*2, y*2
	for dx := 0; dx < 2; dx++ {
		for dy := 0; dy < 2; dy++ {
			s.Dst.Set(startX+dx, startY+dy, c)
		}
	}
}
