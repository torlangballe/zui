//go:build zui

package zlabel

import (
	"strings"

	"github.com/torlangballe/zui/zclipboard"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

//  Created by Tor Langballe on /2/11/15.

type Label struct {
	zview.NativeView
	minWidth          float64
	maxWidth          float64
	maxCalculateWidth float64
	maxLines          int
	margin            zgeo.Rect
	alignment         zgeo.Alignment
	text              string // we need to store the text as NativeView's Text() doesn't work right away
	wrap              ztextinfo.WrapType
	pressed           func()
	longPressed       func()

	Columns                      int
	KeyboardShortcut             zkeyboard.KeyMod
	pressWithModifierToClipboard zkeyboard.Modifier
}

func New(text string) *Label {
	v := &Label{}
	v.Init(v, text)
	return v
}

func NewLink(name, surl string, newWindow bool) *Label {
	v := &Label{}
	v.InitAsLink(v, name, surl, newWindow)
	return v
}

func (v *Label) GetToolTipAddition() string {
	var str string
	if !v.KeyboardShortcut.IsNull() {
		str = zview.GetShortCutTooltipAddition(v.KeyboardShortcut)
	}
	if v.pressWithModifierToClipboard == -1 {
		return str
	}
	str += "\n"

	if v.pressWithModifierToClipboard != zkeyboard.ModifierNone {
		str += v.pressWithModifierToClipboard.AsSymbolsString() + "-"
	}
	str += "press to copy to clipboard"
	return str
}

func (v *Label) SetMaxLinesBasedOnTextLines(notBeyond int) {
	count := strings.Count(v.Text(), "\n")
	m := count + 1
	if notBeyond != 0 {
		m = min(m, notBeyond)
	}
	v.SetMaxLines(m)
}

func (v *Label) SetPressWithModifierToClipboard(mod zkeyboard.Modifier) {
	// zlog.Info("Set LABEL PRESSED2Clip:", zkeyboard.ModifiersAtPress)
	if v.pressWithModifierToClipboard == -1 {
		v.pressWithModifierToClipboard = mod
		tip := v.ToolTip()
		v.SetToolTip(tip) // will add key-press tip even if tip is empty
	}
	v.SetPressedHandler("$copy2clip", mod, func() {
		// zlog.Info("LABEL PRESSED:", zkeyboard.ModifiersAtPress)
		text := v.Text()
		if strings.HasPrefix(text, "📋 ") {
			return
		}
		zclipboard.SetString(text)
		v.SetText("📋 " + text)
		ztimer.StartIn(0.6, func() {
			v.SetText(text)
		})
	})
}

func (v *Label) GetTextInfo() ztextinfo.Info {
	t := ztextinfo.New()
	t.Alignment = v.alignment
	t.Font = v.Font()
	t.Wrap = v.wrap
	if v.Columns == 0 {
		t.Text = v.Text()
	}
	if v.maxWidth != 0 {
		t.SetWidthFreeHight(v.maxWidth)
	}
	t.MaxLines = v.maxLines
	// t.MinLines = v.maxLines
	return *t
}

func (v *Label) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	var widths []float64
	to := v.View.(ztextinfo.Owner)
	ti := to.GetTextInfo()
	if v.Columns != 0 {
		s = ti.GetColumnsSize(v.Columns)
	} else {
		s, _, widths = ti.GetBounds()
	}
	s.Add(v.margin.Size.Negative())
	zfloat.Maximize(&s.W, v.minWidth)
	if v.maxWidth != 0 {
		zfloat.Minimize(&s.W, v.maxWidth)
	}
	if v.maxCalculateWidth != 0 {
		zfloat.Minimize(&s.W, v.maxCalculateWidth)
	}
	if len(widths) == 1 {
	}
	// if strings.HasPrefix(v.ObjectName(), "https://ew-ottcdn") {
	// 	zlog.Info("Label.CalculatedSize:", s, v.MaxLines(), v.Text(), ti.Font.Size/5)
	// }
	s = s.Ceil()
	s.W += 3
	return s, zgeo.SizeD(v.maxWidth, 0) // should we calculate max height?
}

func (v *Label) Text() string {
	return v.text
}

func (v *Label) IsMinimumOneLineHight() bool {
	return v.maxLines > 0
}

func (v *Label) TextAlignment() zgeo.Alignment {
	return v.alignment
}

func (v *Label) MinWidth() float64 {
	return v.minWidth
}

func (v *Label) MaxWidth() float64 {
	return v.maxWidth
}

func (v *Label) MaxLines() int {
	return v.maxLines
}

func (v *Label) SetMinWidth(min float64) {
	v.minWidth = min
}

func (v *Label) SetMaxWidth(max float64) {
	v.maxWidth = max
}

func (v *Label) SetMaxCalculateWidth(max float64) {
	v.maxCalculateWidth = max
}

func (v *Label) SetWidth(w float64) {
	v.SetMinWidth(w)
	v.SetMaxWidth(w)
}

func (v *Label) PressedHandler() func() {
	return v.pressed
}

func (v *Label) LongPressedHandler() func() {
	return v.longPressed
}

func (v *Label) HandleOutsideShortcut(sc zkeyboard.KeyMod) bool {
	if !v.KeyboardShortcut.IsNull() && sc == v.KeyboardShortcut && v.PressedHandler() != nil {
		v.PressedHandler()()
		return true
	}
	return false
}
