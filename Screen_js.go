package zui

import "github.com/torlangballe/zutil/zgeo"

func ScreenMain() Screen {
	var m Screen

	s := WindowJS.Get("screen")
	w := s.Get("width").Float()
	h := s.Get("height").Float()

	dpr := WindowJS.Get("devicePixelRatio").Float()
	m.Rect = zgeo.RectMake(0, 0, w, h)
	m.Scale = dpr
	m.SoftScale = 1
	m.UsableRect = m.Rect

	return m
}
