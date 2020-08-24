package zui

import (
	"math"

	"github.com/torlangballe/zutil/zfloat"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
)

//  Created by Tor Langballe on /22/10/15.

type TextInfoType int
type TextInfoWrap int

const (
	TextInfoFill TextInfoType = iota
	TextInfoStroke
	TextInfoClip
)

const (
	TextInfoWrapWord TextInfoWrap = iota
	TextInfoWrapChar
	TextInfoWrapClip
	TextInfoWrapHeadTruncate
	TextInfoWrapTailTruncate
	TextInfoWrapMiddleTruncate

	TextPaintType   = "paint-type"
	TextWrapType    = "wrap"
	TextColor       = "color"
	TextAlignment   = "alignemnt"
	TextFontName    = "font-name"
	TextFontSize    = "font-size"
	TextFontStyle   = "font-style"
	TextLineSpacing = "line-spacing"
	TextStrokeWidth = "stroke-width"
	TextStrokeColor = "stroke-color"
)

type TextInfo struct {
	Type                  TextInfoType
	Wrap                  TextInfoWrap
	Text                  string
	Color                 zgeo.Color
	Alignment             zgeo.Alignment
	Font                  *Font
	Rect                  zgeo.Rect
	LineSpacing           float64
	StrokeWidth           float64
	MaxLines              int
	MinimumFontScale      float64
	IsMinimumOneLineHight bool
	SplitItems            []string
}

type TextInfoOwner interface {
	GetTextInfo() TextInfo
}

func TextInfoNew() *TextInfo {
	t := &TextInfo{}
	t.Type = TextInfoFill
	t.Wrap = TextInfoWrapWord
	t.Color = zgeo.ColorBlack
	t.Alignment = zgeo.Center
	t.Font = FontNice(FontDefaultSize, FontStyleNormal)
	t.StrokeWidth = 1
	t.MinimumFontScale = 0.5
	t.SplitItems = []string{"\r\n", "\n", "\r"}
	return t
}

func (ti *TextInfo) SetWidthFreeHight(w float64) {
	ti.Rect = zgeo.RectFromSize(zgeo.Size{w, 99999})
}

// GetBounds returns rect of size of text.
// It is placed within ti.Rect using alignment
// TODO: Make it handle multi-line with some home-made wrapping stuff.
func (ti *TextInfo) GetBounds() (zgeo.Size, []string, []float64) {
	var size zgeo.Size
	var allLines []string
	var widths []float64
	lines := zstr.SplitByAnyOf(ti.Text, ti.SplitItems, false)
	for _, str := range lines {
		s := canvasGetTextSize(str, ti.Font)
		// zlog.Info("ti bounds:", str, ti.MaxLines)
		if ti.MaxLines != 1 && ti.Rect.Size.W != 0 {
			split := s.W / ti.Rect.Size.W
			// zlog.Info("TI split:", split, str)
			if split > 1 {
				runes := []rune(str)
				rlen := len(runes)
				each := int(math.Ceil(float64(len(runes)) / split))
				for i := 0; i < rlen; i += each {
					e := zint.Min(rlen, i+each)
					part := string(runes[i:e])
					ratio := float64(e-i) / float64(rlen)
					allLines = append(allLines, part)
					widths = append(widths, s.W*ratio)
				}
			} else {
				allLines = append(allLines, str)
				widths = append(widths, s.W)
			}
			// zlog.Info("SPLIT!", splitToLines, ti.Rect.Size, s)
			s.W = ti.Rect.Size.W
		} else {
			allLines = append(allLines, str)
		}
		zfloat.Maximize(&size.W, s.W)
		zfloat.Maximize(&size.H, s.H)
	}
	if ti.MaxLines == 1 || ti.Rect.Size.W == 0 {
		allLines = []string{ti.Text}
		widths = []float64{size.W}
	} else {
		count := zint.Max(ti.MaxLines, len(allLines))
		// zlog.Info("BOUNDS:", size, count, add)
		//	count = ti.MaxLines
		if count > 1 || ti.IsMinimumOneLineHight {
			size.H = float64(ti.Font.LineHeight()) * float64(zint.Max(count, 1))
		}
	}
	return size, allLines, widths
}

func (ti *TextInfo) getNativeWrapMode(w TextInfoWrap) int {
	switch w {
	case TextInfoWrapWord:
		return 1
	case TextInfoWrapChar:
		return 1
	case TextInfoWrapHeadTruncate:
		return 1
	case TextInfoWrapTailTruncate:
		return 1
	case TextInfoWrapMiddleTruncate:
		return 1
	default:
		return 1
	}
}

