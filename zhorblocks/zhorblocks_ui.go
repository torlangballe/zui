//go:build zui && js

package zhorblocks

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zscrollview"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmath"
	"github.com/torlangballe/zutil/ztimer"
)

type HorBlocksView struct {
	zcontainer.StackView
	CacheDelta      int // CacheDelta is how many more than center view to cache on each side
	IndexWindow     int // IndexWindow is how many more than center view to get
	GetViewFunc     func(index int) zview.View
	ReleaseViewFunc func(index int)
	PanHandler      PanHandler
	IgnoreScroll    bool
	Overlay         *zcontainer.StackView
	Scroller        *zcontainer.StackView

	currentIndex                 int
	maxIndex                     int // this can change over time
	updating                     bool
	viewSize                     zgeo.Size
	fixedHeight                  float64
	flippedAt                    time.Time
	dragXStart                   float64
	flipping                     bool
	presented                    bool
	oldScrollX                   float64
	hasUpdated                   bool
	oldRect                      zgeo.Rect
	queuedGetViews               map[int]bool
	queLock                      sync.Mutex
	scrollToIndexAfterAllUpdates float64
}

type PanHandler interface {
	HandlePan(index float64)
}

func (v *HorBlocksView) Init(indexWindow, cacheDelta int) {
	v.CacheDelta = cacheDelta
	v.IndexWindow = indexWindow
	v.StackView.Init(v, false, "hor-inf")
	// v.SetMargin(zgeo.RectFromXY2(0, 0, 0, -1))
	v.SetPressUpDownMovedHandler(v.handleDrag)

	// v.gutter = gutter
	v.Scroller = zcontainer.StackViewHor("scroller")
	v.ShowScrollBars(true, true)
	v.Scroller.SetJSStyle("display", "flex")
	v.SetScrollHandler(v.handleScroll)
	v.Add(v.Scroller, zgeo.TopLeft|zgeo.Expand).Free = true

	v.Overlay = zcontainer.StackViewHor("overlay")
	v.Overlay.SetJSStyle("position", "sticky")
	// v.Overlay.SetStroke(2, zgeo.ColorBlack, true)
	v.Add(v.Overlay, zgeo.BottomRight|zgeo.Expand, zgeo.SizeD(zscrollview.DefaultBarSize, 0)).Free = true

	v.oldScrollX = zfloat.Undefined
	v.queuedGetViews = map[int]bool{}
	v.maxIndex = math.MaxInt
	v.scrollToIndexAfterAllUpdates = zfloat.Undefined
	v.JSSet("className", "zdarkscroll")
	ztimer.RepeatForever(0.05, func() {
		v.createNextViewInQue()
	})
}

func (v *HorBlocksView) SetContentHeight(h float64) {
	v.fixedHeight = h
	v.viewSize.H = h
}

func (v *HorBlocksView) SetMaxIndex(max int) {
	v.maxIndex = max
}

func NewHorBlocksView(cacheDelta, indexWindow int) *HorBlocksView {
	v := &HorBlocksView{}
	v.Init(cacheDelta, indexWindow)
	return v
}

func (v *HorBlocksView) ReadyToShow(beforeWindow bool) {
	if !beforeWindow {
		v.presented = true
	}
}

func (v *HorBlocksView) CurrentIndex() float64 {
	x := v.ScrollOffsetFromCurrent()
	return float64(v.currentIndex) + x/v.viewSize.W
}

// func (v *HorBlocksView) ArrangeChildren() {
// 	zlog.Info("HB Arrange:", v.Rect())
// 	v.StackView.ArrangeChildren()
// }

func (v *HorBlocksView) SetCurrentIndex(fi float64) {
	i := int(fi)
	if zint.Abs(i-v.currentIndex) > v.IndexWindow {
		v.scrollToIndexAfterAllUpdates = fi
	}
	// zlog.Info("SetCurIndex:", v.currentIndex, "=>", i)
	v.currentIndex = i
	fract := fi - float64(v.currentIndex)
	v.update(fract * v.viewSize.W)
}

func (v *HorBlocksView) handleDrag(pos zgeo.Pos, down zbool.BoolInd) bool {
	switch down {
	case zbool.True:
		v.dragXStart = v.ContentOffset().X + pos.X
	case zbool.Unknown:
		x := v.dragXStart - pos.X
		v.SetXContentOffset(x)
		v.handleScroll(zgeo.PosD(x, 0))
	case zbool.False:
		v.dragXStart = 0
	}
	return true
}

