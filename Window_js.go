package zgo

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

type windowNative struct {
	element js.Value
}

func WindowGetCurrent() *Window {
	w := &Window{}
	w.element = WindowJS
	return w
}

func (w *Window) GetRect() zgeo.Rect {
	var s zgeo.Size
	s.W = w.element.Get("innerWidth").Float()
	s.H = w.element.Get("innerHeight").Float()
	return zgeo.Rect{Size: s}
}
