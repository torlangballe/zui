//go:build zui

package zimageview

import (
	"path"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zfocus"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zstr"
)

//  Created by Tor Langballe on /20/10/15.

type ImageView struct {
	zcontainer.ContainerView
	image              *zimage.Image
	fitSize            zgeo.Size
	alignment          zgeo.Alignment
	imageCorner        float64
	strokeWidth        float64
	strokeColor        zgeo.Color
	strokeInset        bool
	loading            bool
	UseDownsampleCache bool
	CapInsetCorner     zgeo.Size
	EmptyColor         zgeo.Color
	TintColor          zgeo.Color
	MixColor           zgeo.Color // Alpha is used to specify amount to mix, and used as 1
	LoadedFunc         func(img *zimage.Image)
}

func NewWithCachedPath(imagePath string, fitSize zgeo.Size) *ImageView {
	return New(nil, true, imagePath, fitSize)
}

func New(image *zimage.Image, useCache bool, imagePath string, fitSize zgeo.Size) *ImageView {
	v := &ImageView{}
	v.Init(v, useCache, image, imagePath, fitSize)
	return v
}

func (v *ImageView) Init(view zview.View, useCache bool, image *zimage.Image, imagePath string, fitSize zgeo.Size) {
	v.CustomView.Init(view, imagePath)
	v.SetFitSize(fitSize)
	v.UseDownsampleCache = useCache
	v.DownsampleImages = true
	v.SetSelectable(false)
	name := "image"
	if imagePath != "" {
		_, name = path.Split(imagePath)
	}
	v.SetObjectName(name)
	v.alignment = zgeo.Center | zgeo.Proportional
	v.SetDrawHandler(v.Draw)
	if imagePath != "" {
		v.SetImage(image, imagePath, func(i *zimage.Image) {
			if v.LoadedFunc != nil {
				v.LoadedFunc(i)
			}
		})
	} else {
		v.Expose()
	}
}

var PresentTitledViewFunc func(title string, vuew zview.View)

func (v *ImageView) SetPressToShowImage(on bool) {
	const id = "$press.to.show"
	if on {
		v.SetPressedHandler(id, zkeyboard.ModifierNone, func() {
			if v.image != nil {
				path := v.image.Path
				if !zhttp.StringStartsWithHTTPX(path) {
					path = zstr.Concat("/", zrest.AppURLPrefix, path)
				}
				nv := New(v.image, v.UseDownsampleCache, v.image.Path, zgeo.SizeNull)
				PresentTitledViewFunc(path, nv)
			}
		})
	} else {
		v.SetPressedHandler(id, zkeyboard.ModifierNone, nil)
	}
}

func (v *ImageView) SetStroke(width float64, c zgeo.Color, inset bool) {
	v.strokeWidth = width
	v.strokeColor = c
	v.strokeInset = inset
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

func (v *ImageView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	if v.image != nil {
		s = v.image.Size()
	}
	margSize := v.Margin().Size
	if !v.fitSize.IsNull() {
		s = v.fitSize
	}
	s.Add(margSize.Negative())
	s.Maximize(zgeo.SizeD(2, 2))
	// zlog.Info("IV CalcS:", v.Hierarchy(), total, s, v.fitSize, margSize)
	return s, s
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
	v.SetExposed(false)
	if image != nil {
		v.loading = false
		v.image = image
		v.Expose()
		if got != nil {
			got(image)
		}
	} else {
		// zlog.Info("SetImagePath:", v.ObjectName(), path)
		v.loading = true

		zimage.FromPath(path, v.UseDownsampleCache, func(ni *zimage.Image) {
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
	a := v.alignment | zgeo.Scale //| zgeo.Proportional
	// if v.IsFillBox {
	// 	a = AlignmentNone
	// }
	r := inRect.Plus(v.Margin())
	ir := r.Align(v.image.Size(), a, zgeo.SizeNull)
	// zlog.Info(zgeo.RectMarginForSizeAndAlign(zgeo.SizeNull, a), "GetImageRect:", v.Hierarchy(), v.Path(), v.Margin(), a, inRect, "->", r, ir)
	// zlog.Info("IV ImageRect:", v.Rect().Size, inRect, v.Hierarchy(), v.Path, ir, a)
	return ir
}

func (v *ImageView) Draw(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	if rect.Size.Area() == 0 {
		return
	}
	canvas.DownsampleImages = v.DownsampleImages
	if v.image != nil {
		col := v.TintColor
		if v.IsHighlighted() {
			col = zgeo.ColorNewGray(0.2, 1).Mixed(col, 0.5)
		}
		if col.Valid {
			v.image.TintedWithColor(col, 1, func(ti *zimage.Image) { // we tint with 1 because we assume amount is in alpha of col
				v.drawImage(canvas, ti, rect)
			})
		} else if v.MixColor.Valid {
			col = v.MixColor
			amount := col.Colors.A
			col.Colors.A = 1
			v.image.MixedWithColor(col, amount, func(ti *zimage.Image) { // we tint with 1 because we assume amount is in alpha of col
				v.drawImage(canvas, ti, rect)
			})
		} else {
			v.drawImage(canvas, v.image, rect)
		}
		return
	}
	if v.EmptyColor.Valid {
		a := v.alignment | zgeo.Shrink | zgeo.Proportional
		r := rect.Plus(v.Margin())
		ir := r.Align(r.ExpandedD(-2).Size, a, zgeo.SizeNull)
		path := zgeo.PathNewRect(ir, zgeo.SizeBoth(4))
		canvas.SetColor(v.EmptyColor)
		canvas.FillPath(path)
	}
}

func (v *ImageView) drawImage(canvas *zcanvas.Canvas, img *zimage.Image, rect zgeo.Rect) {
	ir := v.GetImageRect(rect)
	if v.imageCorner != 0 {
		canvas.PushState()
		path := zgeo.PathNewRect(ir.Plus(v.Margin()), zgeo.SizeBoth(v.imageCorner))
		canvas.ClipPath(path, true)
	}
	if ir.Size.W == 0 {
		zlog.Info("IV.drawImage null:", v.Hierarchy(), v.Rect(), v.fitSize)
	}
	if !canvas.DrawImage(img, v.UseDownsampleCache, ir, 1, zgeo.Rect{}) {
		//		v.Expose() // causes endless loop of drawing...
	}
	if v.imageCorner != 0 {
		canvas.PopState()
	}
	if v.strokeWidth != 0 {
		corner := v.imageCorner - v.strokeWidth
		path := zgeo.PathNewRect(ir.Expanded(zgeo.SizeBoth(-v.strokeWidth/2+1)), zgeo.SizeBoth(corner))
		canvas.SetColor(v.strokeColor)
		canvas.StrokePath(path, v.strokeWidth, zgeo.PathLineSquare)
	}
	if v.IsFocused() {
		zfocus.Draw(canvas, rect, 15, 0, 1)
	}
}