func (v *HorBlocksView) handleScroll(pos zgeo.Pos) {
	if v.IgnoreScroll {
		return
	}
	if v.oldScrollX == zfloat.Undefined {
		v.oldScrollX = pos.X
		return
	}
	delta := int(pos.X - v.oldScrollX)
	v.oldScrollX = pos.X
	if !v.flipping && time.Since(v.flippedAt) > time.Second {
		offset := v.ScrollOffsetFromCurrent()
		inc := int(offset / v.viewSize.W)
		// zlog.Info("handleScroll:", pos.X, v.currentIndex, offset, inc, ":", v.ContentOffset().X, v.indexToOffset(v.currentIndex), zdebug.CallingStackString())
		if inc != 0 && zmath.Sign(delta) == zmath.Sign(inc) && (inc < 0 || v.currentIndex < v.maxIndex) {
			v.flipping = true
			v.flippedAt = time.Now()
			v.currentIndex += inc
			v.update(zfloat.Undefined)
			// zlog.Info("increaseOneBlock to", inc, v.currentIndex)
			v.flipping = false
			return
		}
	}
	if v.PanHandler != nil {
		v.PanHandler.HandlePan(v.CurrentIndex())
	}
}

func (v *HorBlocksView) FindViewForIndex(index int) (zview.View, int) {
	si := strconv.Itoa(index)
	cell, i := v.Scroller.FindCellWithName(si)
	if cell == nil {
		// zlog.Info("NoCell?", index, len(v.Cells))
		return nil, -1
	}
	return cell.View, i
}

func (v *HorBlocksView) queUpGetAndSetView(i int) {
	view, _ := v.FindViewForIndex(i)
	if view != nil {
		return
	}
	v.queuedGetViews[i] = true
}

func (v *HorBlocksView) createNextViewInQue() {
	const bigDiff = math.MaxInt - 1000000
	// zlog.Info("**createNextViewInQue1", len(v.queuedGetViews), v.currentIndex)
	// defer zlog.Info("**createNextViewInQue done")
	v.queLock.Lock()
	defer v.queLock.Unlock()
	if len(v.queuedGetViews) != 0 {
		closest := bigDiff
		for i := range v.queuedGetViews {
			view, _ := v.FindViewForIndex(i)
			if zint.Abs(i-v.currentIndex) > v.IndexWindow || view != nil {
				// zlog.Info("createNextViewInQue:", i, "has view or distance")
				delete(v.queuedGetViews, i)
				continue
			}
			if zint.Abs(i-v.currentIndex) < zint.Abs(closest-v.currentIndex) {
				// zlog.Info("createNextViewInQue:", zint.Abs(i-v.currentIndex), zint.Abs(closest-v.currentIndex))
				closest = i
			}
		}
		// zlog.Info("createNextViewInQue closest:", closest, v.currentIndex)
		if closest == bigDiff {
			return
		}
		// zlog.Info("createNextViewInQue", closest, v.queuedGetViews)
		v.createAndSetView(closest)
		delete(v.queuedGetViews, closest)
	}
	if v.scrollToIndexAfterAllUpdates != zfloat.Undefined && (len(v.Cells) == 1 || len(v.queuedGetViews) == 0) {
		fi := v.scrollToIndexAfterAllUpdates
		v.scrollToIndexAfterAllUpdates = zfloat.Undefined
		i := int(fi)
		newCurX := v.indexToOffset(i)
		newCurXF := v.floatingIndexToOffset(fi)
		v.SetXContentOffset(newCurX)
		zlog.Info("set scrollToIndexAfterAllUpdates", v.scrollToIndexAfterAllUpdates, newCurX, newCurXF)
	}
}

func (v *HorBlocksView) createAndSetView(i int) {
	si := strconv.Itoa(i)
	view := v.GetViewFunc(i)
	view.SetObjectName(si)
	// zlog.Info("createAndSetView:", i, len(v.queuedGetViews))
	s := v.viewSize
	style := view.Native().JSStyle()
	style.Set("position", "relative")
	style.Set("min-width", fmt.Sprintf("%fpx", s.W))
	style.Set("min-height", fmt.Sprintf("%fpx", s.H))
	if view == nil {
		// zlog.Info("getAndSetView view == nil:", i)
		return
	}
	next, _ := v.Scroller.FindViewWithName(strconv.Itoa(i+1), false)
	view.SetRect(zgeo.Rect{Size: s})
	// zlog.Info("getAndSetView", i, next != nil, s, view.Rect().Size)
	v.Scroller.AddBefore(view, next, zgeo.TopLeft).Alignment = zgeo.AlignmentNone // Set alignment to none, since we set it on add only
	if next != nil && v.scrollToIndexAfterAllUpdates == zfloat.Undefined {        //!!
		x := v.ContentOffset().X
		v.SetXContentOffset(x + v.viewSize.W)
		v.dragXStart += v.viewSize.W
	}
	v.setSizes()
}

