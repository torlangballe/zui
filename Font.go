//  Font.go
//
//  Created by Tor Langballe on /22/10/15.

package zui

import "strings"

type FontStyle int

var scale = 1.0

const (
	FontStyleNormal FontStyle = 0
	FontStyleBold   FontStyle = 1
	FontStyleItalic FontStyle = 2
)

type Font struct {
	Name  string    `json:"name"`
	Style FontStyle `json:"style"`
	Size  float64   `json:"size"`
}

// DefaultSize This is used
var FontDefaultSize = 16.0

func (s FontStyle) String() string {
	var parts []string
	switch s {
	case FontStyleNormal:
		parts = append(parts, "normal")
	case FontStyleBold:
		parts = append(parts, "bold")
	case FontStyleItalic:
		parts = append(parts, "italic")
	}
	return strings.Join(parts, " ")
}

func FontStyleFromStr(str string) FontStyle {
	s := FontStyleNormal
	for _, p := range strings.Split(str, " ") {
		switch p {
		case "bold":
			s |= FontStyleBold
		case "italic":
			s |= FontStyleItalic
		}
	}
	return s
}

func FontNew(name string, size float64, style FontStyle) *Font {
	return &Font{Name: name, Size: size * scale, Style: style}
}

func FontNice(size float64, style FontStyle) *Font {
	return FontNew("Helvetica", size, style)
}

func (f *Font) NewWithSize(size float64) *Font {
	return FontNew(f.Name, size, f.Style)
}

func (f *Font) LineHeight() float64 {
	return f.Size
}

func (f *Font) PointSize() float64 {
	return f.Size
}
