package zui

//  Created by Tor Langballe on /22/10/15.

import (
	"fmt"
	"math"
	"path"

	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
)

type ShapeViewType string

const (
	ShapeViewTypeCircle    ShapeViewType = "circle"
	ShapeViewTypeRectangle               = "rectange"
	ShapeViewTypeRoundRect               = "roundrect"
	ShapeViewTypeStar                    = "star"
	ShapeViewTypeNone                    = ""
)

type ShapeView struct {
	ContainerView
	image        *Image
	Type         ShapeViewType
	StrokeWidth  float64
	TextInfo     TextInfo
	ImageMargin  zgeo.Size //  ZSize(4.0, 1.0) * ZScreen.SoftScale
	TextXMargin  float64
	IsImageFill  bool       //
	ImageOpacity float32    // Float = Float(1)
	Ratio        float32    // = 0.3
	Count        int        // = 5
	StrokeColor  zgeo.Color // = ZColor.White()
	MaxWidth     float64
	ImageAlign   zgeo.Alignment // .Center
	IsFillBox    bool
	IsRoundImage bool
	Value        float64
	PathLineType zgeo.PathLineType
	Proportional bool
}

func ShapeViewNew(shapeType ShapeViewType, minSize zgeo.Size) *ShapeView {
	v := &ShapeView{}
	v.init(shapeType, minSize, string(shapeType))
	return v
}

func (v *ShapeView) init(shapeType ShapeViewType, minSize zgeo.Size, name string) {
	v.CustomView.init(v, name)
	v.TextInfo = *TextInfoNew()
	v.Type = shapeType
	v.ImageMargin = zgeo.Size{4, 1}.TimesD(ScreenMain().SoftScale)
	v.ImageOpacity = 1
	v.Count = 5
	v.StrokeColor = zgeo.ColorWhite
	v.ImageAlign = zgeo.Center
	v.PathLineType = zgeo.PathLineRound
	v.SetColor(zgeo.ColorGray)
	v.Proportional = true

	switch shapeType {
	case ShapeViewTypeRoundRect:
		v.Ratio = 0.495
	case ShapeViewTypeStar:
		v.Ratio = 0.6
	default:
		v.Ratio = 0.3
	}
	v.SetDrawHandler(shapeViewDraw)
	v.CustomView.SetMinSize(minSize)
}

// Text sets the ShapeView's TextInfo.Text string, and exposes. This is also here to avoid underlying NativeView SetText() method being used
func (v *ShapeView) SetText(text string) View {
	v.TextInfo.Text = text
	v.Expose()
	return v
}

func (v *ShapeView) GetImage() *Image {
	return v.image
}

func (v *ShapeView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.MinSize()
	if v.TextInfo.Text != "" {
		ts := v.TextInfo.GetBounds(false).Size
		ts.Add(zgeo.Size{16, 6})
		ts.W *= 1.1 // some strange bug in android doesn't allow *= here...
		s.Maximize(ts)
	}
	s.Add(v.margin.Size.Negative())
	if v.MaxWidth != 0.0 {
		zfloat.Maximize(&s.W, v.MaxWidth)
	}
	if v.Type == ShapeViewTypeCircle {
		//		zmath.Float64Maximize(&s.H, s.W)
	}
	s.MakeInteger()
	return s
}

func (v *ShapeView) SetImage(image *Image, spath string, done func()) *Image {
	v.image = image
	v.exposed = false
	if v.ObjectName() == "" {
		_, name := path.Split(spath)
		v.SetObjectName(name)
	}
	if image == nil && spath != "" {
		v.image = ImageFromPath(spath, func() {
			// println("sv image loaded: " + spath + ": " + v.ObjectName())
			v.Expose()
			if done != nil {
				done()
			}
		})
	}
	return v.image
}

