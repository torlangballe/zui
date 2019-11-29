package zgo

import (
	"fmt"
)

//  Created by Tor Langballe on /14/12/17.

type Button struct {
	ShapeView
}

func ButtonNew(title, imageName string, size Size, insets Size) *Button {
	b := &Button{}
	b.ShapeView.init(ShapeViewTypeNone, size, title)
	if insets.IsNull() {
		insets = Size{6, 13}
	}
	b.CanFocus(true)
	b.ImageAlign = AlignmentExpand | AlignmentCenter | AlignmentNonProp
	b.Color(Color{})
	b.SetImageName(imageName, insets)
	b.TextInfo.Text = title
	b.TextInfo.Font = FontNice(FontDefaultSize, FontStyleNormal)
	b.TextInfo.Color = ColorBlack //White
	b.ImageMargin = Size{}
	b.GetImage()
	b.TextXMargin = 16
	return b
}

func (b *Button) SetImageName(name string, insets Size) {
	s := ""
	if ScreenMain().Scale > 1 {
		s = fmt.Sprintf("@%dx", int(ScreenMain().Scale))
	}
	cimage := b.SetImage(nil, "buttons/"+name+s+".png", nil)
	cimage.CapInsets(RectFromMinMax(insets.Pos(), insets.Pos().Negative()))
}
