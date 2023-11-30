//go:build zui

package ztext

import (
	"fmt"

	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
)

func set(v *zview.NativeView, add, val string) {
	v.JSStyle().Set("text-decoration-"+add, val)
}

func SetTextDecoration(v *zview.NativeView, d ztextinfo.Decoration) {
	if d.LinePos == ztextinfo.DecorationPosNone && d.Style == ztextinfo.DecorationStyleNone && d.Width == 0 && !d.Color.Valid {
		v.JSStyle().Set("text-decoration", "inherit")
		return
	}
	switch d.LinePos {
	case ztextinfo.DecorationUnder:
		set(v, "line", "underline")
	case ztextinfo.DecorationOver:
		set(v, "line", "overline")
	case ztextinfo.DecorationMiddle:
		set(v, "line", "line-though")
	}

	switch d.Style {
	case ztextinfo.DecorationDashed:
		set(v, "style", "dashed")
	case ztextinfo.DecorationWavy:
		set(v, "style", "wavy")
	case ztextinfo.DecorationSolid:
		set(v, "style", "solid")
	}

	if d.Width > 0 {
		set(v, "thickness", fmt.Sprintf("%gpx ", d.Width))
	}
	if d.Color.Valid {
		set(v, "color", d.Color.Hex())
	}
}
