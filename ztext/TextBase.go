//go:build zui

package ztext

import (
	"github.com/torlangballe/zui/zclipboard"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

type TextOwner interface {
	//Font() *Font
	Text() string
	SetText(text string)
	//Text() string
	// SetTextAlignment(a zgeo.Alignment) View
	SetFont(font *zgeo.Font)
	//TextAlignment() zgeo.Alignment
}

func MakeViewPressToClipboard(t TextOwner) {
	p := t.(zview.Pressable) // crash if misused
	p.SetPressedHandler(func() {
		text := t.Text()
		zclipboard.SetString(text)
		t.SetText("📋 " + text)
		ztimer.StartIn(0.6, func() {
			t.SetText(text)
		})
	})
	v, _ := t.(zview.View)
	if v != nil {
		v.Native().SetToolTip("press to copy to clipboard")
	}
}
