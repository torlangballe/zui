package zui

func ClipboardSetString(str string) {
	textArea := DocumentJS.Call("createElement", "textarea")
	textArea.Set("value", str)
	textArea.Get("style").Set("position", "fixed") //avoid scrolling to bottom
	DocumentJS.Get("body").Call("appendChild", textArea)
	textArea.Call("focus")
	textArea.Call("select")
	DocumentJS.Call("execCommand", "copy")
	DocumentJS.Get("body").Call("removeChild", textArea)
}
