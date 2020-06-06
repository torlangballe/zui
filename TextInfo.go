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
	Type             TextInfoType
	Wrap             TextInfoWrap
	Text             string
	Color            zgeo.Color
	Alignment        zgeo.Alignment
	Font             *Font
	Rect             zgeo.Rect
	LineSpacing      float64
	StrokeWidth      float64
	MaxLines         int
	MinimumFontScale float64
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
	return t
}

// GetBounds returns rect of size of text.
// It is placed within ti.Rect using alignment
// TODO: Make it handle multi-line with some home-made wrapping stuff.
func (ti *TextInfo) GetBounds(noWidth bool) zgeo.Size {
	var size zgeo.Size
	lines := zstr.SplitByNewLines(ti.Text, true)
	for _, str := range lines {
		s := canvasGetTextSize(str, ti.Font)
		size.H = s.H
		zfloat.Maximize(&size.W, s.W)
	}
	count := zint.Max(ti.MaxLines, len(lines))
	count = ti.MaxLines
	if count > 1 {
		size.H = float64(ti.Font.LineHeight()) * float64(count)
	}
	// if !ti.Rect.IsNull() {
	// 	zfloat.Minimize(&size.W, ti.Rect.Size.W)
	// }
	return size
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
	t.Draw(canvas)
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
	if ti.Rect.Size.IsNull() {
		canvas.DrawTextInPos(ti.Rect.Pos, ti.Text, w)
		return zgeo.Rect{}
	}
	r := ti.Rect
	var ts = ti.GetBounds(false)
	ts = zgeo.Size{math.Ceil(ts.W), math.Ceil(ts.H)}
	ra := ti.Rect.Align(ts, ti.Alignment, zgeo.Size{}, zgeo.Size{})
	if ti.Alignment&zgeo.Top != 0 {
		r.SetMaxY(ra.Max().Y) // this
	} else if ti.Alignment&zgeo.Bottom != 0 {
		r.SetMinY(r.Max().Y - ra.Size.H) // and this are on r and not ra? Not used??
	} else {
		//r.SetMinY(ra.Pos.Y - float64(ti.Font.LineHeight())/20)
	}
	font := ti.Font
	if ti.Alignment&zgeo.HorCenter != 0 {
		//        r = r.Expanded(ZSize(1, 0))
	}
	if ti.Alignment&zgeo.HorShrink != 0 {
		font = ti.ScaledFontToFit(ti.MinimumFontScale)
	}
	// https://stackoverflow.com/questions/5026961/html5-canvas-ctx-filltext-wont-do-line-breaks/21574562#21574562
	h := font.LineHeight()
	y := ra.Pos.Y + h*0.71
	// zlog.Info("TI.Draw:", ti.Rect, font.Size, ti.Text)
	canvas.SetFont(font, nil)
	zstr.RangeStringLines(ti.Text, false, func(s string) {
		x := ra.Pos.X
		// tsize := canvasGetTextSize(s, ti.Font)
		canvas.DrawTextInPos(zgeo.Pos{x, y}, s, w)
		y += h
	})
	return ra
}

func (ti *TextInfo) ScaledFontToFit(minScale float64) *Font {
	w := ti.Rect.Size.W * 0.99
	noWidth := true
	s := ti.GetBounds(noWidth)

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
	zlog.Info("Scale Font:", r, w, s.W)
	return FontNew(ti.Font.Name, ti.Font.PointSize()*r, ti.Font.Style)
}
