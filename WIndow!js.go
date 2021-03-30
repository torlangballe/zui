// +build !js,zui

// https://github.com/golang/mobile/tree/master/app

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

type windowNative struct {
}

func WindowGetMain() *Window {
	w := &Window{}
	return w
}

func (w *Window) Rect() zgeo.Rect {
	var s zgeo.Size
	return zgeo.Rect{Size: s}
}

func (w *Window) ContentRect() zgeo.Rect {
	return w.Rect()
}

func WindowOpen(o WindowOptions) *Window {
	return nil
}

func (win *Window) GetURL() string {
	return ""
}

func (win *Window) setOnResize() {}

// func (win *Window) SetAddressBarPathAndArgs(path string, args zdict.Dict) {}
func (w *Window) Close()                                                                   {}
func (w *Window) Activate()                                                                {}
func (w *Window) AddView(v View)                                                           {}
func (w *Window) SetTitle(title string)                                                    {}
func (w *Window) SetHandleClosed(closed func())                                            {}
func (win *Window) SetAddressBarURL(surl string)                                           {}
func (win *Window) SetLocation(surl string)                                                {}
func (win *Window) AddKeypressHandler(v View, handler func(KeyboardKey, KeyboardModifier)) {}
