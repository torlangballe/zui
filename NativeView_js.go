package zui

import (
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zutil/zbool"
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

type AddHandler interface {
	HandleAddAsChild()
}

func (v *NativeView) MakeJSElement(view View, etype string) {
	v.Element = zdom.DocumentJS.Call("createElement", etype)
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
	if v.parent != nil && v.parent.View != nil {
		// zlog.Info("PAR:", reflect.ValueOf(v.parent.View).Type())
		nv := ViewGetNative(v.parent.View)
		// zlog.Info("PAR2:", reflect.ValueOf(nv.View).Type())
		if nv != nil {
			return nv
		}
	}
	return v.parent
}

func (v *NativeView) Child(path string) View {
	return ViewChild(v.View, path)
}

func (v *NativeView) GetNative() *NativeView {
	return v
}

func (v *NativeView) SetRect(rect zgeo.Rect) {
	// if rect.Pos.Y > 3000 || rect.Size.H > 3000 {
	// 	zlog.Error(nil, "strange rect for view:", v.Hierarchy(), rect, zlog.GetCallingStackString())
	// }
	// if v.ObjectName() == "v2" {
	// zlog.Info("NV Rect", v.ObjectName(), rect, zlog.GetCallingStackString())
	// }
	rect = rect.ExpandedToInt()
	SetElementRect(v.Element, rect)
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

func SetElementRect(e js.Value, rect zgeo.Rect) {
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
		zlog.Error(nil, "parse empty Coord:", style.Get("left"), style.Get("right"), sw, sh, v.Hierarchy(), zlog.GetCallingStackString())
	}

	return zgeo.RectMake(0, 0, w, h)
}

func (v *NativeView) SetLocalRect(rect zgeo.Rect) {
	zlog.Fatal(nil, "NOT IMPLEMENTED")
}

func (v *NativeView) ObjectName() string {
	return v.JSGet("oname").String()
}

var idCount int

func (v *NativeView) SetObjectName(name string) {
	v.JSSet("oname", name)
	idCount++
	v.JSSet("id", fmt.Sprintf("%s-%d", name, idCount))
}

func (v *NativeView) SetColor(c zgeo.Color) {
	//	v.style().Set("color", zdom.MakeRGBAString(c))
	str := zdom.MakeRGBAString(c)
	v.style().Set("color", str)

}

func (v *NativeView) SetCursor(cursor CursorType) {
	v.style().Set("cursor", string(cursor))
}

func (v *NativeView) Color() zgeo.Color {
	str := v.style().Get("color").String()

	col := zgeo.ColorFromString(str)
	return col

}

func (v *NativeView) style() js.Value {
	return v.JSGet("style")
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
	v.style().Set("backgroundColor", zdom.MakeRGBAString(c))
}

func (v *NativeView) BGColor() zgeo.Color {
	str := v.style().Get("backgroundColor").String()
	return zgeo.ColorFromString(str)
}

func (v *NativeView) SetCorners(radius float64, align zgeo.Alignment) {
	style := v.style()
	for _, a := range []zgeo.Alignment{zgeo.TopLeft, zgeo.TopRight, zgeo.BottomLeft, zgeo.BottomRight} {
		if align&a != a {
			continue
		}
		srad := fmt.Sprintf("%dpx", int(radius))
		pos := "border-" + strings.Replace(a.String(), "|", "-", -1) + "-radius"
		style.Set("-moz-"+pos, srad)
		style.Set("-webkit-"+pos, srad)
		style.Set(pos, srad)
	}
}

func (v *NativeView) SetSelectable(on bool) {
	val := "none"
	if on {
		val = "all"
	}
	v.style().Set("user-select", val)
	v.style().Set("-webkit-user-select", val)
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
		str = fmt.Sprintf("%dpx solid %s", int(width), zdom.MakeRGBAString(c))
	}
	v.style().Set("border", str)
}

func (v *NativeView) Scale(scale float64) {
}

