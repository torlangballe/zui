//go:build zui

package zwidgets

import (
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

type ScreensInfo struct {
	Rects     map[string]zgeo.Rect // screenid:rect
	MainID    string
	CurrentID string
}

type ScreensView struct {
	zcustom.CustomView
	ScreensInfo
}

func NewScreensView(minSize zgeo.Size) *ScreensView {
	// zlog.Info("NewScreensView", zlog.CallingStackString())
	v := &ScreensView{}
	v.CustomView.Init(v, "screens")
	v.SetBGColor(zgeo.ColorWhite)
	v.SetCorner(5)
	v.SetMinSize(minSize)
	v.SetStroke(1, zgeo.ColorGray, false)
	v.SetDrawHandler(v.draw)
	return v
}

func (v *ScreensView) draw(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	// zlog.Info("ScreensView draw", len(v.ScreensInfo.Rects), zlog.Pointer(v))
	if len(v.Rects) == 0 {
		return
	}
	var union zgeo.Rect
	for _, r := range v.Rects {
		if union.IsNull() {
			union = r
		} else {
			union = union.UnionedWith(r)
		}
	}
	intoRect := zgeo.Rect{Size: union.Size.ScaledTo(v.MinSize())}
	tinfo := ztextinfo.New()
	tinfo.Font = zgeo.FontNice(intoRect.Size.H/8, zgeo.FontStyleNormal)
	tinfo.Alignment = zgeo.Center
	for id, r := range v.Rects {
		srect := zgeo.TranslateRectInSystems(union, intoRect, r)
		// zlog.Info("Draw", id, r, union, srect)
		path := zgeo.PathNewRect(srect, zgeo.Size{})
		col := zgeo.ColorNewGray(0.8, 1)
		if id == v.CurrentID {
			col = zgeo.ColorNewGray(0.3, 1)
		}
		canvas.SetColor(col)
		tinfo.Color = col.ContrastingGray()
		canvas.FillPath(path)
		if id == v.MainID {
			canvas.DrawPath(path, zgeo.ColorBlack, 1, zgeo.PathLineButt, false)
		}
		tinfo.Rect = srect
		tinfo.Text = id
		tinfo.Draw(canvas)
	}
}

func (v *ScreensView) SetValue(info ScreensInfo) {
	v.ScreensInfo = info
	// zlog.Info("Sc.SetValue:", len(v.ScreensInfo.Rects), zlog.Pointer(v), zlog.CallingStackString())
	v.Expose()
}
