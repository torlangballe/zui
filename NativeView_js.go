package zui

import (
	"fmt"
	"reflect"
	"strconv"
	"syscall/js"
	"time"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

type baseNativeView struct {
	Element      js.Value
	transparency float32
	parent       *NativeView
}

func (v *NativeView) MakeJSElement(view View, etype string) {
	v.Element = DocumentJS.Call("createElement", etype)
	v.Element.Set("style", "position:absolute")
	v.View = view
}

func (v *NativeView) Parent() *NativeView {
	// e := v.get("parentElement")
	// if e.Type() == js.TypeUndefined || e.Type() == js.TypeNull {
	// 	return nil
	// }
	// //	zlog.Info("ParentElement:", v.ObjectName(), e)
	// n := &NativeView{}
	// v.Element = e
	// v.View = n
	// return n
	return v.parent
}

func (v *NativeView) Child(path string) View {
	return ViewChild(v.View, path)
}

func (v *NativeView) GetNative() *NativeView {
	return v
}

func (v *NativeView) SetRect(rect zgeo.Rect) {
	// if v.ObjectName() == "v2" {
	// zlog.Info("NV Rect", v.ObjectName(), rect, zlog.GetCallingStackString())
	// }
	rect = rect.ExpandedToInt()
	setElementRect(v.Element, rect)
}

func (v *NativeView) HasSize() bool {
	return v.style().Get("left").String() != ""
}

func (v *NativeView) parseElementCoord(value js.Value) float64 {
	var s string
	str := value.String()
	if zstr.HasSuffix(str, "px", &s) {
		n, err := strconv.ParseFloat(s, 32)
		if err != nil {
			zlog.Error(err, "not number", v.ObjectName())
			return 0
		}
		return n
	}
	iv, _ := v.View.(*ImageView)
	if iv != nil {
		zlog.Error(nil, "parseElementCoord For Image: not handled type:", str, iv.Path()) //! Fatal
	}
	//	zlog.Error(nil, "parseElementCoord: not handled type:", str, v.ObjectName(), zlog.GetCallingStackString()) //! Fatal
	return 0
}

func (v *NativeView) Rect() zgeo.Rect {
	var pos zgeo.Pos
	style := v.style()
	// pos.X = v.Element.Get("offsetLeft").Float()
	// pos.Y = v.Element.Get("offsetTop").Float()
	pos.X = v.parseElementCoord(style.Get("left"))
	pos.Y = v.parseElementCoord(style.Get("top"))
	size := v.LocalRect().Size
	return zgeo.Rect{pos, size}
}

func setElementRect(e js.Value, rect zgeo.Rect) {
	style := e.Get("style")
	style.Set("left", fmt.Sprintf("%fpx", rect.Pos.X))
	style.Set("top", fmt.Sprintf("%fpx", rect.Pos.Y))
	style.Set("width", fmt.Sprintf("%fpx", rect.Size.W))
	style.Set("height", fmt.Sprintf("%fpx", rect.Size.H))
}

func (v *NativeView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{10, 10}
}

func (v *NativeView) LocalRect() zgeo.Rect {
	var w, h float64
	style := v.style()
	sw := style.Get("width")
	sh := style.Get("height")
	if sw.String() != "" {
		h = v.parseElementCoord(sh)
		w = v.parseElementCoord(sw)
	} else if v.Presented {
		zlog.Info("parse empty Coord:", v.Hierarchy(), zlog.GetCallingStackString())
	}

	return zgeo.RectMake(0, 0, w, h)
}

func (v *NativeView) SetLocalRect(rect zgeo.Rect) {
	zlog.Fatal(nil, "NOT IMPLEMENTED")
}

func (v *NativeView) ObjectName() string {
	return v.getjs("oname").String()
}

var idCount int

func (v *NativeView) SetObjectName(name string) {
	v.setjs("oname", name)
	idCount++
	v.setjs("id", fmt.Sprintf("%s-%d", name, idCount))
}

func makeRGBAString(c zgeo.Color) string {
	if !c.Valid {
		return "initial"
	}
	return c.Hex()
	//	rgba := c.GetRGBA()
	//	return fmt.Sprintf("rgba(%d,%d,%d,%g)", int(rgba.R*255), int(rgba.G*255), int(rgba.B*255), rgba.A)
}

func (v *NativeView) SetColor(c zgeo.Color) {
	v.style().Set("color", makeRGBAString(c))
}

func (v *NativeView) Color() zgeo.Color {
	str := v.style().Get("color").String()
	return zgeo.ColorFromString(str)
}

func (v *NativeView) style() js.Value {
	return v.getjs("style")
}

func (v *NativeView) SetStyle(key, value string) {
	v.style().Set(key, value)
}

func (v *NativeView) SetAlpha(alpha float32) {
	v.transparency = 1 - alpha
	//	v.style().Set("alpha", alpha)
	v.style().Set("opacity", alpha)
}

func (v *NativeView) Alpha() float32 {
	return 1 - v.transparency
}

func (v *NativeView) SetBGColor(c zgeo.Color) {
	// zlog.Info("SetBGColor:", v.ObjectName(), c)
	v.style().Set("backgroundColor", makeRGBAString(c))
}

func (v *NativeView) BGColor() zgeo.Color {
	str := v.style().Get("backgroundColor").String()
	zlog.Info("nv bgcolor", str)
	return zgeo.ColorFromString(str)
}

func (v *NativeView) SetCorner(radius float64) {
	style := v.style()
	s := fmt.Sprintf("%dpx", int(radius))
	style.Set("-moz-border-radius", s)
	style.Set("-webkit-border-radius", s)
	style.Set("border-radius", s)
}

func (v *NativeView) SetStroke(width float64, c zgeo.Color) {
	str := "none"
	if width != 0 {
		str = fmt.Sprintf("%dpx solid %s", int(width), makeRGBAString(c))
	}
	v.style().Set("border", str)
}

func (v *NativeView) Scale(scale float64) {
}

func (v *NativeView) GetScale() float64 {
	return 1
}

func (v *NativeView) Show(show bool) {
	str := "hidden"
	if show {
		str = "visible"
	}
	v.style().Set("visibility", str)
}

func (v *NativeView) IsShown() bool {
	return v.style().Get("visibility").String() == "visible"
}

func (v *NativeView) SetUsable(usable bool) {
	v.setjs("disabled", !usable)
	style := v.style()
	var alpha float32 = 0.4
	if usable {
		alpha = 1 - v.transparency
	}
	str := "none"
	if usable {
		str = "auto"
	}
	style.Set("pointer-events", str)
	style.Set("opacity", alpha)
	// zlog.Info("NV SetUSABLE:", v.ObjectName(), usable, alpha)
}

func (v *NativeView) Usable() bool {
	dis := v.Element.Get("disabled")
	if dis.IsUndefined() {
		return true
	}
	return !dis.Bool()
}

func (v *NativeView) IsFocused() bool {
	f := DocumentJS.Call("hasFocus")
	return f.Equal(v.Element)
}

func (v *NativeView) Focus(focus bool) {
	// zlog.Info("FOCUS:", v.ObjectName(), focus)
	v.call("focus")
}

func (v *NativeView) SetCanFocus(can bool) {
	// zlog.Info("SetCanFocus:", v.ObjectName())
	val := "-1"
	if can {
		val = "0"
	}
	v.setjs("tabindex", val)
}

func (v *NativeView) SetOpaque(opaque bool) {
}

func (v *NativeView) GetChild(path string) *NativeView {
	zlog.Fatal(nil, "not implemented")
	return nil
}

func (v *NativeView) Hierarchy() string {
	var str string
	for _, p := range v.AllParents() {
		str += "/" + p.ObjectName()
	}
	str += "/" + v.ObjectName()
	return str
}

func (v *NativeView) DumpTree(prefix string) {
	includeCollapsed := true
	zlog.Info(prefix+v.ObjectName(), reflect.ValueOf(v).Type(), reflect.ValueOf(v).Kind())
	ct, _ := v.View.(ContainerType)
	if ct != nil {
		for _, v := range ct.GetChildren(includeCollapsed) {
			nv := ViewGetNative(v)
			nv.DumpTree(prefix + "**")
		}
	}
}

func (v *NativeView) RemoveFromParent() {
	// zlog.Info("RemoveFromParent:", v.ObjectName())
	zlog.Assert(v.parent != nil)
	v.parent.RemoveChild(v)
}

func (v *NativeView) SetFont(font *Font) {
	// zlog.Debug("font:", v.ObjectName(), font.Style&FontStyleItalic, font.Style&FontStyleBold)
	cssStyle := v.style()
	cssStyle.Set("font-style", string(font.Style&FontStyleItalic))
	cssStyle.Set("font-weight", (font.Style & FontStyleBold).String())
	cssStyle.Set("font-family", font.Name)
	cssStyle.Set("font-size", fmt.Sprintf("%gpx", font.Size))
	// cssText := cssStyle.Get("cssText").String()
	// cssText += fmt.Sprintf(";font-style:%v;font-weight:%v;font-family:%s;font-size:%gpx",
	// 	font.Style&FontStyleItalic,
	// 	font.Style&FontStyleBold,
	// 	font.Name,
	// 	font.Size,
	// )
	// // zlog.Info("Font csstext:", cssText)
	// cssStyle.Set("cssText", cssText)
}

func (v *NativeView) Font() *Font {
	cssStyle := v.style()
	fstyle := FontStyleNormal
	name := cssStyle.Get("font-family").String()
	if cssStyle.Get("font-weight").String() == "bold" {
		fstyle |= FontStyleBold
	}
	if cssStyle.Get("font-style").String() == "italic" {
		fstyle |= FontStyleItalic
	}
	ss := cssStyle.Get("font-size")
	size := v.parseElementCoord(ss)

	return FontNew(name, size, fstyle)
}

func (v *NativeView) SetText(text string) {
	//		zlog.Info("NV SETTEXT", v.ObjectName(), zlog.GetCallingStackString())
	v.setjs("innerText", text)
}

func (v *NativeView) Text() string {
	text := v.getjs("innerText").String()
	return text
}

func (v *NativeView) AddChild(child View, index int) {
	n := ViewGetNative(child)
	if n == nil {
		zlog.Fatal(nil, "NativeView AddChild child not native")
	}
	n.parent = v
	if index != -1 {
		nodes := n.parent.getjs("childNodes").Length()
		// zlog.Info("NS AddChild:", v.ObjectName(), child.ObjectName(), index, nodes)
		if nodes == 0 {
			v.call("appendChild", n.Element)
		} else {
			v.call("insertBefore", n.Element, v.getjs("firstChild"))
		}
	} else {
		v.call("appendChild", n.Element)
	}
	n.style().Set("zIndex", 100)
	for _, p := range n.AllParents() {
		p.allChildrenPresented = false
	}
}

func (v *NativeView) ReplaceChild(child, with View) {
	v.AddChild(with, -1) // needs to preserve index, which isn't really supported in AddChild yet anyway
	with.SetRect(child.Rect())
	v.RemoveChild(child)
	et, _ := with.(ExposableType)
	// zlog.Info("ReplaceChild:", et != nil, replace.ObjectName())
	if et != nil {
		et.Expose()
		et.drawIfExposed()
	}
}

func (v *NativeView) AllParents() (all []*NativeView) {
	for v.parent != nil {
		all = append([]*NativeView{v.parent}, all...)
		v = v.parent
	}
	return
}

func (v *NativeView) SetZIndex(index int) {
	v.style().Set("zIndex", index)
}

func (v *NativeView) RemoveChild(child View) {
	// zlog.Info("REMOVE CHILD:", child.ObjectName())
	nv := ViewGetNative(child)
	if nv == nil {
		panic("NativeView AddChild child not native")
	}
	win := v.GetWindow()
	win.removeKeyPressHandlerViews(child)
	nv.StopStoppers()
	nv.Element = v.call("removeChild", nv.Element) // we need to set it since  it might be accessed for ObjectName etc still in collapsed containers
	//!! nv.parent = nil if we don't do this, we can still uncollapse child in container without having to remember comtainer. Testing.
}

func nativeElementSetDropShadow(e js.Value, shadow zgeo.DropShadow) {
	str := fmt.Sprintf("%dpx %dpx %dpx %s", int(shadow.Delta.W), int(shadow.Delta.H), int(shadow.Blur), makeRGBAString(shadow.Color))
	e.Get("style").Set("boxShadow", str)
}

func (v *NativeView) SetDropShadow(shadow zgeo.DropShadow) {
	nativeElementSetDropShadow(v.Element, shadow)
}

func (v *NativeView) SetToolTip(str string) {
	v.setjs("title", str)
}

func (v *NativeView) AbsoluteRect() zgeo.Rect {
	r := v.Element.Call("getBoundingClientRect")
	x := r.Get("x").Float()
	y := r.Get("y").Float()
	w := r.Get("width").Float()
	h := r.Get("height").Float()
	return zgeo.RectFromXYWH(x, y, w, h)
}

type LongPresser struct {
	cancelClick     bool
	downClickedTime time.Time
	longTimer       *ztimer.Timer
}

func (lp *LongPresser) HandleOnClick(view View) {
	if lp.longTimer != nil {
		lp.longTimer.Stop()
	}
	p, _ := view.(Pressable)
	if p != nil && !lp.cancelClick && p.PressedHandler() != nil && view.Usable() {
		// zlog.Info("HANDLE ONCLICK!")
		p.PressedHandler()()
	}
	lp.cancelClick = false
}

func (lp *LongPresser) HandleOnMouseDown(view View) {
	// fmt.Println("MOUSEDOWN")
	lp.downClickedTime = time.Now()
	lp.longTimer = ztimer.StartIn(0.5, func() {
		p, _ := view.(Pressable)
		if p != nil && p.LongPressedHandler() != nil && view.Usable() {
			// fmt.Println("TIMER:", p != nil, p.LongPressedHandler() != nil, view.Usable())
			p.LongPressedHandler()()
		}
		lp.longTimer = nil
		lp.cancelClick = true
	})
}

func (lp *LongPresser) HandleOnMouseUp(view View) {
	if lp.longTimer != nil {
		lp.longTimer.Stop()
	}
}

var mouseStartPos zgeo.Pos
var mouseStartTime time.Time

func getMousePos(e js.Value) (pos zgeo.Pos) {
	pos.X = e.Get("clientX").Float()
	pos.Y = e.Get("clientY").Float()
	return
}

func (v *NativeView) SetSwipeHandler(handler func(pos, dir zgeo.Pos)) {
	const minDiff = 10.0
	v.setjs("onmousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if handler == nil {
			return nil
		}
		mouseStartTime = time.Now()
		mouseStartPos = getMousePos(args[0])
		return nil
	}))
	v.setjs("onmousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if handler == nil {
			return nil
		}
		if time.Since(mouseStartTime) > time.Second/2 {
			return nil
		}
		pos := getMousePos(args[0])
		diff := pos.Minus(mouseStartPos)
		a := diff.Abs()
		if a.X >= minDiff || a.Y > minDiff {
			if a.X < minDiff {
				diff.X = 0
			}
			if a.Y < minDiff {
				diff.Y = 0
			}
			mouseStartTime = time.Time{}
			var vpos zgeo.Pos
			r := v.Element.Call("getBoundingClientRect")
			vpos.X = r.Get("left").Float()
			vpos.Y = r.Get("top").Float()
			pos := mouseStartPos.Minus(vpos)
			handler(pos, diff)
		}
		return nil
	}))
}

