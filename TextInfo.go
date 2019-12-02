package zgo

import (
	"math"

	"github.com/torlangballe/zutil/ustr"
	"github.com/torlangballe/zutil/zgeo"
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
	t.Alignment = zgeo.AlignmentCenter
	t.Font = FontNice(FontDefaultSize, FontStyleNormal)
	t.StrokeWidth = 1
	return t
}

func (ti *TextInfo) GetBounds(noWidth bool) zgeo.Rect {
	var size = canvasGetTextSize(ti.Text, ti.Font)

	if ti.MaxLines != 0 {
		size.H = float64(ti.Font.LineHeight()) * float64(ti.MaxLines)
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
	if style&zgeo.AlignmentLeft != 0 {
		return 0
	} else if style&zgeo.AlignmentRight != 0 {
		return 0
	} else if style&zgeo.AlignmentHorCenter != 0 {
		return 0
	} else if style&zgeo.AlignmentHorJustify != 0 {
		panic("bad text adjust")
	}
	return 0 //NSTextAlignment.left
}

func (ti *TextInfo) MakeAttributes() Dictionary {
	return Dictionary{}
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
	if ti.Alignment&zgeo.AlignmentTop != 0 {
		r.SetMaxY(ra.Max().Y)
	} else if ti.Alignment&zgeo.AlignmentBottom != 0 {
		r.SetMinY(r.Max().Y - ra.Size.H)
	} else {
		//r.SetMinY(ra.Pos.Y - float64(ti.Font.LineHeight())/20)
	}

	if ti.Alignment&zgeo.AlignmentHorCenter != 0 {
		//        r = r.Expanded(ZSize(1, 0))
	}
	if ti.Alignment&zgeo.AlignmentHorShrink != 0 {
		//         ScaleFontToFit()
	}
	// canvas.SetColor(ColorYellow, 0.3)
	// canvas.FillPath(PathNewFromRect(ra, Size{}))

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
	ustr.RangeStringLines(ti.Text, false, func(s string) {
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
