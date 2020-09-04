package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

func WebViewNew(minSize zgeo.Size, content string, isURL bool) *WebView {
	v := &WebView{}
	v.View = v
	v.Element = DocumentJS.Call("createElement", "iframe")
	v.Element.Set("id", "ifrm")
	v.Element.Set("style", "position:absolute")
	v.Element.Set("allow", "encrypted-media")
	if isURL {
		v.Element.Set("src", content)
	} else {
		v.Element.Set("innerHTML", content)
	}
	v.SetObjectName("webview") // must be after creation
	//	v.style().Set("overflow", "hidden") // this clips the canvas, otherwise it is on top of corners etc
	v.View = v
	return v
}
