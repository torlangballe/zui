package zgo

import "math"

//  Created by Tor Langballe on 07-June-2019
//

type HSBA struct {
	H float32 `json:"H"`
	S float32 `json:"S"`
	B float32 `json:"B"`
	A float32 `json:"A"`
}

type RGBA struct {
	R float32 `json:"R"`
	G float32 `json:"G"`
	B float32 `json:"B"`
	A float32 `json:"A"`
}

type Color struct {
	Valid bool
	Rgba  RGBA
	// Tile
}

func NewGray(white, a float32) (c Color) {
	c.Valid = true
	c.Rgba.R = white
	c.Rgba.G = white
	c.Rgba.B = white
	c.Rgba.A = a
	return
}

func NewColor(r, g, b, a float32) (c Color) {
	c.Valid = true
	c.Rgba.R = r
	c.Rgba.G = g
	c.Rgba.B = b
	c.Rgba.A = a
	return
}

func NewHSBAColor(h, s, b, a float32) (c Color) {
	c.Valid = true

	var i = MathFloor(float64(h * 6))
	var f = h*6 - float32(i)
	var p = b * (1 - s)
	var q = b * (1 - f*s)
	var t = b * (1 - (1-f)*s)

	switch math.Mod(i, 6) {
	case 0:
		c.Rgba.R = b
		c.Rgba.G = t
		c.Rgba.B = p

	case 1:
		c.Rgba.R = q
		c.Rgba.G = b
		c.Rgba.B = p

	case 2:
		c.Rgba.R = p
		c.Rgba.G = b
		c.Rgba.B = t

	case 3:
		c.Rgba.R = p
		c.Rgba.G = q
		c.Rgba.B = b

	case 4:
		c.Rgba.R = t
		c.Rgba.G = p
		c.Rgba.B = b

	case 5:
		c.Rgba.R = b
		c.Rgba.G = p
		c.Rgba.B = q
	}
	c.Rgba.A = a
	return
}

func (c Color) GetHSBA() HSBA {
	var h, s, b, a float32
	var hsba HSBA
	hsba.H = float32(h)
	hsba.S = float32(s)
	hsba.B = float32(b)
	hsba.A = float32(a)
	return hsba
}

func (c Color) GetRGBA() RGBA {
	return c.Rgba
}

func (c Color) GetGrayScaleAndAlpha() (float32, float32) { // white, alpha
	return c.GetGrayScale(), c.Rgba.A
}

func (c Color) GetGrayScale() float32 {
	return 0.2126*c.Rgba.R + 0.7152*c.Rgba.G + 0722*c.Rgba.B
}
func (c Color) Opacity() float32 {
	return c.Rgba.A
}

func (c Color) OpacityChanged(opacity float32) Color {
	return NewColor(c.Rgba.R, c.Rgba.G, c.Rgba.B, opacity)
}

func (c Color) Mix(withColor Color, amount float32) Color {
	wc := withColor.GetRGBA()
	col := c.GetRGBA()
	r := (1-amount)*col.R + wc.R*amount
	g := (1-amount)*col.G + wc.G*amount
	b := (1-amount)*col.B + wc.B*amount
	a := (1-amount)*col.A + wc.A*amount
	return NewColor(r, g, b, a)
}

func (c Color) MultipliedBrightness(multiply float32) Color {
	hsba := c.GetHSBA()
	return NewHSBAColor(hsba.H, hsba.S, hsba.B*multiply, hsba.A)
}

func (c Color) AlteredContrast(contrast float32) Color {
	multi := float32(math.Pow((float64(1+contrast))/1, 2.0))
	var col = c.GetRGBA()
	col.R = (col.R-0.5)*multi + 0.5
	col.G = (col.G-0.5)*multi + 0.5
	col.B = (col.B-0.5)*multi + 0.5
	return NewColor(col.R, col.G, col.B, col.A)
}

func (c Color) GetContrastingGray() Color {
	g := c.GetGrayScale()
	if g < 0.5 {
		return ColorWhite
	}
	return ColorBlack
}

var ColorWhite = NewGray(1, 1)
var ColorBlack = NewGray(0, 1)
var ColorGray = NewGray(0.5, 1)
var ColorClear = NewGray(0, 0)
var ColorBlue = NewColor(0, 0, 1, 1)
var ColorRed = NewColor(1, 0, 0, 1)
var ColorYellow = NewColor(1, 1, 0, 1)
var ColorGreen = NewColor(0, 1, 0, 1)
var ColorOrange = NewColor(1, 0.5, 0, 1)
var ColorCyan = NewColor(0, 1, 1, 1)
var ColorMagenta = NewColor(1, 0, 1, 1)
