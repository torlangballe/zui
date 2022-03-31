package zdom

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

var DocumentJS = js.Global().Get("document")
var DocumentElementJS = DocumentJS.Get("documentElement")
var WindowJS = js.Global().Get("window")

func GetFontStyle(font *zgeo.Font) string {
	var parts []string
	if font.Style&zgeo.FontStyleBold != 0 {
		parts = append(parts, "bold")
	}
	if font.Style&zgeo.FontStyleItalic != 0 {
		parts = append(parts, "italic")
	}
	parts = append(parts, fmt.Sprintf("%dpx", int(font.Size)))
	parts = append(parts, font.Name)

	return strings.Join(parts, " ")
}

// func jsAddEventListener(e js.Value, name string, handler func(this func(js.Value, vals []js.Value)) {
// 	err := e.Call("addEventListener", name, js.FuncOf(this func(js.Value, vals []js.Value) interface{} {
// 		handler(this, vals)
// 		zlog.Info("event listener")
// 		return nil
// 	}), false)
// 	if !err.IsUndefined() {
// 		zlog.Info("jsAddEventListener err:", err)
// 	}
// }

func GetBoolIfDefined(e js.Value, get string) bool {
	v := e.Get(get)
	if v.IsUndefined() {
		return false
	}
	return v.Bool()
}

func CreateDotSeparatedObject(f string) js.Value {
	parent := js.Global()
	parts := strings.Split(f, ".")
	for _, p := range parts {
		parent = parent.Get(p)
	}
	return parent
}

func MakeRGBAString(c zgeo.Color) string {
	if !c.Valid {
		return "initial"
	}
	return c.Hex()
	//	rgba := c.GetRGBA()
	//	return fmt.Sprintf("rgba(%d,%d,%d,%g)", int(rgba.R*255), int(rgba.G*255), int(rgba.B*255), rgba.A)
}
