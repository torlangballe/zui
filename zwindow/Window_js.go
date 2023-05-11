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
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/ztimer"
)

var (
	winMain       *Window
	barCalculated bool
	// RemovePresentedWindowFunc func(view zview.View)
)

type windowNative struct {
	hasResized      bool
	Element         js.Value
	AnimationFrames map[int]int // maps random animation id to dom animationFrameID
}

func init() {
	zdom.WindowJS.Set("onbeforeunload", js.FuncOf(func(a js.Value, array []js.Value) interface{} {
		// zlog.Info("Main window closed or refreshed?")
		for w, _ := range windows {
			if w.Element.Equal(zdom.WindowJS) {
				continue
			}
			zlog.Info("onbeforeunload w:", w.ID)
			w.Close()
		}
		windows = map[*Window]bool{} // this might not be necessary, as we're shutting down?
		return nil
	}))
	winMain = New()
	winMain.Element = zdom.WindowJS
	windows[winMain] = true

	zview.RemoveKeyPressHandlerViewsFunc = func(v zview.View) {
		win := FromNativeView(v.Native())
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
	win := New()
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
	// zlog.Info("OPEN WIN:", o.URL, specs)
	win.Element = zdom.WindowJS.Call("open", o.URL, "_blank", strings.Join(specs, ","))
	if win.Element.IsNull() {
		zlog.Error(nil, "open window failed", o.URL)
		return nil
	}
	ztimer.StartIn(0.2, func() { // This is a hack as we don't know browser title bar height. It waits until window is placed, then calculates what title bar height should be, stores and changes for this window.
		if !barCalculated {
			barCalculated = true
			oh := win.Element.Get("outerHeight").Float()
			ih := win.Element.Get("innerHeight").Float()
			// zlog.Info("doc:", oh, ih, originalIH)
			newBar := oh - ih
			if newBar != barHeight {
				ow := win.Element.Get("outerWidth").Float()
				diff := newBar - barHeight
				barHeight = newBar
				win.Element.Call("resizeTo", ow, oh+diff)
			}
		}
	})
	win.ID = o.ID
	windows[win] = true

	// zlog.Info("OPENEDWIN:", o.URL, specs, win.Element, len(windows))
	win.Element.Set("onbeforeunload", js.FuncOf(func(a js.Value, array []js.Value) interface{} {
		// zlog.Info("Other window closed or refreshed")
		if win.ProgrammaticView != nil {
			pnv := win.ProgrammaticView.Native()
			pnv.PerformAddRemoveFuncs(true)
		}
		// zlog.Info("Window Closed!", win.ID, win.AnimationFrames)
		if win.HandleClosed != nil {
			win.HandleClosed()
		}
		delete(windows, win) // do this after HandleClosed
		return nil
	}))
	return win
}

func (win *Window) SetOnResizeHandling() {
	win.Element.Set("onresize", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		// zlog.Info("On Resize1", win.ProgrammaticView.ObjectName(), win.ResizeHandlingView != nil)
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
				// zlog.Info("On Resize", win.ResizeHandlingView.Native().Hierarchy(), r)
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
	v.Native().JSStyle().Set("overflow", "hidden")
	wn.AddChild(v, -1)
}

func (w *Window) Reload() {
	loc := w.Element.Get("location")
	loc.Call("reload")
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
		// zlog.Info("win findForElement:", e, w.Element)
		if w.Element.Equal(e) {
			return w
		}
	}
	return nil
}

func (win *Window) SetAddressBarURL(surl string) {
	win.Element.Get("history").Call("pushState", "", "", surl)
}

func (win *Window) SetOnKeyEvents() {
	setOnKeyEvent(win, true)
	setOnKeyEvent(win, false)
}

func setOnKeyEvent(win *Window, down bool) {
	eventName := "onkeyup"
	if down {
		eventName = "onkeydown"
	}
	doc := win.Element.Get("document")
	doc.Set(eventName, js.FuncOf(func(val js.Value, args []js.Value) interface{} {
		// zlog.Info("KeyWin:", win.Element.Get("outerWidth"), win.Element.Get("document").Call("hasFocus").Bool(), len(win.keyHandlers))
		if !win.Element.Get("document").Call("hasFocus").Bool() {
			return nil
		}
		km := zkeyboard.GetKeyModFromEvent(args[0])
		// zlog.Info("LEN:", len(win.ViewsStack), zlog.CallingStackString())
		top := win.ViewsStack[len(win.ViewsStack)-1].Native()
		if len(win.keyHandlers) != 0 {
			var used bool
			// zlog.Info("win key handlers:", km.Key, len(win.keyHandlers), top.Hierarchy())
			for _, h := range win.keyHandlers {
				// zlog.Info("win key2:", km.Key, h.view.Native().Hierarchy(), top.Hierarchy())
				if top == h.view.Native() || top.IsParentOf(h.view.Native()) {
					focused := top.Native().GetFocusedChildView(true)
					// zlog.Info("win key3:", km.Key, h.view.Native().Hierarchy(), top.Hierarchy(), focused != nil)
					// if focused != nil {
					// 	zlog.Info("win key1: foc:", focused.Native().Hierarchy())
					// }
					// zlog.Info("win key2:", key, view.Native().Hierarchy(), top.Hierarchy())
					if focused != nil { //&& focused != view.Native() {
						kc, _ := focused.View.(zkeyboard.KeyConsumer)
						if kc == nil {
							// if km.Key != zkeyboard.KeyEscape && !km.Key.IsReturnish() {
							// 	continue
							// }
						} else if kc.ConsumesKey(km) {
							continue
						}
					}
					if h.handler(km, down) {
						used = true
						break
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

func deleteKeyHandlers(win *Window, delView zview.View) {
	for i := 0; i < len(win.keyHandlers); {

		if win.keyHandlers[i].view == delView {
			zslice.RemoveAt(&win.keyHandlers, i)
		} else {
			i++
		}
	}
}

func (win *Window) removeKeyPressHandlerViews(root zview.View) {
	// fmt.Printf("removeKeyPressHandlerViews1: %+v\n", root)
	// zlog.Info("removeKeyPressHandlerViews:", root.ObjectName(), reflect.ValueOf(root).Type())
	includeCollapsed := false
	zcontainer.ViewRangeChildren(root, true, includeCollapsed, func(view zview.View) bool {
		// zlog.Info("removeKeyPressHandlerView try:", view.ObjectName(), win != nil)
		if win != nil && win.keyHandlers != nil {
			deleteKeyHandlers(win, view)
		}
		return true
	})
}

func (win *Window) AddKeypressHandler(v zview.View, handler func(km zkeyboard.KeyMod, down bool) bool) {
	if handler == nil {
		deleteKeyHandlers(win, v)
		return
	}
	v.Native().AddOnRemoveFunc(func() {
		deleteKeyHandlers(win, v)
	})
	win.keyHandlers = append(win.keyHandlers, keyHandler{v, handler})
	if v.Native().Presented {
		win.SetOnKeyEvents()
	}
	// zlog.Info("Window AddKeypressHandler", v.ObjectName(), len(win.keyHandlers))
	doc := win.Element.Get("document")
	doc.Set("onvisibilitychange", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		win.SetOnKeyEvents()
		return nil
	}))
}

func FromNativeView(v *zview.NativeView) *Window {
	we := v.GetWindowElement()
	win := findForElement(we)
	return win
}

func Current() *Window {
	for w := range windows {
		if w.Element.Get("document").Call("hasFocus").Bool() {
			return w
		}
	}
	return nil
}

func (win *Window) AddStyle() {
	styleStr := `
input.rounded:focus { border: 2px solid rgb(147,180,248); }
input:focus { border: 3px solid rgb(147,180,248); }
input[type=number]:focus { border: 2px solid rgb(147,180,248); }
input[type=number] { border: 1px solid gray; -webkit-box-shadow:none; }
.zfocus:focus { outline: solid 4px rgb(147,180,248); }
.znofocus:focus { outline: none; }
input::-webkit-outer-spin-button, input::-webkit-inner-spin-button { -webkit-appearance: none; }
input.rounded {
	border: 1px solid #666;
	-moz-border-radius: 10px;
	-webkit-border-radius: 10px;
	border-radius: 10px;
	box-sizing: content-box;
	outline: 0;
	-webkit-appearance: none;
}`
	doc := win.Element.Get("document")
	styleTag := doc.Call("createElement", "style")

	if styleTag.Get("styleSheet").IsUndefined() {
		styleTag.Call("appendChild", doc.Call("createTextNode", styleStr))
	} else {
		styleTag.Get("styleSheet").Set("cssText", styleStr)
	}
	doc.Call("getElementsByTagName", "head").Index(0).Call("appendChild", styleTag)
}