func (v *HorBlocksView) SetRect(r zgeo.Rect) {
	// zlog.Info("HB.SetRect:", r, v.oldRect)
	if v.oldRect == r {
		return
	}
	// if v.IsPresented() && v.presented {
	zlog.Info("HB SetRect", r)
	v.Scroller.RemoveAllChildren()
	// }
	// v.NativeView.SetRect(r) // sets rect so it at least is set
	v.viewSize.W = r.Size.W - zscrollview.DefaultBarSize // 2*v.gutter.Size.W -
	if v.fixedHeight == 0 {
		v.viewSize.H = r.Size.H
	}

	// v.viewSize.H -= zscrollview.DefaultBarSize
	// if v.oldRect.Size.W != 0 {
	v.StackView.SetRect(r) // sets rect as stack, so all parts set
	// v.Overlay.SetRect(v.Scroller.Rect())
	v.update(0)
	// }
	v.oldRect = r
	// zlog.Info("HB.SetRect:", r, v.Scroller.Rect)
}

func (v *HorBlocksView) Reset(update bool) {
	// zlog.Info("HB ReSet:", update, zdebug.CallingStackString())
	// zlog.Info("HB Reset")
	v.Scroller.RemoveAllChildren()
	v.oldRect = zgeo.Rect{}
	if update {
		v.update(0)
	}
}

func (v *HorBlocksView) indexToX(i int) float64 {
	return float64(i) * v.viewSize.W
}

func (v *HorBlocksView) ScrollOffsetFromCurrent() float64 {
	return v.ContentOffset().X - v.indexToOffset(v.currentIndex)
}

func (v *HorBlocksView) indexToOffset(index int) float64 {
	return v.floatingIndexToOffset(float64(index))
}

func (v *HorBlocksView) floatingIndexToOffset(index float64) float64 {
	ii := int(index)
	si := strconv.Itoa(ii)
	fract := index - float64(ii)
	for i, c := range v.Scroller.Cells {
		if si == c.View.ObjectName() {
			// zlog.Info("floatingIndexToOffset", i, fract, zdebug.CallingStackString())
			return (float64(i) + fract) * v.viewSize.W
		}
	}
	return -1
}

func (v *HorBlocksView) getBlocksWidth() float64 {
	return v.viewSize.W * float64(len(v.Scroller.Cells))
}

func (v *HorBlocksView) setSizes() {
	w := v.getBlocksWidth()
	s := zgeo.SizeD(w, v.viewSize.H)
	v.Scroller.SetMinSize(s)
	v.Scroller.SetSize(s)
	// zlog.Info("setSizes:", s)
	s = zgeo.SizeD(10, v.viewSize.H)
	v.Overlay.SetMinSize(s)
	v.Overlay.SetHeight(s.H)
}

func (v *HorBlocksView) update(alterContentOffsetX float64) {
	if v.updating {
		return
	}
	// s := time.Now()
	v.updating = true
	// zlog.Info("update:", v.currentIndex, v.maxIndex, alterContentOffsetX)
	if alterContentOffsetX == zfloat.Undefined {
		alterContentOffsetX = v.ScrollOffsetFromCurrent()
	}
	var lastAddIndex = v.currentIndex - v.IndexWindow - 1
	iMax := min(v.maxIndex, v.currentIndex+v.IndexWindow)
	for i := 0; i < len(v.Scroller.Cells); i++ {
		c := v.Scroller.Cells[i]
		view := c.View
		index, _ := strconv.Atoi(view.ObjectName())
		if !c.Free && (index < v.currentIndex-v.CacheDelta || index > v.currentIndex+v.CacheDelta) {
			zlog.Info("Remove:", index)
			v.Scroller.RemoveChild(c.View, true)
			i--
			continue
		}
		for ji := v.currentIndex - v.IndexWindow; ji <= iMax; ji++ {
			if ji > lastAddIndex && (ji < index || ji > index) {
				// zlog.Info("Add:", ji)
				v.queUpGetAndSetView(ji)
				lastAddIndex = ji
			}
		}
	}
	v.setSizes()
	// now we add any not added in loop above:
	for ji := v.currentIndex - v.IndexWindow; ji <= iMax; ji++ {
		// zlog.Info("Add2:", ji)
		v.queUpGetAndSetView(ji)
	}

	newCurX := v.indexToOffset(v.currentIndex) + alterContentOffsetX
	// zlog.Info("Update2:", v.currentIndex, alterContentOffsetX, "->", newCurX, v.ScrollOffsetFromCurrent(), v.scrollToIndexAfterAllUpdates, len(v.queuedGetViews))
	if v.scrollToIndexAfterAllUpdates == zfloat.Undefined {
		v.SetXContentOffset(newCurX)
	}
	if !v.hasUpdated {
		v.hasUpdated = true
	}
	v.updating = false
	v.createNextViewInQue()
}
