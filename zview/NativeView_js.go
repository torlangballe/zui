package zview

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/torlangballe/zui/zcursor"
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zstyle"
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

var (
	dragEnterView   *NativeView
	idCount         int
	mouseStartPos   zgeo.Pos
	mouseStartTime  time.Time
	lastUploadClick time.Time
	movingPos       *zgeo.Pos
	// moverListeners  = map[*NativeView]js.Value{}

	SetPresentReadyFunc            func(v View, beforeWindow bool)
	RemoveKeyPressHandlerViewsFunc func(v View)
	SkipEnterHandler               bool
)

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
	// zlog.Info("NS Parent:", v.Hierarchy(), v.parent != nil)
	if v.parent != nil && v.parent.View != nil {
		// zlog.Info("PAR:", reflect.ValueOf(v.parent.View).Type())
		nv := v.parent.View.Native()
		// zlog.Info("PAR2:", reflect.ValueOf(nv.View).Type())
		if nv != nil {
			return nv
		}
	}
	return v.parent
}

func (v *NativeView) Child(path string) View {
	return ChildOfViewFunc(v.View, path)
}

func (v *NativeView) Native() *NativeView {
	return v
}

func (v *NativeView) SetRect(rect zgeo.Rect) {
	// if rect.Pos.Y > 3000 || rect.Size.H > 3000 {
	// 	zlog.Error(nil, "strange rect for view:", v.Hierarchy(), rect, zlog.GetCallingStackString())
	// }
	rect = rect.ExpandedToInt()
	SetElementRect(v.Element, rect)
}

func (v *NativeView) HasSize() bool {
	if v.JSStyle().Get("width").String() == "" {
		return false
	}
	if v.JSStyle().Get("height").String() == "" {
		return false
	}
	return true
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
	return 0
}

func (v *NativeView) Rect() zgeo.Rect {
	var pos zgeo.Pos
	style := v.JSStyle()
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
	style := v.JSStyle()
	sw := style.Get("width")
	sh := style.Get("height")
	if sw.String() != "" {
		h = v.parseElementCoord(sh)
		w = v.parseElementCoord(sw)
	} else if v.Presented {
		zlog.Error(nil, "parse empty Coord:", style.Get("left"), style.Get("right"), sw, sh, v.Hierarchy(), zlog.CallingStackString())
	}

	return zgeo.RectMake(0, 0, w, h)
}

func (v *NativeView) SetLocalRect(rect zgeo.Rect) {
	zlog.Fatal(nil, "NOT IMPLEMENTED")
}

func (v *NativeView) ObjectName() string {
	return v.JSGet("oname").String()
}

func (v *NativeView) SetObjectName(name string) {
	v.JSSet("oname", name)
	idCount++
	v.JSSet("id", fmt.Sprintf("%s-%d", name, idCount))
}

func (v *NativeView) SetColor(c zgeo.Color) {
	//	v.JSStyle().Set("color", zdom.MakeRGBAString(c))
	str := zdom.MakeRGBAString(c)
	v.JSStyle().Set("color", str)

}

func (v *NativeView) SetCursor(cursor zcursor.Type) {
	v.JSStyle().Set("cursor", string(cursor))
}

func (v *NativeView) Color() zgeo.Color {
	str := v.JSStyle().Get("color").String()

	col := zgeo.ColorFromString(str)
	return col

}

func (v *NativeView) JSStyle() js.Value {
	return v.JSGet("style")
}

func (v *NativeView) SetJSStyle(key, value string) {
	v.JSStyle().Set(key, value)
}

func (v *NativeView) SetAlpha(alpha float32) {
	v.transparency = 1 - alpha
	//	v.JSStyle().Set("alpha", alpha)
	v.JSStyle().Set("opacity", alpha)
}

func (v *NativeView) Alpha() float32 {
	return 1 - v.transparency
}

