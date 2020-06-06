package zui

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type windowNative struct {
	element js.Value
}

func WindowGetCurrent() *Window {
	w := &Window{}
	w.element = WindowJS
	return w
}

func (w *Window) Rect() zgeo.Rect {
	var s zgeo.Size
	s.W = w.element.Get("innerWidth").Float()
	s.H = w.element.Get("innerHeight").Float()
	return zgeo.Rect{Size: s}
}

// WindowOpenWithURL opens a new window with a given url.
// It can set the *size* if non-zero, and *pos* if non-null.
// Use *loaded* callback before setting title etc, as this is otherwise set during load
func WindowOpenWithURL(surl string, size zgeo.Size, pos *zgeo.Pos) *Window {
	win := &Window{}
	var specs []string
	if !size.IsNull() {
		specs = append(specs, fmt.Sprintf("width=%d,height=%d", int(size.W), int(size.H)))
	}
	if pos != nil {
		specs = append(specs, fmt.Sprintf("left=%d,top=%d", int(pos.X), int(pos.Y)))
	}
	win.element = WindowJS.Call("open", surl, "banana_blank", strings.Join(specs, ","))
	// zlog.Info("OPENEDWIN:", win.element, surl)
	return win
}

func (w *Window) Close() {
	w.element.Call("close")
}

func (w *Window) Activate() {
	w.element.Call("focus")
}

func (w *Window) SetTitle(title string) {
	zlog.Info("setttile", w.element, title)
	w.element.Get("document").Set("title", title)
}
