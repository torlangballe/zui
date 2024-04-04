//go:build zui

package ztext

import (
	"strings"

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

func MakeViewPressToClipboard(view zview.View) {
	t := view.(TextOwner)
	p := view.(zview.Pressable) // crash if misused
	// view.Native().SetSelectable(true) // no point
	p.SetPressedHandler(func() {
		text := t.Text()
		if strings.HasPrefix(text, "📋 ") {
			return
		}
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
