// +build js

package zgo

import (
	"fmt"
	"strconv"
	"syscall"
	"syscall/js"

	"github.com/torlangballe/zutil/ustr"
	"github.com/torlangballe/zutil/zlog"
)

type ViewNative js.Value

func init() {
}

type css js.Value

// https://github.com/siongui/godom
// https://medium.zenika.com/go-1-11-webassembly-for-the-gophers-ae4bb8b1ee03

func parseCoord(value js.Value) float64 {
	var s string
	str := value.String()
	if ustr.HasSuffix(str, "pt", &s) {
		n, err := strconv.ParseFloat(s, 32)
		if err != nil {
			zlog.Error(err, "not number")
			return 0
		}
		return n
	}
	zlog.Error(nil, "not handled type")
	return 0
}

func getCreatedTimeFromStatT(fstat *syscall.Stat_t) Time {
	return TimeNull
}

var DocumentJS = js.Global().Get("document")
var DocumentElementJS = DocumentJS.Get("documentElement")

func AddTextNode(e *ViewNative, text string) {
	textNode := DocumentJS.Call("createTextNode", text)
	e.call("appendChild", textNode)
	//	js.Value(*e).Call("appendChild", textNode)
}

func addView(parent, child *ViewNative) {
	parent.call("appendChild", js.Value(*child))
}

func AddViewToRoot(child AnyView) {
	v := child.GetView()
	DocumentElementJS.Call("appendChild", js.Value(*v))
}

func (n *ViewNative) call(method string, v js.Value) {
	js.Value(*n).Call(method, v)
}

func (n *ViewNative) set(property string, v interface{}) {
	js.Value(*n).Set(property, v)
}

func (n *ViewNative) get(property string) js.Value {
	return js.Value(*n).Get(property)
}

// Text
var measureDiv *js.Value

func (ti *TextInfo) getTextSize(noWidth bool) Size {
	// https://stackoverflow.com/questions/118241/calculate-text-width-with-javascript
	var s Size
	if measureDiv == nil {
		e := DocumentJS.Call("createElement", "div")
		DocumentElementJS.Call("appendChild", e)
		measureDiv = &e
	}
	style := measureDiv.Get("style")

	style.Set("fontSize", fmt.Sprintf("%dpx", int(ti.Font.Size)))
	style.Set("position", "absolute")
	style.Set("left", "-1000")
	style.Set("top", "-1000")
	measureDiv.Set("innerHTML", ti.Text)

	s.W = measureDiv.Get("clientWidth").Float()
	s.H = measureDiv.Get("clientHeight").Float()

	return s
}

// CustomView

func CustomViewInit() *ViewNative {
	e := DocumentJS.Call("createElement", "div")
	e.Set("style", "position:absolute")
	return ViewNative(e)
}

// Alert

func (a *Alert) Show() {
	alert := js.Global().Get("alert")
	alert.Invoke("hello")
}

// Screen

func ScreenMainRect() Rect {
	// win := js.Global().Get("window")
	// s := win.Get("screen")
	// w := s.Get("width").Float()
	// h := s.Get("height").Float()
	// fmt.Println("ScreenMainRect:", w, h)
	// return RectMake(0, 0, w, h)
	return RectMake(0, 0, 400, 400)
}

func zViewAddView(parent ViewSimple, child ViewSimple, index int) {
	parent.GetView().call("appendChild", js.Value(*child.GetView()))
}
