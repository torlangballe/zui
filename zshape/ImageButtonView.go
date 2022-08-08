//go:build zui

package zshape

import (
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /14/12/17.

var ImageButtonViewDefaultName = "gray"

type ImageButtonView struct {
	ShapeView
}

func ImageButtonViewNewSimple(title, imageName string) *ImageButtonView {
	return ImageButtonViewNew(title, imageName, zgeo.Size{20, 22}, zgeo.Size{6, 10})
}

func ImageButtonViewNew(title, imageName string, minSize zgeo.Size, insets zgeo.Size) *ImageButtonView {
	b := &ImageButtonView{}
	b.Init(title, imageName, minSize, insets)
	return b
}

func (v *ImageButtonView) Init(title, imageName string, minSize zgeo.Size, insets zgeo.Size) {
	v.ShapeView.Init(v, TypeNone, minSize, title)
	if insets.IsNull() {
		insets = zgeo.Size{6, 10}
	}
	v.MaxSize.H = minSize.H
	v.SetCanFocus(true)
	v.ImageAlign = zgeo.Expand | zgeo.Center
	v.SetColor(zgeo.Color{})
	v.SetImageName(imageName, insets)
	v.textInfo.Text = title
	v.textInfo.Font = zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	// zlog.Info("ImGButton:", v.Hierarchy(), v.textInfo.Font)
	v.textInfo.Color = zgeo.ColorBlack //White
	v.textInfo.MaxLines = 1
	v.ImageMargin = zgeo.Size{}
}

func (v *ImageButtonView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.ShapeView.CalculatedSize(total)
	// zlog.Info("ButtonCalc:", s, v.ObjectName(), v.textInfo.Text, v.image.Path, v.image.Size())
	return s
}

func (v *ImageButtonView) SetImageName(name string, insets zgeo.Size) {
	if name == "" {
		name = ImageButtonViewDefaultName
	}
	v.SetNamedCapImage("images/buttons/"+name, insets)
}