func (v *NativeView) SetAboveParent(above bool) {
	zlog.Info("SetAboveParent:", v.ObjectName(), above)
	str := "hidden"
	if above {
		str = "visible"
	}
	v.style().Set("overflow", str)
}

func (v *NativeView) call(method string, args ...interface{}) js.Value {
	return v.Element.Call(method, args...)
}

func (v *NativeView) setjs(property string, value interface{}) {
	v.Element.Set(property, value)
}

func (v *NativeView) getjs(property string) js.Value {
	return v.Element.Get(property)
}

func (v *NativeView) SetScrollHandler(handler func(pos zgeo.Pos)) {
	v.setjs("onscroll", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if handler != nil {
			y := v.getjs("scrollTop").Float()
			handler(zgeo.Pos{0, y})
		}
		return nil
	}))
}

func (v *NativeView) RotateDeg(deg float64) {
	v.style().Set("transform", fmt.Sprintf("rotate(%fdeg)", deg))
}

func (v *NativeView) GetWindow() *Window {
	root := v
	all := v.AllParents()
	if len(all) > 1 {
		root = all[0]
	}
	w := root.getjs("ownerDocument").Get("defaultView")
	// zlog.Info("NV.GetWindow:", w, root.ObjectName())
	return windowsFindForElement(w)
}

