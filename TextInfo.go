package zui

import (
	"math"
	"strings"
	"unicode/utf8"

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
	TextInfoWrapNone TextInfoWrap = iota
	TextInfoWrapWord
	TextInfoWrapChar
	TextInfoWrapClip
	TextInfoWrapHeadTruncate
	TextInfoWrapTailTruncate
	TextInfoWrapMiddleTruncate

)

type TextInfo struct {
	Type                  TextInfoType
	Wrap                  TextInfoWrap
	Text                  string
	Color                 zgeo.Color
	Alignment             zgeo.Alignment
	Font                  *zgeo.Font
	Rect                  zgeo.Rect
	LineSpacing           float64
	StrokeWidth           float64
	MaxLines              int
	MinLines              int
	MinimumFontScale      float64
	IsMinimumOneLineHight bool
	SplitItems            []string
	Margin                zgeo.Size
}

type TextInfoOwner interface {
	GetTextInfo() TextInfo
}

type TextSetter interface {
	SetText(etxt string)
}

func TextInfoNew() *TextInfo {
	t := &TextInfo{}
	t.Type = TextInfoFill
	t.Color = zgeo.ColorBlack
	t.Alignment = zgeo.Center
	t.Font = zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	t.StrokeWidth = 1
	t.MinimumFontScale = 0.5
	t.SplitItems = []string{"\r\n", "\n", "\r"}
	return t
}

func reduceStringByOneToWrap(str string, wrap TextInfoWrap) (reduced, withSymbol string) {
	if str == "" {
		return
	}
	switch wrap {
	case TextInfoWrapWord:
		i := strings.IndexAny(str, " \t,.-_!@#$%^&**()=+<>?|:;")
		if i != -1 {
			s := str[:i]
			return s, s
		}
		fallthrough
	case TextInfoWrapChar:
		s := zstr.TruncatedCharsAtEnd(str, 1)
		return s, s

	case TextInfoWrapHeadTruncate:
		s := str[1:]
		return s, "…" + s

	case TextInfoWrapTailTruncate:
		s := zstr.TruncatedCharsAtEnd(str, 1)
		return s, s + "…"

	case TextInfoWrapMiddleTruncate:
		r := []rune(str)
		m := len(r) / 2
		left := string(r[:m])
		right := string(r[m+1:])
		return left + right, left + "…" + right
	}
	return str, str
}

func (ti *TextInfo) SetWidthFreeHight(w float64) {
	ti.Rect = zgeo.RectFromWH(w, 99999)
}

// GetBounds returns rect of size of text.
// It is placed within ti.Rect using alignment
// TODO: Make it handle multi-line with some home-made wrapping stuff.
func (ti *TextInfo) GetBounds() (size zgeo.Size, allLines []string, widths []float64) {
	// zlog.PushTimingLog()
	lines := zstr.SplitByAnyOf(ti.Text, ti.SplitItems, false)
	for _, str := range lines {
		s := canvasGetTextSize(str, ti.Font)
		// zlog.PrintTimingLog("ti bounds:", str, s.W, s.H, ti.Font.Size)
		if ti.MaxLines != 1 && ti.Rect.Size.W != 0 {
			split := s.W / ti.Rect.Size.W
			// zlog.Info("TI split:", split, str)
			if split > 1 {
				rlen := utf8.RuneCountInString(str)
				each := int(math.Ceil(float64(rlen) / split))
				rlines := zstr.BreakIntoRuneLines(str, "", each)
				for _, rline := range rlines {
					allLines = append(allLines, string(rline))
					widths = append(widths, float64(len(rline))/float64(rlen)*s.W)
				}
				s.W = ti.Rect.Size.W
			} else {
				allLines = append(allLines, str)
				widths = append(widths, s.W)
			}
			// zlog.Info("SPLIT!", splitToLines, ti.Rect.Size, s)
		} else {
			allLines = append(allLines, str)
		}
		zfloat.Maximize(&size.W, s.W)
		zfloat.Maximize(&size.H, s.H)
		// zlog.Info("TI GetBounds:", str, size.W, s.W)
	}
	// zlog.PrintTimingLog("ti bounds looped")
	if ti.MaxLines == 1 { //|| ti.Rect.Size.W == 0 {
		allLines = []string{ti.Text}
		widths = []float64{size.W}
	} else {
		count := len(allLines)
		if ti.MaxLines != 0 {
			zint.Minimize(&count, ti.MaxLines)
		}
		if ti.MinLines != 0 {
			zint.Maximize(&count, ti.MinLines)
		}
		//	count = ti.MaxLines
		if count > 1 || ti.IsMinimumOneLineHight {
			size.H = float64(ti.Font.LineHeight()*1.1) * float64(zint.Max(count, 1))
		}
		// zlog.Info("BOUNDS:", size, count, ti.Text)
	}
	// zlog.PopTimingLog()
	// zlog.Info("BOUNDS:", len(allLines), ti.Text)
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
	} else if style&zgeo.HorExpand != 0 {
		panic("bad text adjust")
	}
	return 0 //NSTextAlignment.left
}

