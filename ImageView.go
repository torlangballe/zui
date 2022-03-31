//go:build zui

package zui

import (
	"path"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /20/10/15.

type ImageView struct {
	ContainerView
	image              *zimage.Image
	fitSize            zgeo.Size
	alignment          zgeo.Alignment
	imageCorner        float64
	strokeWidth        float64
	strokeColor        zgeo.Color
	loading            bool
	UseDownsampleCache bool
	CapInsetCorner     zgeo.Size
}

func ImageViewNew(image *zimage.Image, imagePath string, fitSize zgeo.Size) *ImageView {
	v := &ImageView{}
	v.Init(v, image, imagePath, fitSize)
	return v
}

func (v *ImageView) Init(view View, image *zimage.Image, imagePath string, fitSize zgeo.Size) *ImageView {
	v.CustomView.Init(view, imagePath)
	v.SetFitSize(fitSize)
	v.UseDownsampleCache = true
	//	v.DownsampleImages = true
	name := "image"
	if imagePath != "" {
		_, name = path.Split(imagePath)
	}
	v.SetObjectName(name)
	v.alignment = zgeo.Center | zgeo.Proportional
	v.SetDrawHandler(v.Draw)
	if imagePath != "" {
		v.SetImage(image, imagePath, nil)
	}
	return v
}

func (v *ImageView) SetPressToShowImage(on bool) {
	if on {
		v.SetPressedHandler(func() {
			if v.image != nil {
				opts := WindowOptions{
					URL: v.image.Path,
				}
				zlog.Info("Open Image View:", v.image.Path)
				WindowOpen(opts)
			}
		})
	} else {
		v.SetPressedHandler(nil)
	}
}

func (v *ImageView) SetStroke(width float64, c zgeo.Color) {
	v.strokeWidth = width
	v.strokeColor = c
	v.Expose()
}

func (v *ImageView) SetImageCorner(radius float64) {
	v.imageCorner = radius
	v.Expose()
}

func (v *ImageView) GetImage() *zimage.Image {
	return v.image
}

func (v *ImageView) IsLoading() bool {
	return v.loading
}

func (v *ImageView) Path() string {
	if v.image != nil {
		return v.image.Path
	}
	return ""
}

func (v *ImageView) SetRect(rect zgeo.Rect) {
	v.CustomView.SetRect(rect)
	// zlog.Info("IV SR", v.Hierarchy(), p, rect)
	// zlog.Info("ImageView SetRect:", rect, v.JSGet("id"))
	if v.ObjectName() == "zap!" {
		// zlog.Info("ImageView SetRect:", rect, zlog.GetCallingStackString())
	}
}

func (v *ImageView) CalculatedSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	if v.image != nil {
		s = v.image.Size()
	}
	// zlog.Info("IV CS", v.Hierarchy(), s, p, v.image != nil, zlog.GetCallingStackString())
	margSize := v.margin.Size
	if !v.fitSize.IsNull() {
		s = v.fitSize
		// s = s.ShrunkInto(v.fitSize.Plus(margSize))
	}
	// if !v.minSize.IsNull() {
	// 	ms := v.minSize.Plus(margSize)
	// 	if s.IsNull() || v.alignment&zgeo.Proportional == 0 {
	// 		s = ms
	// 	} else {
	// 		s = s.ExpandedInto(ms)
	// 	}
	// }
	s.Add(margSize.Negative())
	s.Maximize(zgeo.Size{2, 2})
	// zlog.Info("IV CalculatedSize:", v.alignment, v.image != nil, v.MinSize(), v.FitSize(), "got:", s)
	return s
}

func (v *ImageView) FitSize() zgeo.Size {
	return v.fitSize
}

func (v *ImageView) SetFitSize(s zgeo.Size) {
	v.fitSize = s
}

func (v *ImageView) Alignment() zgeo.Alignment {
	return v.alignment
}

func (v *ImageView) SetAlignment(a zgeo.Alignment) {
	v.alignment = a
	v.Expose()
}

func (v *ImageView) SetImage(image *zimage.Image, path string, got func(i *zimage.Image)) {
	// zlog.Info("IV SetImage", path, v.JSGet("id").String(), v.Rect(), v.image != nil)
	v.JSSet("href", path)
	v.exposed = false
	if image != nil {
		v.loading = false
		v.image = image
		v.Expose()
		if got != nil {
			got(image)
		}
	} else {
		v.loading = true
		zimage.FromPath(path, func(ni *zimage.Image) {
			v.loading = false
			// zlog.Info(v.ObjectName(), "Image from path gotten:", path, ni != nil)
			// if ni != nil {
			// 	zlog.Info("IV SetImage got", path, ni.Size())
			// }
			if ni != nil && !v.CapInsetCorner.IsNull() {
				ni.SetCapInsetsCorner(v.CapInsetCorner)
			}
			v.image = ni
			v.Expose()
			if got != nil {
				got(ni)
			}
		})
	}
}

func (v *ImageView) GetImageRect(inRect zgeo.Rect) zgeo.Rect {
	a := v.alignment | zgeo.Scale | zgeo.Proportional
	// if v.IsFillBox {
	// 	a = AlignmentNone
	// }
	r := inRect.Plus(v.margin)
	ir := r.Align(v.image.Size(), a, zgeo.Size{})
	return ir
}

func (v *ImageView) Draw(rect zgeo.Rect, canvas *zcanvas.Canvas, view View) {
	canvas.DownsampleImages = v.DownsampleImages
	if v.image != nil {
		if v.IsHighlighted() {
			v.image.TintedWithColor(zgeo.ColorNewGray(0.2, 1), 1, func(ti *zimage.Image) {
				v.drawImage(canvas, ti, rect)
			})
		} else {
			v.drawImage(canvas, v.image, rect)
		}
	}
}

func (v *ImageView) drawImage(canvas *zcanvas.Canvas, img *zimage.Image, rect zgeo.Rect) {
	ir := v.GetImageRect(rect)
	if v.imageCorner != 0 {
		canvas.PushState()
		path := zgeo.PathNewRect(ir.Plus(v.Margin()), zgeo.SizeBoth(v.imageCorner))
		canvas.ClipPath(path, true, true)
	}
	if !canvas.DrawImage(img, v.UseDownsampleCache, ir, 1, zgeo.Rect{}) {
		//		v.Expose() // causes endless loop of drawing...
	}
	if v.imageCorner != 0 {
		canvas.PopState()
	}
	if v.strokeWidth != 0 {
		corner := v.imageCorner - v.strokeWidth
		path := zgeo.PathNewRect(ir.Expanded(zgeo.SizeBoth(-v.strokeWidth/2)), zgeo.SizeBoth(corner))
		canvas.SetColor(v.strokeColor)
		canvas.StrokePath(path, v.strokeWidth, zgeo.PathLineSquare)
	}
	if v.IsFocused() {
		FocusDraw(canvas, rect, 15, 0, 1)
	}
}
