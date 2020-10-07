package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

func WebViewNew(minSize zgeo.Size, isFrame bool) *WebView {
	v := &WebView{}
	v.View = v
	v.init(minSize, isFrame)
	return v
}

func (v *WebView) init(minSize zgeo.Size, isFrame bool) {
	v.minSize = minSize
	stype := "div"
	if isFrame {
		stype = "iframe"
	}
	v.Element = DocumentJS.Call("createElement", stype)
	v.Element.Set("id", "ifrm")
	if isFrame {
		v.Element.Set("allow", "encrypted-media")
	}
	v.Element.Set("style", "position:absolute")
	v.SetObjectName("webview") // must be after creation
	v.View = v
	//	v.style().Set("overflow", "hidden") // this clips the canvas, otherwise it is on top of corners etc
}

func (v *WebView) SetURL(surl string) {
	v.Element.Set("src", surl)
}

func (v *WebView) SetHTMLContent(html string) {
	v.Element.Set("innerHTML", html)
}
