//  Font.go
//
//  Created by Tor Langballe on /22/10/15.

package zgo

type FontStyle string

var scale = 1.0

const (
	FontNormal FontStyle = ""
	FontBold             = "bold"
	FontItalic           = "italic"
)

type Font struct {
	Name  string    `json:"name"`
	Style FontStyle `json:"style"`
	Size  float64   `json:"size"`
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
	return f.Size * 1.5
}

func (f *Font) PointSize() float64 {
	return f.Size
}
