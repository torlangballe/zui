package zwindow

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

var winMain *Window

type windowNative struct {
	hasResized      bool
	Element         js.Value
	AnimationFrames map[int]int // maps random animation id to dom animationFrameID
}

func init() {
	zdom.WindowJS.Set("onbeforeunload", js.FuncOf(func(a js.Value, array []js.Value) interface{} {
		// zlog.Info("Main window closed or refreshed?")
		for w, _ := range windows {
			w.Close()
		}
		windows = map[*Window]bool{} // this might not be necessary, as we're shutting down?
		return nil
	}))
	winMain = New()
	winMain.Element = zdom.WindowJS
	windows[winMain] = true
	zview.RemoveKeyPressHandlerViewsFunc = func(v zview.View) {
		win := GetFromNativeView(v.Native())
		win.removeKeyPressHandlerViews(v)
	}
}

func GetMain() *Window {
	return winMain
}

func (w *Window) Rect() zgeo.Rect {
	var r zgeo.Rect
	r.Pos.X = w.Element.Get("screenX").Float()
	r.Pos.Y = w.Element.Get("screenY").Float()
	// r.Size.W = w.Element.Get("innerWidth").Float()
	// r.Size.H = w.Element.Get("innerHeight").Float()
	r.Size.W = w.Element.Get("outerWidth").Float()
	r.Size.H = w.Element.Get("outerHeight").Float()
	return r
}

func (w *Window) ContentRect() zgeo.Rect {
	var r zgeo.Rect
	r.Pos.X = w.Element.Get("screenX").Float()
	r.Pos.Y = w.Element.Get("screenY").Float()
	r.Size.W = w.Element.Get("innerWidth").Float()
	r.Size.H = w.Element.Get("innerHeight").Float()
	return r
}

// Open opens a new window, currently with a URL, which can be blank.
// It can set the *o.Size* if non-zero, and *o.Pos* if non-null.
// Use *loaded* callback before setting title etc, as this is otherwise set during load
func Open(o Options) *Window {
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
		o.URL = GetMain().GetURLWithNewPathAndArgs(o.URL, nil)
	}
	// zlog.Info("OPEN WIN:", o.URL, zlog.GetCallingStackString())
	win.Element = zdom.WindowJS.Call("open", o.URL, "_blank", strings.Join(specs, ","))
	if win.Element.IsNull() {
		zlog.Error(nil, "open window failed", o.URL)
		return nil
	}
	win.ID = o.ID
	windows[win] = true
	// zlog.Info("OPENEDWIN:", o.URL, specs, win.Element, len(windows))
	win.Element.Set("onbeforeunload", js.FuncOf(func(a js.Value, array []js.Value) interface{} {
		if win.ProgrammaticView != nil {
			pnv := win.ProgrammaticView.Native()
			pnv.PerformAddRemoveFuncs(true)
		}
		// zlog.Info("Window Closed!", win.ID, win.AnimationFrames)
		delete(windows, win)
		if win.HandleClosed != nil {
			win.HandleClosed()
		}
		return nil
	}))
	return win
}

func (win *Window) SetOnResizeHandling() {
	win.Element.Set("onresize", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		// fmt.Println("On Resize1", win.ProgrammaticView.ObjectName())
		// if !win.hasResized { // removing this so we can get first resize... what was it for?
		// 	win.hasResized = true
		// 	return nil
		// }
		ztimer.StartIn(0.2, func() {
			r := win.ContentRect()
			if win.HandleBeforeResized != nil {
				win.HandleBeforeResized(r)
			}
			r.Pos = zgeo.Pos{}
			if win.ResizeHandlingView != nil {
				// zlog.Info("On Resized: to", win.ProgrammaticView.ObjectName(), r.Size, reflect.ValueOf(win.ProgrammaticView).Type(), "from:", win.ProgrammaticView.Rect().Size)
				win.ResizeHandlingView.SetRect(r)
				// SetElementRect(win.Element, r)
				if win.HandleAfterResized != nil {
					win.HandleAfterResized(r)
				}
			}
		})
		return nil
	}))
}

func (win *Window) GetURL() string {
	return win.Element.Get("location").Get("href").String()
}

func (win *Window) SetLocation(surl string) {
	win.Element.Get("location").Set("href", surl)
}

func (w *Window) Close() {
	w.Element.Call("close")
}

func (w *Window) Activate() {
	w.Element.Call("focus")
}

