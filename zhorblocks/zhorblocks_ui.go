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
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/ztimer"
)

type HorBlocksView struct {
	zcontainer.StackView
	CacheDelta            int // CacheDelta is how many more than center view to cache on each side
	IndexWindow           int // IndexWindow is how many more than center view to get
	GetViewFunc           func(blockIndex int) zview.View
	RemovedViewFunc       func(blockIndex int)
	CreateHeaderBlockView func(blockIndex int, w float64) zview.View
	PanHandler            PanHandler
	IgnoreScroll          bool
	VertOverlay           *zcontainer.StackView
	Scroller              *zcontainer.StackView
	HorScrollHeaderHeight float64
	CenterSpin            *zwidgets.ActivityView

	VertStack    *zcontainer.StackView
	currentIndex int
	maxIndex     int // this can change over time
	Updating     bool
	viewSize     zgeo.Size
	fixedHeight  float64
	flippedAt    time.Time
	dragXStart   float64
	flipping     bool
	presented    bool
	oldScrollX   float64
	// hasUpdated                   bool
	oldRect                      zgeo.Rect
	queuedGetViews               map[int]bool
	queLock                      sync.Mutex
	scrollToIndexAfterAllUpdates float64
	horScrollHeader              *zcontainer.StackView
	horHeader                    *zcontainer.StackView
	subViewSlaveScrollerRepeater *ztimer.Repeater
}

type PanHandler interface {
	HandlePan(index float64)
}

const spinnerName = "zhorblocks.CenterSpin"

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
	v.horHeader.Add(v.horScrollHeader, zgeo.AlignmentNone)

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
	v.VertStack.Add(v.Scroller, zgeo.AlignmentNone) //.Free = true

	v.VertOverlay = zcontainer.StackViewHor("vert-overlay")
	// v.VertOverlay.SetJSStyle("position", "sticky")
	// v.VertOverlay.SetDimUsable(false)
	//	v.VertOverlay.SetUsable(false)
	v.VertOverlay.SetInteractive(false)
	v.VertOverlay.SetZIndex(500)
	//	v.VertStack.Add(v.VertOverlay, zgeo.BottomRight|zgeo.Expand, zgeo.SizeD(zscrollview.DefaultBarSize, 0)).Free = true
	v.Add(v.VertOverlay, zgeo.TopLeft|zgeo.Expand).Free = true // huge hack whereby adding it to v, and not updating it's hight works. Changing hight disabled scrolling!!!

	v.oldScrollX = zfloat.Undefined
	v.queuedGetViews = map[int]bool{}
	v.maxIndex = math.MaxInt
	v.scrollToIndexAfterAllUpdates = zfloat.Undefined
	v.subViewSlaveScrollerRepeater = ztimer.RepeatForever(0.02, v.scrollSubViews)
	v.AddOnRemoveFunc(v.subViewSlaveScrollerRepeater.Stop)
}

func (v *HorBlocksView) SetContentHeight(h float64) {
	// zlog.Info("SetContentHeight", h)
	v.fixedHeight = h
	v.viewSize.H = h
	v.setSizes()
}

