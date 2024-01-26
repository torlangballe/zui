//go:build zui

package zwidgets

import (
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

type ScreensInfo struct {
	Rects     map[string]zgeo.Rect
	MainID    string
	CurrentID string
}

type ScreensView struct {
	zcustom.CustomView
	ScreensInfo
	minSize zgeo.Size
}

func NewScreensView(minSize zgeo.Size) *ScreensView {
	v := &ScreensView{}
	v.CustomView.Init(v, "screens")
	v.minSize = minSize
	v.SetDrawHandler(v.draw)
	return v
}

func (v *ScreensView) draw(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	if len(v.Rects) == 0 {
		return
	}
	var union zgeo.Rect
	for _, r := range v.Rects {
		if union.IsNull() {
			union = r
		} else {
			union = rect.UnionedWith(r)
		}
	}
	intoRect := zgeo.Rect{Size: union.Size.ShunkToFill(v.minSize)}
	for id, r := range v.Rects {
		srect := zgeo.TranslateRectInSystems(r, intoRect, r)
		path := zgeo.PathNewRect(srect, zgeo.Size{})
		col := zgeo.ColorNewGray(0.8, 1)
		if id == v.CurrentID {
			col = zgeo.ColorNewGray(0.3, 1)
		}
		canvas.SetColor(col)
		canvas.FillPath(path)
		if id == v.MainID {
			canvas.DrawPath(path, zgeo.ColorBlack, 1, zgeo.PathLineButt, false)
		}
	}
}

func (v *ScreensView) SetValue(info ScreensInfo) {
	v.ScreensInfo = info
}