// func (w *Window) GetFocusedView() *zview.NativeView {
// 	e := w.Element.Get("document").Get("activeElement")
// 	if e.IsUndef() {
// 		return nil
// 	}
// 	zlog.Info("GetFocusedView:", w.ProgrammaticView != nil)
// 	return nil
// }

func (w *Window) SetTitle(title string) {
	// zlog.Info("setttile", w.Element, title)
	w.Element.Get("document").Set("title", title)
}

func setDarkCSSStylings(doc js.Value) {
	css := `
*::selection {
background-color: #774433;
color: #ddd;
}`
	head := doc.Call("getElementsByTagName", "head").Index(0)
	style := doc.Call("createElement", "style")
	style.Set("type", "text/css")
	style.Call("appendChild", doc.Call("createTextNode", css))
	head.Call("appendChild", style)
}

func (w *Window) AddView(v zview.View) {
	// ftrans := js.FuncOf(func(js.Value, []js.Value) interface{} {
	// 	return nil
	// })
	// zlog.Info("Win:AddView", v.ObjectName(), reflect.ValueOf(v).Type())
	w.ProgrammaticView = v
	wn := &zview.NativeView{}
	//	wn.Element = w.Element.Get("document").Get("documentElement")
	doc := w.Element.Get("document")
	wn.Element = doc.Get("body")
	wn.View = wn
	if zstyle.Dark {
		//		setDarkCSSStylings(doc)
	}
	wn.SetObjectName("window")
	v.Native().JSStyle().Set("overflowX", "hidden")
	wn.AddChild(v, -1)
}

func (w *Window) SetScrollHandler(handler func(pos zgeo.Pos)) {
	w.Element.Set("scroll", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if handler != nil {
			y := w.Element.Get("scrollY").Float()
			handler(zgeo.Pos{0, y})
		}
		return nil
	}))
}

func findForElement(e js.Value) *Window {
	for w, _ := range windows {
		if w.Element.Equal(e) {
			return w
		}
	}
	return nil
}

func (win *Window) SetAddressBarURL(surl string) {
	win.Element.Get("history").Call("pushState", "", "", surl)
}

// func setKeyHandler(doc js.Value, handler func(zkeyboard.Key, zkeyboard.Modifier)) {
// }

func (win *Window) setOnKeyDown() {
	doc := win.Element.Get("document")
	doc.Set("onkeydown", js.FuncOf(func(val js.Value, args []js.Value) interface{} {
		key, mods := zkeyboard.GetKeyAndModsFromEvent(args[0])
		// zlog.Info("win key:", key)
		if len(win.keyHandlers) != 0 {
			var used bool
			for view, h := range win.keyHandlers {
				if PresentedViewCurrentIsParentFunc(view) {
					if h(key, mods) {
						used = true
					}
				}
			}
			if used {
				event := args[0]
				event.Call("preventDefault") // so they don't scroll scrollview with other stuff on top of it
			}
		}
		return nil
	}))
}

func (win *Window) removeKeyPressHandlerViews(root zview.View) {
	// fmt.Printf("removeKeyPressHandlerViews1: %+v\n", root)
	// zlog.Info("removeKeyPressHandlerViews:", root.ObjectName(), reflect.ValueOf(root).Type())
	includeCollapsed := false
	zcontainer.ViewRangeChildren(root, true, includeCollapsed, func(view zview.View) bool {
		// zlog.Info("removeKeyPressHandlerView try:", view.ObjectName(), win != nil)
		if win != nil && win.keyHandlers != nil {
			// zlog.Info("removeKeyPressHandlerView:", view.ObjectName())
			delete(win.keyHandlers, view) // I guess we could just call delete without checking if it exists first, faster?
		}
		return true
	})
}

func (win *Window) AddKeypressHandler(v zview.View, handler func(zkeyboard.Key, zkeyboard.Modifier) bool) {
	if handler == nil {
		delete(win.keyHandlers, v)
		return
	}
	win.keyHandlers[v] = handler
	win.setOnKeyDown()
	// zlog.Info("Window AddKeypressHandler", v.ObjectName(), len(win.keyHandlers))
	doc := win.Element.Get("document")
	doc.Set("onvisibilitychange", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		win.setOnKeyDown()
		zlog.Info("WIN activate!")
		return nil
	}))
}

func GetFromNativeView(v *zview.NativeView) *Window {
	we := v.GetWindowElement()
	return findForElement(we)
}
