package zdom

import (
	"fmt"
	"strings"

	"github.com/torlangballe/zutil/zgeo"
)

func GetFontStyle(font *zgeo.Font) string {
	var parts []string
	if font.Style&zgeo.FontStyleBold != 0 {
		parts = append(parts, "bold")
	}
	if font.Style&zgeo.FontStyleItalic != 0 {
		parts = append(parts, "italic")
	}
	parts = append(parts, fmt.Sprintf("%dpx", int(font.Size)))
	parts = append(parts, font.Name)

	return strings.Join(parts, " ")
}
