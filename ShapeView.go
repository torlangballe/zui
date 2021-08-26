// +build zui

package zui

//  Created by Tor Langballe on /22/10/15.

import (
	"math"
	"path"

	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zscreen"
)

type ShapeViewType string

const (
	ShapeViewTypeCircle    ShapeViewType = "circle"
	ShapeViewTypeRectangle ShapeViewType = "rectange"
	ShapeViewTypeRoundRect ShapeViewType = "roundrect"
	ShapeViewTypeStar      ShapeViewType = "star"
	ShapeViewTypeNone      ShapeViewType = ""
)

type ShapeView struct {
	ContainerView
	Type         ShapeViewType
	StrokeWidth  float64
	StrokeColor  zgeo.Color // = ZColor.White()
	textInfo     TextInfo
	ImageMargin  zgeo.Size //  ZSize(4.0, 1.0) * ZScreen.SoftScale
	ImageGap     float64
	ImageAlign   zgeo.Alignment // .Center
	ImageMaxSize zgeo.Size
	IsRoundImage bool
	IsImageFill  bool

	TextXMargin  float64
	ImageOpacity float32 // Float = Float(1)
	Ratio        float32 // = 0.3
	Count        int     // = 5
	MaxSize      zgeo.Size
	Value        float64
	PathLineType zgeo.PathLineType
	//	Proportional bool
	DropShadow zgeo.DropShadow

	image   *Image
	loading bool
}

func ShapeViewNew(shapeType ShapeViewType, minSize zgeo.Size) *ShapeView {
	v := &ShapeView{}
	v.Init(v, shapeType, minSize, string(shapeType))
	return v
}

func (v *ShapeView) Init(view View, shapeType ShapeViewType, minSize zgeo.Size, name string) {
	v.CustomView.Init(view, name)
	v.textInfo = *TextInfoNew()
	v.Type = shapeType
	v.ImageMargin = zgeo.Size{4, 1}.TimesD(zscreen.MainSoftScale)
	v.ImageOpacity = 1
	v.ImageGap = 4
	v.Count = 5
	v.StrokeColor = zgeo.ColorWhite
	v.ImageAlign = zgeo.Center | zgeo.Proportional
	v.PathLineType = zgeo.PathLineRound
	v.TextXMargin = 4
	v.SetColor(zgeo.ColorGray)
	v.SetTextColor(StyleDefaultFGColor())

	switch shapeType {
	case ShapeViewTypeRoundRect:
		v.Ratio = 0.495
	case ShapeViewTypeStar:
		v.Ratio = 0.6
	default:
		v.Ratio = 0.3
	}
	v.SetDrawHandler(v.draw)
	v.CustomView.SetMinSize(minSize)
	f := FontNice(FontDefaultSize, FontStyleNormal)
	v.SetFont(f)
}

// Text sets the ShapeView's TextInfo.Text string, and exposes. This is also here to avoid underlying NativeView SetText() method being used
func (v *ShapeView) SetText(text string) {
	v.textInfo.Text = text
	v.Expose()
}

func (v *ShapeView) Text() string {
	return v.textInfo.Text
}

func (v *ShapeView) SetTextAlignment(a zgeo.Alignment) {
	v.textInfo.Alignment = a
}

func (v *ShapeView) SetTextWrap(w TextInfoWrap) {
	v.textInfo.Wrap = w
}

func (v *ShapeView) SetTextColor(col zgeo.Color) {
	v.textInfo.Color = col
	v.Expose()
}

func (v *ShapeView) MinWidth() float64 {
	return v.MinWidth()
}

func (v *ShapeView) MaxLines() int {
	return v.textInfo.MaxLines
}

func (v *ShapeView) SetMinWidth(min float64) {
	s := v.MinSize()
	s.W = min
	v.SetMinSize(s)
}

func (v *ShapeView) SetMaxLines(max int) {
	v.textInfo.MaxLines = max
}

func (v *ShapeView) GetImage() *Image {
	return v.image
}

