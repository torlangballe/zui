//go:build zui && js

package zhorblocks

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
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
	CacheDelta            int // CacheDelta is how many more than center view to cache on each side
	IndexWindow           int // IndexWindow is how many more than center view to get
	GetViewFunc           func(blockIndex int) zview.View
	CreateHeaderBlockView func(blockIndex int, w float64) zview.View
	PanHandler            PanHandler
	IgnoreScroll          bool
	VertOverlay           *zcontainer.StackView
	Scroller              *zcontainer.StackView
	HorScrollHeaderHeight float64

	VertStack                    *zcontainer.StackView
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
	horScrollHeader              *zcontainer.StackView
	horHeader                    *zcontainer.StackView
}

type PanHandler interface {
	HandlePan(index float64)
}

func (v *HorBlocksView) Init(indexWindow, cacheDelta int) {
	v.CacheDelta = cacheDelta
	v.IndexWindow = indexWindow
	v.StackView.Init(v, true, "hor-inf")
	// v.SetMargin(zgeo.RectFromXY2(0, 0, 0, -1))
	v.SetPressUpDownMovedHandler(v.handleDrag)
	v.SetSpacing(0)

	// v.gutter = gutter

	v.HorScrollHeaderHeight = 20

	v.horHeader = zcontainer.StackViewHor("hor-header")
	v.horHeader.ShowScrollBars(true, false)
	v.horHeader.SetMinSize(zgeo.SizeD(10, v.HorScrollHeaderHeight))
	v.horHeader.JSSet("className", "znoscrollbar")
	// v.horHeader.SetScrollHandler(func(pos zgeo.Pos) { // this makes things sluggish, must be some kind of feedback loop
	// 	v.VertStack.SetXContentOffset(pos.X)
	// })
	v.Add(v.horHeader, zgeo.TopLeft|zgeo.HorExpand)

	v.horScrollHeader = zcontainer.StackViewHor("hor-scroll-header")
	v.horScrollHeader.SetJSStyle("display", "flex")
	v.horScrollHeader.SetMinSize(zgeo.SizeD(10, v.HorScrollHeaderHeight))
	v.horHeader.Add(v.horScrollHeader, zgeo.TopLeft|zgeo.HorExpand)

	v.VertStack = zcontainer.StackViewVert("vstack")
	v.VertStack.SetSpacing(0)
	v.VertStack.SetMinSize(zgeo.SizeBoth(40))
	v.VertStack.ShowScrollBars(true, true)
	// v.VertStack.SetBGColor(zgeo.ColorRed)
	v.VertStack.SetScrollHandler(v.handleScroll)
	v.VertStack.JSSet("className", "zdarkscroll")
	v.Add(v.VertStack, zgeo.TopLeft|zgeo.Expand)

	v.Scroller = zcontainer.StackViewHor("scroller")
	v.Scroller.SetJSStyle("display", "flex")
	v.VertStack.Add(v.Scroller, zgeo.TopLeft|zgeo.Expand).Free = true

	v.VertOverlay = zcontainer.StackViewHor("vert-overlay")
	v.VertOverlay.SetJSStyle("position", "sticky")
	v.VertStack.Add(v.VertOverlay, zgeo.BottomRight|zgeo.Expand, zgeo.SizeD(zscrollview.DefaultBarSize, 0)).Free = true

	v.oldScrollX = zfloat.Undefined
	v.queuedGetViews = map[int]bool{}
	v.maxIndex = math.MaxInt
	v.scrollToIndexAfterAllUpdates = zfloat.Undefined
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
		v.dragXStart = v.VertStack.ContentOffset().X + pos.X
	case zbool.Unknown:
		x := v.dragXStart - pos.X
		v.VertStack.SetXContentOffset(x)
		v.handleScroll(zgeo.PosD(x, 0))
	case zbool.False:
		v.dragXStart = 0
	}
	return true
}

