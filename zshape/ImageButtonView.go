//go:build zui

package zshape

import (
	"github.com/torlangballe/zutil/zgeo"
)

// Originally created by Tor Langballe on /14/12/17.

var (
	ImageButtonViewDefaultName = "gray"
	DefaultInsets              = zgeo.SizeF(6, 10)
)

type ImageButtonView struct {
	ShapeView
}

func ImageButtonViewSimpleInsets(title, imageName string) *ImageButtonView {
	return ImageButtonViewNew(title, imageName, zgeo.SizeD(20, 22), DefaultInsets)
}

func ImageButtonViewNew(title, imageName string, minSize zgeo.Size, insets zgeo.Size) *ImageButtonView {
	b := &ImageButtonView{}
	b.Init(title, imageName, minSize, insets)
	return b
}

func (v *ImageButtonView) Init(title, imageName string, minSize zgeo.Size, insets zgeo.Size) {
	v.ShapeView.Init(v, TypeNone, minSize, title)
	v.MaxSize.H = minSize.H
	if insets.IsNull() {
		insets = DefaultInsets
	}
	v.SetCanTabFocus(false)
	v.ImageAlign = zgeo.Expand | zgeo.Center
	v.SetNativePadding(zgeo.RectFromXY2(-8, -8, 8, 8))
	v.SetColor(zgeo.Color{})
	v.SetImageName(imageName, insets)
	v.textInfo.Text = title
	v.textInfo.Font = zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	v.textInfo.Color = zgeo.ColorBlack
	v.SetMaxLines(1) // we force label update here
	v.ImageMargin = zgeo.SizeNull
}

func (v *ImageButtonView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	s, _ = v.ShapeView.CalculatedSize(total)
	return s, zgeo.SizeD(0, s.H)
}

func (v *ImageButtonView) SetImageName(name string, insets zgeo.Size) {
	if name == "" {
		name = ImageButtonViewDefaultName
	}
	if insets.IsNull() {
		insets = DefaultInsets
	}
	v.SetNamedCapImage("images/zbuttons/"+name, insets)
}
