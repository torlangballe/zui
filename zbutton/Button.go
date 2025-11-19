//go:build zui

package zbutton

import (
	"github.com/torlangballe/zui/zdocs"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /19/apr/21.

type Button struct {
	zview.NativeView
	KeyboardShortcut zkeyboard.ShortCut

	minWidth float64
	maxWidth float64
	margin   zgeo.Rect
}

func (v *Button) GetTextInfo() ztextinfo.Info {
	t := ztextinfo.New()
	t.Font = v.Font()
	t.Text = v.InnerText()
	if v.maxWidth != 0 {
		t.SetWidthFreeHight(v.maxWidth)
	}
	t.MaxLines = 1
	t.IsMinimumOneLineHight = true
	return *t
}

func (v *Button) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	to := v.View.(ztextinfo.Owner)
	ti := to.GetTextInfo()
	s, _, _ = ti.GetBounds()
	s.Add(v.margin.Size.Negative())
	zfloat.Maximize(&s.W, v.minWidth)
	if v.maxWidth != 0 {
		zfloat.Minimize(&s.W, v.maxWidth)
	}
	s = s.Ceil()
	s.W += 4
	// zlog.Info("Button CS:", v.ObjectName(), s)
	return s, zgeo.SizeD(0, s.H-2)
}

func (v *Button) MinWidth() float64 {
	return v.minWidth
}

func (v *Button) MaxWidth() float64 {
	return v.maxWidth
}

func (v *Button) SetMinWidth(min float64) {
	v.minWidth = min
}

func (v *Button) SetMaxWidth(max float64) {
	v.maxWidth = max
}

func (v *Button) GetToolTipAddition() string {
	var str string
	if !v.KeyboardShortcut.IsNull() {
		str = zview.GetShortCutTooltipAddition(v.KeyboardShortcut.KeyMod)
	}
	return str
}

func (v *Button) HandleShortcut(sc zkeyboard.KeyMod, inFocus bool) bool {
	if !v.KeyboardShortcut.IsNull() && sc == v.KeyboardShortcut.KeyMod {
		v.ClickAll()
		return true
	}
	return false
}

// func (v *Button) HandleOutsideShortcut(sc zkeyboard.KeyMod, isWithinFocus bool) bool {
// 	if !isWithinFocus {
// 		return false
// 	}
// 	if !v.KeyboardShortcut.IsNull() && sc == v.KeyboardShortcut {
// 		v.ClickAll()
// 		return true
// 	}
// 	return false
// }

func (v *Button) GetSearchableItems(currentPath []zdocs.PathPart) []zdocs.SearchableItem {
	var parts []zdocs.SearchableItem
	item := zdocs.MakeSearchableItem(currentPath, zdocs.StaticField, "", "", v.Text())
	parts = append(parts, item)
	tip := v.ToolTip()
	if tip != "" {
		item = zdocs.MakeSearchableItem(currentPath, zdocs.StaticField, "Tip", "Tip", tip)
		parts = append(parts, item)
	}
	return []zdocs.SearchableItem{item}
}