func dragEvent(event js.Value, dtype DragType, handler func(dtype DragType, data []byte, name string) bool) {
	handler(dtype, nil, "")
	event.Call("preventDefault")
}

func (v *NativeView) SetDraggable(getData func() (data string, mime string)) {
	// https://www.digitalocean.com/community/tutorials/js-drag-and-drop-vanilla-js
	v.setjs("draggable", true)
	v.setjs("ondragstart", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		data, mime := getData()
		//		array := js.Global().Get("Uint8Array").New(len(data))
		//		js.CopyBytesToJS(array, []byte(data))
		// zlog.Info("Dtrans:", mime, array.Length())
		zlog.Info("Dtrans:", mime, data)
		//		mime = "text/plain"
		event.Get("dataTransfer").Call("setData", mime, data) //event.Get("target").Get("id"))
		return nil
	}))
}

func jsFileToGo(file js.Value, got func(data []byte, name string)) {
	reader := js.Global().Get("FileReader").New()
	reader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		array := js.Global().Get("Uint8Array").New(this.Get("result"))
		data := make([]byte, array.Length())
		js.CopyBytesToGo(data, array)
		name := file.Get("name").String()
		got(data, name)
		return nil
	}))
	reader.Call("readAsArrayBuffer", file)
}

func (v *NativeView) SetPointerDragHandler(handler func(dtype DragType, data []byte, name string) bool) {
	if zlog.IsInTests {
		return
	}
	v.setjs("ondragenter", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		dragEvent(args[0], DragEnter, handler)
		return nil
	}))
	v.setjs("ondragleave", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		dragEvent(args[0], DragLeave, handler)
		return nil
	}))
	v.setjs("ondragover", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		dragEvent(args[0], DragOver, handler)
		return nil
	}))
	v.setjs("ondrop", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		dt := event.Get("dataTransfer")
		files := dt.Get("files")
		if files.Length() == 0 {
			dragEvent(event, DragDrop, handler)
			return nil
		}
		file := files.Index(0)
		jsFileToGo(file, func(data []byte, name string) {
			handler(DragDropFile, data, name)
		})
		event.Call("preventDefault")
		return nil
	}))
}

