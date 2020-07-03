package zui

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

type NativeView struct {
	Element      js.Value
	View         View
	presented    bool
	transparency float32
	parent       *NativeView
}

func (v *NativeView) Parent() *NativeView {
	// e := v.get("parentElement")
	// if e.Type() == js.TypeUndefined || e.Type() == js.TypeNull {
	// 	return nil
	// }
	// //	zlog.Info("ParentElement:", v.ObjectName(), e)
	// n := &NativeView{}
	// n.Element = e
	// n.View = n
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
	// zlog.Info("NV Rect", v.ObjectName())
	setElementRect(v.Element, rect)
	return v
}

func (v *NativeView) Rect() zgeo.Rect {
	var pos zgeo.Pos
	style := v.style()
	// pos.X = v.Element.Get("offsetLeft").Float()
	// pos.Y = v.Element.Get("offsetTop").Float()
	pos.X = parseElementCoord(style.Get("left"))
	pos.Y = parseElementCoord(style.Get("top"))
	size := v.GetLocalRect().Size
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

func (v *NativeView) GetLocalRect() zgeo.Rect {
	var w, h float64
	style := v.style()
	sw := style.Get("width")
	sh := style.Get("height")
	if sw.String() != "" {
		h = parseElementCoord(sh)
		w = parseElementCoord(sw)
	} else {
		zlog.Info("parse empty Coord:", v.ObjectName(), zlog.GetCallingStackString())
	}

	return zgeo.RectMake(0, 0, w, h)
}

func (v *NativeView) LocalRect(rect zgeo.Rect) {

}

func (v *NativeView) ObjectName() string {
	return v.get("id").String()
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

func (n *NativeView) style() js.Value {
	return n.get("style")
}

func (v *NativeView) SetAlpha(alpha float32) View {
	v.transparency = 1 - alpha
	v.style().Set("alpha", alpha)
	return v
}

func (v *NativeView) Alpha() float32 {
	return 1 - v.transparency
}

func (v *NativeView) SetObjectName(name string) View {
	v.set("id", name)
	return v
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
	str := fmt.Sprintf("%dpx solid %s", int(width), makeRGBAString(c))
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
	v.set("disabled", !usable)
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
	return v
}

func (v *NativeView) CanFocus(can bool) View {
	v.set("tabindex", "0")
	return v
}

func (v *NativeView) Opaque(opaque bool) View {
	return v
}

func (v *NativeView) GetChild(path string) *NativeView {
	return nil
}

func (v *NativeView) DumpTree() {
}

func (v *NativeView) RemoveFromParent() {
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
	size := parseElementCoord(ss)

	return FontNew(name, size, fstyle)
}

func (v *NativeView) SetText(text string) View {
	//		zlog.Info("NV SETTEXT", v.ObjectName(), zlog.GetCallingStackString())
	v.set("innerHTML", text)
	return v
}

func (v *NativeView) Text() string {
	text := v.get("innerHTML").String()
	return text
}

func (v *NativeView) AddChild(child View, index int) {
	o, _ := child.(NativeViewOwner)
	if o == nil {
		panic("NativeView AddChild child not native")
	}
	n := o.GetNative()
	n.parent = v
	v.call("appendChild", n.Element)
	n.style().Set("zIndex", 100)
}

func (v *NativeView) SetStokeWidth(width float64) *NativeView {

	return v
}

func (v *NativeView) SetZIndex(index int) {
	v.style().Set("zIndex", index)
}

func (v *NativeView) RemoveChild(child View) {
	o, _ := child.(NativeViewOwner)
	if o == nil {
		panic("NativeView AddChild child not native")
	}
	v.call("removeChild", o.GetNative().Element)
}

func (v *NativeView) SetDropShadow(deltaSize zgeo.Size, blur float32, color zgeo.Color) {
	str := fmt.Sprintf("%dpx %dpx %dpx %s", int(deltaSize.W), int(deltaSize.H), int(blur), makeRGBAString(color))
	v.style().Set("boxShadow", str)
}

func (v *NativeView) SetToolTip(str string) {
	v.set("title", str)
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
