//go:build zui

package zweb

import (
	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

var DefaultBarIconSize float64 = 20

type WebView struct {
	nativeWebView
	zcontainer.ContainerView
	url               string
	minSize           zgeo.Size
	History           []string
	URLChangedHandler func(surl string)

	TitleLabel *zlabel.Label
	Back       *zimageview.ImageView
	Forward    *zimageview.ImageView
	Refresh    *zimageview.ImageView
	Bar        *zcontainer.StackView

	// Padding zgeo.Size
}

func NewView(minSize zgeo.Size, isFrame, makeBar bool) (webView *WebView) {
	webView = &WebView{}
	webView.View = webView
	webView.init(minSize, isFrame)
	if makeBar {
		webView.Bar = webView.MakeBar()
	}
	return webView
}

func (v *WebView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return v.minSize
}

func (v *WebView) MakeBar() *zcontainer.StackView {
	v.Bar = zcontainer.StackViewHor("bar")
	v.Bar.SetSpacing(8)
	v.Bar.SetMarginS(zgeo.Size{8, 6})
	v.Bar.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
		colors := []zgeo.Color{zgeo.ColorNew(0.85, 0.88, 0.91, 1), zgeo.ColorNew(0.69, 0.72, 0.76, 1)}
		path := zgeo.PathNewRect(rect, zgeo.Size{})
		canvas.DrawGradient(path, colors, rect.Min(), rect.BottomLeft(), nil)
	})
	v.TitleLabel = zlabel.New("")
	v.TitleLabel.SetFont(zgeo.FontNew("Arial", 16, zgeo.FontStyleBold))
	v.TitleLabel.SetTextAlignment(zgeo.Center)
	v.TitleLabel.SetColor(zgeo.ColorNewGray(0.3, 1))
	v.Bar.Add(v.TitleLabel, zgeo.Center|zgeo.HorExpand)

	// v.Back = ImageViewNew(nil, "images/triangle-left-gray.png", zgeo.SizeBoth(DefaultBarIconSize))
	// v.Back.DownsampleImages = true
	// v.Back.SetPressedHandler(func() {
	// 	zlog.Info("back")
	// })
	// v.Back.SetUsable(false)
	// v.Bar.Add(v.Back, zgeo.CenterLeft)
	// v.Forward = ImageViewNew(nil, "images/triangle-right-gray.png", zgeo.SizeBoth(DefaultBarIconSize))
	// v.Forward.SetPressedHandler(func() {
	// 	zlog.Info("forward")
	// })
	// v.Forward.DownsampleImages = true
	// v.Forward.SetUsable(false)
	// v.Bar.Add(v.Forward, zgeo.CenterLeft)

	if zui.DebugOwnerMode {
		v.Refresh = zimageview.New(nil, "images/refresh.png", zgeo.SizeBoth(DefaultBarIconSize))
		v.Refresh.SetPressedHandler(func() {
			v.SetURL(v.url)
			// zlog.Info("refresh")
		})
		v.Refresh.DownsampleImages = true
		v.Bar.Add(v.Refresh, zgeo.CenterRight)
	}
	return v.Bar
}

func OpenFullScreenWebViewInScreenID(screenID int64, surl string) {
	o := zwindow.Options{}
	o.FullScreenID = screenID

	win := zwindow.Open(o)
	size := win.Rect().Size
	wv := NewView(size, false, false)
	win.AddView(wv)
	zlog.Info("Win:", win, size)
	//	win.SetTitle("Full")
	win.Activate()
	wv.SetURL(surl)
}