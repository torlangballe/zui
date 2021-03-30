// +build zui

package zui

import (
	"path"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /20/10/15.

type ImageView struct {
	ContainerView
	image            *Image
	maxSize          zgeo.Size
	alignment        zgeo.Alignment
	imageCorner      float64
	strokeWidth      float64
	strokeColor      zgeo.Color
	DownsampleImages bool
}

func ImageViewNew(image *Image, imagePath string, maxSize zgeo.Size) *ImageView {
	v := &ImageView{}
	v.CustomView.Init(v, imagePath)
	v.SetMaxSize(maxSize)
	name := "image"
	if imagePath != "" {
		_, name = path.Split(imagePath)
	}
	v.SetObjectName(name)
	v.alignment = zgeo.Center | zgeo.Proportional
	v.SetDrawHandler(ImageViewDraw)
	if imagePath != "" {
		v.SetImage(image, imagePath, nil)
	}
	//        isAccessibilityElement = true
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

func (v *ImageView) SetStroke(width float64, c zgeo.Color) View {
	v.strokeWidth = width
	v.strokeColor = c
	v.Expose()
	return v
}

func (v *ImageView) SetImageCorner(radius float64) View {
	v.imageCorner = radius
	v.Expose()
	return v
}

func (v *ImageView) GetImage() *Image {
	return v.image
}

func (v *ImageView) Path() string {
	if v.image != nil {
		return v.image.Path
	}
	return ""
}

func (v *ImageView) SetRect(rect zgeo.Rect) {
	v.CustomView.SetRect(rect)
	// zlog.Info("ImageView SetRect:", rect, v.getjs("id"))
	if v.ObjectName() == "zap!" {
		// zlog.Info("ImageView SetRect:", rect, zlog.GetCallingStackString())
	}
}

func (v *ImageView) CalculatedSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	if v.image != nil {
		s = v.image.Size()
	}
	margSize := v.margin.Size
	if !v.maxSize.IsNull() {
		s = s.ShrunkInto(v.maxSize.Plus(margSize))
	}
	if !v.minSize.IsNull() {
		ms := v.minSize.Plus(margSize)
		if s.IsNull() || v.alignment&zgeo.Proportional == 0 {
			s = ms
		} else {
			s = s.ExpandedInto(ms)
		}
	}
	s.Add(margSize.Negative())
	s.Maximize(zgeo.Size{2, 2})
	// zlog.Info("IV CalculatedSize:", v.alignment, v.getjs("id"), v.image != nil, v.MinSize(), v.MaxSize(), "got:", s)
	return s
}

func (v *ImageView) MaxSize() zgeo.Size {
	return v.maxSize
}

func (v *ImageView) SetMaxSize(s zgeo.Size) *ImageView {
	v.maxSize = s
	return v
}

func (v *ImageView) Alignment() zgeo.Alignment {
	return v.alignment
}

func (v *ImageView) SetAlignment(a zgeo.Alignment) *ImageView {
	v.alignment = a
	v.Expose()
	return v
}

func (v *ImageView) SetImage(image *Image, path string, got func(i *Image)) {
	// zlog.Info("IV SetImage", path, v.getjs("id").String(), v.Rect(), v.image != nil)
	v.setjs("href", path)
	v.exposed = false
	if image != nil {
		v.image = image
		v.Expose()
		if got != nil {
			got(image)
		}
	} else {
		v.image = ImageFromPath(path, func(ni *Image) {
			// zlog.Info(v.ObjectName(), "Image from path gotten:", path, ni != nil)
			// if ni != nil {
			// 	zlog.Info("IV SetImage got", path, ni.Size())
			// }
			v.image = ni
			v.Expose()
			if got != nil {
				got(ni)
			}
		})
	}
}

func ImageViewDraw(rect zgeo.Rect, canvas *Canvas, view View) {
	v := view.(*ImageView)
	canvas.DownsampleImages = v.DownsampleImages
	if v.image != nil {
		var path *zgeo.Path
		drawImage := v.image
		if v.IsHighlighted {
			drawImage = drawImage.TintedWithColor(zgeo.ColorNewGray(0.2, 1))
		}
		// o := 1.0
		// v.ImageOpacity
		// if !v.Usable() {
		// 	o *= 0.6
		// }
		a := v.alignment | zgeo.Shrink | zgeo.Proportional
		// if v.IsFillBox {
		// 	a = AlignmentNone
		// }
		r := rect.Plus(v.margin)
		ir := r.Align(v.image.Size(), a, zgeo.Size{}, zgeo.Size{})
		// zlog.Info("IV Draw:", v.image.Size(), view.ObjectName(), r, v.image.Path, rect, "->", ir)
		if v.imageCorner != 0 {
			canvas.PushState()
			path = zgeo.PathNewRect(ir.Plus(v.Margin()), zgeo.SizeBoth(v.imageCorner))
			canvas.ClipPath(path, true, true)
		}
		// zlog.Info(v.ObjectName(), "IV.DrawImage:", v.getjs("id").String())
		// zlog.Info(v.ObjectName(), "IV.DrawImage22:", v.Rect(), v.image.imageJS.IsUndefined(), v.image.imageJS.IsNull())
		canvas.DrawImage(drawImage, ir, 1, zgeo.Rect{})
		if v.imageCorner != 0 {
			canvas.PopState()
		}
		if v.strokeWidth != 0 {
			corner := v.imageCorner - v.strokeWidth
			path := zgeo.PathNewRect(ir.Expanded(zgeo.SizeBoth(-v.strokeWidth/2)), zgeo.SizeBoth(corner))
			canvas.SetColor(v.strokeColor)
			canvas.StrokePath(path, v.strokeWidth, zgeo.PathLineSquare)
		}
	}
	if v.IsFocused() {
		FocusDraw(canvas, rect, 15, 0, 1)
	}
}
