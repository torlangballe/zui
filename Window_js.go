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

var winMain *Window

type windowNative struct {
	hasResized      bool
	element         js.Value
	animationFrames map[int]int // maps random animation id to dom animationFrameID
}

func init() {
	WindowJS.Set("onbeforeunload", js.FuncOf(func(a js.Value, array []js.Value) interface{} {
		// zlog.Info("Main window closed or refreshed?")
		for w, _ := range windows {
			w.Close()
		}
		windows = map[*Window]bool{} // this might not be necessary, as we're shutting down?
		return nil
	}))
	winMain = WindowNew()
	winMain.element = WindowJS
	windows[winMain] = true
}

func WindowGetMain() *Window {
	return winMain
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
	// zlog.Info("OPEN WIN:", o.URL, zlog.GetCallingStackString())
	win.element = WindowJS.Call("open", o.URL, "_blank", strings.Join(specs, ","))
	if win.element.IsNull() {
		zlog.Error(nil, "open window failed", o.URL)
		return nil
	}
	win.ID = o.ID
	windows[win] = true
	// zlog.Info("OPENEDWIN:", o.URL, specs, win.element, len(windows))
	win.element.Set("onbeforeunload", js.FuncOf(func(a js.Value, array []js.Value) interface{} {
		if win.ProgrammaticView != nil {
			pnv := ViewGetNative(win.ProgrammaticView)
			pnv.PerformAddRemoveFuncs(true)
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

func (win *Window) setOnResizeHandling() {
	win.element.Set("onresize", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
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
			if win.resizeHandlingView != nil {
				// zlog.Info("On Resized: to", win.ProgrammaticView.ObjectName(), r.Size, reflect.ValueOf(win.ProgrammaticView).Type(), "from:", win.ProgrammaticView.Rect().Size)
				win.resizeHandlingView.SetRect(r)
				// setElementRect(win.element, r)
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

func (w *Window) AddView(v View) {
	// ftrans := js.FuncOf(func(js.Value, []js.Value) interface{} {
	// 	return nil
	// })
	// zlog.Info("Win:AddView", v.ObjectName(), reflect.ValueOf(v).Type())
	w.ProgrammaticView = v
	wn := &NativeView{}
	//	wn.Element = w.element.Get("document").Get("documentElement")
	doc := w.element.Get("document")
	wn.Element = doc.Get("body")
	wn.View = wn
	if StyleDark {
		//		setDarkCSSStylings(doc)
	}
	wn.SetObjectName("window")
	ViewGetNative(v).style().Set("overflowX", "hidden")
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

// func setKeyHandler(doc js.Value, handler func(KeyboardKey, KeyboardModifier)) {
// }

func (win *Window) setOnKeyUp() {
	doc := win.element.Get("document")
	// zlog.Info("win keydown")
	doc.Set("onkeydown", js.FuncOf(func(val js.Value, args []js.Value) interface{} {
		if len(win.keyHandlers) != 0 {
			key, mods := getKeyAndModsFromEvent(args[0])
			for view, h := range win.keyHandlers {
				// zlog.Info("win key:", key, view.ObjectName())
				if PresentedViewCurrentIsParent(view) {
					// zlog.Info("win key2:", key, ViewGetNative(view).Hierarchy())
					h(key, mods)
				}
			}
			if key == KeyboardKeyDownArrow || key == KeyboardKeyUpArrow {
				event := args[0]
				event.Call("preventDefault") // so they don't scroll scrollview with other stuff on top of it
			}
		}
		return nil
	}))
}

func (win *Window) removeKeyPressHandlerViews(root View) {
	// fmt.Printf("removeKeyPressHandlerViews1: %+v\n", root)
	// zlog.Info("removeKeyPressHandlerViews:", root.ObjectName(), reflect.ValueOf(root).Type())
	ct, is := root.(ContainerType)
	if !is {
		return
	}
	includeCollapsed := false
	ContainerTypeRangeChildren(ct, true, includeCollapsed, func(view View) bool {
		// zlog.Info("removeKeyPressHandlerView try:", view.ObjectName(), win != nil)
		if win != nil && win.keyHandlers != nil {
			// zlog.Info("removeKeyPressHandlerView:", view.ObjectName())
			delete(win.keyHandlers, view) // I guess we could just call delete without checking if it exists first, faster?
		}
		return true
	})
}

func (win *Window) AddKeypressHandler(v View, handler func(KeyboardKey, KeyboardModifier)) {
	// zlog.Info("Window AddKeypressHandler", v.ObjectName())
	if handler == nil {
		delete(win.keyHandlers, v)
		return
	}
	win.keyHandlers[v] = handler
	win.setOnKeyUp()
	doc := win.element.Get("document")
	doc.Set("onvisibilitychange", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		win.setOnKeyUp()
		// zlog.Info("WIN activate!")
		return nil
	}))
}
