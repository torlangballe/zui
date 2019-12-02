// +build !js

// https://github.com/golang/mobile/tree/master/app

package zgo

import "github.com/torlangballe/zutil/zgeo"

type windowNative struct {
}

func WindowGetCurrent() *Window {
	w := &Window{}
	return w
}

func (w *Window) GetRect() zgeo.Rect {
	var s zgeo.Size
	return zgeo.Rect{Size: s}
}
