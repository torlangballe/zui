//go:build zui && js

package zview

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zcursor"
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zbits"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdebug"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

type baseNativeView struct {
	Element      js.Value
	transparency float32
	parent       *NativeView
	jsFuncs      map[string]js.Func
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

	SetPresentReadyFunc func(v View, beforeWindow bool)
	// RemoveKeyPressHandlerViewsFunc func(v View)
	FindLayoutCellForView func(v View) *zgeo.LayoutCell
	customStylePrefixes   = []string{"-moz-", "-webkit-", ""}
)

func (v *NativeView) MakeJSElement(view View, etype string) {
	v.Element = zdom.DocumentJS.Call("createElement", etype)
	v.Element.Set("style", "position:absolute")
	v.JSStyle().Set("zIndex", BaseZIndex)
	v.View = view
	v.SetPressedHandler("$zdebug", zkeyboard.MetaModifier|zkeyboard.ModifierAlt|zkeyboard.ModifierShift, func() {
		if globalForceClick {
			return
		}
		var str string
		cell := FindLayoutCellForView(v.View)
		if cell != nil {
			str = zstr.Spaced(cell.Alignment, "cmarg:", cell.Margin, "cmin:", cell.MinSize, "cmax:", cell.MaxSize)
		}
		mo, _ := view.(MarginOwner)
		if mo != nil {
			str += fmt.Sprint(" vmarg: ", mo.Margin())
		}
		s, max := v.View.CalculatedSize(zgeo.SizeBoth(9999))
		zlog.Info("NativeView:", v.Hierarchy(), zlog.Pointer(view), reflect.TypeOf(view), v.Rect(), str, "calc:", s, max)
	})
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

func (v *NativeView) SetTop(top float64) {
	v.SetJSStyle("top", fmt.Sprintf("%fpx", top))
}

func (v *NativeView) SetLeft(left float64) {
	v.SetJSStyle("left", fmt.Sprintf("%fpx", left))
}

func (v *NativeView) SetWidth(w float64) {
	v.SetJSStyle("width", fmt.Sprintf("%fpx", w))
}

func (v *NativeView) SetHeight(h float64) {
	v.SetJSStyle("height", fmt.Sprintf("%fpx", h))
}

func (v *NativeView) SetPos(pos zgeo.Pos) {
	style := v.JSGet("style")
	style.Set("left", fmt.Sprintf("%fpx", pos.X))
	style.Set("top", fmt.Sprintf("%fpx", pos.Y))
}

func (v *NativeView) SetSize(size zgeo.Size) {
	style := v.JSGet("style")
	style.Set("width", fmt.Sprintf("%fpx", size.W))
	style.Set("height", fmt.Sprintf("%fpx", size.H))
}

func (v *NativeView) SetRect(rect zgeo.Rect) {
	if rect.Pos.Y < -10 {
		zlog.Error("strange rect for view:", v.Hierarchy(), rect, zdebug.CallingStackString())
	}
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
			zlog.Error("not number", v.ObjectName(), err)
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
	return zgeo.Rect{Pos: pos, Size: size}
}

func SetElementRect(e js.Value, rect zgeo.Rect) {
	style := e.Get("style")
	style.Set("left", fmt.Sprintf("%fpx", rect.Pos.X))
	style.Set("top", fmt.Sprintf("%fpx", rect.Pos.Y))
	style.Set("width", fmt.Sprintf("%fpx", rect.Size.W))
	style.Set("height", fmt.Sprintf("%fpx", rect.Size.H))
}

func (v *NativeView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	return zgeo.SizeD(10, 10), zgeo.Size{}
}

func (v *NativeView) LocalRect() zgeo.Rect {
	var w, h float64
	style := v.JSStyle()
	sw := style.Get("width")
	sh := style.Get("height")
	if sw.String() != "" {
		h = v.parseElementCoord(sh)
		w = v.parseElementCoord(sw)
	} else if v.IsPresented() {
		zlog.Error("parse empty Coord:", style.Get("left"), style.Get("right"), sw, sh, v.Hierarchy(), zdebug.CallingStackString())
	}

	return zgeo.RectMake(0, 0, w, h)
}

func (v *NativeView) SetLocalRect(rect zgeo.Rect) {
	zlog.Fatal("NOT IMPLEMENTED")
}

func (v *NativeView) ObjectName() string {
	if v == nil {
		return "<nil>"
	}
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
	a := alpha
	if !v.IsUsable() {
		a *= 0.4
	}
	v.transparency = 1 - alpha
	v.JSStyle().Set("opacity", a)
}

func (v *NativeView) Alpha() float32 {
	return 1 - v.transparency
}

func (v *NativeView) SetBGColor(c zgeo.Color) {
	v.JSStyle().Set("backgroundColor", zdom.MakeRGBAString(c))
}

func (v *NativeView) BGColor() zgeo.Color {
	str := v.JSStyle().Get("backgroundColor").String()
	if str == "" {
		return zgeo.ColorClear
	}
	return zgeo.ColorFromString(str)
}

func (v *NativeView) SetStyleForAllPlatforms(key, value string) {
	style := v.JSStyle()
	for _, pre := range customStylePrefixes {
		style.Set(pre+key, value)
	}
}

func (v *NativeView) SetSelectable(on bool) {
	val := "none" // https://stackoverflow.com/questions/3779534/how-do-i-disable-text-selection-with-css-or-javascript
	if on {
		val = "all"
	}
	v.SetStyleForAllPlatforms("user-select", val)
}

func (v *NativeView) SetCorner(radius float64) {
	s := fmt.Sprintf("%dpx", int(radius))
	v.SetStyleForAllPlatforms("border-radius", s)
}

func (v *NativeView) SetCorners(radius float64, cAlignment ...zgeo.Alignment) {
	srad := fmt.Sprintf("%dpx", int(radius))
	for a, s := range map[zgeo.Alignment]string{
		zgeo.TopLeft:     "border-top-left-radius",
		zgeo.TopRight:    "border-top-right-radius",
		zgeo.BottomLeft:  "border-bottom-left-radius",
		zgeo.BottomRight: "border-bottom-right-radius",
	} {
		for _, ca := range cAlignment {
			if a == ca {
				v.SetStyleForAllPlatforms(s, srad)
			}
		}
	}
}

func (v *NativeView) Corner() float64 {
	var corner float64
	style := v.JSStyle()
	for _, pre := range customStylePrefixes {
		if zdom.GetIfFloat(style, pre+"border-radius", &corner) {
			return corner
		}
	}
	return 0
}

func (v *NativeView) SetStroke(width float64, c zgeo.Color, inset bool) {
	if inset {
		if width == 0 || !c.Valid {
			v.SetDropShadow()
			return
		}
		d := zstyle.MakeDropShadow(0, 0, 0, c)
		d.Inset = true
		d.Spread = width
		v.SetDropShadow(d)
		v.SetJSStyle("border", "none")
		return
	}
	str := fmt.Sprintf("%dpx solid %s", int(width), c.Hex())
	if inset {
		str += " inset"
	}
	v.SetJSStyle("border", str)
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
	// zlog.Info("Show", v.Hierarchy(), show, zlog.CallingStackString())
	// }
	str := "hidden"
	if show {
		str = "inherit" //visible"
	}
	v.JSStyle().Set("visibility", str)
}

func (v *NativeView) IsShown() bool {
	return v.JSStyle().Get("visibility").String() != "hidden"
}

func (v *NativeView) IsUsable() bool {
	dis := v.Element.Get("disabled")
	if dis.IsNull() || dis.IsUndefined() {
		return true
	}
	// zlog.Info("Usable:", v.Hierarchy(), dis, dis.Bool())
	return !dis.Bool()
}

func (v *NativeView) SetUsable(usable bool) bool {
	if usable == v.IsUsable() {
		return false
	}
	zbits.ChangeBit((*int64)(&v.Flags), ViewUsableFlag, usable)
	v.setUsableAttributes(usable)
	// zlog.Info("SetUsable:", v.Hierarchy(), usable, "->", v.Element.Get("disabled"))
	RangeChildrenFunc(v, false, true, func(view View) bool {
		view.Native().setUsableAttributes(usable)
		return true
	})
	return true
}

func (v *NativeView) setUsableAttributes(usable bool) {
	u := usable //&& v.IsUsable()
	if usable {
		v.JSSet("disabled", nil)
	} else {
		v.JSSet("disabled", true)
	}
	// zlog.Info("setUsableAttributes", v.ObjectName(), usable)
	style := v.JSStyle()
	var alpha float32 = 0.4
	if u || v.Flags&ViewNoDimUsableFlag != 0 {
		alpha = 1 - v.transparency
	}
	// str := "none"
	// if usable {
	// 	str = "auto"
	// }
	// style.Set("pointerEvents", str)
	style.Set("opacity", alpha)
}

func (v *NativeView) SetInteractive(interactive bool) {
	str := "none"
	if interactive {
		str = "auto"
	}
	// zlog.Info("NV.SetInteractive", v.Hierarchy(), str)
	v.SetJSStyle("pointer-events", str)
}

func (v *NativeView) IsInteractive() bool {
	inter := v.JSStyle().Get("pointer-events")
	// fmt.Printf("Inter? %p %s %v\n", v, v.ObjectName(), inter.String())
	if inter.IsUndefined() || inter.String() != "none" {
		return true
	}
	return false
}

func getActiveElement(v *NativeView) js.Value {
	return v.Document().Get("activeElement")
}

func (v *NativeView) IsFocused() bool {
	return getActiveElement(v).Equal(v.Element)
}

func (v *NativeView) Focus(focus bool) {
	// v.JSSet("contenteditable", focus) ?
	if focus {
		v.JSCall("focus")
		// zlog.Info("NV Focus:", v.Hierarchy())
	} else {
		v.JSCall("blur")
		v.Document().Get("body").Call("focus")
		// zlog.Info("NV unFocus:", v.Hierarchy(), v.IsFocused(), v.Document().Get("activeElement").Get("id"))
	}
	// ztimer.StartIn(2, func() {
	// 	v.EnvokeFocusIn()
	// })
}

func (v *NativeView) CanTabFocus() bool {
	val := v.JSGet("tabIndex")
	if val.IsUndefined() || val.String() == "" {
		return false
	}
	if val.Int() < 0 {
		return false
	}
	return true
}

func (v *NativeView) SetKeepFocusOnOutsideClick() {
	v.SetListenerJSFunc("blur:$refocus", func(this js.Value, args []js.Value) any { // if user clicks elsewhere, let's re-set focus if not a focusable relTarget
		rel := args[0].Get("relatedTarget")
		if rel.IsNull() {
			v.Element.Call("focus")
		}
		return nil
	})
}

func (v *NativeView) SetCanTabFocus(can bool) {
	// if v.ObjectName() == "layout" {
	// 	zlog.Info("SetCanFocus:", can, v.Hierarchy(), zdebug.CallingStackString())
	// }
	if can {
		v.JSSet("tabIndex", "0") // Note the capital I in tabIndex !!!!!!
		v.JSSet("className", "zfocus")
		v.SetKeepFocusOnOutsideClick()
		return
	}
	v.JSSet("tabIndex", "-1")
	v.JSSet("className", "")
}

func (v *NativeView) SetFocusHandler(focused func(focus bool)) {
	v.SetListenerJSFunc("focus", func(this js.Value, args []js.Value) any {
		focused(true)
		return nil
	})
	v.SetListenerJSFunc("blur", func(this js.Value, args []js.Value) any {
		focused(false)
		return nil
	})
}

func (root *NativeView) HandleFocusInChildren(in, out bool, handle func(view View, focused bool)) {
	// zlog.Info("Set HandleFocusInChildren", in, root.Hierarchy())
	if in {
		// zlog.Info("Set HandleFocusInChildren", root.Hierarchy())
		handleFocusInChildren(root, "focusin", true, handle)
	}
	if out {
		handleFocusInChildren(root, "focusout", false, handle)
	}
}

func handleFocusInChildren(root *NativeView, eventName string, forFocused bool, handle func(view View, focused bool)) {
	root.SetListenerJSFunc(eventName, func(this js.Value, args []js.Value) any {
		// zlog.Info("child focused", eventName)
		e := args[0].Get("target")
		if e.IsNull() || e.IsUndefined() {
			return nil
		}
		found := FindChildWithElement(root, e)
		// zlog.Info("child focused", root.Hierarchy(), eventName, e.Get("id").String(), found)
		if found != nil {
			handle(found, forFocused)
		}
		return nil
	})
}

func (root *NativeView) GetFocusedChildView(andSelf bool) View {
	e := getActiveElement(root)
	if e.IsUndefined() {
		return nil
	}
	if andSelf && root.IsFocused() {
		return root
	}
	return FindChildWithElement(root, e)
}

func FindChildWithElement(root *NativeView, e js.Value) View {
	var found View
	foundID := e.Get("id").String()
	// zlog.Info("FindChildWithElement:", root.Hierarchy())
	RangeChildrenFunc(root.View, true, true, func(view View) bool {
		n := view
		id := n.Native().JSGet("id").String()
		// zlog.Info("FindChildWithElement:", root.Hierarchy(), id, foundID)
		if id == foundID {
			found = n
			return false
		}
		return true
	})
	return found
}

func (v *NativeView) SetOpaque(opaque bool) {
}

func (v *NativeView) Hierarchy() string {
	if v == nil {
		return "nil"
	}
	return v.HierarchyToRoot(nil)
}

func (v *NativeView) HierarchyToRoot(root *NativeView) string {
	var str string
	var found bool
	var debug bool
	if root == nil {
		found = true
		str = "/"
	}
	for _, p := range v.AllParents() {
		if !found {
			found = (root == p)
		}
		if found {
			str += p.ObjectName()
			if debug {
				str += ":" + zlog.Pointer(p)
			}
		}
	}
	str += v.ObjectName()
	if debug {
		str += ":" + zlog.Pointer(v)
	}
	return str
}

func (v *NativeView) RemoveFromParent(callRemoveFuncs bool) {
	zlog.Assert(!zui.DebugOwnerMode || v.parent != nil, v.Hierarchy())
	// zlog.Info("NV.RemoveFromParent:", v.Hierarchy())
	v.parent.RemoveChild(v.View, callRemoveFuncs)
	v.parent = nil
}

func (v *NativeView) SetFont(font *zgeo.Font) {
	cssStyle := v.JSStyle()
	if font.Style == zgeo.FontStyleUndef {
		font.Style = zgeo.FontStyleNormal
	}
	// zlog.Debug("font-style:", v.ObjectName(), (font.Style & zgeo.FontStyleItalic).String())
	cssStyle.Set("font-style", (font.Style & zgeo.FontStyleItalic).String())
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
	v.JSSet("innerText", text)
}

func (v *NativeView) Text() string {
	text := v.JSGet("innerText").String()
	return text
}

func (v *NativeView) InsertBefore(before View) {
	v.Element.Get("parentNode").Call("insertBefore", v.Element, before.Native().Element)
}

func (v *NativeView) AddChild(child, before View) {
	n := child.Native()
	if n == nil {
		zlog.Fatal("NativeView AddChild child not native", v.Hierarchy(), child.ObjectName())
	}
	n.parent = v
	// if child.ObjectName() == "tab-separator" {
	// 	zlog.Info("Call On Add:", n.Hierarchy(), len(n.DoOnAdd))
	// }
	n.PerformAddRemoveFuncs(true)
	if before != nil {
		// zlog.Info("AddBefore:", v.ObjectName(), child.ObjectName(), "before:", before.ObjectName())
		v.JSCall("insertBefore", n.Element, before.Native().Element)
	} else {
		v.JSCall("appendChild", n.Element)
	}
	// for _, p := range n.AllParents() {
	// 	p.allChildrenPresented = false
	// }
	// zlog.Info("ADDCHILD:", v.ObjectName(), child.ObjectName(), v.Rect())
	if v.IsPresented() {
		SetPresentReadyFunc(child, true)
		SetPresentReadyFunc(child, false)
	}
}

func (v *NativeView) GetPathOfChild(child View) string {
	path := child.Native().HierarchyToRoot(v)
	if path == "" {
		return path
	}
	zstr.HeadUntilWithRest(path, "/", &path) // remove first path component, which is v's
	return path
}

func (v *NativeView) ReplaceChild(child, with View) {
	var focusedPath string
	focused := v.GetFocusedChildView(false)
	if focused != nil {
		focusedPath = child.Native().GetPathOfChild(focused)
		// focused.HierarchyToRoot(child.Native())
		// zstr.HeadUntilWithRest(focusedPath, "/", &focusedPath) // remove first path component, which is v's
	}
	v.AddChild(with, nil) // needs to preserve index, which isn't really supported in AddChild yet anyway
	with.SetRect(child.Rect())
	v.RemoveChild(child, true)
	if focusedPath != "" {
		f := ChildOfViewFunc(with, focusedPath)
		if f != nil {
			f.Native().Focus(true)
		}
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

func (v *NativeView) RemoveChild(child View, callRemoveFuncs bool) {
	// zlog.Info("RemoveChild:", child.Native().Hierarchy(), len(v.DoOnRemove))
	nv := child.Native()
	if nv == nil {
		panic("NativeView AddChild child not native")
	}
	if callRemoveFuncs {
		// RemoveKeyPressHandlerViewsFunc(child)
		nv.PerformAddRemoveFuncs(false)
	}
	nv.Element = v.JSCall("removeChild", nv.Element) // we need to set it since  it might be accessed for ObjectName etc still in collapsed containers
	//!! nv.parent = nil if we don't do this, we can still uncollapse child in container without having to remember comtainer. Testing.
}

func (v *NativeView) SetDropShadow(shadow ...zstyle.DropShadow) {
	var parts []string
	var ss string
	for _, d := range shadow {
		if !d.Color.Valid {
			continue
		}
		if d.Spread != 0 {
			ss = fmt.Sprintf("%dpx", int(d.Spread))
		}
		str := fmt.Sprintf("%dpx %dpx %dpx %s %s", int(d.Delta.W), int(d.Delta.H), int(d.Blur), ss, zdom.MakeRGBAString(d.Color))
		if d.Inset {
			str += " inset"
		}
		parts = append(parts, str)
	}
	str := "none"
	if len(parts) > 0 {
		str = strings.Join(parts, ", ")
	}
	v.SetJSStyle("boxShadow", str)
}

func (v *NativeView) SetToolTip(str string) {
	tta, _ := v.View.(ToolTipAdder)
	if tta != nil {
		str += tta.GetToolTipAddition()
	}
	v.JSSet("title", str)
}

func (v *NativeView) ToolTip() string {
	return v.JSGet("title").String()
}

func (v *NativeView) AbsoluteRect() zgeo.Rect {
	r := v.Element.Call("getBoundingClientRect")
	x := r.Get("left").Float() // x
	y := r.Get("top").Float()  // y
	w := r.Get("width").Float()
	h := r.Get("height").Float()
	return zgeo.RectFromXYWH(x, y, w, h)
}

func (v *NativeView) AbsoluteRectWithParentOffset() zgeo.Rect {
	offset := v.parent.ContentOffset()
	r := v.AbsoluteRect()
	r.Pos.Add(offset)
	return r
}

var globalForceClick bool

const pressMouseDownPrefix = "mousedown:$press-"

func (v *NativeView) DoMouseDown(id string, mods zkeyboard.Modifier) {
	mid := makeDownPressKey(id+"$down", false, mods)
	globalForceClick = true
	for n, f := range v.jsFuncs {
		// zlog.Info("ClickDown", n, "==", mid, n == mid)
		if n == mid {
			f.Value.Invoke(f.Value)
			break
		}
	}
	globalForceClick = false
}

func (v *NativeView) ClickAll() {
	globalForceClick = true
	for n, f := range v.jsFuncs {
		if strings.HasPrefix(n, pressMouseDownPrefix) {
			f.Value.Invoke(f.Value)
		}
	}
	globalForceClick = false
}

func (v *NativeView) Click(id string, long bool, mods zkeyboard.Modifier) {
	mid := makeDownPressKey(id, long, mods)
	globalForceClick = true
	for n, f := range v.jsFuncs {
		// zlog.Info("Click", n, "==", mid, n == mid)
		if n == mid {
			f.Value.Invoke(f.Value)
			break
		}
	}
	globalForceClick = false
}

func (v *NativeView) SetPressedHandler(id string, mods zkeyboard.Modifier, handler func()) {
	v.setMouseDownForPress(id, mods, handler, nil)
}

func (v *NativeView) SetLongPressedHandler(id string, mods zkeyboard.Modifier, handler func()) {
	v.setMouseDownForPress(id, mods, nil, handler)
}

func (v *NativeView) CallPressHandlers() {
}

type longPresser struct {
	cancelPress     bool
	downPressedTime time.Time
	longTimer       *ztimer.Timer
}

var globalLongPressState longPresser
var lastPressedEvent js.Value

func makeDownPressKey(id string, long bool, mods zkeyboard.Modifier) string {
	if id == "" {
		id = "$general"
	}
	if long {
		id += ".$long"
	} else {
		id += ".$short"
	}

	return fmt.Sprintf("%s%s^%s", pressMouseDownPrefix, id, mods)
}

func (v *NativeView) setMouseDownForPress(id string, mods zkeyboard.Modifier, press func(), long func()) {
	v.JSSet("className", "widget")
	_, got := v.jsFuncs[id]
	if got {
		return
	}
	invokeFunc := zdebug.FileLineAndCallingFunctionString(4, true)
	mid := makeDownPressKey(id, long != nil, mods)
	v.SetListenerJSFunc(mid, func(this js.Value, args []js.Value) any {
		if globalForceClick {
			press()
			return nil
		}
		event := args[0]
		v.SetStateOnPress(event)
		if zkeyboard.ModifiersAtPress != mods {
			return nil // don't call stopPropagation, we aren't handling it
		}
		// zlog.Info("Pressed2:", v.Hierarchy(), mid)
		target := event.Get("target")
		var foundChild *NativeView
		if !target.Equal(v.Element) {
			// zlog.Info("Pressed Child:", v.Hierarchy(), mid, target.Get("id").String())
			RangeChildrenFunc(v.View, true, false, func(view View) bool {
				// zlog.Info("Pressed Child Find?:", view.Native().Hierarchy(), view.Native().Element.Get("id").String())
				if view.Native().Element.Equal(target) {
					foundChild = view.Native()
					return false
				}
				return true
			})
			if foundChild != nil {
				// zlog.Info("Pressed Child Found:", found.Hierarchy(), foundChild.IsUsable(), foundChild.IsInteractive())
				if foundChild.IsUsable() && foundChild.IsInteractive() {
					return nil
				}
			}
		}
		// if foundChild != nil {
		// 	LastPressedPos.Add(foundChild.AbsoluteRect().Pos)
		// }
		// zlog.Info("Press:", foundChild != nil, LastPressedPos, v.Hierarchy(), id)
		// zlog.Info("Pressed:", v.Hierarchy(), v.AbsoluteRect(), LastPressedPos)
		if long != nil {
			globalLongPressState = longPresser{}
			globalLongPressState.downPressedTime = time.Now()
			globalLongPressState.longTimer = ztimer.StartIn(0.5, func() {
				if long != nil {
					if v.IsUsable() {
						defer zdebug.RecoverFromPanic(true, invokeFunc)
						long()
					}
					globalLongPressState.cancelPress = true
				}
				globalLongPressState.longTimer = nil
			})
		}
		var fup js.Func
		fup = js.FuncOf(func(this js.Value, args []js.Value) any {
			lastPressedEvent = args[0]
			if !globalLongPressState.cancelPress && press != nil && v.IsUsable() {
				defer zdebug.RecoverFromPanic(true, invokeFunc)
				// args[0].Call("stopPropagation") // this one canceled up-listener in SetPressUpDownMovedHandler for some reason
				v.SetStateOnPress(args[0])
				press()
				lastPressedEvent = js.Value{}
				v.ClearStateOnUpPress()
			}
			if globalLongPressState.longTimer != nil {
				globalLongPressState.longTimer.Stop()
				globalLongPressState.longTimer = nil
			}
			globalLongPressState.cancelPress = false
			v.JSCall("removeEventListener", "mouseup", fup)
			fup.Release()
			return nil
		})
		v.JSCall("addEventListener", "mouseup", fup)
		// args[0].Call("stopPropagation")
		return nil
	})
}

func StopPropagationOfLastPressedEvent() {
	if lastPressedEvent.IsUndefined() {
		return
	}
	lastPressedEvent.Call("stopPropagation")
}

func getMousePos(e js.Value) (pos zgeo.Pos) {
	pos.X = e.Get("clientX").Float()
	pos.Y = e.Get("clientY").Float()
	return
}

// AttachJSFunc calls setJSFunc below with isListener false.
func (v *NativeView) AttachJSFunc(name string, fn func(this js.Value, args []js.Value) any) js.Func {
	return v.setJSFunc(name, false, fn)
}

// SetListenerJSFunc calls setJSFunc below with isListener true.
func (v *NativeView) SetListenerJSFunc(name string, fn func(this js.Value, args []js.Value) any) js.Func {
	return v.setJSFunc(name, true, fn)
}

// setJSFunc adds a named js func created with js.FuncOf to a view's jsFuncs map.
// These are released on view remove, or if set again for the same name.
// If isListener is set, the func is added as a listener to the view, and removed as well.
// anything after : in the name is removed first, to allow multiple listeners with same name to co-exist.
func (v *NativeView) setJSFunc(name string, isListener bool, fn func(this js.Value, args []js.Value) any) js.Func {
	// zlog.Info("setJSFunc:", v.Hierarchy(), name, isListener)
	var outFunc js.Func
	nameEventPart := zstr.HeadUntil(name, ":") // we allow for multiple funcs with same name to exists at same time, if they have different :xxx suffixes.
	if v.jsFuncs == nil {
		v.jsFuncs = map[string]js.Func{}
		v.AddOnRemoveFunc(func() {
			for _, f := range v.jsFuncs {
				// zlog.Info("NV: Release func:", v.Native(), name)
				f.Release()
			}
		})
	} else {
		f, got := v.jsFuncs[name]
		if got {
			outFunc = f
			if isListener {
				v.JSCall("removeEventListener", nameEventPart, f)
			}
			f.Release()
		}
	}
	if fn == nil {
		delete(v.jsFuncs, name)
	} else {
		// outFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		// 	// zlog.Info("ListenerFired:", v.Hierarchy(), name)
		// 	return fn(this, args)
		// })
		outFunc = js.FuncOf(fn)
		if isListener {
			// zlog.Info("addEventListener", name, v.Hierarchy())
			v.JSCall("addEventListener", nameEventPart, outFunc)
		}
		v.jsFuncs[name] = outFunc
	}
	return outFunc
}

func (v *NativeView) SetSwipeHandler(handler func(pos, dir zgeo.Pos)) {
	const minDiff = 10.0
	v.SetListenerJSFunc("mousedown:swipe", func(this js.Value, args []js.Value) any {
		if handler == nil {
			return nil
		}
		mouseStartTime = time.Now()
		mouseStartPos = getMousePos(args[0])
		return nil
	})
	v.SetListenerJSFunc("mousemove:swipe", func(this js.Value, args []js.Value) any {
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
	})
}

func (v *NativeView) SetChildrenAboveParent(above bool) {
	// zlog.Info("SetChildrenAboveParent:", v.ObjectName(), above)
	str := "hidden"
	if above {
		str = "visible"
	}
	v.JSStyle().Set("overflow", str)
}

func (v *NativeView) JSCall(method string, args ...any) js.Value {
	return v.Element.Call(method, args...)
}

func (v *NativeView) JSSet(property string, value any) {
	v.Element.Set(property, value)
}

func (v *NativeView) JSGet(property string) js.Value {
	return v.Element.Get(property)
}

func (v *NativeView) SetScrollHandler(handler func(pos zgeo.Pos)) {
	v.SetListenerJSFunc("scroll", func(js.Value, []js.Value) any {
		if handler != nil {
			handler(v.ContentOffset())
		}
		return nil
	})
}

func (v *NativeView) ContentOffset() zgeo.Pos {
	x := v.JSGet("scrollLeft").Float()
	y := v.JSGet("scrollTop").Float()
	return zgeo.PosD(x, y)
}

func (v *NativeView) SetXContentOffsetAnimated(x float64, done func()) {
	setContentOffsetAnimated(v, false, x, done)
}

func (v *NativeView) SetYContentOffsetAnimated(x float64, done func()) {
	setContentOffsetAnimated(v, true, x, done)
}

func (v *NativeView) SetXContentOffset(x float64) {
	v.JSSet("scrollLeft", x)
}

func (v *NativeView) SetYContentOffset(y float64) {
	v.JSSet("scrollTop", y)
}

func setContentOffsetAnimated(v *NativeView, vertical bool, n float64, animateDone func()) {
	o := v.ContentOffset()
	args := map[string]any{
		"behavior": "smooth",
	}
	if vertical {
		args["top"] = n - o.Y
	} else {
		args["left"] = n - o.X
	}
	if animateDone != nil {
		ztimer.StartIn(0.6, animateDone) // this is a hack, should test when not changing anymore?
	}
	v.JSCall("scrollBy", args)
	return
}

func (v *NativeView) SetRootYContentOffset(y float64) {
	toModalWindowOnly := true
	v.RootParent(toModalWindowOnly).SetYContentOffset(y)
}

func (v *NativeView) ShowScrollBars(x, y bool) {
	str := "hidden"
	if x {
		str = "auto"
	}
	v.SetJSStyle("overflow-x", str)
	str = "hidden"
	if y {
		str = "auto"
	}
	v.SetJSStyle("overflow-y", str)
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
	item.Call("getAsString", zdom.MakeSingleCallJSCallback(func(this js.Value, args []js.Value) any {
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
	v.SetListenerJSFunc("dragstart", func(this js.Value, args []js.Value) any {
		event := args[0]
		data, mime := getData()
		//		array := js.Global().Get("Uint8Array").New(len(data))
		//		js.CopyBytesToJS(array, []byte(data))
		// zlog.Info("Dtrans:", mime, array.Length())
		// zlog.Info("Dtrans:", mime, data)
		//		mime = "text/plain"
		event.Get("dataTransfer").Call("setData", mime, data) //event.Get("target").Get("id"))
		return nil
	})
}

func (v *NativeView) SetPointerDropHandler(handler func(dtype DragType, data []byte, name string, pos zgeo.Pos) bool) {
	if zdebug.IsInTests {
		return
	}
	//	v.JSSet("className", "zdropper")
	v.SetListenerJSFunc("dragenter", func(this js.Value, args []js.Value) any {
		if dragEnterView == nil {
			dragEvent(args[0], DragEnter, handler)
		}
		// zlog.Info("ondragenter:", v.ObjectName())
		dragEnterView = v
		return nil
	})
	v.SetListenerJSFunc("dragleave", func(this js.Value, args []js.Value) any {
		// zlog.Info("ondragleave1:", dragEnterView != nil, dragEnterView == v, v.ObjectName())
		// zlog.Info("ondragleave:", v.ObjectName(), dragEnterView != v)
		// if dragEnterView != v {
		// 	return nil
		// }
		dragEnterView = nil
		dragEvent(args[0], DragLeave, handler)
		return nil
	})
	v.SetListenerJSFunc("dragover", func(this js.Value, args []js.Value) any {
		dragEvent(args[0], DragOver, handler)
		return nil
	})
	v.SetListenerJSFunc("drop", func(this js.Value, args []js.Value) any {
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
		if !handler(DragDropFilePreflight, nil, name, zgeo.Pos{}) {
			return nil
		}
		// zlog.Info("FileProcessing")
		zdom.JSFileToGo(file, func(data []byte, name string) {
			var pos zgeo.Pos
			pos.X = event.Get("offsetX").Float()
			pos.Y = event.Get("offsetY").Float()
			// zlog.Info("Drop offset:", pos)
			// zlog.Info("nv.DragDropFile")
			handler(DragDropFile, data, name, pos)
		}, nil)
		return nil
	})
}

func (v *NativeView) SetUploader(got func(data []byte, name string), skip func(name string) bool, progress func(p float64)) {
	e := v.Document().Call("createElement", "input")
	e.Set("type", "file")
	e.Set("style", "opacity: 0.0; position: absolute; top: 0; left: 0; bottom: 0; right: 0; width: 100%; height:100%;")
	// e.Set("accept", "*/*")

	var changeFunc js.Func
	changeFunc = js.FuncOf(func(this js.Value, args []js.Value) any { // was onchange????
		files := e.Get("files")
		if files.Length() > 0 {
			file := files.Index(0)
			name := file.Get("name").String()
			if skip != nil && skip(name) {
				return nil
			}
			zdom.JSFileToGo(file, got, progress)
		}
		changeFunc.Release()
		return nil
	})
	e.Set("onchange", changeFunc)

	v.SetListenerJSFunc("click", func(this js.Value, args []js.Value) any {
		if time.Since(lastUploadClick) < time.Millisecond*100 { // e.Call("click") below causes v onclick to be re-called, bail + preventDefault important or it doesn't work (error started on Tor's M1 Mac Pro)
			args[0].Call("preventDefault")
			return nil
		}
		lastUploadClick = time.Now()
		// zlog.Info("uploader clickthru")
		e.Call("click")
		return nil
	})
	v.JSCall("appendChild", e)
}

func (v *NativeView) HasPressedDownHandler() bool {
	return v.JSGet("onmousedown").IsNull()
}

func (v *NativeView) SetPressedDownHandler(id string, mods zkeyboard.Modifier, handler func()) {
	// zlog.Info("nv.SetPressedDownHandler:", v.Hierarchy())
	v.JSSet("className", "widget")
	mid := makeDownPressKey(id+"$down", false, mods)
	v.SetListenerJSFunc(mid, func(this js.Value, args []js.Value) any {
		if globalForceClick {
			handler()
			return nil
		}
		e := args[0]
		v.SetStateOnPress(e)
		if zkeyboard.ModifiersAtPress != mods {
			return nil // don't call stopPropagation, we aren't handling it
		}
		e.Call("stopPropagation")
		zlog.Assert(len(args) > 0)
		handler()
		return nil
	})
}

func (v *NativeView) SetDoublePressedHandler(handler func()) {
	// zlog.Info("SetDoublePressedHandler", v.Hierarchy(), zlog.CallingStackString())
	v.SetListenerJSFunc("dblclick", func(this js.Value, args []js.Value) any {
		handler()
		return nil
	})
}

func getMousePosRelative(v *NativeView, e js.Value) zgeo.Pos {
	pos := getMousePos(e)
	return pos.Minus(v.AbsoluteRect().Pos)
}

func (v *NativeView) SetPointerEnterHandler(handleMoves bool, handler func(pos zgeo.Pos, inside zbool.BoolInd)) {
	if zdebug.IsInTests {
		return
	}
	v.SetListenerJSFunc("mouseenter", func(this js.Value, args []js.Value) any {
		if SkipEnterHandler {
			return nil
		}
		handler(getMousePosRelative(v, args[0]), zbool.True)
		if handleMoves {
			// we := v.GetWindowElement()
			v.SetListenerJSFunc("mousemove:enter", func(this js.Value, args []js.Value) any {
				// zlog.Info("mousemove:enter")
				e := args[0]
				if e.Get("movementX").Float() == 0 && e.Get("movementY").Float() == 0 { // if we don't do this, we get weird move events when modifier key is pressed. Maybe we want that one day, so follow this space.
					return nil
				}
				handler(getMousePosRelative(v, e), zbool.Unknown)
				return nil
			})
		}
		return nil
	})
	v.SetListenerJSFunc("mouseleave", func(this js.Value, args []js.Value) any {
		// zlog.Info("Mouse Leave")
		if SkipEnterHandler {
			return nil
		}
		handler(getMousePosRelative(v, args[0]), zbool.False)
		if handleMoves {
			v.SetListenerJSFunc("mousemove:enter", nil)
		}
		return nil
	})
}

func setKeyHandler(down bool, v *NativeView, handler func(km zkeyboard.KeyMod, down bool) bool) {
	event := "keyup"
	if down {
		event = "keydown"
	}
	v.SetListenerJSFunc(event, func(val js.Value, args []js.Value) any {
		if !v.Document().Call("hasFocus").Bool() {
			return nil
		}
		if handler != nil {
			event := args[0]
			km := zkeyboard.GetKeyModFromEvent(event)
			if down {
				zkeyboard.CurrentKeyDown = km
			} else {
				zkeyboard.CurrentKeyDown = zkeyboard.KeyMod{}
			}
			// zlog.Info("Key!", v.Hierarchy(), v.IsFocused())
			if handler(km, down) {
				event.Call("preventDefault")
				event.Call("stopPropagation")
			}
		}
		return nil
	})
}

func (v *NativeView) SetKeyHandler(handler func(km zkeyboard.KeyMod, down bool) bool) {
	setKeyHandler(true, v, handler)
	setKeyHandler(false, v, handler)
}

func (v *NativeView) SetOnInputHandler(handler func()) {
	v.SetListenerJSFunc("input", func(js.Value, []js.Value) any {
		handler()
		return nil
	})
}

func (v *NativeView) SetStateOnPress(event js.Value) {
	var pos zgeo.Pos
	pos.X = event.Get("offsetX").Float()
	pos.Y = event.Get("offsetY").Float()
	LastPressedPos = pos //.Minus(v.AbsoluteRect().Pos)
	zkeyboard.ModifiersAtPress = zkeyboard.GetKeyModFromEvent(event).Modifier
}

func (v *NativeView) ClearStateOnUpPress() {
	LastPressedPos = zgeo.Pos{}
	zkeyboard.ModifiersAtPress = zkeyboard.ModifierNone
}

func (v *NativeView) SetPressUpDownMovedHandler(handler func(pos zgeo.Pos, down zbool.BoolInd) bool) {
	const minDiff = 10.0
	v.SetListenerJSFunc("mousedown:$updown", func(this js.Value, args []js.Value) any {
		var moveFunc, upFunc js.Func
		// zlog.Info("Mouse up/down down")
		we := v.GetWindowElement()
		// we := v.Element
		if we.IsUndefined() {
			return nil
		}
		e := args[0]
		target := e.Get("target")
		if !target.Equal(v.Element) && target.Get("tagName").String() == "INPUT" {
			e.Call("stopPropagation")
			return false
		}
		pos := getMousePosRelative(v, e)
		v.SetStateOnPress(e)
		// pos = getMousePos(e).Minus(v.AbsoluteRect().Pos)
		movingPos = &pos
		zkeyboard.ModifiersAtPress = zkeyboard.GetKeyModFromEvent(e).Modifier
		if handler(*movingPos, zbool.True) {
			e.Call("stopPropagation")
			e.Call("preventDefault")
		}
		// pos := getMousePos(e).Minus(v.AbsoluteRect().Pos)
		upFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
			// zlog.Info("SetPressUpDownMovedHandler up")
			e := args[0]
			upPos := getMousePosRelative(v, e)
			movingPos = nil
			v.ClearStateOnUpPress()
			we.Call("removeEventListener", "mouseup", upFunc)
			we.Call("removeEventListener", "mousemove", moveFunc)
			upFunc.Release()
			moveFunc.Release()
			zkeyboard.ModifiersAtPress = zkeyboard.GetKeyModFromEvent(e).Modifier
			if handler(upPos, zbool.False) {
				// e.Call("stopPropagation")
				// e.Call("preventDefault")
				// }
			}
			e.Call("stopPropagation")
			return nil
		})
		we.Call("addEventListener", "mouseup", upFunc)
		moveFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
			// v.SetListenerJSFunc("mousemove:updown", func(this js.Value, args []js.Value) any {
			if movingPos != nil {
				e := args[0]
				pos := getMousePosRelative(v, e)
				zkeyboard.ModifiersAtPress = zkeyboard.GetKeyModFromEvent(e).Modifier
				if handler(pos, zbool.Unknown) {
					e.Call("stopPropagation")
					e.Call("preventDefault")
				}
			}
			//			e.Call("stopPropagation")
			return nil
		})
		we.Call("addEventListener", "mousemove", moveFunc)
		return nil
	})
}

func (v *NativeView) MakeLink(surl, name string) {
	stype := strings.ToLower(v.Element.Get("nodeName").String())
	zlog.Assert(stype == "a", stype)
	v.JSSet("download", name)
	v.JSSet("href", surl)
}

func (v *NativeView) SetTilePath(spath string) {
	// spath2 := zimage.MakeImagePathWithAddedScale(spath, 2)
	// format := `-webkit-image-set(url("%s") 1x, url("%s") 2x)`
	// s := fmt.Sprintf(format, spath, spath2)
	// v.JSStyle().Set("backgroundImage", s)
}

// SetHandleExposed sets a handler for v that is called with intersectsViewport=true, when v becomes visible.
// It calls with intersectsViewport=false when it becomes inivisble.
// In this js implementation is uses the IntersectionObserver, and removed the observation on view removal.
// Note that it has to be called AFTER the window v will be in is opened, so v.GetWindow() gives correct window observe with.
// If called twice. the js.Func will leak.
func (v *NativeView) SetHandleExposed(handle func(intersectsViewport bool)) {
	f := js.FuncOf(func(this js.Value, args []js.Value) any {
		entries := args[0]
		for i := 0; i < entries.Length(); i++ {
			e := entries.Index(i)
			inter := e.Get("isIntersecting").Bool()
			handle(inter)
		}
		return nil
	})
	e := v.GetWindowElement()
	// opts := map[string]any{
	// 	"root": all[0].Element,
	// }
	observer := e.Get("IntersectionObserver").New(f) //, js.ValueOf(opts))
	observer.Call("observe", v.Element)
	v.AddOnRemoveFunc(func() {
		// zlog.Info("remove expose observer:", v.Hierarchy())
		observer.Call("disconnect")
		f.Release()
		handle(false)
	})
}

func AddTextNode(v *NativeView, text string) {
	textNode := v.Document().Call("createTextNode", text)
	v.JSCall("appendChild", textNode)
	//	js.Value(*e).Call("appendChild", textNode)
}

func AddView(parent, child *NativeView) {
	parent.JSCall("appendChild", child.Element)
}

func (v *NativeView) Document() js.Value {
	root := v
	all := v.AllParents()
	if len(all) > 1 {
		root = all[0]
	}
	return root.JSGet("ownerDocument")
}

func (v *NativeView) GetWindowElement() js.Value {
	d := v.Document()
	return d.Get("defaultView")
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
		// zlog.Info("SetStyling:", v.Hierarchy(), style.StrokeWidth, style.StrokeColor)
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

func (nv *NativeView) SetNativePadding(p zgeo.Rect) {
	style := nv.JSStyle()
	style.Set("padding-top", fmt.Sprintf("%dpx", int(p.Min().Y)))
	style.Set("padding-left", fmt.Sprintf("%dpx", int(p.Min().X)))
	style.Set("padding-bottom", fmt.Sprintf("%dpx", -int(p.Max().Y))) // trying remove minus
	style.Set("padding-right", fmt.Sprintf("%dpx", -int(p.Max().X)))  // trying remove minus
}

func (nv *NativeView) SetNativeMargin(m zgeo.Rect) {
	style := nv.JSStyle()
	style.Set("margin-top", fmt.Sprintf("%dpx", int(m.Min().Y)))
	style.Set("margin-left", fmt.Sprintf("%dpx", int(m.Min().X)))
	style.Set("margin-bottom", fmt.Sprintf("%dpx", -int(m.Max().Y)))
	style.Set("margin-right", fmt.Sprintf("%dpx", -int(m.Max().X)))
}

func (nv *NativeView) ShowBackface(visible bool) {
	str := "hidden"
	if visible {
		str = "visible"
	}
	nv.SetJSStyle("backfaceVisibility", str)
}

func (nv *NativeView) EnvokeFocusIn() {
	// zlog.Info("EnvokeFocusIn", nv.Hierarchy())
	fin := js.Global().Get("Event").New("focusin")
	nv.JSCall("dispatchEvent", fin)
}

func (v *NativeView) RootParent(toModalWindowOnly bool) *NativeView {
	all := v.AllParents()
	if len(all) == 0 {
		return v
	}
	i := 0
	if all[i].ObjectName() == "window" {
		i++
	}
	if len(all) > i && toModalWindowOnly && all[i].ObjectName() == "$blocker" {
		i++
	}
	return all[i]
}

func DownloadURI(uri, name string) {
	link := zdom.DocumentJS.Call("createElement", "a")
	link.Set("href", uri)
	link.Set("download", name)
	zdom.DocumentJS.Get("body").Call("appendChild", link)
	link.Call("click")
	zdom.DocumentJS.Get("body").Call("removeChild", link)
	// link.Delete() // What is this?
}

func (v *NativeView) AddTest() {
	fdown := js.FuncOf(func(this js.Value, args []js.Value) any {
		zlog.Info("DAUWN!")
		fup := js.FuncOf(func(this js.Value, args []js.Value) any {
			zlog.Info("UP1!")
			return nil
		})
		v.JSCall("addEventListener", "mouseup", fup)
		return nil
	})
	fup2 := js.FuncOf(func(this js.Value, args []js.Value) any {
		zlog.Info("UP2!")
		args[0].Call("stopPropagation")
		return nil
	})

	v.JSCall("addEventListener", "mousedown", fdown)
	v.JSCall("addEventListener", "mouseup", fup2)
}
