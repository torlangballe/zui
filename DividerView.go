//go:build zui
// +build zui

package zui

import (
	"time"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/ztime"
)

type DividerView struct {
	CustomView
	valueChanged   func(view View)
	Vertical       bool
	startDelta     float64
	Delta          float64
	downAt         time.Time
	doubleClicking bool
	storeKey       string
}

func newDiv(storeKey string) *DividerView {
	v := &DividerView{}
	v.CustomView.Init(v, "div")
	v.SetCursor(CursorRowResize)
	v.SetDrawHandler(v.draw)
	if storeKey != "" {
		v.storeKey = storeKey
		v.Delta, _ = zkeyvalue.DefaultStore.GetDouble(storeKey, 0)
	}
	v.SetUpDownMovedHandler(func(pos zgeo.Pos, down zbool.BoolInd) {
		switch down {
		case zbool.False:
		case zbool.True:
			v.startDelta = v.Delta
			since := ztime.Since(v.downAt)
			if since > 1 {
				v.downAt = time.Time{}
			}
			if since < 0.4 {
				v.doubleClicking = true
				v.downAt = time.Time{}
				v.Delta = 0
				v.storeDelta()
				ct, _ := v.Parent().View.(ContainerType)
				ct.ArrangeChildren()
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
			v.Delta = v.startDelta + pos.Vertice(v.Vertical)
			v.storeDelta()
			ct, _ := v.Parent().View.(ContainerType)
			ct.ArrangeChildren()
		}
	})
	return v
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

func (v *DividerView) draw(rect zgeo.Rect, canvas *zcanvas.Canvas, view View) {
	canvas.SetColor(StyleDefaultFGColor())
	path := zgeo.PathNew()
	path.Circle(rect.Center(), zgeo.SizeBoth(4))
	canvas.FillPath(path)
}

func (v *DividerView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{10, 10}
}
