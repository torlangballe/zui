//go:build zui

package zshape

//  Created by Tor Langballe on /22/10/15.

import (
	"math"
	"path"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zfocus"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zscreen"
)

type Type string

const (
	TypeCircle    Type = "circle"
	TypeRectangle Type = "rectange"
	TypeRoundRect Type = "roundrect"
	TypeStar      Type = "star"
	TypeNone      Type = ""
)

type ShapeView struct {
	zcontainer.ContainerView
	Type         Type
	StrokeWidth  float64
	StrokeColor  zgeo.Color // = ZColor.White()
	textInfo     ztextinfo.Info
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
	DropShadow zstyle.DropShadow

	image   *zimage.Image
	loading bool
}

func NewView(shapeType Type, minSize zgeo.Size) *ShapeView {
	v := &ShapeView{}
	v.Init(v, shapeType, minSize, string(shapeType))
	return v
}

func (v *ShapeView) Init(view zview.View, shapeType Type, minSize zgeo.Size, name string) {
	v.CustomView.Init(view, name)
	v.textInfo = *ztextinfo.New()
	v.Type = shapeType
	v.ImageMargin = zgeo.Size{4, 1}.TimesD(zscreen.MainSoftScale)
	v.ImageOpacity = 1
	v.ImageGap = 4
	v.Count = 5
	v.StrokeColor = zgeo.ColorWhite
	v.ImageAlign = zgeo.Center | zgeo.Proportional
	v.PathLineType = zgeo.PathLineRound
	v.TextXMargin = 8
	v.SetColor(zgeo.ColorGray)
	v.SetTextColor(zstyle.DefaultFGColor())

	switch shapeType {
	case TypeRoundRect:
		v.Ratio = 0.495
	case TypeStar:
		v.Ratio = 0.6
	default:
		v.Ratio = 0.3
	}
	v.SetDrawHandler(v.draw)
	v.CustomView.SetMinSize(minSize)
	f := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	v.SetFont(f)
}

// Text sets the ShapeView's textInfo.Text string, and exposes. This is also here to avoid underlying NativeView SetText() method being used
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

