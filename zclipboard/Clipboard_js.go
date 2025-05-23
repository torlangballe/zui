package zclipboard

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zlog"
)

func SetString(str string) {
	oldFocused := zdom.DocumentJS.Get("activeElement")
	textArea := zdom.DocumentJS.Call("createElement", "textarea")
	textArea.Set("value", str)
	textArea.Get("style").Set("position", "fixed") //avoid scrolling to bottom
	zdom.DocumentJS.Get("body").Call("appendChild", textArea)
	textArea.Call("focus")
	textArea.Call("select")
	zdom.DocumentJS.Call("execCommand", "copy")
	zdom.DocumentJS.Get("body").Call("removeChild", textArea)
	if !oldFocused.IsUndefined() {
		oldFocused.Call("focus")
	}
}

func GetString(got func(s string)) {
	clip := js.Global().Get("navigator").Get("clipboard")
	if clip.IsUndefined() {
		return
	}
	promise := clip.Call("readText")
	zdom.Resolve(promise, func(resolved js.Value, err error) {
		if err != nil {
			zlog.Error("play", err)
			return
		}
		got(resolved.String())
	})
}
