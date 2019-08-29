package zgo

//  Created by Tor Langballe on /20/10/15.

type ImageView struct {
	ContainerView
	image     *Image
	maxSize   Size
	alignment Alignment
}

func ImageViewFromPath(path string, minSize Size) *ImageView {
	v := &ImageView{}
	v.CustomView.init(v, path)
	v.CustomView.MinSize(minSize)
	s := Pos{4, 1}.TimesD(ScreenMain().SoftScale)
	v.margin = RectFromMinMax(s, s.Negative())
	v.alignment = AlignmentCenter
	v.DrawHandler(imageViewDraw)
	if path != "" {
		v.SetImage(nil, path, nil)
	}
	//        isAccessibilityElement = true
	return v
}

func (v *ImageView) GetImage() *Image {
	return v.image
}

func (v *ImageView) GetCalculatedSize(total Size) Size {
	s := Size{0, 0}
	if v.image != nil {
		s.Maximize(v.image.Size())
	}
	s.Maximize(v.maxSize)
	if !v.minSize.IsNull() {
		s.Minimize(v.minSize)
	}
	return s
}

func (v *ImageView) MinSize(s Size) *ImageView {
	v.minSize = s
	return v
}

func (v *ImageView) MaxSize(s Size) *ImageView {
	v.maxSize = s
	return v
}

func (v *ImageView) Alignment(a Alignment) *ImageView {
	v.alignment = a
	return v
}

func ImageViewFromImage(image *Image) *ImageView {
	v := ImageViewFromPath("", Size{})
	v.SetImage(image, "", nil)

	return v
}

func (v *ImageView) SetImage(image *Image, path string, got func()) {
	if image != nil {
		v.image = image
		v.ObjectName(image.path)
		v.Expose()
		got()
	} else {
		v.ObjectName(path)
		v.image = ImageFromPath(path, func() {
			//			src := i.image.imageJS.Get("src").String()
			//			i.GetView().set("src", src)
			v.Expose()
			if got != nil {
				got()
			}
		})
	}
}

func imageViewDraw(rect Rect, canvas *Canvas, view View) {
	v := view.(*ImageView)
	if v.image != nil {
		drawImage := v.image
		if v.IsHighlighted {
			drawImage = drawImage.TintedWithColor(ColorNewGray(0.2, 1))
		}
		// o := 1.0
		// v.ImageOpacity
		// if !v.IsUsable() {
		// 	o *= 0.6
		// }
		a := v.alignment | AlignmentShrink
		// if v.IsFillBox {
		// 	a = AlignmentNone
		// }
		r := rect.Plus(v.margin)
		ir := r.Align(v.image.Size(), a, Size{}, Size{})
		canvas.DrawImage(drawImage, ir, 1, CanvasBlendModeNormal, Rect{})
	}
	if v.IsFocused() {
		FocusDraw(canvas, rect, 15, 0, 1)
	}
}