func (v *NativeView) Rotate(deg float64) {
	rot := fmt.Sprintf("rotate(%ddeg)", int(deg))
	v.style().Set("webkitTransform", rot)
}

func (v *NativeView) GetScale() float64 {
	return 1
}

func (v *NativeView) Show(show bool) {
	str := "hidden"
	if show {
		str = "inherit" //visible"
	}
	v.style().Set("visibility", str)
}

func (v *NativeView) IsShown() bool {
	return v.style().Get("visibility").String() == "visible"
}

func (v *NativeView) SetUsable(usable bool) {
	v.JSSet("disabled", !usable)
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
}

func (v *NativeView) Usable() bool {
	dis := v.Element.Get("disabled")
	if dis.IsUndefined() {
		return true
	}
	return !dis.Bool()
}

func (v *NativeView) IsFocused() bool {
	f := zdom.DocumentJS.Call("hasFocus")
	return f.Equal(v.Element)
}

func (v *NativeView) Focus(focus bool) {
	v.JSCall("focus")
}

func (v *NativeView) SetCanFocus(can bool) {
	val := "-1"
	if can {
		val = "0"
	}
	v.JSSet("tabindex", val)
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
	zlog.Assert(v.parent != nil)
	v.parent.RemoveChild(v.View)
	v.parent = nil
}