func getNativeTextAdjustment(style zgeo.Alignment) int {
	if style&zgeo.Left != 0 {
		return 0
	} else if style&zgeo.Right != 0 {
		return 0
	} else if style&zgeo.HorCenter != 0 {
		return 0
	} else if style&zgeo.HorJustify != 0 {
		panic("bad text adjust")
	}
	return 0 //NSTextAlignment.left
}

func (ti *TextInfo) MakeAttributes() zdict.Dict {
	return zdict.Dict{}
}

func (ti *TextInfo) FillAndStroke(canvas *Canvas, strokeColor zgeo.Color, width float64) zgeo.Rect {
	t := *ti
	t.Type = TextInfoFill
	//	t.Draw(canvas)
	t.Color = strokeColor
	t.Type = TextInfoStroke
	t.StrokeWidth = width
	return t.Draw(canvas)
}

func (ti *TextInfo) Draw(canvas *Canvas) zgeo.Rect {
	w := 0.0
	if ti.Text == "" {
		return zgeo.Rect{ti.Rect.Pos, zgeo.Size{}}
	}
	canvas.SetColor(ti.Color, 1)
	// switch ti.Type {
	//     case ZTextDrawType.fill
	//         //                    CGContextSetFillColorWithColor(canvas.context, canvaspfcolor)
	//         canvas.context.setTextDrawingMode(CGTextDrawingMode.fill)

	//     case ZTextDrawType.stroke
	//         canvas.context.setLineWidth(CGFloat(strokeWidth))
	//         //                CGContextSetFillColorWithColor(canvas.context, canvaspfcolor)
	//         // CGContextSetStrokeColorWithColor(canvas.context, canvaspfcolor)
	//         canvas.context.setTextDrawingMode(CGTextDrawingMode.stroke)

	//     case ZTextDrawType.clip
	//         canvas.context.setTextDrawingMode(CGTextDrawingMode.clip)
	// }
	if ti.Type == TextInfoStroke {
		w = ti.StrokeWidth
	}
	font := ti.Font
	if ti.Alignment&zgeo.HorCenter != 0 {
		//        r = r.Expanded(ZSize(1, 0))
	}
	if ti.Alignment&zgeo.HorShrink != 0 {
		zlog.Assert(!ti.Rect.Size.IsNull())
		// fmt.Println("CANVAS TI SetFont B4 Scale:", font)
		font = ti.ScaledFontToFit(ti.MinimumFontScale)
	}
	// fmt.Println("CANVAS TI SetFont:", font)
	canvas.SetFont(font, nil)
	if ti.Rect.Size.IsNull() {
		canvas.DrawTextInPos(ti.Rect.Pos, ti.Text, w)
		return zgeo.Rect{}
	}
	ts, lines, widths := ti.GetBounds()
	ts = zgeo.Size{math.Ceil(ts.W), math.Ceil(ts.H)}
	ra := ti.Rect.Align(ts, ti.Alignment, zgeo.Size{}, zgeo.Size{})
	// https://stackoverflow.com/questions/5026961/html5-canvas-ctx-filltext-wont-do-line-breaks/21574562#21574562
	h := font.LineHeight()
	y := ra.Pos.Y + h*0.71
	// zlog.Info("TI.Draw:", ti.Rect, font.Size, ti.Text)
	zlog.Assert(len(lines) == len(widths), len(lines), len(widths))
	for i, s := range lines {
		x := ra.Pos.X + (ra.Size.W-widths[i])/2
		// tsize := canvasGetTextSize(s, ti.Font)
		canvas.DrawTextInPos(zgeo.Pos{x, y}, s, w)
		y += h
	}
	return ra
}

func (ti *TextInfo) ScaledFontToFit(minScale float64) *Font {
	w := ti.Rect.Size.W * 0.99
	s, _, _ := ti.GetBounds()

	var r float64
	if s.W > w {
		r = w / s.W
		if r < 0.94 {
			r = math.Max(r, minScale)
		}
	} else if s.H > ti.Rect.Size.H {
		r = math.Max(5, (ti.Rect.Size.H/s.H)*1.01) // max was for all three args!!!
	} else {
		return ti.Font
	}
	return FontNew(ti.Font.Name, ti.Font.PointSize()*r, ti.Font.Style)
}