func (v *HorBlocksView) SetMaxIndex(max int) {
	// zlog.Info("v.horInfinite.SetMaxIndex", max)
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

func (v *HorBlocksView) CurrentFloatingIndex() float64 {
	x := v.ScrollOffsetInBlock()
	return float64(v.currentIndex) + x/v.viewSize.W
}

func (v *HorBlocksView) CurrentIndex() int {
	return v.currentIndex
}

func (v *HorBlocksView) SetFloatingCurrentIndex(fi float64) {
	ni := int(fi)
	sci := strconv.Itoa(ni)
	v.currentIndex = int(fi)
	_, fract := math.Modf(fi)
	// zlog.Info("SetFloatingCurrentIndex1", fi, "->", fract)
	v.SetMaxIndex(ni + 1)
	v.update() // fract * v.viewSize.W)
	if fract != 0 {
		_, i := v.Scroller.FindViewWithName(sci, true)
		var o float64
		if i == -1 {
			i = 0
		}
		o = (float64(i) + fract) * v.viewSize.W
		// zlog.Info("SetFloatingCurrentIndex", ni, fi, v.VertStack.ContentOffset().X, "->", oi, o)
		v.VertStack.SetXContentOffset(o)
	}
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

func (v *HorBlocksView) scrollSubViews() {
	pos := v.VertStack.ContentOffset()
	v.horHeader.SetXContentOffset(pos.X)
	v.VertOverlay.SetYContentOffset(pos.Y)
	// zlog.Info("ScrollOffset:", pos, v.VertOverlay.ContentOffset().Y)
}

func (v *HorBlocksView) handleScroll(pos zgeo.Pos) {
	// zlog.Info("Scroll:", pos.X)
	if v.IgnoreScroll {
		return
	}
	v.update()
	if v.oldScrollX == zfloat.Undefined {
		v.oldScrollX = pos.X
		return
	}
	if v.PanHandler != nil {
		v.PanHandler.HandlePan(v.CurrentFloatingIndex())
	}
}

func (v *HorBlocksView) FindViewForBlockIndex(index int) (zview.View, int) {
	si := strconv.Itoa(index)
	cell, i := v.Scroller.FindCellWithName(si)
	if cell == nil {
		// zlog.Info("NoCell?", index, len(v.Cells))
		return nil, -1
	}
	return cell.View, i
}

func (v *HorBlocksView) createAndSetView(i int) {
	si := strconv.Itoa(i)
	view := v.GetViewFunc(i)
	view.SetObjectName(si)

	aa, _ := view.(zcontainer.AdvancedAdder)
	if aa != nil {
		centerSpin := zwidgets.NewActivityView(zgeo.SizeBoth(50), zgeo.ColorWhite)
		centerSpin.SetObjectName(spinnerName)
		centerSpin.SetZIndex(22000)
		aa.AddAdvanced(centerSpin, zgeo.TopCenter, zgeo.RectFromMarginSize(zgeo.SizeD(0, 40)), zgeo.SizeNull, -1, true)
	}
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
	// if next != nil && v.scrollToIndexAfterAllUpdates == zfloat.Undefined {        //!!
	// 	// x := v.VertStack.ContentOffset().X
	// 	// v.VertStack.SetXContentOffset(x + v.viewSize.W)
	// 	v.dragXStart += v.viewSize.W
	// }
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
	// zlog.Info("HB.SetRect1:", r, v.viewSize.W)

	v.StackView.SetRect(r) // sets rect as stack, so all parts set
	// zlog.Info("HB SetRect", v.VertOverlay.Rect().Size.H)
	if v.viewSize.W != 0 {
		v.update()
	}
	// v.oldRect = r
}

func (v *HorBlocksView) Reset(update bool) {
	// zlog.Info("HB ReSet:", update, zdebug.CallingStackString())
	// v.currentIndex = 0
	// v.VertStack.SetXContentOffset(0)
	// zlog.Info("HB Reset", zdebug.CallingStackString())
	v.Scroller.RemoveAllChildren()
	v.horScrollHeader.RemoveAllChildren()
	v.oldRect = zgeo.Rect{}
	if update {
		v.update()
	}
}

// func (v *HorBlocksView) indexToX(i int) float64 {
// 	return float64(i) * v.viewSize.W
// }

func (v *HorBlocksView) ScrollOffsetInBlock() float64 {
	return v.VertStack.ContentOffset().X - v.indexToOffset(v.currentIndex)
}

func (v *HorBlocksView) indexToOffset(index int) float64 {
	return v.floatingIndexToOffset(float64(index))
}

func (v *HorBlocksView) floatingIndexToOffset(index float64) float64 {
	si := strconv.Itoa(v.currentIndex)
	fract := index - float64(v.currentIndex)
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
	// s.H -= zscrollview.DefaultBarSize
	v.Scroller.SetMinSize(s)
	v.Scroller.SetSize(s)

	v.horScrollHeader.SetWidth(w)
	// v.horScrollHeader.SetTop(v.AbsoluteRect().Pos.Y)

	s.W = 10
	v.VertOverlay.SetMinSize(s)
	// v.VertOverlay.SetHeight(s.H)
	// zlog.Info("setSizes:", s.H, zdebug.CallingStackString())
	// zlog.Info("setSizes2", v.VertOverlay.Rect().Size.H)
}

func (v *HorBlocksView) DebugPrintCells() string {
	str := "[ "
	for _, c := range v.Scroller.Cells {
		str += c.View.ObjectName() + " "
	}
	return str + " ]"
}

// func intCurrentIndex(ci float64) int {
// 	return int(math.Floor(ci))
// }

func (v *HorBlocksView) update() {
	if v.Updating {
		return
	}
	// s := time.Now()
	v.Updating = true
	// zlog.Info("** update1: cur:", v.currentIndex)
fullLoop:
	for {
		var adjustedOffset bool
		for i := 0; i < len(v.Scroller.Cells); i++ {
			c := v.Scroller.Cells[i]
			view := c.View
			index, _ := strconv.Atoi(view.ObjectName())
			if !c.Free && zint.Abs(index-v.currentIndex) > v.IndexWindow {
				// zlog.Info("Remove:", index, v.currentIndex, v.currentIndex)
				v.Scroller.RemoveChild(c.View, true)
				v.horScrollHeader.RemoveNamedChild(c.View.ObjectName(), false, true)
				v.setSizes()
				if v.RemovedViewFunc != nil {
					v.RemovedViewFunc(index)
				}
				continue fullLoop
			}
		}
		if time.Since(v.flippedAt) > time.Second {
			sci := strconv.Itoa(v.currentIndex)
			_, i := v.Scroller.FindViewWithName(sci, true)
			if i != -1 {
				w := float64(i) * v.viewSize.W
				o := v.VertStack.ContentOffset().X
				blocksDiff := int((w - o) / v.viewSize.W)
				// zlog.Info("update1?:", "cur:", v.currentIndex, sci, "bdiff:", blocksDiff, v.DebugPrintCells(), "max:", v.maxIndex, o, o/v.viewSize.W)
				if blocksDiff != 0 {
					// zlog.Info("update:", "cur:", v.currentIndex, sci, "bdiff:", blocksDiff, v.DebugPrintCells(), "max:", v.maxIndex, o, o/v.viewSize.W, "i:", i)
					if blocksDiff == -1 && i < len(v.Scroller.Cells)-1 || blocksDiff == 1 && i > 0 {
						newOffset := o + float64(blocksDiff)*v.viewSize.W
						// flipped = true
						v.flippedAt = time.Now()
						v.currentIndex = v.currentIndex - blocksDiff
						// zlog.Info("ChangeCur:", v.currentIndex, newOffset)
						v.VertStack.SetXContentOffset(newOffset)
						adjustedOffset = true
						continue fullLoop
					}
					// v.currentIndex, _ = strconv.Atoi(v.Scroller.Cells[ni].View.ObjectName())
					if v.currentIndex < v.maxIndex || blocksDiff < 0 {
						_, fract := math.Modf(o / v.viewSize.W)
						newOffset := (float64(i) + fract) * v.viewSize.W
						// zlog.Info("moreInc:", v.maxIndex, w, o, blocksDiff, i, v.currentIndex, newOffset)
						v.VertStack.SetXContentOffset(newOffset)
						adjustedOffset = true
					}
				}
			}
		}
		iMax := min(v.maxIndex, v.currentIndex+v.IndexWindow)
		// zlog.Info("Add?: start:", v.currentIndex-v.IndexWindow, "max", iMax, "ci", v.currentIndex)
		for ji := v.currentIndex - v.IndexWindow; ji <= iMax; ji++ {
			view, _ := v.FindViewForBlockIndex(ji)
			if view != nil {
				continue
			}
			// zlog.Info("Create:", ji, v.currentIndex)
			v.createAndSetView(ji)
			v.setSizes()
			continue fullLoop
		}
		if adjustedOffset {
			continue
		}
		break
	}
	v.Updating = false
	// for i, c := range v.Scroller.Cells {
	// 	if c.View.ObjectName() != "0" {
	// 		continue
	// 	}
	// 	cc := c.View.(zcontainer.CellsCounter)
	// 	if cc != nil {
	// 		view, _ := v.FindViewForBlockIndex(0)
	// 		zlog.Info(i, "Updated cur:", v.currentIndex, zlog.Pointer(c.View), cc.CountChildren(), zlog.Pointer(view))
	// 	}
	// }
}

func (v *HorBlocksView) IsBlockInWindow(blockIndex int) bool {
	return blockIndex >= v.currentIndex-v.IndexWindow && blockIndex <= v.currentIndex+v.IndexWindow
}

func (v *HorBlocksView) BlockViews() []zview.View {
	return v.Scroller.GetChildren(false)
}

func (v *HorBlocksView) SpinActivity(blockIndex int, spin bool) {
	// zlog.Info("SPIN", blockIndex, spin)
	sci := strconv.Itoa(blockIndex)
	view, _ := v.Scroller.FindViewWithName(sci, true)
	bv, _ := view.(*zcontainer.ContainerView)
	if bv != nil {
		view, _ := bv.FindViewWithName(spinnerName, true)
		av := view.(*zwidgets.ActivityView)
		av.StopOrStart(spin)
	}
}
