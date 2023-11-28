package zdom

import (
	"errors"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

// https://github.com/teamortix/golang-wasm

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
		if parent.IsUndefined() {
			zlog.Error(nil, "Unknown dot-sep part:", p)
			return js.Undefined()
		}
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

func New(stype string, args ...any) js.Value {
	return js.Global().Get(stype).New(args)
}

func JSFileToGo(file js.Value, got func(data []byte, name string), progress func(p float64)) {
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

func Resolve(val js.Value, done func(resolved js.Value, err error)) {
	then := val.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if done != nil {
			done(args[0], nil)
		}
		return nil
	}))

	then.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		str := fmt.Sprint(args[0].Call("toString").String()) // ???
		zlog.Info("CATCH!", args[0].String(), this)
		if done != nil {
			done(js.Undefined(), errors.New(str))
		}
		return nil
	}))
}