func (v *NativeView) MakeUploader(got func(data []byte, name string)) {
	e := DocumentJS.Call("createElement", "input")
	e.Set("type", "file")
	e.Set("style", "opacity: 0.0; position: absolute; top: 0; left: 0; bottom: 0; right: 0; width: 100%; height:100%;")
	v.call("appendChild", e)

	e.Set("onchange", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		files := e.Get("files")
		if files.Length() > 0 {
			file := files.Index(0)
			jsFileToGo(file, got)
		}
		return nil
	}))
}

func MakeUploadButton() *ShapeView {
	v := ShapeViewNew(ShapeViewTypeRoundRect, zgeo.Size{68, 22})
	v.SetColor(zgeo.ColorWhite)
	v.StrokeColor = zgeo.ColorNew(0, 0.6, 0, 1)
	v.StrokeWidth = 2
	v.Ratio = 0.3
	v.SetBGColor(zgeo.ColorClear)
	v.SetText("Upload")
	return v
}

func (v *NativeView) SetPointerEnterHandler(handler func(pos zgeo.Pos, inside bool)) {
	if zlog.IsInTests {
		return
	}
	v.setjs("onmouseenter", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if handler != nil {
			// zlog.Info("Mouse enter", v.ObjectName())
			handler(getMousePos(args[0]), true)
		}
		return nil
	}))
	v.setjs("onmouseleave", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if handler != nil {
			// zlog.Info("Mouse leave", v.ObjectName())
			handler(getMousePos(args[0]), false)
		}
		return nil
	}))
}