func (v *NativeView) SetBGColor(c zgeo.Color) {
	v.JSStyle().Set("backgroundColor", zdom.MakeRGBAString(c))
}

func (v *NativeView) BGColor() zgeo.Color {
	str := v.JSStyle().Get("backgroundColor").String()
	return zgeo.ColorFromString(str)
}

func (v *NativeView) SetCorners(radius float64, align zgeo.Alignment) {
	style := v.JSStyle()
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
	v.JSStyle().Set("user-select", val)
	v.JSStyle().Set("-webkit-user-select", val)
}

func (v *NativeView) SetCorner(radius float64) {
	style := v.JSStyle()
	s := fmt.Sprintf("%dpx", int(radius))
	style.Set("-moz-border-radius", s)
	style.Set("-webkit-border-radius", s)
	style.Set("border-radius", s)
}

func (v *NativeView) SetStroke(width float64, c zgeo.Color, inset bool) {
	str := "none"
	if width != 0 {
		str = fmt.Sprintf("%dpx solid %s", int(width), zdom.MakeRGBAString(c))
	}
	style := v.JSStyle()
	style.Set("border", str)
	str = "content-box"
	if inset {
		str = "border-box"
	}
	style.Set("boxSizing", str)
}

func (v *NativeView) SetStrokeSide(width float64, c zgeo.Color, a zgeo.Alignment, inset bool) {
	str := "none"
	if width != 0 {
		str = fmt.Sprintf("%dpx solid %s", int(width), zdom.MakeRGBAString(c))
	}
	style := v.JSStyle()
	if a&zgeo.Left != 0 {
		style.Set("borderLeft", str)
	}
	if a&zgeo.Right != 0 {
		style.Set("borderRight", str)
	}
	if a&zgeo.Top != 0 {
		style.Set("borderTop", str)
	}
	if a&zgeo.Bottom != 0 {
		style.Set("borderBottom", str)
	}
	str = "content-box"
	if inset {
		str = "border-box"
	}
	style.Set("boxSizing", str)
}

func (v *NativeView) SetOutline(width float64, c zgeo.Color, offset float64) {
	str := "none"
	if width != 0 {
		str = fmt.Sprintf("%dpx solid %s", int(width), zdom.MakeRGBAString(c))
	}
	v.JSStyle().Set("outline", str)
	v.JSStyle().Set("outline-offset", fmt.Sprintf("%gpx", offset))
}

func (v *NativeView) Scale(scale float64) {
}

func (v *NativeView) Rotate(deg float64) {
	rot := fmt.Sprintf("rotate(%ddeg)", int(deg))
	v.JSStyle().Set("webkitTransform", rot)
}

func (v *NativeView) GetScale() float64 {
	return 1
}

func (v *NativeView) Show(show bool) {
	// if strings.HasSuffix(v.Hierarchy(), "activity.png") {
	// 	zlog.Info("Show", v.Hierarchy(), show, zlog.CallingStackString())
	// }
	str := "hidden"
	if show {
		str = "inherit" //visible"
	}
	v.JSStyle().Set("visibility", str)
}

func (v *NativeView) IsShown() bool {
	return v.JSStyle().Get("visibility").String() == "visible" // or inherit???
}

func (v *NativeView) Usable() bool {
	dis := v.Element.Get("disabled")
	// zlog.Info("Usable:", v.ObjectName(), dis)
	if dis.IsUndefined() {
		return true
	}
	return !dis.Bool()
}

func (v *NativeView) SetUsable(usable bool) {
	v.JSSet("disabled", !usable)
	style := v.JSStyle()
	// zlog.Info("SetUsable:", v.ObjectName(), v.Element.Get("disabled"))
	var alpha float32 = 0.4
	if usable {
		alpha = 1 - v.transparency
	}
	// str := "none"
	// if usable {
	// 	str = "auto"
	// }
	//style.Set("pointer-events", str)
	style.Set("opacity", alpha)
}

