package zgo

import "syscall/js"

type windowNative struct {
	element js.Value
}

func WindowGetCurrent() *Window {
	w := &Window{}
	w.element = WindowJS
	return w
}

func (w *Window) GetRect() Rect {
	var s Size
	s.W = w.element.Get("innerWidth").Float()
	s.H = w.element.Get("innerHeight").Float()
	return Rect{Size: s}
}
