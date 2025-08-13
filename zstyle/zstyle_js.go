package zstyle

import (
	"github.com/torlangballe/zui/zview"
)

func SetStyling(v zview.View, style Styling) {
	nv := v.Native()
	if style.DropShadow.Color.Valid {
		nv.SetDropShadow(style.DropShadow)
	}
	if style.BGColor.Valid {
		nv.SetBGColor(style.BGColor)
	}
	if style.Corner != -1 {
		nv.SetCorner(style.Corner)
	}
	if style.StrokeColor.Valid {
		// zlog.Info("SetStyling:", nv.Hierarchy(), style.StrokeWidth, style.StrokeColor)
		nv.SetStroke(style.StrokeWidth, style.StrokeColor, style.StrokeIsInset.IsTrue())
	}
	if style.OutlineColor.Valid {
		nv.SetOutline(style.OutlineWidth, style.OutlineColor, style.OutlineOffset)
	}
	if style.FGColor.Valid {
		nv.SetColor(style.FGColor)
	}
	if style.Font.Name != "" {
		nv.SetFont(&style.Font)
	}
	if !style.Margin.IsUndef() {
		m, _ := v.(zview.Marginalizer)
		if m != nil {
			m.SetMargin(style.Margin)
		}
	}
}