func (ti *TextInfo) MakeAttributes() zdict.Dict {
	return zdict.Dict{}
}

// StrokeAndFill strokes the text with *strokeColor* and width, then fills with ti.Color
// canvas' width, color and stroke style are changed
func (ti *TextInfo) StrokeAndFill(canvas *Canvas, strokeColor zgeo.Color, width float64) zgeo.Rect {
	t := *ti
	old := t.Color
	t.Color = strokeColor
	t.Type = TextInfoStroke
	t.StrokeWidth = width
	t.Draw(canvas)
	t.Color = old
	t.Type = TextInfoFill
	return t.Draw(canvas)
}

func (ti *TextInfo) Draw(canvas *Canvas) zgeo.Rect {
	w := 0.0
	if ti.Text == "" {
		return zgeo.Rect{ti.Rect.Pos, zgeo.Size{}}
	}
	canvas.SetColor(ti.Color)
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
	var ts zgeo.Size
	var lines []string
	var widths []float64
	text := ti.Text
	for {
		ts, lines, widths = ti.GetBounds()
		if ti.Wrap == TextInfoWrapClip || ti.Wrap == TextInfoWrapNone || len(lines) != 1 || ts.W <= ti.Rect.Size.W {
			// zlog.Info("REDUCED:", lines)
			break
		}
		text, ti.Text = reduceStringByOneToWrap(text, ti.Wrap)
		// zlog.Info("REDUCE:", ti.Wrap, ti.Text, ts.W, ti.Rect.Size.W, lines)
	}
	ts = zgeo.Size{math.Ceil(ts.W), math.Ceil(ts.H)}
	ra := ti.Rect.Align(ts, ti.Alignment, ti.Margin)
	// https://stackoverflow.com/questions/5026961/html5-canvas-ctx-filltext-wont-do-line-breaks/21574562#21574562
	h := font.LineHeight()
	y := ra.Pos.Y + h*0.95 // 0.71
	// zlog.Info("TI.Draw:", ti.Rect, font.Size, ti.Text)
	zlog.Assert(len(lines) == len(widths), len(lines), len(widths))
	for i, s := range lines {
		// zlog.Info("DRAW:", i, len(lines), s, ti.Rect, ts, ra, "h:", h, "y:", y)
		x := ra.Pos.X
		if ti.Alignment&zgeo.HorCenter != 0 {
			x += (ra.Size.W - widths[i]) / 2
		} else if ti.Alignment&zgeo.Right != 0 {
			x += (ra.Size.W - widths[i])
		}
		canvas.DrawTextInPos(zgeo.Pos{x, y}, s, w)
		y += h
	}
	return ra
}

func (ti *TextInfo) ScaledFontToFit(minScale float64) *zgeo.Font {
	w := ti.Rect.Size.W
	w -= ti.Margin.W
	if ti.Alignment&zgeo.HorCenter != 0 {
		w -= ti.Margin.W
	}
	s, _, _ := ti.GetBounds()

	var r float64
	if s.W > w {
		r = w / s.W
		if r < 0.94 {
			r = math.Max(r, minScale)
		}
	} else if s.H > ti.Rect.Size.H {
		r = math.Max(5, (w/s.H)*1.01)
	} else {
		return ti.Font
	}
	return zgeo.FontNew(ti.Font.Name, ti.Font.PointSize()*r, ti.Font.Style)
}

func TextInfoWidthOfString(str string, font *zgeo.Font) float64 {
	ti := TextInfoNew()
	ti.Alignment = zgeo.Left
	ti.Text = str
	ti.IsMinimumOneLineHight = true
	ti.Font = font
	ti.MaxLines = 1
	s, _, _ := ti.GetBounds()
	return s.W
}

func (ti *TextInfo) GetColumnsSize(cols int) zgeo.Size {
	var temp = *ti
	temp.Text = ""
	const letters = "etaoinsrhdlucmfywgpbvkxqjz"
	len := len(letters)
	for i := 0; i < cols; i++ {
		c := string(letters[i%len])
		if i%8 == 4 {
			c = strings.ToUpper(c)
		}
		temp.Text += c
	}
	s, _, _ := temp.GetBounds()
	return s
}
