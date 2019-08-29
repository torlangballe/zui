package zgo

var focusColor = ColorNew(0.5, 0.5, 1, 1)

func FocusDraw(canvas *Canvas, rect Rect, corner, width float64, opacity float32) {
	if corner == 0 {
		corner = 7
	}
	if width == 0 {
		width = 4
	}
	ss := ScreenMain().SoftScale
	w := width * ss
	r := rect.ExpandedD(-width / 2 * ss)
	path := PathNewFromRect(r, Size{corner, corner})
	canvas.SetColor(focusColor, opacity)
	canvas.StrokePath(path, w, PathLineRound)
}
