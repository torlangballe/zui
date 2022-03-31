package zui

import (
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

type nativeWebView struct {
}

func (v *WebView) init(minSize zgeo.Size, isFrame bool) {
	v.minSize = minSize
	stype := "div"
	if isFrame {
		stype = "iframe"
	}
	v.Element = zdom.DocumentJS.Call("createElement", stype)
	v.Element.Set("id", "ifrm")
	if isFrame {
		v.Element.Set("allow", "encrypted-media")
		v.Element.Set("frameBorder", "0")
	}
	v.Element.Set("style", "position:absolute")
	v.SetObjectName("webview") // must be after creation
	v.View = v

	repeater := ztimer.RepeatIn(0.5, func() bool {
		// zlog.Info("cddoc:", v.JSGet("contentDocument"))
		contentDoc := v.JSGet("contentDocument")
		// zlog.Info("CDOC:", contentDoc, zdom.DocumentJS)
		if !contentDoc.IsUndefined() && !contentDoc.IsNull() {
			// zlog.Info("LOC:", contentDoc.Get("location"))
			newURL := contentDoc.Get("location").Get("href").String()
			if newURL != v.url {
				v.url = newURL
				// zlog.Info("cddoc:", newURL)
				v.History = append(v.History, newURL)
				if v.Back != nil {
					v.Back.SetUsable(true)
				}
				v.setTitle()
				if v.URLChangedHandler != nil {
					v.URLChangedHandler(newURL)
				}
			}
		}
		return true
	})
	v.AddOnRemoveFunc(repeater.Stop)
	//	v.style().Set("overflow", "hidden") // this clips the canvas, otherwise it is on top of corners etc
}

func (v *WebView) setTitle() {
	str := v.url
	win := v.GetWindow()
	if win != nil {
		// zlog.Info("SetWebViewURLTitle:", str)
		win.SetTitle(str)
	}
	if v.TitleLabel != nil {
		v.TitleLabel.SetText(str)
	}
}

func (v *WebView) SetURL(surl string) {
	v.Element.Set("src", surl)
	old := v.url
	v.url = surl
	if old != "" && v.URLChangedHandler != nil {
		v.URLChangedHandler(surl)
	}
	v.setTitle()
}

func (v *WebView) FetchHTMLAndSet(surl string) {
	var html string
	zlog.Info("WebView.FetchHTMLAndSet:", surl)
	params := zhttp.MakeParameters()
	_, err := zhttp.Get(surl, params, &html)
	if err != nil {
		AlertShowError(err)
		return
	}
	v.SetHTMLContent(html)
}

func (v *WebView) SetHTMLContent(html string) {
	v.Element.Set("innerHTML", html)
}
