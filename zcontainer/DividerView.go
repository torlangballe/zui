//go:build zui

package zcontainer

import (
	"time"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcursor"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztime"
)

type DividerView struct {
	zcustom.CustomView
	valueChanged   func(view zview.View)
	Vertical       bool
	startDelta     float64
	Delta          float64
	downAt         time.Time
	doubleClicking bool
	storeKey       string
}

func newDiv(storeKey string) *DividerView {
	v := &DividerView{}
	v.CustomView.Init(v, "divider")
	v.SetCursor(zcursor.RowResize)
	v.SetDrawHandler(v.draw)
	v.storeKey = storeKey
	v.SetPressUpDownMovedHandler(func(pos zgeo.Pos, down zbool.BoolInd) bool {
		abs := v.AbsoluteRect().Pos.Vertice(v.Vertical) + pos.Vertice(v.Vertical)
		switch down {
		case zbool.False:
			zview.SkipEnterHandler = false
		case zbool.True:
			zview.SkipEnterHandler = true
			v.startDelta = v.Delta - abs
			since := ztime.Since(v.downAt)
			if since > 1 {
				v.downAt = time.Time{}
			}
			if since < 0.4 {
				v.doubleClicking = true
				v.downAt = time.Time{}
				v.Delta = 0
				v.storeDelta()
				at, _ := v.Parent().View.(Arranger)
				at.ArrangeChildren()
			} else {
				v.doubleClicking = false
				if v.downAt.IsZero() {
					v.downAt = time.Now()
				} else {
					v.downAt = time.Time{}
				}
			}
		case zbool.Unknown:
			if v.doubleClicking {
				break
			}
			zfloat.Maximize(&abs, v.Parent().AbsoluteRect().Pos.Vertice(v.Vertical))
			v.Delta = v.startDelta + abs
			// zlog.Info("DivDelta:", v.startDelta, pos.Vertice(v.Vertical), v.Delta)
			v.storeDelta()
			at, _ := v.Parent().View.(Arranger)
			at.ArrangeChildren()
		}
		return true
	})
	return v
}

func (v *DividerView) ReadyToShow(beforeWindow bool) {
	v.CustomView.ReadyToShow(beforeWindow)
	if beforeWindow {
		if v.storeKey != "" {
			delta, got := zkeyvalue.DefaultStore.GetDouble(v.storeKey, 0)
			if got {
				v.Delta = delta
				// ArrangeChildrenAtRootContainer(v)
				// v.Expose()
			}
			zlog.Info("DIV Delta:", delta)
		}
	}
}

func (v *DividerView) storeDelta() {
	if v.storeKey != "" {
		zkeyvalue.DefaultStore.SetDouble(v.Delta, v.storeKey, true)
	}
}

func DividerViewNewVert(storeKey string) *DividerView {
	v := newDiv(storeKey)
	v.Vertical = true
	return v
}

func DividerViewNewHor(storeKey string) *DividerView {
	v := newDiv(storeKey)
	v.Vertical = false
	return v
}

func (v *DividerView) draw(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	canvas.SetColor(zstyle.DefaultFGColor())
	path := zgeo.PathNew()
	path.Circle(rect.Center(), zgeo.SizeBoth(4))
	canvas.FillPath(path)
}

func (v *DividerView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{10, 10}
}
