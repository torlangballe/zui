package zui

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"syscall/js"
	"time"

	"github.com/torlangballe/zutil/ustr"
	"github.com/torlangballe/zutil/zlog"
)

// https://github.com/siongui/godom
// https://medium.zenika.com/go-1-11-webassembly-for-the-gophers-ae4bb8b1ee03

var DocumentJS = js.Global().Get("document")
var DocumentElementJS = DocumentJS.Get("documentElement")
var WindowJS = js.Global().Get("window")

type css js.Value

func init() {
}

func parseElementCoord(value js.Value) float64 {
	var s string
	str := value.String()
	if ustr.HasSuffix(str, "px", &s) {
		n, err := strconv.ParseFloat(s, 32)
		if err != nil {
			zlog.Error(err, "not number")
			return 0
		}
		return n
	}
	zlog.Error(nil, "not handled type: "+str)
	return 0
}

func getCreatedTimeFromStatT(fstat *syscall.Stat_t) time.Time {
	return time.Time{}
}

func AddTextNode(e *NativeView, text string) {
	textNode := DocumentJS.Call("createTextNode", text)
	e.call("appendChild", textNode)
	//	js.Value(*e).Call("appendChild", textNode)
}

func addView(parent, child *NativeView) {
	parent.call("appendChild", child.Element)
}

func jsAddEventListener(e js.Value, name string, handler func()) {
	err := e.Call("addEventListener", name, js.FuncOf(func(js.Value, []js.Value) interface{} {
		fmt.Println("event listener")
		return nil
	}))
	fmt.Println("jsAddEventListener err:", err)
}

func NativeViewAddToRoot(v View) {
	// ftrans := js.FuncOf(func(js.Value, []js.Value) interface{} {
	// 	return nil
	// })

	n := &NativeView{}
	n.Element = DocumentElementJS
	n.View = n
	// s := WindowGetCurrent().Rect().Size.DividedByD(2)

	// o, _ := v.(NativeViewOwner)
	// if o == nil {
	// 	panic("NativeView AddChild child not native")
	// }
	// nv := o.GetNative()
	// nv.style().Set("display", "inline-block")

	// scale := fmt.Sprintf("scale(%f)", ScreenMain().Scale)
	// n.style().Set("-webkit-transform", scale)

	// trans := fmt.Sprintf("translate(-%f,%f)", s.W, 0.0)
	// fmt.Println("TRANS:", trans)
	// n.style().Set("-webkit-transform", trans)
	n.AddChild(v, -1)
}

func (n *NativeView) call(method string, v js.Value) {
	n.Element.Call(method, v)
}

func (n *NativeView) set(property string, v interface{}) {
	n.Element.Set(property, v)
}

func (n *NativeView) get(property string) js.Value {
	return n.Element.Get(property)
}

func getFontStyle(font *Font) string {
	var parts []string
	if font.Style&FontStyleBold != 0 {
		parts = append(parts, "bold")
	}
	if font.Style&FontStyleItalic != 0 {
		parts = append(parts, "italic")
	}
	parts = append(parts, fmt.Sprintf("%dpx", int(font.Size)))
	parts = append(parts, font.Name)

	return strings.Join(parts, " ")
}

// func (ti *TextInfo) getTextSize(noWidth bool) Size {
// 	var canvas = DocumentJS.Call("createElement", "canvas")
// 	var context = canvas.Call("getContext", "2d")
// 	context.Set("font", getFontStyle(ti.Font))
// 	var metrics = context.Call("measureText", ti.Text)
// 	fmt.Println("CALCD:", metrics.Get("width"), ti.Font.Name, ti.Font.Size)
// 	return Size{metrics.Get("width").Float(), metrics.Get("height").Float()}
// }