func (v *NativeView) SetKeyHandler(handler func(key KeyboardKey, mods KeyboardModifier) bool) {
	jsSetKeyHandler(v.Element, func(key KeyboardKey, mods KeyboardModifier) bool {
		return handler(key, mods)
	})
}

func (v *NativeView) SetOnInputHandler(handler func()) {
	v.Element.Set("oninput", js.FuncOf(func(js.Value, []js.Value) interface{} {
		handler()
		return nil
	}))
}
func (v *NativeView) SetOnPointerMoved(handler func(pos zgeo.Pos)) {
	v.Element.Set("onmousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		x := e.Get("offsetX").Float()
		y := e.Get("offsetY").Float()
		handler(zgeo.Pos{x, y})
		return nil
	}))
}

func (v *NativeView) GetFocusedView() (found View) {
	ct := v.View.(ContainerType)
	e := DocumentJS.Get("activeElement")
	// zlog.Info("GetFocusedView:", e, e.IsUndefined())
	if e.IsUndefined() {
		return nil
	}

	foundID := e.Get("id").String()
	// zlog.Info("GetFocusedView2:", foundID)
	recursive := true
	includeCollapsed := false
	ContainerTypeRangeChildren(ct, recursive, includeCollapsed, func(view View) bool {
		n := ViewGetNative(view)
		id := n.getjs("id").String()
		if id == foundID {
			found = n.View
			return false
		}
		return true
	})
	return found
}
