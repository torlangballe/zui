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
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/ztime"
)

type DividerView struct {
	zcustom.CustomView
	valueChanged   func(view zview.View)
	Vertical       bool
	startRatio     float64
	startPos       float64
	delta          float64
	Ratio          float64
	downAt         time.Time
	doubleClicking bool
	storeKey       string
}

func newDiv(storeKey string, defaultRatio float64) *DividerView {
	v := &DividerView{}
	v.CustomView.Init(v, "divider")
	v.SetCursor(zcursor.RowResize)
	v.SetDrawHandler(v.draw)
	v.storeKey = "zcontainer.Div." + storeKey
	v.startRatio = defaultRatio
	v.SetPressUpDownMovedHandler(func(pos zgeo.Pos, down zbool.BoolInd) bool {
		ar := v.Parent().AbsoluteRect()
		height := ar.Size.Element(v.Vertical)
		divPos := v.AbsoluteRect().Pos.Y - ar.Pos.Y
		switch down {
		case zbool.False:
			zview.SkipEnterHandler = false
		case zbool.True:
			zview.SkipEnterHandler = true
			since := ztime.Since(v.downAt)
			if since > 1 {
				v.downAt = time.Time{}
			}
			if since < 0.4 {
				v.doubleClicking = true
				v.downAt = time.Time{}
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
			pos := pos.Element(v.Vertical) + divPos
			v.Ratio = (pos / height)
			v.storeRatio()
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
			ratio, got := zkeyvalue.DefaultSessionStore.GetDouble(v.storeKey, 0)
			if got {
				v.startRatio = ratio
			}
		}
		v.Ratio = v.startRatio
	}
}

func (v *DividerView) storeRatio() {
	if v.storeKey != "" {
		zkeyvalue.DefaultSessionStore.SetDouble(v.Ratio, v.storeKey, true)
	}
}

func DividerViewNewVert(storeKey string, defaultRatio float64) *DividerView {
	v := newDiv(storeKey, defaultRatio)
	v.Vertical = true
	return v
}

func DividerViewNewHor(storeKey string, defaultRatio float64) *DividerView {
	v := newDiv(storeKey, defaultRatio)
	v.Vertical = false
	return v
}

func (v *DividerView) draw(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	canvas.SetColor(zstyle.DefaultFGColor())
	path := zgeo.PathNew()
	path.Circle(rect.Center(), zgeo.SizeBoth(4))
	canvas.FillPath(path)
}

func (v *DividerView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	const h = 10
	return zgeo.SizeD(10, h), zgeo.SizeD(0, h)
}