func (v *NativeView) SetFont(font *zgeo.Font) {
	// zlog.Debug("font:", v.ObjectName(), font.Style&FontStyleItalic, font.Style&FontStyleBold)
	cssStyle := v.style()
	cssStyle.Set("font-style", string(font.Style&zgeo.FontStyleItalic))
	cssStyle.Set("font-weight", (font.Style & zgeo.FontStyleBold).String())
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

func (v *NativeView) Font() *zgeo.Font {
	cssStyle := v.style()
	fstyle := zgeo.FontStyleNormal
	name := cssStyle.Get("font-family").String()
	if cssStyle.Get("font-weight").String() == "bold" {
		fstyle |= zgeo.FontStyleBold
	}
	if cssStyle.Get("font-style").String() == "italic" {
		fstyle |= zgeo.FontStyleItalic
	}
	ss := cssStyle.Get("font-size")
	size := v.parseElementCoord(ss)

	return zgeo.FontNew(name, size, fstyle)
}

func (v *NativeView) SetText(text string) {
	//		zlog.Info("NV SETTEXT", v.ObjectName(), zlog.GetCallingStackString())
	v.JSSet("innerText", text)
}

func (v *NativeView) Text() string {
	text := v.JSGet("innerText").String()
	return text
}

func (v *NativeView) AddChild(child View, index int) {
	n := ViewGetNative(child)
	if n == nil {
		zlog.Fatal(nil, "NativeView AddChild child not native", v.Hierarchy(), child.ObjectName())
	}
	n.parent = v
	// if child.ObjectName() == "tab-separator" {
	// 	zlog.Info("Call On Add:", n.Hierarchy(), len(n.doOnAdd))
	// }
	n.PerformAddRemoveFuncs(true)
	if index != -1 {
		nodes := n.parent.JSGet("childNodes").Length()
		// zlog.Info("NS InsertChild:", v.ObjectName(), child.ObjectName(), index, nodes)
		if nodes == 0 {
			v.JSCall("appendChild", n.Element)
		} else {
			v.JSCall("insertBefore", n.Element, v.JSGet("firstChild"))
		}
	} else {
		v.JSCall("appendChild", n.Element)
	}
	n.style().Set("zIndex", 100)
	// for _, p := range n.AllParents() {
	// 	p.allChildrenPresented = false
	// }
	if v.Presented {
		// zlog.Info("Set Presented For New Add:", n.Hierarchy(), len(n.doOnReady))
		PresentViewCallReady(child, true)
		PresentViewCallReady(child, false)
	}
}

func (v *NativeView) ReplaceChild(child, with View) {
	v.AddChild(with, -1) // needs to preserve index, which isn't really supported in AddChild yet anyway
	with.SetRect(child.Rect())
	// zlog.Info("RemoveFromParent:", v.ObjectName(), reflect.ValueOf(v.View).Type())
	v.RemoveChild(child)
	ExposeView(with)
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
	nv := ViewGetNative(child)
	if nv == nil {
		panic("NativeView AddChild child not native")
	}
	win := v.GetWindow()
	// zlog.Info("RemoveChild:", v.ObjectName(), child.ObjectName(), reflect.ValueOf(child).Type())

	win.removeKeyPressHandlerViews(child)
	nv.PerformAddRemoveFuncs(false)
	nv.Element = v.JSCall("removeChild", nv.Element) // we need to set it since  it might be accessed for ObjectName etc still in collapsed containers
	//!! nv.parent = nil if we don't do this, we can still uncollapse child in container without having to remember comtainer. Testing.
}

func nativeElementSetDropShadow(e js.Value, shadow zgeo.DropShadow) {
	str := fmt.Sprintf("%dpx %dpx %dpx %s", int(shadow.Delta.W), int(shadow.Delta.H), int(shadow.Blur), zdom.MakeRGBAString(shadow.Color))
	e.Get("style").Set("boxShadow", str)
}

func (v *NativeView) SetDropShadow(shadow zgeo.DropShadow) {
	nativeElementSetDropShadow(v.Element, shadow)
}

func (v *NativeView) SetToolTip(str string) {
	v.JSSet("title", str)
}

func (v *NativeView) AbsoluteRect() zgeo.Rect {
	r := v.Element.Call("getBoundingClientRect")
	x := r.Get("left").Float() // x
	y := r.Get("top").Float()  // y
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
	v.JSSet("onmousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if handler == nil {
			return nil
		}
		mouseStartTime = time.Now()
		mouseStartPos = getMousePos(args[0])
		return nil
	}))
	v.JSSet("onmousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
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

func (v *NativeView) JSCall(method string, args ...interface{}) js.Value {
	return v.Element.Call(method, args...)
}

func (v *NativeView) JSSet(property string, value interface{}) {
	v.Element.Set(property, value)
}

func (v *NativeView) JSGet(property string) js.Value {
	return v.Element.Get(property)
}

func (v *NativeView) SetScrollHandler(handler func(pos zgeo.Pos)) {
	v.JSSet("onscroll", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if handler != nil {
			y := v.JSGet("scrollTop").Float()
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
	w := root.JSGet("ownerDocument").Get("defaultView")
	// zlog.Info("NV.GetWindow:", w, root.ObjectName())
	return windowsFindForElement(w)
}

func dragEvent(event js.Value, dtype DragType, handler func(dtype DragType, data []byte, name string, pos zgeo.Pos) bool) {
	dt := event.Get("dataTransfer")
	item := dt.Get("items").Index(0)
	mime := item.Get("type").String()

	var pos zgeo.Pos
	pos.X = event.Get("offsetX").Float()
	pos.Y = event.Get("offsetY").Float()

	if dtype != DragDrop {
		handler(dtype, nil, mime, pos)
	} else {
		item.Call("getAsString", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			val := []byte(args[0].String())
			handler(DragDrop, val, mime, pos)
			return nil
		}))
	}
	event.Call("preventDefault")
}

func (v *NativeView) SetDraggable(getData func() (data string, mime string)) {
	// https://www.digitalocean.com/community/tutorials/js-drag-and-drop-vanilla-js
	v.JSSet("draggable", true)
	v.JSSet("ondragstart", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		data, mime := getData()
		//		array := js.Global().Get("Uint8Array").New(len(data))
		//		js.CopyBytesToJS(array, []byte(data))
		// zlog.Info("Dtrans:", mime, array.Length())
		// zlog.Info("Dtrans:", mime, data)
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

var dragEnterView *NativeView

func (v *NativeView) SetPointerDropHandler(handler func(dtype DragType, data []byte, name string, pos zgeo.Pos) bool) {
	if zlog.IsInTests {
		return
	}
	//	v.JSSet("className", "zdropper")
	v.JSSet("ondragenter", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if dragEnterView == nil {
			dragEvent(args[0], DragEnter, handler)
		}
		// zlog.Info("ondragenter:", v.ObjectName())
		dragEnterView = v
		return nil
	}))
	v.JSSet("ondragleave", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// zlog.Info("ondragleave1:", dragEnterView != nil, dragEnterView == v, v.ObjectName())
		// zlog.Info("ondragleave:", v.ObjectName(), dragEnterView != v)
		// if dragEnterView != v {
		// 	return nil
		// }
		dragEnterView = nil
		dragEvent(args[0], DragLeave, handler)
		return nil
	}))
	v.JSSet("ondragover", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		dragEvent(args[0], DragOver, handler)
		return nil
	}))
	v.JSSet("ondrop", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		dt := event.Get("dataTransfer")
		files := dt.Get("files")
		dragEnterView = nil

		if files.Length() == 0 {
			dragEvent(event, DragDrop, handler)
			return nil
		}
		file := files.Index(0)
		jsFileToGo(file, func(data []byte, name string) {
			var pos zgeo.Pos
			pos.X = event.Get("offsetX").Float()
			pos.Y = event.Get("offsetY").Float()
			// zlog.Info("Drop offset:", pos)
			if handler(DragDropFile, data, name, pos) {
				event.Call("preventDefault")
			}
		})
		return nil
	}))
}

