package zui

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/torlangballe/zutil/ztimer"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zslice"
)

var windows []*Window

func init() {
	WindowGetCurrent().SetHandleClosed(func() {
		// zlog.Info("Main window closed or refreshed?")
		for _, w := range windows {
			w.Close()
		}
		windows = windows[:0]
	})
}

type windowNative struct {
	element js.Value
}

func WindowGetCurrent() *Window {
	w := &Window{}
	w.element = WindowJS
	return w
}

func (w *Window) Rect() zgeo.Rect {
	var r zgeo.Rect
	r.Pos.X = w.element.Get("screenX").Float()
	r.Pos.Y = w.element.Get("screenY").Float()
	r.Size.W = w.element.Get("innerWidth").Float()
	r.Size.H = w.element.Get("innerHeight").Float()
	return r
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
	win.element = WindowJS.Call("open", surl, "_blank", strings.Join(specs, ","))
	// zlog.Info("OPENEDWIN:", win.element, surl)
	windows = append(windows, win)
	return win
}

func WindowOpenOtherURLType(surl string) {
	// zlog.Info("OPEN URL:", surl)
	WindowJS.Get("location").Set("href", surl)
}

func (w *Window) Close() {
	w.element.Call("close")
}

func (w *Window) Activate() {
	w.element.Call("focus")
}

func (w *Window) SetTitle(title string) {
	// zlog.Info("setttile", w.element, title)
	w.element.Get("document").Set("title", title)
}

func (w *Window) SetHandleClosed(closed func()) {
	w.element.Set("onbeforeunload", js.FuncOf(func(a js.Value, array []js.Value) interface{} {
		for i, win := range windows {
			if win == w {
				zslice.RemoveAt(&windows, i)
			}
		}
		if closed != nil {
			closed()
		}
		return nil
	}))
}

var resizeTimer *ztimer.Timer

func (w *Window) AddView(v View) {
	// ftrans := js.FuncOf(func(js.Value, []js.Value) interface{} {
	// 	return nil
	// })
	wn := &NativeView{}
	wn.Element = w.element.Get("document").Get("documentElement")
	wn.View = wn

	// s := WindowGetCurrent().Rect().Size.DividedByD(2)

	// o, _ := v.(NativeViewOwner)
	// if o == nil {
	// 	panic("NativeView AddChild child not native")
	// }
	// nv := o.GetNative()
	// nv.style().Set("display", "inline-block")

	// scale := fmt.Sprintf("scale(%f)", ScreenMain().Scale)
	// n.style().Set("-webkit-transform", scale)

	// trans := fmt.Sprintf("translate(-%f,%f)", s.W, 0.0)
	// zlog.Info("TRANS:", trans)
	// n.style().Set("-webkit-transform", trans)
	wn.AddChild(v, -1)
	w.element.Set("onresize", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		if resizeTimer == nil {
			resizeTimer = ztimer.StartIn(0.2, func() {
				r := w.Rect()
				if w.HandleBeforeResized != nil {
					w.HandleBeforeResized(r)
				}
				r.Pos = zgeo.Pos{}
				v.SetRect(r)
				if w.HandleAfterResized != nil {
					w.HandleAfterResized(r)
				}
				resizeTimer = nil
			})
		}
		return nil
	}))
}
