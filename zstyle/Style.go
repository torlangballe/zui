package zstyle

import (
	"math"
	"path"

	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
)

type Styling struct {
	DropShadow    zgeo.DropShadow
	BGColor       zgeo.Color
	FGColor       zgeo.Color
	Font          zgeo.Font
	Corner        float64
	StrokeWidth   float64
	StrokeColor   zgeo.Color
	StrokeIsInset zbool.BoolInd
	OutlineWidth  float64
	OutlineColor  zgeo.Color
	OutlineOffset float64
	Margin        zgeo.Rect
	Spacing       float64
}

const DarkModeKey = "zstyle.DarkMode"

var (
	Dark               bool
	DefaultFGColorFunc = GrayFunc(0.2, 0.8)
	DefaultBGColorFunc = GrayFunc(0.8, 0.2)
	DefaultFGColor     = GrayF(DefaultFGColorFunc)
	DefaultBGColor     = GrayF(DefaultBGColorFunc)
	DefaultHoverColor  = Col(zgeo.ColorNew(0.2, 0.6, 1, 1), zgeo.Color{})
	DefaultFocusColor  = Col(zgeo.ColorNew(0.58, 0.71, 0.97, 1), zgeo.Color{})

	DebugBackgroundColor  = Col(zgeo.ColorNew(1, 0.9, 0.9, 1), zgeo.ColorNew(0.8, 0.4, 0.4, 1))
	DefaultRowRightMargin = 6.0
	DefaultRowLeftMargin  = 6.0
)

var EmptyStyling = Styling{
	Corner:        -1,
	StrokeWidth:   -1,
	StrokeIsInset: zbool.Unknown,
	OutlineWidth:  -1,
	OutlineOffset: -1,
	Margin:        zgeo.RectUndef,
	Spacing:       zfloat.Undefined,
}

func useInvertedIfInvalid(c, alt zgeo.Color) zgeo.Color {
	if c.Valid {
		return c
	}
	return alt.BrightnessInverted()
}

func Col(l, d zgeo.Color) zgeo.Color {
	return ColFor(l, d, Dark)
}

func Col1(c zgeo.Color) zgeo.Color {
	return Col1For(c, Dark)
}

func Col1For(c zgeo.Color, dark bool) zgeo.Color {
	if dark {
		return c.BrightnessInverted()
	}
	return c
}

func ColFor(l, d zgeo.Color, dark bool) zgeo.Color {
	if dark {
		return useInvertedIfInvalid(d, l)
	}
	// zlog.Error(zlog.StackAdjust(1), "ColLight:", l)
	return useInvertedIfInvalid(l, d)
}

func ColCur(l, d zgeo.Color) func() zgeo.Color {
	return func() zgeo.Color {
		return ColFor(l, d, Dark)
	}
}

func ColF(l, d zgeo.Color) func() zgeo.Color {
	return func() zgeo.Color {
		return Col(l, d)
	}
}

func Gray(l, d float32) zgeo.Color {
	return GrayFor(l, d, Dark)
}

func Gray1(g float32) zgeo.Color {
	return Gray1For(g, Dark)
}

func Gray1For(g float32, dark bool) zgeo.Color {
	if dark {
		g = 1 - g
	}
	return zgeo.ColorNewGray(g, 1)
}

func GrayFor(l, d float32, dark bool) zgeo.Color {
	return ColFor(zgeo.ColorNewGray(l, 1), zgeo.ColorNewGray(d, 1), dark)
}

func GrayCur(l, d float32) func() zgeo.Color {
	return func() zgeo.Color {
		return GrayFor(l, d, Dark)
	}
}

func GrayF(f func(dark bool) zgeo.Color) func() zgeo.Color {
	return func() zgeo.Color {
		return f(Dark)
	}
}

func GrayFunc(l, d float32) func(dark bool) zgeo.Color {
	return func(dark bool) zgeo.Color {
		return GrayFor(l, d, dark)
	}
}

func ColFunc(l, d zgeo.Color) func(dark bool) zgeo.Color {
	return func(dark bool) zgeo.Color {
		return ColFor(l, d, dark)
	}
}

func ImagePath(spath string) string {
	if !Dark {
		return spath
	}
	dir, _, stub, ext := zfile.Split(spath)
	size := ""
	if zstr.HasSuffix(stub, "@2x", &stub) {
		size = "@2x"
	}
	return path.Join(dir, stub+"_dark"+size+ext)
}

func (s Styling) MergeWith(m Styling) Styling {
	if m.DropShadow.Color.Valid {
		s.DropShadow.Color = m.DropShadow.Color
	}
	if m.DropShadow.Blur != -1 {
		s.DropShadow.Blur = m.DropShadow.Blur
	}
	if !m.DropShadow.Delta.IsUndef() {
		s.DropShadow.Delta = m.DropShadow.Delta
	}
	if m.DropShadow.Color.Valid {
		s.DropShadow.Color = m.DropShadow.Color
	}
	if m.BGColor.Valid {
		s.BGColor = m.BGColor
	}
	if m.FGColor.Valid {
		s.FGColor = m.FGColor
	}
	if m.Font.Name != "" {
		s.Font = m.Font
	}
	if m.Corner != -1 {
		s.Corner = m.Corner
	}
	if m.StrokeWidth != -1 {
		s.StrokeWidth = m.StrokeWidth
	}
	if m.StrokeColor.Valid {
		s.StrokeColor = m.StrokeColor
	}
	if !m.StrokeIsInset.IsUnknown() {
		s.StrokeIsInset = m.StrokeIsInset
	}
	if m.OutlineWidth != -1 {
		s.OutlineWidth = m.OutlineWidth
	}
	if m.OutlineColor.Valid {
		s.OutlineColor = m.OutlineColor
	}
	if m.OutlineOffset != -1 {
		s.OutlineOffset = m.OutlineOffset
	}
	if !m.Margin.IsUndef() {
		s.Margin = m.Margin
	}
	if m.Spacing != zfloat.Undefined {
		s.Spacing = m.Spacing
	}
	return s
}

// SpacingOrMax returns s's spacing if defined, or max
func (s Styling) SpacingOrMax(max float64) float64 {
	if s.Spacing == zfloat.Undefined {
		return max
	}
	return math.Max(max, s.Spacing)
}

func DropShadowDefault() zgeo.DropShadow {
	gray := zgeo.ColorNewGray(0, 0.7)
	if Dark {
		gray = zgeo.ColorNewGray(0.5, 0.2)
	}
	return zgeo.DropShadow{Delta: zgeo.SizeBoth(3), Blur: 3, Color: gray}
}
