//go:build zui

package zshape

//  Created by Tor Langballe on /22/10/15.

import (
	"math"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zdocs"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zgeo"
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
	zcontainer.StackView
	Type        Type
	StrokeWidth float64
	StrokeColor zgeo.Color // = ZColor.White()
	textInfo    ztextinfo.Info
	ImageMargin zgeo.Rect
	TextMargin  zgeo.Rect
	ImageAlign  zgeo.Alignment
	// ImageFitSize zgeo.Size
	ImageMaxSize zgeo.Size
	IsImageFill  bool
	Ratio        float32 // = 0.3
	Count        int     // = 5
	MaxSize      zgeo.Size
	Value        float64
	PathLineType zgeo.PathLineType
	DropShadow   zgeo.DropShadow

	ImageView *zimageview.ImageView
	TextLabel *zlabel.Label
	loading   bool

	SubPart any // points to a MenuedOwner if any
}

func NewView(shapeType Type, minSize zgeo.Size) *ShapeView {
	v := &ShapeView{}
	v.Init(v, shapeType, minSize, string(shapeType))
	return v
}

func (v *ShapeView) Init(view zview.View, shapeType Type, minSize zgeo.Size, name string) {
	v.StackView.Init(view, false, name)
	v.textInfo = *ztextinfo.New()
	v.Type = shapeType
	v.ImageMargin = zgeo.RectFromMarginSize(zgeo.SizeD(12, 1))
	v.Count = 5
	v.StrokeColor = zgeo.ColorWhite
	v.ImageAlign = zgeo.Center | zgeo.Proportional
	v.PathLineType = zgeo.PathLineRound
	v.SetColor(zgeo.ColorGray)
	v.SetTextColor(zstyle.DefaultFGColor())
	v.textInfo.Alignment = zgeo.Center

	switch shapeType {
	case TypeRoundRect:
		v.Ratio = 0.495
	case TypeStar:
		v.Ratio = 0.6
	default:
		v.Ratio = 0.3
	}
	if shapeType != TypeNone {
		v.SetDrawHandler(v.draw)
	}
	v.CustomView.SetMinSize(minSize)
	f := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	v.SetFont(f)
	v.NativeView.SetKeyHandler(func(km zkeyboard.KeyMod, down bool) bool {
		if !down && km.Key.IsReturnish() && km.Modifier == zkeyboard.ModifierNone {
			if v.PressedHandler() != nil {
				v.PressedHandler()()
				return true
			}
		}
		return false
	})
	// style := v.JSStyle()
	// style.Set("display", "flex")
	// style.Set("flexDirection", "column")
	// style.Set("alignItems", "center")
	v.UpdateText()
}

// Text sets the ShapeView's textInfo.Text string, and exposes. This is also here to avoid underlying NativeView SetText() method being used
func (v *ShapeView) SetText(text string) {
	v.textInfo.Text = text
	v.UpdateText()
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
	v.UpdateText()
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
	v.UpdateText()
}

func (v *ShapeView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	s, max = v.StackView.CalculatedSize(total)
	// zlog.Info("SV Calc:", v.ObjectName(), s, max, v.MinSize())
	return s, max
}

func (v *ShapeView) SetImage(image *zimage.Image, useCache bool, fitSize zgeo.Size, spath string, insets zgeo.Size, done func(i *zimage.Image)) {
	if v.ImageView == nil {
		v.ImageView = zimageview.New(nil, useCache, "", fitSize)
		v.ImageView.SetInteractive(false)
		v.ImageView.CapInsetCorner = insets
		if v.IsImageFill {
			v.ImageView.SetAlignment(zgeo.Expand | zgeo.Center)
			// a := zgeo.Expand | zgeo.Center
			v.AddAdvanced(v.ImageView, zgeo.AlignmentNone, zgeo.RectNull, zgeo.SizeNull, -1, true)
			// zlog.Info("SVSetImage:", v.ObjectName(), spath)
			// v.Add(v.ImageView, zgeo.AlignmentNone)
		} else {
			v.ImageView.SetAlignment(v.ImageAlign)
			// zlog.Info("SVSetImage:", v.ObjectName(), spath, v.ImageAlign, v.ImageView.FitSize())
			v.Add(v.ImageView, v.ImageAlign, v.ImageMargin, v.ImageMaxSize)
		}
	}
	v.ImageView.SetImage(image, spath, done)
}

func (v *ShapeView) ArrangeChildren() {
	v.StackView.ArrangeChildren()
	if v.ImageView != nil {
		if v.IsImageFill {
			v.ImageView.SetRect(v.LocalRect())
		} else {
			// zlog.Info("SV Arrange", v.Rect(), v.ObjectName(), v.ImageView.Rect(), v.ImageAlign, v.ImageView.FitSize(), v.ImageView.MinSize())
		}
	}
}

func (v *ShapeView) IsLoading() bool {
	return v.loading
}