// func (v *NativeView) SetInteractive(interactive bool) {
// 	str := "none"
// 	if interactive {
// 		str = "auto"
// 	}
// 	// fmt.Printf("NV SetInteractive: %p %s %v %s\n", v, v.ObjectName(), interactive, str)
// 	v.JSStyle().Set("pointer-events", str)
// }

// func (v *NativeView) Interactive() bool {
// 	inter := v.JSStyle().Get("pointer-events")
// 	// fmt.Printf("Inter? %p %s %v\n", v, v.ObjectName(), inter)
// 	if inter.IsUndefined() || inter.String() != "none" {
// 		return true
// 	}
// 	return false
// }

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
	return v.HierarchyToRoot(nil)
}

func (v *NativeView) HierarchyToRoot(root *NativeView) string {
	var str string
	var found bool
	if root == nil {
		found = true
		str = "/"
	}
	for _, p := range v.AllParents() {
		if !found {
			found = (root == p)
		}
		if found {
			str += p.ObjectName() + "/"
		}
	}
	str += v.ObjectName()
	return str
}

func (v *NativeView) RemoveFromParent() {
	zlog.Assert(v.parent != nil, v.Hierarchy())
	v.parent.RemoveChild(v.View)
	v.parent = nil
}

func (v *NativeView) SetFont(font *zgeo.Font) {
	// zlog.Debug("font:", v.ObjectName(), font.Style&FontStyleItalic, font.Style&FontStyleBold)
	cssStyle := v.JSStyle()
	if font.Style == zgeo.FontStyleUndef {
		font.Style = zgeo.FontStyleNormal
	}
	cssStyle.Set("font-style", string(font.Style&zgeo.FontStyleItalic))
	str := (font.Style & (zgeo.FontStyleBold | zgeo.FontStyleNormal)).String()
	cssStyle.Set("font-weight", str)
	// zlog.Info("NS font-weight", v.Hierarchy(), str)
	cssStyle.Set("font-family", font.Name)
	cssStyle.Set("font-size", fmt.Sprintf("%gpx", font.Size))

	// zlog.Info("NS font-style", v.Hierarchy(), font.Style&zgeo.FontStyleItalic)

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
	cssStyle := v.JSStyle()
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
	// zlog.Info("NV SetText:", v.Hierarchy(), text, zlog.CallingStackString())
	v.JSSet("innerText", text)
}

func (v *NativeView) Text() string {
	text := v.JSGet("innerText").String()
	return text
}

func (v *NativeView) AddChild(child View, index int) {
	n := child.Native()
	if n == nil {
		zlog.Fatal(nil, "NativeView AddChild child not native", v.Hierarchy(), child.ObjectName())
	}
	n.parent = v
	// if child.ObjectName() == "tab-separator" {
	// 	zlog.Info("Call On Add:", n.Hierarchy(), len(n.DoOnAdd))
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
	n.JSStyle().Set("zIndex", 100)
	// for _, p := range n.AllParents() {
	// 	p.allChildrenPresented = false
	// }
	// zlog.Info("ADDCHILD:", v.ObjectName(), child.ObjectName(), v.Rect())
	if v.Presented {
		SetPresentReadyFunc(child, true)
		SetPresentReadyFunc(child, false)
	}
}

func (v *NativeView) ReplaceChild(child, with View) {
	var focusedPath string
	focused := v.GetFocusedChildView()
	if focused != nil {
		focusedPath = focused.HierarchyToRoot(child.Native())
		zstr.HeadUntilWithRest(focusedPath, "/", &focusedPath) // remove first path component, which is v's
	}
	v.AddChild(with, -1) // needs to preserve index, which isn't really supported in AddChild yet anyway
	with.SetRect(child.Rect())
	v.RemoveChild(child)
	if focusedPath != "" {
		f := ChildOfViewFunc(with, focusedPath) // use v.View here to get proper underlying container type in ChildOfViewFunc
		zlog.Assert(f != nil)
		f.Native().Focus(true)
		// zlog.Info("ReplaceChild old focus:", focusedPath, f.Native().Hierarchy())
	}
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
	v.JSStyle().Set("zIndex", index)
}