func shapeViewDraw(rect zgeo.Rect, canvas *Canvas, view View) {
	path := zgeo.PathNew()
	v := view.(*ShapeView)
	// fmt.Println("shapeViewDraw:", v.canvas != nil, v.GetMinSize(), rect, view.ObjectName())

	switch v.Type {
	case ShapeViewTypeStar:
		path.AddStar(rect, v.Count, v.Ratio)

	case ShapeViewTypeCircle:
		s := rect.Size.MinusD(v.StrokeWidth).DividedByD(2).TimesD(ScreenMain().SoftScale).MinusD(1)
		w := s.Min()
		path.ArcDegFromCenter(rect.Center(), zgeo.Size{w, w}, 0, 360)

	case ShapeViewTypeRoundRect:
		r := rect.Expanded(zgeo.Size{-1, -1}.TimesD(ScreenMain().SoftScale))
		corner := math.Min(math.Min(r.Size.W, r.Size.H)*float64(v.Ratio), 15)
		path.AddRect(r, zgeo.Size{corner, corner})

	case ShapeViewTypeRectangle:
		path.AddRect(rect, zgeo.Size{})
	}
	col := v.color
	if col.Valid {
		var o = col.Opacity()
		if !v.IsUsable() {
			o *= 0.6
		}
		canvas.SetColor(v.getStateColor(col), o)
		canvas.FillPath(path)
	}
	if v.StrokeWidth != 0 {
		fmt.Println("shapeViewDraw stroke:", view.ObjectName())
		var o = v.StrokeColor.Opacity()
		if !v.IsUsable() {
			o *= 0.6
		}
		canvas.SetColor(v.getStateColor(v.StrokeColor), o)
		canvas.StrokePath(path, v.StrokeWidth, v.PathLineType)
	}
	imarg := v.ImageMargin
	// if IsTVBox() {
	// 	imarg.Maximize(Size{5.0, 5.0}.TimesD(ScreenMain().SoftScale))
	// }
	textRect := rect
	if v.image != nil {
		drawImage := v.image
		if v.IsHighlighted {
			drawImage = drawImage.TintedWithColor(zgeo.ColorNewGray(0.2, 1))
		}
		o := v.ImageOpacity
		if !v.IsUsable() {
			o *= 0.6
		}
		if v.IsImageFill {
			canvas.PushState()
			canvas.ClipPath(path, false, false)
			canvas.DrawImage(drawImage, rect, o, CanvasBlendModeNormal, zgeo.Rect{})
			canvas.PopState()
		} else {
			a := v.ImageAlign | zgeo.Shrink
			// if v.IsFillBox {
			// 	a = AlignmentNone
			// }
			ir := rect.Align(v.image.Size(), a, v.ImageMargin, zgeo.Size{})
			var corner float64
			if v.IsRoundImage {
				if v.Type == ShapeViewTypeRoundRect {
					corner = math.Min(15, rect.Size.Min()*float64(v.Ratio)) - imarg.Min()
				} else if v.Type == ShapeViewTypeCircle {
					corner = v.image.Size().Max() / 2
				}
				clipPath := zgeo.PathNewFromRect(ir, zgeo.Size{corner, corner})
				canvas.PushState()
				canvas.ClipPath(clipPath, false, false)
			}
			textRect = ir
			canvas.DrawImage(drawImage, ir, o, CanvasBlendModeNormal, zgeo.Rect{})
			if v.IsRoundImage {
				canvas.PopState()
			}
		}
	}
	if v.TextInfo.Text != "" {
		t := v.TextInfo // .Copy()
		t.Color = v.getStateColor(t.Color)
		t.Rect = textRect.Expanded(zgeo.Size{-v.TextXMargin * ScreenMain().SoftScale, 0})
		t.Rect.Pos.Y -= 2
		if v.IsImageFill {
			canvas.SetDropShadow(zgeo.Size{}, 2, zgeo.ColorBlack)
		}
		t.Draw(canvas)
		if v.IsImageFill {
			canvas.SetDropShadowOff(1)
		}
	}
	if v.IsFocused() {
		FocusDraw(canvas, rect, 15, 0, 1)
	}
}