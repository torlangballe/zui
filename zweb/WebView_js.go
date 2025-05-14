package zweb

import (
	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
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
	v.Element.Set("overflow", "auto")
	v.Element.Set("style", "position:absolute")
	// doc := v.Element.Get("document")
	// zlog.Info("Doc:", doc.IsUndefined())
	v.SetObjectName("webview") // must be after creation
	v.View = v

	repeater := ztimer.Repeat(0.5, func() bool {
		// zlog.Info("cddoc:", v.JSGet("contentDocument").IsUndefined(), v.Element.Get("document").IsUndefined())
		contentDoc := v.JSGet("contentDocument")
		// zlog.Info("CDOC:", contentDoc, zdom.DocumentJS)
		if !contentDoc.IsUndefined() && !contentDoc.IsNull() {
			newURL := contentDoc.Get("location").Get("href").String()
			zlog.Info("LOC:", newURL)
			if newURL != v.url {
				old := v.url
				v.url = newURL
				v.History = append(v.History, newURL)
				v.updateWidgets()
				v.setTitle()
				if v.URLChangedFunc != nil {
					v.URLChangedFunc(newURL, old)
				}
			}
		}
		return true
	})
	v.AddOnRemoveFunc(repeater.Stop)
}

func (v *WebView) updateWidgets() {
	if v.Back != nil {
		v.Back.SetUsable(len(v.History) > 1)
	}
}

func (v *WebView) SetCookies(cookieMap map[string]string) {
	if cookieMap != nil {
		str := zstr.GetArgsAsURLParameters(cookieMap)
		contentDoc := v.JSGet("contentDocument")
		zlog.Info("SetCookieDoc:", contentDoc.IsUndefined())
		contentDoc.Set("cookie", str)
		// v.Element.Get("document").Set("cookie", str)
	}
}

func (v *WebView) setTitle() {
	str := v.url
	win := zwindow.FromNativeView(&v.NativeView)
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
	if len(v.History) == 0 || zslice.Top(v.History) != surl {
		v.History = append(v.History, surl)
		v.updateWidgets()
	}
	if old != "" && v.URLChangedFunc != nil {
		v.URLChangedFunc(surl, old)
	}
	v.setTitle()
}

func (v *WebView) FetchHTMLAndSet(surl string) {
	var html string
	zlog.Info("WebView.FetchHTMLAndSet:", surl)
	params := zhttp.MakeParameters()
	_, err := zhttp.Get(surl, params, &html)
	if err != nil {
		zalert.ShowError(err)
		return
	}
	v.SetHTMLContent(html)
}

func (v *WebView) SetHTMLContent(html string) {
	v.Element.Set("innerHTML", html)
}

func (v *WebView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	h := v.JSGet("scrollHeight").Float()
	zfloat.Maximize(&h, v.minSize.H)
	return zgeo.SizeD(v.minSize.W, h), zgeo.Size{}
}