func (v *ShapeView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.MinSize()
	if v.textInfo.Text != "" && v.textInfo.Alignment != zgeo.AlignmentNone {
		ts, _, _ := v.textInfo.GetBounds()
		ts.Add(zgeo.Size{16, 6})
		ts.W *= 1.1
		s.Maximize(ts)
	}
	if v.image != nil {
		is := v.image.Size()
		if !v.ImageMaxSize.IsNull() {
			is.Minimize(v.ImageMaxSize)
		}
		if !v.MaxSize.IsNull() {
			ms := v.MaxSize
			if ms.W == 0 {
				ms.W = 99999
			}
			if ms.H == 0 {
				ms.H = 99999
			}
			ms.Subtract(v.ImageMargin.TimesD(2))
			// zlog.Info("SV MS:", v.ObjectName(), ms)
			is = is.ShrunkInto(ms)
		} else {
			is.Maximize(v.image.CapInsets().Size.Negative())
		}
		is.Add(v.ImageMargin.TimesD(2))
		// zlog.Info("SV IS:", v.IsImageFill, s, v.ObjectName(), is, v.MaxSize, v.ImageMargin, v.image.Size(), v.image.Path, "\n")
		if v.ImageAlign&(zgeo.Left|zgeo.Right) != 0 {
			s.W += is.W + v.ImageGap
		} else if v.ImageAlign&(zgeo.Top|zgeo.Bottom) != 0 {
			s.H += is.H + v.ImageGap
		} else {
			s.Maximize(is)
		}
	}
	s.Add(v.margin.Size.Negative())

	if v.MaxSize.W != 0.0 {
		zfloat.Minimize(&s.W, v.MaxSize.W)
	}
	if v.MaxSize.H != 0.0 {
		zfloat.Minimize(&s.H, v.MaxSize.H)
	}

	// zlog.Info(v.margin, v.MaxSize, v.ImageMargin, "SV Calcsize:", v.ObjectName(), s)
	if v.Type == ShapeViewTypeCircle {
		//		zmath.Float64Maximize(&s.H, s.W)
	}
	// if v.image.loading {
	// 	zlog.Info("SH image loading:", v.ObjectName(), v.image.capInsets.Size.Negative())
	// }
	// zlog.Info("ShapeView CalcSize:", v.ObjectName(), v.textInfo.Text, s, v.MinSize(), v.MaxSize.H)
	s = s.Ceil()
	return s
}

func (v *ShapeView) SetImage(image *Image, spath string, done func(i *Image)) {
	v.image = image
	v.exposed = false
	if v.ObjectName() == "" {
		_, name := path.Split(spath)
		v.SetObjectName(name)
	}
	v.loading = false
	if image == nil && spath != "" {
		v.loading = true
		ImageFromPath(spath, func(i *Image) {
			v.loading = false
			// zlog.Info("sv image loaded: "+spath+": "+v.ObjectName(), i.loading)
			v.image = i // we must set it here, or it's not set yet in done() below
			v.Expose()
			if done != nil {
				done(i)
			}
		})
	}
}

func (v *ShapeView) IsLoading() bool {
	return v.loading
}

func (v *ShapeView) SetNamedCapImage(pathedName string, insets zgeo.Size) {
	s := ""
	if zscreen.MainScale >= 2 {
		s = "@2x"
	}
	str := pathedName + s + ".png"

	// zlog.Info("SetImageButtonName:", str)
	v.SetImage(nil, str, func(image *Image) {
		v.image = image
		if image != nil {
			if v.image.Size().W < insets.W*2 || v.image.Size().H < insets.H*2 {
				zlog.Error(nil, "Button: Small image for inset:", v.ObjectName(), pathedName, v.image.Size(), insets)
				return
			}
			v.image.SetCapInsetsCorner(insets)
		}
	})
}

