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

func FontNice(size float64, style FontStyle) *Font {
	return &Font{Name: "Helvetica", Size: size * scale, Style: style}
}

func (f *Font) NewWithSize(size float64) *Font {
	return &Font{Name: f.Name, Size: size, Style: f.Style}
}
