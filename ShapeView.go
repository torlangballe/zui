package zui

//  Created by Tor Langballe on /22/10/15.

import (
	"math"
	"path"

	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
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
	textInfo     TextInfo
	ImageMargin  zgeo.Size //  ZSize(4.0, 1.0) * ZScreen.SoftScale
	TextXMargin  float64
	IsImageFill  bool       //
	ImageOpacity float32    // Float = Float(1)
	Ratio        float32    // = 0.3
	Count        int        // = 5
	StrokeColor  zgeo.Color // = ZColor.White()
	maxWidth     float64
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
	v.textInfo = *TextInfoNew()
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
	f := FontNice(FontDefaultSize, FontStyleNormal)
	v.SetFont(f)
}

func (v *ShapeView) SetColor(c zgeo.Color) View {
	v.textInfo.Color = c
	return v
}

func (v *ShapeView) Color() zgeo.Color {
	return v.textInfo.Color
}

// Text sets the ShapeView's TextInfo.Text string, and exposes. This is also here to avoid underlying NativeView SetText() method being used
func (v *ShapeView) SetText(text string) View {
	v.textInfo.Text = text
	v.Expose()
	return v
}

func (v *ShapeView) Text() string {
	return v.textInfo.Text
}

func (v *ShapeView) SetTextAlignment(a zgeo.Alignment) View {
	v.textInfo.Alignment = a
	return v
}

func (v *ShapeView) MinWidth() float64 {
	return v.MinWidth()
}

func (v *ShapeView) MaxWidth() float64 {
	return v.maxWidth
}

func (v *ShapeView) MaxLines() int {
	return v.textInfo.MaxLines
}

func (v *ShapeView) SetMinWidth(min float64) View {
	s := v.MinSize()
	s.W = min
	v.SetMinSize(s)
	return v
}

func (v *ShapeView) SetMaxWidth(max float64) View {
	v.maxWidth = max
	return v
}

func (v *ShapeView) SetMaxLines(max int) View {
	v.textInfo.MaxLines = max
	return v
}

func (v *ShapeView) GetImage() *Image {
	return v.image
}

func (v *ShapeView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.MinSize()
	if v.textInfo.Text != "" {
		ts := v.textInfo.GetBounds(false)
		ts.Add(zgeo.Size{16, 6})
		ts.W *= 1.1 // some strange bug in android doesn't allow *= here...
		s.Maximize(ts)
	}
	if v.image != nil {
		s.Maximize(v.image.Size())
	}
	s.Add(v.margin.Size.Negative())
	if v.maxWidth != 0.0 {
		zfloat.Maximize(&s.W, v.maxWidth)
	}
	if v.Type == ShapeViewTypeCircle {
		//		zmath.Float64Maximize(&s.H, s.W)
	}
	// zlog.Info("ShapeView CalcSize:", v.ObjectName(), s)
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
		v.image = ImageFromPath(spath, func(*Image) {
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
	// zlog.Info("shapeViewDraw:", v.canvas != nil, v.MinSize(), rect, view.ObjectName())
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
		if !v.Usable() {
			o *= 0.6
		}
		canvas.SetColor(v.getStateColor(col), o)
		canvas.FillPath(path)
	}
	if v.StrokeWidth != 0 {
		zlog.Info("shapeViewDraw stroke:", view.ObjectName())
		var o = v.StrokeColor.Opacity()
		if !v.Usable() {
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
		if !v.Usable() {
			o *= 0.6
		}
		if v.IsImageFill {
			canvas.PushState()
			canvas.ClipPath(path, false, false)
			canvas.DrawImage(drawImage, rect, o, zgeo.Rect{})
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
				clipPath := zgeo.PathNewRect(ir, zgeo.Size{corner, corner})
				canvas.PushState()
				canvas.ClipPath(clipPath, false, false)
			}
			textRect = ir
			canvas.DrawImage(drawImage, ir, o, zgeo.Rect{})
			if v.IsRoundImage {
				canvas.PopState()
			}
		}
	}
	if v.textInfo.Text != "" {
		t := v.textInfo // .Copy()
		t.Color = v.getStateColor(t.Color)
		exp := zgeo.Size{-v.TextXMargin * ScreenMain().SoftScale, 0}
		t.Rect = textRect.Expanded(exp)
		t.Rect.Pos.Y += 3
		// zlog.Info("Shape draw:", exp, v.textInfo.Text, t.Color)
		t.Font = v.Font()
		if v.IsImageFill {
			canvas.SetDropShadow(zgeo.Size{}, 2, zgeo.ColorBlack)
		}
		// if v.textInfotextInfo.Text == "On" {
		// 	zlog.Info("ShapeView draw text:", textRect, t.Rect, v.TextXMargin, t.Text)
		// }

		t.Draw(canvas)
		if v.IsImageFill {
			canvas.SetDropShadowOff(1)
		}
	}
	if v.IsFocused() {
		FocusDraw(canvas, rect, 15, 0, 1)
	}
}
