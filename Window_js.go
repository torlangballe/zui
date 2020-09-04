package zui

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"

	"github.com/torlangballe/zutil/zgeo"
)

func init() {
	WindowJS.Set("onbeforeunload", js.FuncOf(func(a js.Value, array []js.Value) interface{} {
		// zlog.Info("Main window closed or refreshed?")
		for w, _ := range windows {
			w.Close()
		}
		windows = map[*Window]bool{} // this might not be necessary, as we're shutting down?
		return nil
	}))
}

type windowNative struct {
	hasResized bool
	element    js.Value
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

// WindowOpen opens a new window, currently with a URL, which can be blank.
// It can set the *o.Size* if non-zero, and *o.Pos* if non-null.
// Use *loaded* callback before setting title etc, as this is otherwise set during load
func WindowOpen(o WindowOptions) *Window {
	win := &Window{}
	var specs []string
	if !o.Size.IsNull() {
		specs = append(specs, fmt.Sprintf("width=%d,height=%d", int(o.Size.W), int(o.Size.H)))
	}
	if o.Pos != nil {
		specs = append(specs, fmt.Sprintf("left=%d,top=%d", int(o.Pos.X), int(o.Pos.Y)))
	}
	win.element = WindowJS.Call("open", o.URL, "_blank", strings.Join(specs, ","))
	win.ID = o.ID
	// zlog.Info("OPENEDWIN:", win.element, surl)
	windows[win] = true
	win.element.Set("onbeforeunload", js.FuncOf(func(a js.Value, array []js.Value) interface{} {
		// zlog.Info("Window Closed!")
		delete(windows, win)
		if win.HandleClosed != nil {
			win.HandleClosed()
		}
		return nil
	}))
	return win
}

func WindowCurrentSetLocation(surl string) {
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
	//	ztimer.StartIn(0.1, func() {
	w.element.Set("onresize", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		if !w.hasResized {
			w.hasResized = true
			return nil
		}
		if resizeTimer == nil {
			resizeTimer = ztimer.StartIn(0.2, func() {
				r := w.Rect()
				if w.HandleBeforeResized != nil {
					w.HandleBeforeResized(r)
				}
				r.Pos = zgeo.Pos{}
				zlog.Info("On Resized: to", r.Size, "from:", v.Rect().Size)
				v.SetRect(r)
				if w.HandleAfterResized != nil {
					w.HandleAfterResized(r)
				}
				resizeTimer = nil
			})
		}
		return nil
	}))
	//	})
}

func (w *Window) SetScrollHandler(handler func(pos zgeo.Pos)) {
	w.element.Set("scroll", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if handler != nil {
			y := w.element.Get("scrollY").Float()
			handler(zgeo.Pos{0, y})
		}
		return nil
	}))
}