func (v *NativeView) RemoveChild(child View) {
	// zlog.Info("RemoveChild:", child.Native().Hierarchy(), len(v.DoOnRemove))
	nv := child.Native()
	if nv == nil {
		panic("NativeView AddChild child not native")
	}
	RemoveKeyPressHandlerViewsFunc(child)
	nv.PerformAddRemoveFuncs(false)
	nv.Element = v.JSCall("removeChild", nv.Element) // we need to set it since  it might be accessed for ObjectName etc still in collapsed containers
	//!! nv.parent = nil if we don't do this, we can still uncollapse child in container without having to remember comtainer. Testing.
}

func nativeElementSetDropShadow(e js.Value, shadow zstyle.DropShadow) {
	str := fmt.Sprintf("%dpx %dpx %dpx %s", int(shadow.Delta.W), int(shadow.Delta.H), int(shadow.Blur), zdom.MakeRGBAString(shadow.Color))
	e.Get("style").Set("boxShadow", str)
}

func (v *NativeView) SetDropShadow(shadow zstyle.DropShadow) {
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
	v.JSStyle().Set("overflow", str)
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
	v.JSStyle().Set("transform", fmt.Sprintf("rotate(%fdeg)", deg))
}

func dragEvent(event js.Value, dtype DragType, handler func(dtype DragType, data []byte, name string, pos zgeo.Pos) bool) bool {
	var mime string
	var handled bool
	dt := event.Get("dataTransfer")
	items := dt.Get("items")
	var pos zgeo.Pos
	pos.X = event.Get("offsetX").Float()
	pos.Y = event.Get("offsetY").Float()
	if dtype != DragDrop || items.Length() == 0 {
		event.Call("preventDefault")
		return handler(dtype, nil, "", pos)
	}
	item := dt.Get("items").Index(0)
	mime = item.Get("type").String()
	item.Call("getAsString", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		val := []byte(args[0].String())
		handled = handler(DragDrop, val, mime, pos)
		return nil
	}))
	event.Call("preventDefault")
	return handled
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

func jsFileToGo(file js.Value, got func(data []byte, name string), progress func(p float64)) {
	// TODO progress: https://developer.mozilla.org/en-US/docs/Web/API/FileReader/progress_event
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
		name := file.Get("name").String()
		event.Call("preventDefault")
		if handler(DragDropFilePreflight, nil, name, zgeo.Pos{}) {
			return nil
		}
		// zlog.Info("FileProcessing")
		jsFileToGo(file, func(data []byte, name string) {
			var pos zgeo.Pos
			pos.X = event.Get("offsetX").Float()
			pos.Y = event.Get("offsetY").Float()
			// zlog.Info("Drop offset:", pos)
			// zlog.Info("nv.DragDropFile")
			handler(DragDropFile, data, name, pos)
		}, nil)
		return nil
	}))
}

