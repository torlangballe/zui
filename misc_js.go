package zui

import (
	"fmt"
	"strings"
	"syscall"
	"syscall/js"
	"time"

	"github.com/torlangballe/zutil/zdevice"
)

// https://github.com/siongui/godom
// https://medium.zenika.com/go-1-11-webassembly-for-the-gophers-ae4bb8b1ee03
// https://github.com/golang/go/wiki/WebAssembly

var DocumentJS = js.Global().Get("document")
var DocumentElementJS = DocumentJS.Get("documentElement")
var WindowJS = js.Global().Get("window")

// type css js.Value

func init() {
	if zdevice.OS() == zdevice.MacOSType && zdevice.WasmBrowser() == "safari" {
		FontDefaultName = "-apple-system"
	}
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

func jsSetKeyHandler(e js.Value, handler func(key KeyboardKey, mods KeyboardModifier) bool) {
	//!!	v.keyPressed = handler
	e.Set("onkeyup", js.FuncOf(func(val js.Value, args []js.Value) interface{} {
		// zlog.Info("KeyUp")
		if handler != nil {
			event := args[0]
			key, mods := getKeyAndModsFromEvent(event)
			if handler(key, mods) {
				event.Call("stopPropagation")
			}
			// zlog.Info("KeyUp:", key, mods)
		}
		return nil
	}))
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

func jsCreateDotSeparatedObject(f string) js.Value {
	parent := js.Global()
	parts := strings.Split(f, ".")
	for _, p := range parts {
		parent = parent.Get(p)
	}
	return parent
}