func (v *HorBlocksView) handleScroll(pos zgeo.Pos) {
	v.horHeader.SetXContentOffset(pos.X)
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
		v.VertStack.SetXContentOffset(newCurX)
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
	si1 := strconv.Itoa(i + 1)
	next, _ := v.Scroller.FindViewWithName(si1, false)
	view.SetRect(zgeo.Rect{Size: s})
	// zlog.Info("getAndSetView", i, next != nil, s, view.Rect().Size)
	v.Scroller.AddBefore(view, next, zgeo.TopLeft).Alignment = zgeo.AlignmentNone // Set alignment to none, since we set it on add only
	if next != nil && v.scrollToIndexAfterAllUpdates == zfloat.Undefined {        //!!
		x := v.VertStack.ContentOffset().X
		v.VertStack.SetXContentOffset(x + v.viewSize.W)
		v.dragXStart += v.viewSize.W
	}
	if v.CreateHeaderBlockView != nil {
		next, _ := v.horScrollHeader.FindViewWithName(si1, false)
		over := v.CreateHeaderBlockView(i, v.viewSize.W)
		nv := over.Native()
		nv.SetJSStyle("position", "relative")
		nv.SetObjectName(si)
		nv.SetZIndex(92999)
		nv.SetTop(0)
		nv.SetSize(zgeo.SizeD(v.viewSize.W, v.HorScrollHeaderHeight))
		v.horScrollHeader.AddBefore(nv, next, zgeo.AlignmentNone)
		cv, _ := over.(*zcustom.CustomView)
		if cv != nil {
			cv.ForceDrawSelf()
		}
	}
	v.setSizes()
}

func (v *HorBlocksView) SetRect(r zgeo.Rect) {
	// if v.oldRect == r {
	// 	return
	// }
	// // if v.IsPresented() && v.presented {
	v.Scroller.RemoveAllChildren()
	v.horScrollHeader.RemoveAllChildren()
	// old := v.viewSize.W
	v.viewSize.W = r.Size.W - zscrollview.DefaultBarSize // 2*v.gutter.Size.W -
	if v.fixedHeight == 0 {
		v.viewSize.H = r.Size.H
	}
	// zlog.Info("HB.SetRect1:", r, old, v.viewSize.W)

	v.StackView.SetRect(r) // sets rect as stack, so all parts set
	// zlog.Info("HB SetRect", r, v.VertStack.Rect())
	v.update(0)
	// v.oldRect = r
}

func (v *HorBlocksView) Reset(update bool) {
	// zlog.Info("HB ReSet:", update, zdebug.CallingStackString())
	// zlog.Info("HB Reset")
	v.Scroller.RemoveAllChildren()
	v.horScrollHeader.RemoveAllChildren()
	v.oldRect = zgeo.Rect{}
	if update {
		v.update(0)
	}
}

func (v *HorBlocksView) indexToX(i int) float64 {
	return float64(i) * v.viewSize.W
}

func (v *HorBlocksView) ScrollOffsetFromCurrent() float64 {
	return v.VertStack.ContentOffset().X - v.indexToOffset(v.currentIndex)
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

	s = zgeo.SizeD(5000, v.HorScrollHeaderHeight)
	// v.horScrollHeader.SetMinSize(s)
	v.horScrollHeader.SetWidth(w)
	// v.horScrollHeader.SetTop(v.AbsoluteRect().Pos.Y)
	// zlog.Info("setSizes:", s)

	s = zgeo.SizeD(10, v.viewSize.H)
	v.VertOverlay.SetMinSize(s)
	v.VertOverlay.SetHeight(s.H)
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
			// zlog.Info("Remove:", index)
			v.Scroller.RemoveChild(c.View, true)
			v.horScrollHeader.RemoveNamedChild(c.View.ObjectName(), false, true)
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
		v.VertStack.SetXContentOffset(newCurX)
	}
	if !v.hasUpdated {
		v.hasUpdated = true
	}
	v.updating = false
	v.createNextViewInQue()
}
