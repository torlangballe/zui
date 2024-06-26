//go:build zui

package zfocus

import (
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zscreen"
)

var focusColor = zgeo.ColorNew(0.5, 0.5, 1, 1)

func Draw(canvas *zcanvas.Canvas, rect zgeo.Rect, corner, width float64, opacity float32) {
	if corner == 0 {
		corner = 7
	}
	if width == 0 {
		width = 4
	}
	ss := zscreen.MainSoftScale()
	w := width * ss
	r := rect.ExpandedD(-width / 2 * ss)
	path := zgeo.PathNewRect(r, zgeo.SizeBoth(corner))
	canvas.SetColor(focusColor.WithOpacity(opacity))
	canvas.StrokePath(path, w, zgeo.PathLineRound)
}
