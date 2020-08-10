package zui

import (
	"fmt"
	"strings"
	"syscall"
	"syscall/js"
	"time"

	"github.com/torlangballe/zutil/zlog"
)

// https://github.com/siongui/godom
// https://medium.zenika.com/go-1-11-webassembly-for-the-gophers-ae4bb8b1ee03

var DocumentJS = js.Global().Get("document")
var DocumentElementJS = DocumentJS.Get("documentElement")
var WindowJS = js.Global().Get("window")

type css js.Value

func init() {
	if DeviceWasmBrowser() != "Safari" {
		menuViewHeight = 25
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

func jsAddEventListener(e js.Value, name string, handler func()) {
	err := e.Call("addEventListener", name, js.FuncOf(func(js.Value, []js.Value) interface{} {
		zlog.Info("event listener")
		return nil
	}), false)
	if !err.IsUndefined() {
		zlog.Info("jsAddEventListener err:", err)
	}
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
