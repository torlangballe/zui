package zdom

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

// https://github.com/teamortix/golang-wasm
// https://pkg.go.dev/honnef.co/go/js/dom/v2

var DocumentJS = js.Global().Get("document")
var DocumentElementJS = DocumentJS.Get("documentElement")
var WindowJS = js.Global().Get("window")

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

func GetIfFloat(e js.Value, name string, f *float64) bool {
	v := e.Get(name)
	if v.IsUndefined() {
		return false
	}
	n, err := strconv.ParseFloat(v.String(), 64)
	if err != nil {
		return false
	}
	*f = n
	return true
}

func GetBoolIfDefined(e js.Value, name string) bool {
	v := e.Get(name)
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
			zlog.Error("Unknown dot-sep part:", p)
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
	return js.Global().Get(stype).New(args...)
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
		zlog.Info("CATCH!", str, args[0].String(), this)
		if done != nil {
			done(args[0], errors.New(str))
		}
		return nil
	}))
}

func ResolveInPlace(val js.Value) (js.Value, error) {
	done := make(chan struct{})
	var resolved js.Value
	var err error
	Resolve(val, func(res js.Value, e error) {
		resolved = res
		err = e
		close(done)
	})
	<-done
	return resolved, err
}

func MakeSingleCallJSCallback(call func(this js.Value, args []js.Value) any) js.Func {
	var f js.Func
	f = js.FuncOf(func(this js.Value, args []js.Value) any {
		a := call(this, args)
		f.Release()
		return a
	})
	return f
}

func ObjectKeys(obj js.Value) []string {
	var out []string
	keys := js.Global().Get("Object").Call("keys", obj)
	len := keys.Length()
	for i := range len {
		key := keys.Index(i)
		out = append(out, key.String())
	}
	return out
}

func DebugString(obj js.Value) string {
	var out string
	for _, key := range ObjectKeys(obj) {
		val := obj.Get(key)
		str := fmt.Sprint(val)
		out += fmt.Sprintf("%s=%s\n", key, str)
	}
	return out
}

func ObjectToMap(obj js.Value) map[string]any {
	out := make(map[string]any)
	for _, key := range ObjectKeys(obj) {
		out[key] = obj.Get(key).String()
	}
	return out
}

func ObjectToJSONString(obj js.Value) string {
	return js.Global().Get("JSON").Call("stringify", obj).String()
}

// ForEach calls forEach on an an item. This is not the same as getting key/values from an object or items in an array. I think.
func ForEach(item js.Value, got func(v js.Value)) {
	jfunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		got(args[0])
		return nil
	})
	item.Call("forEach", jfunc)
	jfunc.Release()
}
