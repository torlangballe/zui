package ztextinfo

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zutil/zfloat"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
)

//  Created by Tor Langballe on /22/10/15.

type ActionType int
type WrapType int

const (
	Fill ActionType = iota
	Stroke
	Clip
)

const (
	WrapNone WrapType = iota
	WrapWord
	WrapChar
	WrapClip
	WrapHeadTruncate
	WrapTailTruncate
	WrapMiddleTruncate
)

type Info struct {
	Type                  ActionType
	Wrap                  WrapType
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

type DecorationPos int
type DecorationStyle int
type Decoration struct {
	LinePos DecorationPos
	Style   DecorationStyle
	Width   float64
	Color   zgeo.Color
}

const (
	DecorationPosNone DecorationPos = 0
	DecorationUnder   DecorationPos = 1 << iota
	DecorationOver
	DecorationMiddle

	DecorationStyleNone DecorationStyle = 0
	DecorationSolid     DecorationStyle = 1 << iota
	DecorationWavy
	DecorationDashed
)

type Owner interface {
	GetTextInfo() Info
}

type TextSetter interface {
	SetText(text string)
}

var DecorationUnderlined = Decoration{
	LinePos: DecorationUnder,
	Style:   DecorationSolid,
	Width:   1,
}

func New() *Info {
	t := &Info{}
	t.Type = Fill
	t.Color = zgeo.ColorBlack
	t.Alignment = zgeo.Center
	t.Font = zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	t.StrokeWidth = 1
	t.MinimumFontScale = 0.5
	t.SplitItems = []string{"\r\n", "\n", "\r"}
	return t
}

var count int
var breakSet = []rune(` \t,.-_!@#$%^&**()=+<>?|:;\/`)

func Trim1FromRunes(runes []rune, wrap WrapType) []rune {
	// zlog.Info("reduceStringByOneToWrap")
	length := len(runes)
	if length == 0 {
		return runes
	}
	switch wrap {
	case WrapWord:
		var white, wasBlack bool
		var i int
		for {
			for i = length - 1; i >= 0; i-- {
				if zstr.IndexOfRuneInSet(runes[i], breakSet) == -1 { // black (non-breaking text)
					if white && wasBlack {
						return runes[:i+1]
					}
					wasBlack = true
					white = false
				} else {
					white = true
				}
			}
			if i == -1 {
				break
			}
		}
		fallthrough

	case WrapChar:
	case WrapTailTruncate:
		return runes[:length-1]

	case WrapHeadTruncate:
		return runes[1:]

	case WrapMiddleTruncate:
		m := length / 2
		return append(runes[:m], runes[m+1:]...)
	}
	return nil
}

func (ti *Info) SetWidthFreeHight(w float64) {
	ti.Rect = zgeo.RectFromWH(w, 99999)
}

// GetBounds returns rect of size of text.
// It is placed within ti.Rect using alignment
// TODO: Make it handle multi-line with some home-made wrapping stuff.
func (ti *Info) GetBounds() (size zgeo.Size, allLines []string, widths []float64) {
	// zlog.PushTimingLog()
	lines := zstr.SplitByAnyOf(ti.Text, ti.SplitItems, false)
	for _, str := range lines {
		s := zcanvas.GetTextSize(str, ti.Font)
		// zlog.Info("TI GetBounds:", str, ti.Font, s)
		// zlog.PrintTimingLog("ti bounds:", str, s.W, s.H, ti.Font.Size)
		if ti.MaxLines != 1 && ti.Rect.Size.W != 0 {
			split := s.W / ti.Rect.Size.W
			if split > 1 {
				rlen := utf8.RuneCountInString(str)
				each := int(math.Ceil(float64(rlen) / split))
				rlines := zstr.BreakRunesIntoLines([]rune(str), "", each)
				for _, rline := range rlines {
					allLines = append(allLines, string(rline))
					widths = append(widths, float64(len(rline))/float64(rlen)*s.W)
				}
				s.W = ti.Rect.Size.W
			} else {
				allLines = append(allLines, str)
				widths = append(widths, s.W)
			}
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
			size.H = float64(ti.Font.LineHeight()*1.2) * float64(zint.Max(count, 1))
		}
		// zlog.Info("BOUNDS:", size, count, ti.Text)
	}
	// zlog.PopTimingLog()
	// zlog.Info("BOUNDS:", len(allLines), ti.Text)
	return size, allLines, widths
}

func (ti *Info) MakeAttributes() zdict.Dict {
	return zdict.Dict{}
}

func reduceTextToFit(ti *Info) {
	if ti.Wrap == WrapClip || ti.Wrap == WrapNone {
		return
	}
	runes := []rune(ti.Text)
	for {
		s := zcanvas.GetTextSize(ti.Text, ti.Font)
		zlog.Info("REDUCE1:", ti.Wrap, ti.Text, s, ti.Rect.Size.W)
		if s.W <= ti.Rect.Size.W {
			// zlog.Info("REDUCED:", lines)
			return
		}
		runes = Trim1FromRunes(runes, ti.Wrap)
		ti.Text = string(runes) + "…"
		if len(runes) == 0 {
			return
		}
		// if ti.Text == "no…" {
		// 	zlog.Info("REDUCE:", ti.Wrap, ti.Text, s, ti.Rect.Size.W, zlog.CallingStackString())
		// }
	}
}

// StrokeAndFill strokes the text with *strokeColor* and width, then fills with ti.Color
// canvas' width, color and stroke style are changed
func (ti *Info) StrokeAndFill(canvas *zcanvas.Canvas, strokeColor zgeo.Color, width float64) zgeo.Rect {
	t := *ti
	old := t.Color
	t.Color = strokeColor
	t.Type = Stroke
	t.StrokeWidth = width
	t.Draw(canvas)
	t.Color = old
	t.Type = Fill
	return t.Draw(canvas)
}

func (tin *Info) Draw(canvas *zcanvas.Canvas) zgeo.Rect {
	ti := *tin
	w := 0.0
	if ti.Text == "" {
		return zgeo.Rect{ti.Rect.Pos, zgeo.Size{}}
	}
	canvas.SetColor(ti.Color)
	if ti.Type == Stroke {
		w = ti.StrokeWidth
	}
	if ti.Alignment&zgeo.HorShrink != 0 {
		zlog.Assert(!ti.Rect.Size.IsNull())
		// s := ti.Font.Size
		ti.Font = ti.ScaledFontToFit(ti.MinimumFontScale)
		// zlog.Warn("Font Scaled:", s, "->", ti.Font.Size, ti.Text)
	}
	// fmt.Println("CANVAS TI SetFont:", font)
	canvas.SetFont(ti.Font, nil)
	if ti.Rect.Size.IsNull() {
		canvas.DrawTextInPos(ti.Rect.Pos, ti.Text, w)
		return zgeo.Rect{}
	}
	reduceTextToFit(&ti)
	ts, lines, widths := ti.GetBounds()
	ts = zgeo.Size{math.Ceil(ts.W), math.Ceil(ts.H)}
	ra := ti.Rect.Align(ts, ti.Alignment, ti.Margin)
	// https://stackoverflow.com/questions/5026961/html5-canvas-ctx-filltext-wont-do-line-breaks/21574562#21574562
	h := ti.Font.LineHeight()
	y := ra.Pos.Y + h*0.73 // 0.71
	zlog.Assert(len(lines) == len(widths) || len(widths) == 0, len(lines), len(widths))
	for i, s := range lines {
		x := ra.Pos.X
		if len(widths) > 0 {
			if ti.Alignment&zgeo.HorCenter != 0 {
				x += (ra.Size.W - widths[i]) / 2
			} else if ti.Alignment&zgeo.Right != 0 {
				x += (ra.Size.W - widths[i])
			}
		}
		canvas.DrawTextInPos(zgeo.Pos{x, y}, s, w)
		y += h
	}
	return ra
}

func (ti *Info) ScaledFontToFit(minScale float64) *zgeo.Font {
	mr := ti.Rect.Expanded(ti.Margin.Negative())
	w := mr.Size.W
	if ti.Alignment&zgeo.HorCenter != 0 {
		w -= ti.Margin.W // ???
	}
	t2 := *ti
	t2.Rect = zgeo.Rect{}
	s, _, _ := t2.GetBounds()

	// zlog.Warn("XScale1:", s.W, w, ti.Text)
	var r float64
	if s.W > w {
		r = math.Floor(w) / math.Ceil(s.W+1)
		if r < 0.94 {
			r = math.Max(r, minScale)
		}
		// zlog.Warn("XScale:", s.W*r, w, r, ti.Text)
	} else if s.H > mr.Size.H {
		r = math.Max(5, (w/s.H)*1.01)
	} else {
		return ti.Font
	}
	return zgeo.FontNew(ti.Font.Name, ti.Font.PointSize()*r, ti.Font.Style)
}

func WidthOfString(str string, font *zgeo.Font) float64 {
	ti := New()
	ti.Alignment = zgeo.Left
	ti.Text = str
	ti.IsMinimumOneLineHight = true
	ti.Font = font
	ti.MaxLines = 1
	s, _, _ := ti.GetBounds()
	return s.W
}

func (ti *Info) GetColumnsSize(cols int) zgeo.Size {
	var temp = *ti
	temp.Text = ""
	const letters = "etaoinsrhdlucmfywgpbvkxqjz"
	len := len(letters)
	for i := 0; i < cols; i++ {
		c := string(letters[i%len])
		if i%8 == 4 || cols <= 4 {
			c = strings.ToUpper(c)
		}
		temp.Text += c
	}
	s, _, _ := temp.GetBounds()
	return s
}

func (ti *Info) GetRect() zgeo.Rect {
	zlog.Assert(!ti.Rect.IsNull())
	box, _, _ := ti.GetBounds()
	return ti.Rect.Align(box, ti.Alignment, ti.Margin)
}
