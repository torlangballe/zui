// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type WebView struct {
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

func WebViewNew(minSize zgeo.Size, isFrame bool, makeBar bool) (webView *WebView) {
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
	v.Bar.SetSpacing(2)
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
	v.Bar.Add(zgeo.Center|zgeo.HorExpand, v.TitleLabel)

	v.Back = ImageViewNew(nil, "images/triangle-left-gray.png", zgeo.Size{20, 20})
	v.Back.SetPressedHandler(func() {
		zlog.Info("back")
	})
	v.Back.SetUsable(false)
	v.Bar.Add(zgeo.CenterLeft, v.Back)

	v.Forward = ImageViewNew(nil, "images/triangle-right-gray.png", zgeo.Size{20, 20})
	v.Forward.SetPressedHandler(func() {
		zlog.Info("forward")
	})
	v.Forward.SetUsable(false)
	v.Bar.Add(zgeo.CenterLeft, v.Forward)

	v.Refresh = ImageViewNew(nil, "images/refresh.png", zgeo.Size{18, 18})
	v.Refresh.SetPressedHandler(func() {
		v.SetURL(v.url)
		zlog.Info("refresh")
	})
	v.Bar.Add(zgeo.CenterRight, v.Refresh)

	return v.Bar
}
