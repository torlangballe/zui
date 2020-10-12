package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /14/12/17.

type Button struct {
	ShapeView
}

func ButtonNewSimple(title, imageName string) *Button {
	return ButtonNew(title, imageName, zgeo.Size{20, 26}, zgeo.Size{6, 12})
}

func ButtonNew(title, imageName string, minSize zgeo.Size, insets zgeo.Size) *Button {
	b := &Button{}
	b.ShapeView.init(ShapeViewTypeNone, minSize, title)
	if insets.IsNull() {
		insets = zgeo.Size{6, 12}
	}
	b.SetCanFocus(true)
	b.ImageAlign = zgeo.Expand | zgeo.Center
	b.SetColor(zgeo.Color{})
	b.SetImageName(imageName, insets)
	b.textInfo.Text = title
	b.textInfo.Font = FontNice(FontDefaultSize, FontStyleNormal)
	b.textInfo.Color = zgeo.ColorBlack //White
	b.textInfo.MaxLines = 1
	b.ImageMargin = zgeo.Size{}
	b.TextXMargin = 16
	return b
}

func (v *Button) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.ShapeView.CalculatedSize(total)
	if v.image != nil {
		if true { //v.image.path == "images/buttons/gray@2x.png" {
			// zlog.Info("ButtonCalc:", s, v.ObjectName(), v.image.path, v.image.Size())
		}
		s.Maximize(v.image.Size())
	}
	return s
}

func (b *Button) SetImageName(name string, insets zgeo.Size) {
	s := ""
	if ScreenMain().Scale >= 2 {
		s = "@2x"
	}
	str := "images/buttons/" + name + s + ".png"

	// zlog.Info("SetImageButtonName:", str)
	cimage := b.SetImage(nil, str, func() {
		if b.image.Size().W < insets.W*2 || b.image.Size().H < insets.H*2 {
			zlog.Error(nil, "Button: Small image for inset:", b.ObjectName(), name, b.image.Size(), insets)
			return
		}
	})
	cimage.SetCapInsets(zgeo.RectFromMinMax(insets.Pos(), insets.Pos().Negative()))
}
