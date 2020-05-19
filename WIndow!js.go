// +build !js

// https://github.com/golang/mobile/tree/master/app

package zui

import "github.com/torlangballe/zutil/zgeo"

type windowNative struct {
}

func WindowGetCurrent() *Window {
	w := &Window{}
	return w
}

func (w *Window) Rect() zgeo.Rect {
	var s zgeo.Size
	return zgeo.Rect{Size: s}
}

func WindowOpenWithURL(surl string, size zgeo.Size, pos *zgeo.Pos) *Window {
	return nil
}

func (w *Window) Close() {
}

func (w *Window) Activate() {
}