var lastUploadClick time.Time

func (v *NativeView) SetUploader(got func(data []byte, name string)) {
	e := zdom.DocumentJS.Call("createElement", "input")
	e.Set("type", "file")
	e.Set("style", "opacity: 0.0; position: absolute; top: 0; left: 0; bottom: 0; right: 0; width: 100%; height:100%;")
	// e.Set("accept", "*/*")

	e.Set("onchange", js.FuncOf(func(this js.Value, args []js.Value) interface{} { // was onchange????
		// zlog.Info("uploader on change")
		files := e.Get("files")
		if files.Length() > 0 {
			file := files.Index(0)
			jsFileToGo(file, got)
		}
		return nil
	}))

	// zlog.Info("NV SetUploader", v.ObjectName())
	v.JSSet("onclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if time.Since(lastUploadClick) < time.Millisecond*100 { // e.Call("click") below causes v onclick to be re-called, bail + preventDefault important or it doesn't work (error started on Tor's M1 Mac Pro)
			args[0].Call("preventDefault")
			// zlog.Info("cancel clickthru")
			return nil
		}
		lastUploadClick = time.Now()
		// zlog.Info("uploader clickthru")
		e.Call("click")
		return nil
	}))

	v.JSCall("appendChild", e)
}

func (v *NativeView) SetPointerEnterHandler(moves bool, handler func(pos zgeo.Pos, inside zbool.BoolInd)) {
	if zlog.IsInTests {
		return
	}
	v.JSSet("onmouseenter", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// zlog.Info("Mouse enter", v.ObjectName())
		handler(getMousePos(args[0]), zbool.True)
		if moves {
			v.GetWindow().element.Set("onmousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				x := args[0].Get("offsetX").Float()
				y := args[0].Get("offsetY").Float()
				handler(zgeo.Pos{x, y}, zbool.Unknown)
				return nil
			}))
		}
		return nil
	}))
	v.JSSet("onmouseleave", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// zlog.Info("Mouse leave", v.ObjectName())
		handler(getMousePos(args[0]), zbool.False)
		if moves {
			v.GetWindow().element.Set("onmousemove", nil)
		}
		return nil
	}))
}

func (v *NativeView) SetKeyHandler(handler func(key zkeyboard.Key, mods zkeyboard.Modifier) bool) {
	zkeyboard.SetKeyHandler(v.Element, func(key zkeyboard.Key, mods zkeyboard.Modifier) bool {
		return handler(key, mods)
	})
}

