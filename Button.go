package zgo

import (
	"fmt"

	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /14/12/17.

type Button struct {
	ShapeView
}

func ButtonNew(title, imageName string, size zgeo.Size, insets zgeo.Size) *Button {
	b := &Button{}
	b.ShapeView.init(ShapeViewTypeNone, size, title)
	if insets.IsNull() {
		insets = zgeo.Size{6, 13}
	}
	b.CanFocus(true)
	b.ImageAlign = zgeo.AlignmentExpand | zgeo.AlignmentCenter | zgeo.AlignmentNonProp
	b.Color(zgeo.Color{})
	b.SetImageName(imageName, insets)
	b.TextInfo.Text = title
	b.TextInfo.Font = FontNice(FontDefaultSize, FontStyleNormal)
	b.TextInfo.Color = zgeo.ColorBlack //White
	b.ImageMargin = zgeo.Size{}
	b.GetImage()
	b.TextXMargin = 16
	return b
}

func (b *Button) SetImageName(name string, insets zgeo.Size) {
	s := ""
	if ScreenMain().Scale > 1 {
		s = fmt.Sprintf("@%dx", int(ScreenMain().Scale))
	}
	cimage := b.SetImage(nil, "buttons/"+name+s+".png", nil)
	cimage.CapInsets(zgeo.RectFromMinMax(insets.Pos(), insets.Pos().Negative()))
}
