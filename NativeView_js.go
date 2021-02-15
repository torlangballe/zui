package zui

import (
	"fmt"
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
	View         View
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

func (v *NativeView) SetRect(rect zgeo.Rect) View {
	// zlog.Info("NV Rect", v.ObjectName(), rect, zlog.GetCallingStackString())
	rect = rect.ExpandedToInt()
	setElementRect(v.Element, rect)
	return v
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
	} else {
		zlog.Info("parse empty Coord:", v.ObjectName(), zlog.GetCallingStackString())
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

func (v *NativeView) SetObjectName(name string) View {
	v.setjs("oname", name)
	idCount++
	v.setjs("id", fmt.Sprintf("%s-%d", name, idCount))
	return v
}

func makeRGBAString(c zgeo.Color) string {
	if !c.Valid {
		return "initial"
	}
	return c.GetHex()
	//	rgba := c.GetRGBA()
	//	return fmt.Sprintf("rgba(%d,%d,%d,%g)", int(rgba.R*255), int(rgba.G*255), int(rgba.B*255), rgba.A)
}

func (v *NativeView) SetColor(c zgeo.Color) View {
	v.style().Set("color", makeRGBAString(c))
	return v
}

func (v *NativeView) Color() zgeo.Color {
	str := v.style().Get("color").String()
	return zgeo.ColorFromString(str)
}

func (v *NativeView) style() js.Value {
	return v.getjs("style")
}

func (v *NativeView) SetAlpha(alpha float32) View {
	v.transparency = 1 - alpha
	//	v.style().Set("alpha", alpha)
	v.style().Set("opacity", alpha)
	return v
}

func (v *NativeView) Alpha() float32 {
	return 1 - v.transparency
}

func (v *NativeView) SetBGColor(c zgeo.Color) View {
	// zlog.Info("SetBGColor:", v.ObjectName(), c)
	v.style().Set("backgroundColor", makeRGBAString(c))
	return v
}

func (v *NativeView) BGColor() zgeo.Color {
	str := v.style().Get("backgroundColor").String()
	zlog.Info("nv bgcolor", str)
	return zgeo.ColorFromString(str)
}

func (v *NativeView) SetCorner(radius float64) View {
	style := v.style()
	s := fmt.Sprintf("%dpx", int(radius))
	style.Set("-moz-border-radius", s)
	style.Set("-webkit-border-radius", s)
	style.Set("border-radius", s)
	return v
}

func (v *NativeView) SetStroke(width float64, c zgeo.Color) View {
	str := "none"
	if width != 0 {
		str = fmt.Sprintf("%dpx solid %s", int(width), makeRGBAString(c))
	}
	v.style().Set("border", str)
	return v
}

func (v *NativeView) Scale(scale float64) View {
	return v
}

func (v *NativeView) GetScale() float64 {
	return 1
}

func (v *NativeView) Show(show bool) View {
	str := "hidden"
	if show {
		str = "visible"
	}
	v.style().Set("visibility", str)
	return v
}

func (v *NativeView) IsShown() bool {
	return v.style().Get("visibility").String() == "visible"
}

func (v *NativeView) SetUsable(usable bool) View {
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
	return v
}

func (v *NativeView) Usable() bool {
	dis := v.Element.Get("disabled")
	if dis.IsUndefined() {
		return true
	}
	return !dis.Bool()
}

func (v *NativeView) IsFocused() bool {
	return false
}

func (v *NativeView) Focus(focus bool) View {
	zlog.Assert(focus)
	// zlog.Info("FOCUS:", v.ObjectName(), focus)
	v.call("focus")
	return v
}

func (v *NativeView) SetCanFocus(can bool) View {
	// zlog.Info("SetCanFocus:", v.ObjectName())
	val := "-1"
	if can {
		val = "0"
	}
	v.setjs("tabindex", val)
	return v
}

func (v *NativeView) Opaque(opaque bool) View {
	return v
}

func (v *NativeView) GetChild(path string) *NativeView {
	zlog.Fatal(nil, "not implemented")
	return nil
}

func (v *NativeView) DumpTree() {
}

func (v *NativeView) RemoveFromParent() {
	win := v.GetWindow()
	zlog.Info("RemoveFromParent:", v.ObjectName())
	win.removeKeyPressHandlerViews(v.View)

	v.StopStoppers()
	zlog.Assert(v.parent != nil)
	v.parent.RemoveChild(v)
}

func (v *NativeView) SetFont(font *Font) View {
	cssStyle := v.style()
	cssStyle.Set("font-style", string(font.Style&FontStyleItalic))
	// zlog.Debug("font:", v.ObjectName(), font.Style&FontStyleItalic, font.Style&FontStyleBold)
	cssStyle.Set("font-weight", (font.Style & FontStyleBold).String())
	cssStyle.Set("font-family", font.Name)
	cssStyle.Set("font-size", fmt.Sprintf("%gpx", font.Size))
	return v
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

func (v *NativeView) SetText(text string) View {
	//		zlog.Info("NV SETTEXT", v.ObjectName(), zlog.GetCallingStackString())
	v.setjs("innerText", text)
	return v
}

func (v *NativeView) Text() string {
	text := v.getjs("innerText").String()
	return text
}

func (v *NativeView) AddChild(child View, index int) {
	zlog.ErrorIf(index != -1, "non -1 insert not implemented")
	n := ViewGetNative(child)
	if n == nil {
		zlog.Fatal(nil, "NativeView AddChild child not native")
	}
	n.parent = v
	v.call("appendChild", n.Element)
	n.style().Set("zIndex", 100)
	for _, p := range n.AllParents() {
		p.allChildrenPresented = false
	}
}

func (v *NativeView) ReplaceChild(child, replace View) {
	v.AddChild(replace, -1) // needs to preserve index, which isn't really supported in AddChild yet anyway
	replace.SetRect(child.Rect())
	v.RemoveChild(child)
	et, _ := replace.(ExposableType)
	// zlog.Info("ReplaceChild:", et != nil, replace.ObjectName())
	if et != nil {
		et.Expose()
		et.drawIfExposed()
	}
}

func (v *NativeView) AllParents() (all []*NativeView) {
	for v.parent != nil {
		all = append(all, v.parent)
		v = v.parent
	}
	return
}

// func (v *NativeView) SetStokeWidth(width float64) *NativeView { // spelt wrong too
// 	return v
// }

func (v *NativeView) SetZIndex(index int) {
	v.style().Set("zIndex", index)
}

func (v *NativeView) RemoveChild(child View) {
	nv := ViewGetNative(child)
	if nv == nil {
		panic("NativeView AddChild child not native")
	}
	// zlog.Info("REMOVE CHILD:", child.ObjectName())
	nv.Element = v.call("removeChild", nv.Element) // we need to set it since  it might be accessed for ObjectName etc still in collapsed containers
	nv.parent = nil
}

func (v *NativeView) SetDropShadow(shadow zgeo.DropShadow) {
	str := fmt.Sprintf("%dpx %dpx %dpx %s", int(shadow.Delta.W), int(shadow.Delta.H), int(shadow.Blur), makeRGBAString(shadow.Color))
	v.style().Set("boxShadow", str)
}

func (v *NativeView) SetToolTip(str string) {
	v.setjs("title", str)
}

func (v *NativeView) GetAbsoluteRect() zgeo.Rect {
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

func (v *NativeView) SetPointerEnterHandler(handler func(inside bool)) {
	v.setjs("onmouseenter", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if handler != nil {
			handler(true)
		}
		return nil
	}))
	v.setjs("onmouseleave", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if handler != nil {
			handler(false)
		}
		return nil
	}))
}

func (v *NativeView) SetKeyHandler(handler func(view View, key KeyboardKey, mods KeyboardModifier) bool) {
	jsSetKeyHandler(v.Element, func(key KeyboardKey, mods KeyboardModifier) bool {
		return handler(v.View, key, mods)
	})
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
	ContainerTypeRangeChildren(ct, recursive, func(view View) bool {
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