func (v *ShapeView) SetTextWrap(w ztextinfo.WrapType) {
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

func (v *ShapeView) GetImage() *zimage.Image {
	return v.image
}

func (v *ShapeView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.MinSize()
	if v.textInfo.Text != "" && v.textInfo.Alignment != zgeo.AlignmentNone {
		ts, _, _ := v.textInfo.GetBounds()
		// zlog.Info("ShapeView.CalculatedSize:", v.ObjectName(), v.textInfo.Text, s, ts)

		ts.Add(zgeo.Size{12, 6})
		// ts.W *= 1.1
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
	s.Add(v.Margin().Size.Negative())

	if v.MaxSize.W != 0.0 {
		zfloat.Minimize(&s.W, v.MaxSize.W)
	}
	if v.MaxSize.H != 0.0 {
		zfloat.Minimize(&s.H, v.MaxSize.H)
	}

	// zlog.Info(v.margin, v.MaxSize, v.ImageMargin, "SV Calcsize:", v.ObjectName(), s)
	if v.Type == TypeCircle {
		//		zmath.Float64Maximize(&s.H, s.W)
	}
	// if v.image.loading {
	// 	zlog.Info("SH image loading:", v.ObjectName(), v.image.capInsets.Size.Negative())
	// }
	// zlog.Info("ShapeView CalcSize:", v.ObjectName(), v.textInfo.Text, s, v.MinSize(), v.MaxSize.H)
	s = s.Ceil()
	return s
}

func (v *ShapeView) SetImage(image *zimage.Image, spath string, done func(i *zimage.Image)) {
	// zlog.Info("SVSetImage:", v.ObjectName(), spath)
	v.image = image
	v.SetExposed(false)
	if v.ObjectName() == "" {
		_, name := path.Split(spath)
		v.SetObjectName(name)
	}
	v.loading = false
	if image == nil && spath != "" {
		v.loading = true
		zimage.FromPath(spath, func(i *zimage.Image) {
			v.loading = false
			// zlog.Info("sv image loaded: " + spath + ": " + v.ObjectName())
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
	v.SetImage(nil, str, func(image *zimage.Image) {
		v.image = image
		if image != nil {
			// zlog.Info("SetImageButtonName:", str)
			if v.image.Size().W < insets.W*2 || v.image.Size().H < insets.H*2 {
				zlog.Error(nil, "Button: Small image for inset:", v.ObjectName(), pathedName, v.image.Size(), insets)
				return
			}
			v.image.SetCapInsetsCorner(insets)
		}
	})
}

func (v *ShapeView) draw(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	path := zgeo.PathNew()
	switch v.Type {
	case TypeStar:
		path.AddStar(rect, v.Count, v.Ratio)

	case TypeCircle:
		s := rect.Size.MinusD(v.StrokeWidth + 0.5).DividedByD(2).TimesD(zscreen.MainSoftScale)
		w := s.Min()
		path.ArcDegFromCenter(rect.Center(), zgeo.Size{w, w}, 0, 360)

	case TypeRoundRect:
		r := rect.Expanded(zgeo.Size{-1, -1}) //.TimesD(zscreen.GetMain().SoftScale))
		corner := math.Round(math.Min(math.Min(r.Size.W, r.Size.H)*float64(v.Ratio), 15))
		path.AddRect(r, zgeo.Size{corner, corner})

	case TypeRectangle:
		path.AddRect(rect, zgeo.Size{})
	}
	col := v.Color()
	if col.Valid {
		var o = col.Opacity()
		if !v.Usable() {
			o *= 0.6
		}
		canvas.SetColor(v.GetStateColor(col.WithOpacity(o)))
		canvas.FillPath(path)
	}
	if v.StrokeWidth != 0 {
		var o = v.StrokeColor.Opacity()
		if !v.Usable() {
			o *= 0.6
		}
		canvas.SetColor(v.GetStateColor(v.StrokeColor).WithOpacity(o))
		canvas.StrokePath(path, v.StrokeWidth, v.PathLineType)
	}
	textRect := rect.Plus(v.Margin())
	if v.image != nil && !v.image.Loading {
		if v.IsHighlighted() {
			v.image.TintedWithColor(zgeo.ColorNewGray(0.2, 1), 1, func(ni *zimage.Image) {
				v.drawImage(canvas, ni, path, rect, textRect)
			})
		} else {
			v.drawImage(canvas, v.image, path, rect, textRect)
		}
	}
	if v.textInfo.Text != "" && v.textInfo.Alignment != zgeo.AlignmentNone {
		t := v.textInfo // .Copy()
		t.Color = v.GetStateColor(t.Color)
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
		zfocus.Draw(canvas, rect, 15, 0, 1)
	}
}

func (v *ShapeView) Font() *zgeo.Font {
	return v.textInfo.Font
}

func (v *ShapeView) SetFont(font *zgeo.Font) {
	// zlog.Info("SH SetFont:", v.Hierarchy(), v.textInfo.Text, *font)
	v.textInfo.Font = font
	v.NativeView.SetFont(font)
}

func (v *ShapeView) drawImage(canvas *zcanvas.Canvas, img *zimage.Image, shapePath *zgeo.Path, rect, textRect zgeo.Rect) {
	useDownsampleCache := true
	imarg := v.ImageMargin
	o := v.ImageOpacity
	if !v.Usable() {
		o *= 0.6
	}
	if v.IsImageFill {
		canvas.PushState()
		canvas.ClipPath(shapePath, false, false)
		canvas.DrawImage(img, useDownsampleCache, rect, o, zgeo.Rect{})
		canvas.PopState()
	} else {
		a := v.ImageAlign | zgeo.Shrink
		ir := rect.AlignPro(v.image.Size(), a, v.ImageMargin, v.ImageMaxSize, zgeo.Size{})
		var corner float64
		if v.IsRoundImage {
			if v.Type == TypeRoundRect {
				corner = math.Min(15, rect.Size.Min()*float64(v.Ratio)) - imarg.Min()
			} else if v.Type == TypeCircle {
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
		canvas.DrawImage(img, useDownsampleCache, ir, o, zgeo.Rect{})
		if v.IsRoundImage {
			canvas.PopState()
		}
	}
}

func (v *ShapeView) SetColor(c zgeo.Color) {
	v.NativeView.SetColor(c)
	// zlog.Info("SV.SetColor:", v.Hierarchy(), c, v.Color())
	v.Expose()
}
