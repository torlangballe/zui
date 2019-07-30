package zgo

import (
	"math"
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
	Color       Color
	Alignment   Alignment
	Font        *Font
	Rect        Rect
	Pos         *Pos
	LineSpacing float32
	StrokeWidth float32
	MaxLines    int
}

func NewTextInfo() *TextInfo {
	t := &TextInfo{}
	t.Type = TextInfoFill
	t.Wrap = TextInfoWrapWord
	t.Color = ColorBlack
	t.Alignment = AlignmentCenter
	t.Font = FontNice(18, FontNormal)
	t.StrokeWidth = 1
	return t
}

func (ti *TextInfo) GetBounds(noWidth bool) Rect {
	var size = ti.getTextSize(noWidth)

	if ti.MaxLines != 0 {
		size.H = float64(ti.Font.LineHeight()) * float64(ti.MaxLines)
	}
	return ti.Rect.Align(size, ti.Alignment, Size{}, Size{})
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

func getNativeTextAdjustment(style Alignment) int {
	if style&AlignmentLeft != 0 {
		return 0
	} else if style&AlignmentRight != 0 {
		return 0
	} else if style&AlignmentHorCenter != 0 {
		return 0
	} else if style&AlignmentHorJustify != 0 {
		panic("bad text adjust")
	}
	return 0 //NSTextAlignment.left
}

func (ti *TextInfo) MakeAttributes() Dictionary {
	return Dictionary{}
}

func (ti *TextInfo) Draw(canvas *Canvas) Rect {
	/*
		        if text.isEmpty {
		            return Rect(pos rect.pos, size ZSize(0, 0))
		        }

		        //!        let attributes = MakeAttributes()

		        switch type {
		            case ZTextDrawType.fill
		                //                    CGContextSetFillColorWithColor(canvas.context, canvaspfcolor)
		                canvas.context.setTextDrawingMode(CGTextDrawingMode.fill)

		            case ZTextDrawType.stroke
		                canvas.context.setLineWidth(CGFloat(strokeWidth))
		                //                CGContextSetFillColorWithColor(canvas.context, canvaspfcolor)
		                // CGContextSetStrokeColorWithColor(canvas.context, canvaspfcolor)
		                canvas.context.setTextDrawingMode(CGTextDrawingMode.stroke)

		            case ZTextDrawType.clip
		                canvas.context.setTextDrawingMode(CGTextDrawingMode.clip)
		        }
		        if pos == nil {
		            var r = rect
		            var ts = GetBounds().size
		            ts = ZSize(ceil(ts.w), ceil(ts.H))
		            let ra = rect.Align(ts, align alignment)
		            if (alignment & AlignmentTop) {
		                r.Max.y = ra.Max.y
		            } else if (alignment & AlignmentBottom) {
		                r.pos.y = r.Max.y - ra.size.H
		            } else {
		                r.pos.y = ra.pos.y -float64 (font.lineHeight) / 20
		            }

		            if (alignment & AlignmentHorCenter) {
		                //        r = r.Expanded(ZSize(1, 0))
		            }
		            if (alignment & AlignmentHorShrink) {
		                //         ScaleFontToFit()
		            }
		            NSString(string text).draw(in  r.GetCGRect(), withAttributes MakeAttributes())
		            return rect.Align(ts, align alignment)
		        } else {
		            NSString(string text).draw(at pos!.GetCGPoint(), withAttributes MakeAttributes())
		            return Rect.Null
				}
	*/
	return Rect{}
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

/*
    #if os(iOS)
    func CreateLayer(margin Rect = Rect())  ZTextLayer {
        let textLayer = ZTextLayer()
        textLayer.font = font
        textLayer.fontSize = font.pointSize
        textLayer.string = text
        textLayer.contentsScale = CGFloat(ZScreen.Scale)

        if alignment & .HorCenter {
          textLayer.alignmentMode = CATextLayerAlignmentMode.center
        }
        if alignment & .Left {
          textLayer.alignmentMode = CATextLayerAlignmentMode.left
        }
        if alignment & .Right {
          textLayer.alignmentMode = CATextLayerAlignmentMode.right
        }
        textLayer.foregroundColor = color.color.cgColor
        let s = (GetBounds().size + margin.size)
        textLayer.frame = Rect(size s).GetCGRect()

        return textLayer
    }
	#endif
*/