func (v *NativeView) SetOnInputHandler(handler func()) {
	v.Element.Set("oninput", js.FuncOf(func(js.Value, []js.Value) interface{} {
		handler()
		return nil
	}))
}

var movingPos *zgeo.Pos

func (v *NativeView) SetUpDownMovedHandler(handler func(pos zgeo.Pos, down zbool.BoolInd)) {
	const minDiff = 10.0
	v.JSSet("onmousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		m := getMousePos(e)
		movingPos = &m
		handler(*movingPos, zbool.True)
		e.Call("preventDefault")
		v.GetWindow().element.Set("onmousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if movingPos != nil {
				pos := getMousePos(args[0])
				pos2 := pos.Minus(*movingPos)
				handler(pos2, zbool.Unknown)
			}
			return nil
		}))
		return nil
	}))
	win := v.GetWindow()
	if win == nil {
		return
	}
	win.element.Set("onmouseup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// zlog.Info("MOUSEUP")
		movingPos = nil
		v.GetWindow().element.Set("onmousemove", nil)
		pos := getMousePos(args[0])
		pos = pos.Minus(v.AbsoluteRect().Pos)
		handler(pos, zbool.False)
		return nil
	}))
}

func (v *NativeView) GetFocusedView() (found View) {
	ct := v.View.(ContainerType)
	e := zdom.DocumentJS.Get("activeElement")
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
		id := n.JSGet("id").String()
		if id == foundID {
			found = n.View
			return false
		}
		return true
	})
	return found
}

func (v *NativeView) SetDownloader(surl, name string) {
	if name == "" {
		_, name = path.Split(surl)
	}
	// v.JSSet("download", name)
	v.JSSet("href", surl)
}

func (v *NativeView) MakeLink(surl, name string) {
	stype := strings.ToLower(v.Element.Get("nodeName").String())
	zlog.Assert(stype == "a", stype)
	v.JSSet("download", name)
	v.JSSet("href", surl)
}

func (v *NativeView) SetTilePath(spath string) {
	spath2 := zimage.MakeImagePathWithAddedScale(spath, 2)
	format := `-webkit-image-set(url("%s") 1x, url("%s") 2x)`
	s := fmt.Sprintf(format, spath, spath2)
	v.style().Set("backgroundImage", s)
}

// SetHandleExposed sets a handler for v that is called with intersectsViewport=true, when v becomes visible.
// It calls with intersectsViewport=false when it becomes inivisble.
// In this js implementation is uses the IntersectionObserver, and removed the observation on view removal.
// Note that it has to be called AFTER the window v will be in is opened, so v.GetWindow() gives correct window observe with.
func (v *NativeView) SetHandleExposed(handle func(intersectsViewport bool)) {
	f := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		entries := args[0]
		for i := 0; i < entries.Length(); i++ {
			e := entries.Index(i)
			inter := e.Get("isIntersecting").Bool()
			handle(inter)
		}
		return nil
	})
	observer := v.GetWindow().element.Get("IntersectionObserver").New(f) //, opts)
	observer.Call("observe", v.Element)
	v.AddOnRemoveFunc(func() {
		observer.Call("disconnect")
		handle(false)
	})
}

func (v *NativeView) hasElement(e js.Value) (got *NativeView) {
	if v.Element.Equal(e) {
		return v
	}
	zlog.Info("FIND:", reflect.ValueOf(v).Type(), v.Hierarchy())
	ViewRangeChildren(v, true, true, func(view View) bool {
		nv, _ := (view.(*NativeView))
		fmt.Printf("FIND: %+v\n", nv.Element)
		if nv != nil && nv.Element.Equal(e) {
			got = nv
			return false
		}
		return true
	})
	return
}

func AddTextNode(e *NativeView, text string) {
	textNode := zdom.DocumentJS.Call("createTextNode", text)
	e.JSCall("appendChild", textNode)
	//	js.Value(*e).Call("appendChild", textNode)
}

func AddView(parent, child *NativeView) {
	parent.JSCall("appendChild", child.Element)
}
