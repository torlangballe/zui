package zdom

import (
	"fmt"
	"github.com/torlangballe/zutil/zgeo"
	"html"
	"strings"
)

func CSSFontStyle(s zgeo.FontStyle) string {
	if s&zgeo.FontStyleBold != 0 {
		return "bold"
	}
	if s&zgeo.FontStyleItalic != 0 {
		return "italic"
	}
	return "normal"
}

func GetFontCSSKeyValues(font *zgeo.Font) map[string]string {
	return map[string]string{
		"font-style":  CSSFontStyle(font.Style),
		"font-family": font.Name,
		// "font-family": "-apple-system",
		"font-size": fmt.Sprintf("%dpx", int(font.Size)),
	}
}

func CSSStringFromMap(m map[string]string) string {
	var out string
	for k, v := range m {
		v = html.EscapeString(v)
		if strings.Contains(v, " ") {
			v = `"` + v + `"`
		}
		a := fmt.Sprintf("%s: %s; ", k, v)
		out += a
	}
	return out
}
