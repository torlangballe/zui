package zclipboard

import "github.com/torlangballe/zui/zdom"

func SetString(str string) {
	textArea := zdom.DocumentJS.Call("createElement", "textarea")
	textArea.Set("value", str)
	textArea.Get("style").Set("position", "fixed") //avoid scrolling to bottom
	zdom.DocumentJS.Get("body").Call("appendChild", textArea)
	textArea.Call("focus")
	textArea.Call("select")
	zdom.DocumentJS.Call("execCommand", "copy")
	zdom.DocumentJS.Get("body").Call("removeChild", textArea)
}

// "clipboard-read" permission.
// https://web.dev/async-clipboard/