func (v *NativeView) SetUploader(got func(data []byte, name string), skip func(name string) bool, progress func(p float64)) {
	e := zdom.DocumentJS.Call("createElement", "input")
	e.Set("type", "file")
	e.Set("style", "opacity: 0.0; position: absolute; top: 0; left: 0; bottom: 0; right: 0; width: 100%; height:100%;")
	// e.Set("accept", "*/*")

	e.Set("onchange", js.FuncOf(func(this js.Value, args []js.Value) interface{} { // was onchange????
		// zlog.Info("uploader on change")
		files := e.Get("files")
		if files.Length() > 0 {
			file := files.Index(0)
			name := file.Get("name").String()
			if skip != nil && skip(name) {
				return nil
			}
			jsFileToGo(file, got, progress)
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

func (v *NativeView) HasPressedDownHandler() bool {
	return v.JSGet("onmousedown").IsNull()
}

func (v *NativeView) SetPressedDownHandler(handler func()) {
	// zlog.Info("NV.SetPressedDownHandler:", v.Hierarchy())
	v.JSSet("className", "widget")
	v.JSCall("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		v.SetStateOnDownPress(e)
		e.Call("stopPropagation")
		zlog.Assert(len(args) > 0)
		handler()
		return nil
	}))
}

func getMousePosRelative(v *NativeView, e js.Value) zgeo.Pos {
	pos := getMousePos(e)
	return pos.Minus(v.AbsoluteRect().Pos)
}

func (v *NativeView) SetPointerEnterHandler(moves bool, handler func(pos zgeo.Pos, inside zbool.BoolInd)) {
	if zlog.IsInTests {
		return
	}
	v.JSSet("onmouseenter", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if SkipEnterHandler {
			return nil
		}
		handler(getMousePosRelative(v, args[0]), zbool.True)
		if moves {
			// we := v.GetWindowElement()
			v.JSSet("onmousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				handler(getMousePosRelative(v, args[0]), zbool.Unknown)
				return nil
			}))
			// listener := we.Call("addEventListener", "mousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// 	handler(getMousePosRelative(v, args[0]), zbool.Unknown)
			// 	return nil
			// }))
			// moverListeners[v] = listener
			// v.AddOnRemoveFunc(func() {
			// 	delete(moverListeners, v)
			// })
		}
		return nil
	}))
	v.JSSet("onmouseleave", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if SkipEnterHandler {
			return nil
		}
		handler(getMousePosRelative(v, args[0]), zbool.False)
		if moves {
			v.JSSet("onmousemove", nil)
			// 	listener, got := moverListeners[v]
			// 	zlog.Info("Remove mover:", got, v.Hierarchy())
			// 	if got {
			// 		we := v.GetWindowElement()
			// 		we.Call("removeEventListener", "mousemove", listener)
			// 		delete(moverListeners, v)
			// 	}
		}
		return nil
	}))
}

func (v *NativeView) SetKeyHandler(handler func(key zkeyboard.Key, mods zkeyboard.Modifier) bool) {
	v.JSSet("onkeyup", js.FuncOf(func(val js.Value, args []js.Value) interface{} {
		if !v.GetWindowElement().Get("document").Call("hasFocus").Bool() {
			return nil
		}
		// zlog.Info("KeyUp")
		if handler != nil {
			event := args[0]
			key, mods := zkeyboard.GetKeyAndModsFromEvent(event)
			if handler(key, mods) {
				event.Call("stopPropagation")
			}
			// zlog.Info("KeyUp:", key, mods)
		}
		return nil
	}))
}

func (v *NativeView) SetOnInputHandler(handler func()) {
	v.Element.Set("oninput", js.FuncOf(func(js.Value, []js.Value) interface{} {
		handler()
		return nil
	}))
}

func (v *NativeView) SetStateOnDownPress(event js.Value) {
	var pos zgeo.Pos
	pos.X = event.Get("offsetX").Float()
	pos.Y = event.Get("offsetY").Float()
	LastPressedPos = pos //.Minus(v.AbsoluteRect().Pos)
	// zlog.Info("SetStateOnDownPress", v.Hierarchy(), pos.X, v.AbsoluteRect().Pos.X)
	_, zkeyboard.ModifiersAtPress = zkeyboard.GetKeyAndModsFromEvent(event)
}

var oldMouseMove js.Value

