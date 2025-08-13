package ztextinfo

import (
	"fmt"

	"github.com/torlangballe/zui/zview"
)

func set(v *zview.NativeView, add, val string) {
	v.JSStyle().Set("text-decoration-"+add, val)
}

func SetTextDecoration(v zview.View, d Decoration) {
	nv := v.Native()
	if d.LinePos == DecorationPosNone && d.Style == DecorationStyleNone && d.Width == 0 && !d.Color.Valid {
		nv.JSStyle().Set("text-decoration", "inherit")
		return
	}
	switch d.LinePos {
	case DecorationUnder:
		set(nv, "line", "underline")
	case DecorationOver:
		set(nv, "line", "overline")
	case DecorationMiddle:
		set(nv, "line", "line-though")
	}

	switch d.Style {
	case DecorationDashed:
		set(nv, "style", "dashed")
	case DecorationWavy:
		set(nv, "style", "wavy")
	case DecorationSolid:
		set(nv, "style", "solid")
	}

	if d.Width > 0 {
		set(nv, "thickness", fmt.Sprintf("%gpx ", d.Width))
	}
	if d.Color.Valid {
		set(nv, "color", d.Color.Hex())
	}
}
