// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

var WebViewDefaultBarIconSize float64 = 20

type WebView struct {
	nativeWebView
	ContainerView
	url               string
	minSize           zgeo.Size
	History           []string
	URLChangedHandler func(surl string)

	TitleLabel *Label
	Back       *ImageView
	Forward    *ImageView
	Refresh    *ImageView
	Bar        *StackView

	// Padding zgeo.Size
}

func WebViewNew(minSize zgeo.Size, isFrame, makeBar bool) (webView *WebView) {
	webView = &WebView{}
	webView.View = webView
	webView.init(minSize, isFrame)
	if makeBar {
		webView.Bar = webView.makeBar()
	}
	return webView
}

func (v *WebView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return v.minSize
}

func (v *WebView) makeBar() *StackView {
	v.Bar = StackViewHor("bar")
	v.Bar.SetSpacing(8)
	v.Bar.SetMarginS(zgeo.Size{8, 6})
	v.Bar.SetDrawHandler(func(rect zgeo.Rect, canvas *Canvas, view View) {
		colors := []zgeo.Color{zgeo.ColorNew(0.85, 0.88, 0.91, 1), zgeo.ColorNew(0.69, 0.72, 0.76, 1)}
		path := zgeo.PathNewRect(rect, zgeo.Size{})
		canvas.DrawGradient(path, colors, rect.Min(), rect.BottomLeft(), nil)
	})
	v.TitleLabel = LabelNew("")
	v.TitleLabel.SetFont(FontNew("Arial", 16, FontStyleBold))
	v.TitleLabel.SetTextAlignment(zgeo.Center)
	v.TitleLabel.SetColor(zgeo.ColorNewGray(0.3, 1))
	v.Bar.Add(v.TitleLabel, zgeo.Center|zgeo.HorExpand)

	v.Back = ImageViewNew(nil, "images/triangle-left-gray.png", zgeo.SizeBoth(WebViewDefaultBarIconSize))
	v.Back.DownsampleImages = true
	v.Back.SetPressedHandler(func() {
		zlog.Info("back")
	})
	v.Back.SetUsable(false)
	v.Bar.Add(v.Back, zgeo.CenterLeft)

	v.Forward = ImageViewNew(nil, "images/triangle-right-gray.png", zgeo.SizeBoth(WebViewDefaultBarIconSize))
	v.Forward.SetPressedHandler(func() {
		zlog.Info("forward")
	})
	v.Forward.DownsampleImages = true
	v.Forward.SetUsable(false)
	v.Bar.Add(v.Forward, zgeo.CenterLeft)

	v.Refresh = ImageViewNew(nil, "images/refresh.png", zgeo.SizeBoth(WebViewDefaultBarIconSize))
	v.Refresh.SetPressedHandler(func() {
		v.SetURL(v.url)
		// zlog.Info("refresh")
	})
	v.Refresh.DownsampleImages = true
	v.Bar.Add(v.Refresh, zgeo.CenterRight)

	return v.Bar
}

func OpenFullScreenWebViewInScreenID(screenID int64, surl string) {
	o := WindowOptions{}
	o.FullScreenID = screenID

	win := WindowOpen(o)
	size := win.Rect().Size
	wv := WebViewNew(size, false, false)
	win.AddView(wv)
	zlog.Info("Win:", win, size)
	//	win.SetTitle("Full")
	win.Activate()
	wv.SetURL(surl)
}
