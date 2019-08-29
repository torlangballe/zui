package zgo

type windowNative struct {
}

func WindowGetCurrent() *Window {
	w := &Window{}
	return w
}

func (w *Window) GetRect() Rect {
	var s Size
	return Rect{Size: s}
}
