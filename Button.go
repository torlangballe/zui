package zgo

import "fmt"

//  Created by Tor Langballe on /14/12/17.

type Button struct {
	ShapeView
}

func ButtonNew(title, colorName string, size Size, insets Size) *Button {
	b := &Button{}
	b.ShapeView.init(ShapeViewTypeNone, size)
	if insets.IsNull() {
		insets = Size{6, 13}
	}
	b.CanFocus(true)
	b.ImageAlign = AlignmentHorExpand | AlignmentCenter | AlignmentNonProp
	b.Color(Color{})
	b.SetNamedColor(colorName, insets)
	b.TextInfo.Text = title
	b.TextInfo.Font = FontNice(18, FontStyleNormal)
	b.TextInfo.Color = ColorWhite
	//    b.FillBox = true
	b.ImageMargin = Size{0, 5}.TimesD(ScreenMain().SoftScale)
	b.GetImage()
	return b
}

func (b *Button) SetNamedColor(col string, insets Size) {
	s := ""
	if ScreenMain().Scale > 1 {
		s = fmt.Sprintf("@%dx", int(ScreenMain().Scale))
	}
	cimage := b.SetImage(nil, col+"Button"+s+".png", nil)
	cimage.CapInsets(RectFromMinMax(insets.Pos(), insets.Pos().Negative()))
}
