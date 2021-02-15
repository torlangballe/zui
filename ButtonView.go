// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /14/12/17.

type ButtonView struct {
	ShapeView
}

func ButtonViewNewSimple(title, imageName string) *ButtonView {
	return ButtonViewNew(title, imageName, zgeo.Size{20, 26}, zgeo.Size{6, 12})
}

func ButtonViewNew(title, imageName string, minSize zgeo.Size, insets zgeo.Size) *ButtonView {
	b := &ButtonView{}
	b.Init(title, imageName, minSize, insets)
	return b
}

func (v *ButtonView) Init(title, imageName string, minSize zgeo.Size, insets zgeo.Size) {
	v.ShapeView.init(ShapeViewTypeNone, minSize, title)
	if insets.IsNull() {
		insets = zgeo.Size{6, 12}
	}
	v.MaxSize.H = minSize.H
	v.SetCanFocus(true)
	v.ImageAlign = zgeo.Expand | zgeo.Center
	v.SetColor(zgeo.Color{})
	v.SetImageName(imageName, insets)
	v.textInfo.Text = title
	v.textInfo.Font = FontNice(FontDefaultSize, FontStyleNormal)
	v.textInfo.Color = zgeo.ColorBlack //White
	v.textInfo.MaxLines = 1
	v.ImageMargin = zgeo.Size{}
	v.TextXMargin = 16
}

func (v *ButtonView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.ShapeView.CalculatedSize(total)
	// zlog.Info("ButtonCalc:", s, v.ObjectName(), v.image.Path, v.image.Size())
	return s
}

func (v *ButtonView) SetImageName(name string, insets zgeo.Size) {
	v.SetNamedCapImage("images/buttons/"+name, insets)
}
