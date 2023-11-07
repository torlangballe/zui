//go:build !js && zui && !catalyst

// https://github.com/golang/mobile/tree/master/app

package zwindow

import (
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

type windowNative struct {
}

// func GetMain() *Window {
// 	w := &Window{}
// 	return w
// }

func (w *Window) Rect() zgeo.Rect {
	var s zgeo.Size
	return zgeo.Rect{Size: s}
}

func (w *Window) ContentRect() zgeo.Rect {
	return w.Rect()
}

func Open(o Options) *Window {
	return nil
}

func (win *Window) GetURL() string {
	return ""
}

func (win *Window) setOnResize() {}

// func (win *Window) SetAddressBarPathAndArgs(path string, args zdict.Dict) {}
func (w *Window) Close()                         {}
func (w *Window) Activate()                      {}
func (w *Window) AddView(v zview.View)           {}
func (w *Window) SetTitle(title string)          {}
func (w *Window) SetHandleClosed(closed func())  {}
func (win *Window) SetAddressBarURL(surl string) {}
func (win *Window) SetLocation(surl string)      {}
func (win *Window) AddKeypressHandler(v zview.View, handler func(km zkeyboard.KeyMod, down bool) bool) {
}
func (win *Window) SetOnResizeHandling()         {}
func (win *Window) SetOnKeyEvents()              {}
func FromNativeView(v *zview.NativeView) *Window { return nil }
func Current() *Window                           { return nil }
func (win *Window) AddStyle()                    {}
