package zui

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/torlangballe/zutil/zhttp"
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
	hasResized      bool
	element         js.Value
	animationFrames map[int]int // maps random animation id to dom animationFrameID
}

func WindowGetMain() *Window {
	w := &Window{}
	w.element = WindowJS
	return w
}

func (w *Window) Rect() zgeo.Rect {
	var r zgeo.Rect
	r.Pos.X = w.element.Get("screenX").Float()
	r.Pos.Y = w.element.Get("screenY").Float()
	// r.Size.W = w.element.Get("innerWidth").Float()
	// r.Size.H = w.element.Get("innerHeight").Float()
	r.Size.W = w.element.Get("outerWidth").Float()
	r.Size.H = w.element.Get("outerHeight").Float()
	return r
}

func (w *Window) ContentRect() zgeo.Rect {
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
	}

	rect, gotPos, gotSize := getRectFromOptions(o)
	if gotPos {
		specs = append(specs, fmt.Sprintf("left=%d,top=%d", int(rect.Pos.X), int(rect.Pos.Y)))
	}
	if gotSize {
		specs = append(specs, fmt.Sprintf("width=%d,height=%d", int(rect.Size.W), int(rect.Size.H)))
	}
	if o.URL != "" && !zhttp.StringStartsWithHTTPX(o.URL) {
		o.URL = WindowGetMain().GetURLWithNewPathAndArgs(o.URL, nil)
	}
	// zlog.Info("OPEN WIN:", o.URL)
	win.element = WindowJS.Call("open", o.URL, "_blank", strings.Join(specs, ","))
	win.ID = o.ID
	windows[win] = true
	// zlog.Info("OPENEDWIN:", o.URL, specs, len(windows))
	win.element.Set("onbeforeunload", js.FuncOf(func(a js.Value, array []js.Value) interface{} {
		if win.ProgrammaticView != nil {
			pnv := ViewGetNative(win.ProgrammaticView)
			pnv.StopStoppers()
		}
		// zlog.Info("Window Closed!", win.ID, win.animationFrames)
		delete(windows, win)
		if win.HandleClosed != nil {
			win.HandleClosed()
		}
		return nil
	}))
	return win
}

func (win *Window) setOnResize() {
	win.element.Set("onresize", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		// zlog.Info("On Resize", win.hasResized)
		if !win.hasResized {
			win.hasResized = true
			return nil
		}
		ztimer.StartIn(0.2, func() {
			r := win.ContentRect()
			if win.HandleBeforeResized != nil {
				win.HandleBeforeResized(r)
			}
			r.Pos = zgeo.Pos{}
			if win.ProgrammaticView != nil {
				// zlog.Info("On Resized: to", win.ProgrammaticView.ObjectName(), r.Size, reflect.ValueOf(win.ProgrammaticView).Type(), "from:", win.ProgrammaticView.Rect().Size)
				win.ProgrammaticView.SetRect(r)
				if win.HandleAfterResized != nil {
					win.HandleAfterResized(r)
				}
			}
		})
		return nil
	}))
}

func (win *Window) GetURL() string {
	return win.element.Get("location").Get("href").String()
}

func (win *Window) SetLocation(surl string) {
	win.element.Get("location").Set("href", surl)
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

func (w *Window) AddView(v View) {
	// ftrans := js.FuncOf(func(js.Value, []js.Value) interface{} {
	// 	return nil
	// })
	w.ProgrammaticView = v
	wn := &NativeView{}
	wn.Element = w.element.Get("document").Get("documentElement")
	wn.View = wn

	wn.AddChild(v, -1)
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

func windowsFindForElement(e js.Value) *Window {
	for w, _ := range windows {
		if w.element.Equal(e) {
			return w
		}
	}
	return nil
}

func (win *Window) SetAddressBarURL(surl string) {
	win.element.Get("history").Call("pushState", "", "", surl)
}

func setKeyHandler(doc js.Value, handler func(KeyboardKey, KeyboardModifier)) {
}

func (win *Window) SetKeypressHandler(handler func(KeyboardKey, KeyboardModifier)) {
	win.keyPressedHandler = handler
	doc := win.element.Get("document")
	doc.Set("onkeyup", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		if handler != nil {
			key, mods := getKeyAndModsFromEvent(vs[0])
			handler(key, mods)
		}
		return nil
	}))
	doc.Set("onvisibilitychange", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		win.SetKeypressHandler(handler)
		zlog.Info("WIN activate!")
		return nil
	}))
}
