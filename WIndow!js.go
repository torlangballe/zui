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

func WindowOpen(o WindowOptions) *Window {
	return nil
}

func (w *Window) Close()                        {}
func (w *Window) Activate()                     {}
func (w *Window) AddView(v View)                {}
func (w *Window) SetTitle(title string)         {}
func (w *Window) SetHandleClosed(closed func()) {}
