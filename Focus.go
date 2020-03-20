package zui

import "github.com/torlangballe/zutil/zgeo"

var focusColor = zgeo.ColorNew(0.5, 0.5, 1, 1)

func FocusDraw(canvas *Canvas, rect zgeo.Rect, corner, width float64, opacity float32) {
	if corner == 0 {
		corner = 7
	}
	if width == 0 {
		width = 4
	}
	ss := ScreenMain().SoftScale
	w := width * ss
	r := rect.ExpandedD(-width / 2 * ss)
	path := zgeo.PathNewRect(r, zgeo.Size{corner, corner})
	canvas.SetColor(focusColor, opacity)
	canvas.StrokePath(path, w, zgeo.PathLineRound)
}