func (v *ShapeView) getStateColor(col zgeo.Color) zgeo.Color {
	if v.IsHighlighted() {
		g := col.GrayScale()
		if g < 0.5 {
			col = col.Mixed(zgeo.ColorWhite, 0.5)
		} else {
			col = col.Mixed(zgeo.ColorBlack, 0.5)
		}
	}
	if !v.IsUsable() {
		col = col.WithOpacity(0.3)
	}
	return col
}

func (v *ShapeView) SetNamedCapImage(pathedName string, insets zgeo.Size) {
	s := ""
	if zscreen.MainScale() >= 2 {
		s = "@2x"
	}
	str := pathedName + s + ".png"
	v.SetImage(nil, true, zgeo.SizeNull, str, insets, nil)
}

func (v *ShapeView) draw(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	path := zgeo.PathNew()
	switch v.Type {
	case TypeStar:
		path.AddStar(rect, v.Count, v.Ratio)

	case TypeCircle:
		s := rect.Size.MinusD(v.StrokeWidth + 0.5).DividedByD(2).TimesD(zscreen.MainSoftScale())
		w := s.Min()
		path.ArcDegFromCenter(rect.Center(), zgeo.SizeBoth(w), 0, 360)

	case TypeRoundRect:
		r := rect.Expanded(zgeo.SizeBoth(-1)) //.TimesD(zscreen.GetMain().SoftScale))
		corner := math.Round(math.Min(math.Min(r.Size.W, r.Size.H)*float64(v.Ratio), 15))
		path.AddRect(r, zgeo.SizeBoth(corner))

	case TypeRectangle:
		path.AddRect(rect, zgeo.SizeNull)
	}
	col := v.Color()
	if col.Valid {
		var o = col.Opacity()
		if !v.IsUsable() {
			o *= 0.6
		}
		fillCol := v.getStateColor(col.WithOpacity(o))
		canvas.SetColor(fillCol)
		canvas.FillPath(path)
	}
	if v.StrokeWidth != 0 {
		var o = v.StrokeColor.Opacity()
		if !v.IsUsable() {
			o *= 0.6
		}
		canvas.SetColor(v.getStateColor(v.StrokeColor).WithOpacity(o))
		canvas.StrokePath(path, v.StrokeWidth, v.PathLineType)
	}
}

func (v *ShapeView) UpdateText() {
	if (v.textInfo.Text != "" || v.TextLabel != nil) && v.textInfo.Alignment != zgeo.AlignmentNone {
		a := zgeo.VertCenter | v.textInfo.Alignment
		if v.TextLabel == nil {
			v.TextLabel = zlabel.New("")
			v.TextLabel.SetInteractive(false)
			v.TextLabel.SetObjectName("title")
			if v.textInfo.Alignment.Has(zgeo.HorCenter) && zdevice.CurrentWasmBrowser != zdevice.Safari && zdevice.OS() != zdevice.MacOSType {
				v.TextLabel.SetMargin(zgeo.RectFromXY2(0, 0, -5, 0))
			}
			v.AddAdvanced(v.TextLabel, a, v.TextMargin, zgeo.SizeNull, -1, false)
		} else {
			c, _ := v.FindCellWithView(v.TextLabel)
			c.Alignment = a
			c.Margin = v.TextMargin
		}
		// zlog.Info("SV.SetTXT:", v.textInfo.Alignment, v.ObjectName(), v.textInfo.Text)
		v.TextLabel.SetTextAlignment(v.textInfo.Alignment)
		v.TextLabel.SetFont(v.Font())
		v.TextLabel.SetColor(v.getStateColor(v.textInfo.Color))
		v.TextLabel.SetText(v.textInfo.Text)
		if v.IsPresented() {
			v.ArrangeChildren()
		}
	}
}

func (v *ShapeView) Font() *zgeo.Font {
	return v.textInfo.Font
}

func (v *ShapeView) SetFont(font *zgeo.Font) {
	// zlog.Info("SH SetFont:", v.Hierarchy(), v.textInfo.Text, *font)
	v.textInfo.Font = font
	v.NativeView.SetFont(font)
	v.UpdateText()
}

func (v *ShapeView) SetColor(c zgeo.Color) {
	v.NativeView.SetColor(c)
	// zlog.Info("SV.SetColor:", zlog.Pointer(v), v.Hierarchy(), c, v.Color(), zdebug.CallingStackString())
}

func (v *ShapeView) GetSearchableItems(currentPath []zdocs.PathPart) []zdocs.SearchableItem {
	var parts []zdocs.SearchableItem
	if v.SubPart != nil {
		sig, _ := v.SubPart.(zdocs.SearchableItemsGetter)
		if sig != nil {
			subPath := zdocs.AddedPath(currentPath, zdocs.StaticField, v.ObjectName(), v.ObjectName())
			subParts := sig.GetSearchableItems(subPath)
			parts = append(parts, subParts...)
		}
	}
	if len(parts) == 0 && v.textInfo.Text != "" {
		ctype := zdocs.StaticField
		if v.PressedHandler() != nil {
			ctype = zdocs.PressField
		}
		item := zdocs.MakeSearchableItem(currentPath, ctype, v.ObjectName(), v.ObjectName(), v.textInfo.Text)
		parts = append(parts, item)
	}
	return parts
}
