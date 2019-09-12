package zgo

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"syscall/js"

	"github.com/torlangballe/zutil/ustr"
	"github.com/torlangballe/zutil/zlog"
)

func init() {
}

type css js.Value

// https://github.com/siongui/godom
// https://medium.zenika.com/go-1-11-webassembly-for-the-gophers-ae4bb8b1ee03

func parseCoord(value js.Value) float64 {
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

func getCreatedTimeFromStatT(fstat *syscall.Stat_t) Time {
	return TimeNull
}

var DocumentJS = js.Global().Get("document")
var DocumentElementJS = DocumentJS.Get("documentElement")
var WindowJS = js.Global().Get("window")

func AddTextNode(e *NativeView, text string) {
	textNode := DocumentJS.Call("createTextNode", text)
	e.call("appendChild", textNode)
	//	js.Value(*e).Call("appendChild", textNode)
}

func addView(parent, child *NativeView) {
	parent.call("appendChild", child.Element)
}

func NativeViewAddToRoot(v View) {
	n := &NativeView{}
	n.Element = DocumentElementJS
	n.View = n
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

// Alert

func (a *Alert) Show() {
	alert := js.Global().Get("alert")
	alert.Invoke("hello")
}

// Screen

func ScreenMain() Screen {
	var m Screen

	s := WindowJS.Get("screen")
	w := s.Get("width").Float()
	h := s.Get("height").Float()

	dpr := WindowJS.Get("devicePixelRatio").Float()
	m.Rect = RectMake(0, 0, w, h)
	m.Scale = dpr
	m.SoftScale = 1
	m.UsableRect = m.Rect

	return m
}