func (v *ShapeView) draw(rect zgeo.Rect, canvas *Canvas, view View) {
	// zlog.Info("shapeViewDraw:", view.ObjectName(), rect)
	path := zgeo.PathNew()
	switch v.Type {
	case ShapeViewTypeStar:
		path.AddStar(rect, v.Count, v.Ratio)

	case ShapeViewTypeCircle:
		s := rect.Size.MinusD(v.StrokeWidth + 0.5).DividedByD(2).TimesD(zscreen.MainSoftScale)
		w := s.Min()
		path.ArcDegFromCenter(rect.Center(), zgeo.Size{w, w}, 0, 360)

	case ShapeViewTypeRoundRect:
		r := rect.Expanded(zgeo.Size{-1, -1}) //.TimesD(zscreen.GetMain().SoftScale))
		corner := math.Round(math.Min(math.Min(r.Size.W, r.Size.H)*float64(v.Ratio), 15))
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
		canvas.SetColor(v.getStateColor(col.WithOpacity(o)))
		canvas.FillPath(path)
	}
	if v.StrokeWidth != 0 {
		var o = v.StrokeColor.Opacity()
		if !v.Usable() {
			o *= 0.6
		}
		canvas.SetColor(v.getStateColor(v.StrokeColor).WithOpacity(o))
		canvas.StrokePath(path, v.StrokeWidth, v.PathLineType)
	}
	imarg := v.ImageMargin
	useDownsampleCache := true
	textRect := rect.Plus(v.margin)
	if v.image != nil && !v.image.loading {
		drawImage := v.image
		if v.IsHighlighted() {
			drawImage = drawImage.TintedWithColor(zgeo.ColorNewGray(0.2, 1))
		}
		o := v.ImageOpacity
		if !v.Usable() {
			o *= 0.6
		}
		if v.IsImageFill {
			canvas.PushState()
			canvas.ClipPath(path, false, false)
			canvas.DrawImage(drawImage, true, useDownsampleCache, rect, o, zgeo.Rect{})
			canvas.PopState()
		} else {
			a := v.ImageAlign | zgeo.Shrink
			ir := rect.AlignPro(v.image.Size(), a, v.ImageMargin, v.ImageMaxSize, zgeo.Size{})
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
			if v.textInfo.Text != "" {
				if v.ImageAlign&zgeo.Right != 0 {
					textRect.SetMaxX(ir.Min().X - v.ImageGap)
				} else if v.ImageAlign&zgeo.Left != 0 {
					textRect.SetMinX(ir.Max().X + v.ImageGap)
				}
			}
			// zlog.Info("SV DrawImage:", v.ObjectName(), drawImage.Path, ir, o)
			canvas.DrawImage(drawImage, true, useDownsampleCache, ir, o, zgeo.Rect{})
			if v.IsRoundImage {
				canvas.PopState()
			}
		}
	}
	if v.textInfo.Text != "" && v.textInfo.Alignment != zgeo.AlignmentNone {
		t := v.textInfo // .Copy()
		t.Color = v.getStateColor(t.Color)
		exp := zgeo.Size{-v.TextXMargin * zscreen.MainSoftScale, 0}
		t.Rect = textRect.Expanded(exp)
		t.Rect.Pos.Y -= 1
		//		zlog.Info("Shape draw:", v.ObjectName(), exp, v.textInfo.Text, t.Color)
		t.Font = v.Font()
		if v.IsImageFill {
			canvas.SetDropShadow(zgeo.Size{}, 2, zgeo.ColorBlack) // why do we do this????
		}
		// if v.textInfotextInfo.Text == "On" {
		// zlog.Info("ShapeView draw text:", textRect, t.Rect, v.TextXMargin, t.Text)
		// }

		// canvas.SetColor(zgeo.ColorGreen)
		// canvas.FillRect(t.Rect)
		// zlog.Info("shapeViewDraw text:", view.ObjectName(), v.margin, rect, t.Rect, v.TextXMargin)
		t.Draw(canvas)
		if v.IsImageFill {
			canvas.SetDropShadowOff(1)
		}
	}
	if v.IsFocused() {
		FocusDraw(canvas, rect, 15, 0, 1)
	}
}
