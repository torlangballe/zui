package zgo

import "github.com/torlangballe/zutil/zgeo"

//  Created by Tor Langballe on /20/10/15.

type ImageView struct {
	ContainerView
	image     *Image
	maxSize   zgeo.Size
	alignment zgeo.Alignment
}

func ImageViewNew(path string, minSize zgeo.Size) *ImageView {
	v := &ImageView{}
	v.CustomView.init(v, path)
	v.CustomView.MinSize(minSize)
	s := zgeo.Pos{4, 1}.TimesD(ScreenMain().SoftScale)
	v.margin = zgeo.RectFromMinMax(s, s.Negative())
	v.alignment = zgeo.AlignmentCenter
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

func (v *ImageView) GetCalculatedSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	if v.image != nil {
		s = v.image.Size()
	}
	if !v.maxSize.IsNull() {
		s.Minimize(v.maxSize)
	}
	if !v.minSize.IsNull() {
		s.Maximize(v.minSize)
	}
	return s
}

func (v *ImageView) MinSize(s zgeo.Size) *ImageView {
	v.minSize = s
	return v
}

func (v *ImageView) MaxSize(s zgeo.Size) *ImageView {
	v.maxSize = s
	return v
}

func (v *ImageView) Alignment(a zgeo.Alignment) *ImageView {
	v.alignment = a
	return v
}

func ImageViewFromImage(image *Image) *ImageView {
	v := ImageViewNew("", zgeo.Size{})
	v.SetImage(image, "", nil)

	return v
}

func (v *ImageView) SetImage(image *Image, path string, got func()) {
	v.exposed = false
	if image != nil {
		v.image = image
		v.Expose()
		got()
	} else {
		v.image = ImageFromPath(path, func() {
			v.Expose()
			if got != nil {
				got()
			}
		})
	}
}

func imageViewDraw(rect zgeo.Rect, canvas *Canvas, view View) {
	v := view.(*ImageView)
	if v.image != nil {
		drawImage := v.image
		if v.IsHighlighted {
			drawImage = drawImage.TintedWithColor(zgeo.ColorNewGray(0.2, 1))
		}
		// o := 1.0
		// v.ImageOpacity
		// if !v.IsUsable() {
		// 	o *= 0.6
		// }
		a := v.alignment | zgeo.AlignmentShrink
		// if v.IsFillBox {
		// 	a = AlignmentNone
		// }
		r := rect.Plus(v.margin)
		ir := r.Align(v.image.Size(), a, zgeo.Size{}, zgeo.Size{})
		canvas.DrawImage(drawImage, ir, 1, CanvasBlendModeNormal, zgeo.Rect{})
	}
	if v.IsFocused() {
		FocusDraw(canvas, rect, 15, 0, 1)
	}
}
