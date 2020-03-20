package zui

import (
	"math"

	"github.com/torlangballe/zutil/zfloat"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
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
)

type TextInfo struct {
	Type        TextInfoType
	Wrap        TextInfoWrap
	Text        string
	Color       zgeo.Color
	Alignment   zgeo.Alignment
	Font        *Font
	Rect        zgeo.Rect
	LineSpacing float32
	StrokeWidth float32
	MaxLines    int
}

func TextInfoNew() *TextInfo {
	t := &TextInfo{}
	t.Type = TextInfoFill
	t.Wrap = TextInfoWrapWord
	t.Color = zgeo.ColorBlack
	t.Alignment = zgeo.Center
	t.Font = FontNice(FontDefaultSize, FontStyleNormal)
	t.StrokeWidth = 1
	return t
}

// GetBounds returns rect of size of text.
// It is placed within ti.Rect using alignment
// TODO: Make it handle multi-line with some home-made wrapping stuff.
func (ti *TextInfo) GetBounds(noWidth bool) zgeo.Rect {
	var size = canvasGetTextSize(ti.Text, ti.Font)

	if ti.MaxLines != 0 {
		size.H = float64(ti.Font.LineHeight()) * float64(ti.MaxLines)
	}
	if !ti.Rect.IsNull() {
		zfloat.Minimize(&size.W, ti.Rect.Size.W)
	}
	return ti.Rect.Align(size, ti.Alignment, zgeo.Size{}, zgeo.Size{})
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

func (ti *TextInfo) Draw(canvas *Canvas) zgeo.Rect {
	if ti.Text == "" {
		return zgeo.Rect{ti.Rect.Pos, zgeo.Size{}}
	}
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
	if ti.Rect.Size.IsNull() {
		canvas.DrawTextInPos(ti.Rect.Pos, ti.Text, ti.MakeAttributes())
		return zgeo.Rect{}
	}
	r := ti.Rect
	var ts = ti.GetBounds(false).Size
	ts = zgeo.Size{math.Ceil(ts.W), math.Ceil(ts.H)}
	ra := ti.Rect.Align(ts, ti.Alignment, zgeo.Size{}, zgeo.Size{})
	if ti.Alignment&zgeo.Top != 0 {
		r.SetMaxY(ra.Max().Y)
	} else if ti.Alignment&zgeo.Bottom != 0 {
		r.SetMinY(r.Max().Y - ra.Size.H)
	} else {
		//r.SetMinY(ra.Pos.Y - float64(ti.Font.LineHeight())/20)
	}

	if ti.Alignment&zgeo.HorCenter != 0 {
		//        r = r.Expanded(ZSize(1, 0))
	}
	if ti.Alignment&zgeo.HorShrink != 0 {
		//         ScaleFontToFit()
	}
	// canvas.SetColor(ColorYellow, 0.3)
	// canvas.FillPath(PathNewRect(ra, Size{}))

	//	fmt.Println("drawTextInRect:", ts, ra, ti.Rect, ti.Alignment)

	return ti.drawTextInRect(canvas, ra)
}

func (ti *TextInfo) drawTextInRect(canvas *Canvas, rect zgeo.Rect) zgeo.Rect {
	// https://stackoverflow.com/questions/5026961/html5-canvas-ctx-filltext-wont-do-line-breaks/21574562#21574562
	h := ti.Font.LineHeight()
	y := rect.Pos.Y + h*0.9
	attributes := ti.MakeAttributes()
	canvas.SetFont(ti.Font, nil)
	canvas.SetColor(zgeo.ColorWhite, 1)
	zstr.RangeStringLines(ti.Text, false, func(s string) {
		x := rect.Pos.X
		// tsize := canvasGetTextSize(s, ti.Font)
		canvas.DrawTextInPos(zgeo.Pos{x, y}, s, attributes)
		y += h
	})
	return rect
}

func (ti *TextInfo) ScaleFontToFit(minScale float64) {
	w := ti.Rect.Size.W * 0.95
	noWidth := true
	s := ti.GetBounds(noWidth).Size

	var r float64
	if s.W > w {
		r = w / s.W
		if r < 0.94 {
			r = math.Max(r, minScale)
		}
	} else if s.H > ti.Rect.Size.H {
		r = math.Max(5, (ti.Rect.Size.H/s.H)*1.01) // max was for all three args!!!
	}
	ti.Font = FontNew(ti.Font.Name, ti.Font.PointSize()*r, ti.Font.Style)
}