func (v *NativeView) SetPressUpDownMovedHandler(handler func(pos zgeo.Pos, down zbool.BoolInd) bool) {
	// zlog.Info("NV.SetPressUpDownMovedHandler:", v.Hierarchy())
	const minDiff = 10.0
	v.JSSet("onmousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// zlog.Info("NV.PressUpDownMovedHandler down:", v.Hierarchy())
		we := v.GetWindowElement()
		if we.IsUndefined() {
			return nil
		}
		e := args[0]
		pos := getMousePosRelative(v, e)
		// pos := getMousePos(e).Minus(v.AbsoluteRect().Pos)
		we.Set("onmouseup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// v.JSSet("onmouseup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			movingPos = nil
			v.GetWindowElement().Set("onmousemove", oldMouseMove)
			oldMouseMove = js.Null()
			v.GetWindowElement().Set("onmouseup", nil)
			if handler(pos, zbool.False) {
				e.Call("preventDefault")

			}
			return nil
		}))
		v.SetStateOnDownPress(e)
		// pos = getMousePos(e).Minus(v.AbsoluteRect().Pos)
		movingPos = &pos
		if handler(*movingPos, zbool.True) {
			e.Call("preventDefault")
		}
		oldMouseMove = v.GetWindowElement().Get("onmousemove")
		v.GetWindowElement().Set("onmousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if movingPos != nil {
				pos := getMousePosRelative(v, args[0])
				// zlog.Info("MM:", pos)
				if handler(pos, zbool.Unknown) {
					e.Call("preventDefault")
				}
			}
			return nil
		}))
		return nil
	}))
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
	v.JSStyle().Set("backgroundImage", s)
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
	e := v.GetWindowElement()
	// opts := map[string]interface{}{
	// 	"root": all[0].Element,
	// }
	observer := e.Get("IntersectionObserver").New(f) //, js.ValueOf(opts))
	observer.Call("observe", v.Element)
	v.AddOnRemoveFunc(func() {
		// zlog.Info("remove expose observer:", v.Hierarchy())
		observer.Call("disconnect")
		handle(false)
	})
}

func AddTextNode(v *NativeView, text string) {
	textNode := zdom.DocumentJS.Call("createTextNode", text)
	v.JSCall("appendChild", textNode)
	//	js.Value(*e).Call("appendChild", textNode)
}

func AddView(parent, child *NativeView) {
	parent.JSCall("appendChild", child.Element)
}

func (v *NativeView) GetWindowElement() js.Value {
	root := v
	all := v.AllParents()
	if len(all) > 1 {
		root = all[0]
	}
	w := root.JSGet("ownerDocument").Get("defaultView")
	// zlog.Info("NV.GetWindow:", w, root.ObjectName())
	//	return windowsFindForElement(w)
	return w
}

func (v *NativeView) SetStyling(style zstyle.Styling) {
	if style.DropShadow.Color.Valid {
		v.SetDropShadow(style.DropShadow)
	}
	if style.BGColor.Valid {
		v.SetBGColor(style.BGColor)
	}
	if style.Corner != -1 {
		v.SetCorner(style.Corner)
	}
	if style.StrokeColor.Valid {
		v.SetStroke(style.StrokeWidth, style.StrokeColor, style.StrokeIsInset.IsTrue())
	}
	if style.OutlineColor.Valid {
		v.SetOutline(style.OutlineWidth, style.OutlineColor, style.OutlineOffset)
	}
	if style.FGColor.Valid {
		v.SetColor(style.FGColor)
	}
	if style.Font.Name != "" {
		v.SetFont(&style.Font)
	}
	if !style.Margin.IsUndef() {
		m, _ := v.View.(Marginalizer)
		if m != nil {
			m.SetMargin(style.Margin)
		}
	}
}

func (root *NativeView) GetFocusedChildView() (found *NativeView) {
	e := zdom.DocumentJS.Get("activeElement")
	// zlog.Info("GetFocusedView:", e, e.IsUndefined())
	if e.IsUndefined() {
		return nil
	}
	foundID := e.Get("id").String()
	RangeAllVisibleChildrenFunc(root.View, func(view View) bool {
		n := view.Native()
		id := n.JSGet("id").String()
		if id == foundID {
			found = n
			return false
		}
		return true
	})
	return
}
