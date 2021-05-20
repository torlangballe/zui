package zui

import (
	"syscall/js"
)

type nativeCodeEditorView struct {
	editor js.Value
}

func (v *CodeEditorView) Init(view View, name string) {
	// v.MakeJSElement(view, "div")
	v.Element = DocumentJS.Call("createElement", "div")
	//	v.Element.Set("style", "position:absolute")
	v.View = view
	v.Element.Set("id", "editor")
}

/*
func (v *CodeEditorView) InitEdit() {
	ace := js.Global().Get("ace")
	zlog.Info("InitEditing ace:", ace)
	v.editor = ace.Call("edit", v.Element)
	v.editor.Call("setTheme", "ace/theme/monokai")
	v.editor.Get("session").Call("setMode", "ace/mode/markdown")
	v.editor.Get("renderer").Call("setShowGutter", false)
	v.editor.Call("setValue", "Hello, world of editors.")
	// v.editor.Call("setOption", "maxLines", 22)
}

func (v *CodeEditorView) SetRect(rect zgeo.Rect) {
	if !v.editor.IsUndefined() {
		r := v.editor.Call("resize", true)
		zlog.Info("Resize", r, rect)
	}
	// v.NativeView.SetRect(rect)
}
*/
