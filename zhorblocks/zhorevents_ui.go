//go:build zui && js

package zhorblocks

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zgridlist"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zshape"
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdraw"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmath"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
)

type HorEventsView struct {
	zcontainer.StackView
	BlockDuration             time.Duration
	Bar                       *zcontainer.StackView
	LastEventTimeForBlock     map[int]time.Time
	LockedTime                time.Time
	TestMode                  bool
	Updater                   Updater
	MixOddBlocksColor         zgeo.Color // If Valid, is mixed a tiny bit with odd background block color to show borders
	GutterWidth               zmath.RangeF64
	OddLaneColor              zgeo.Color
	GetEventViewsFunc         func(blockIndex int, isNewBlockView bool, got func(childView zview.View, x int, cellBox zgeo.Size, laneID, rowType int64, blockDone bool) bool)
	MakeRowBackgroundViewFunc func(blockIndex int, laneID int64, row *Row, size zgeo.Size) *zcontainer.ContainerView
	MakeLaneActionIconFunc    func(laneID int64) *zimageview.ImageView
	MakeLaneRowActionIconFunc func(laneID int64, rowID int64) *zimageview.ImageView
	LockChildViewFunc         func(child zview.View)

	nowLine               *zcustom.CustomView
	lanes                 []Lane
	horInfinite           *HorBlocksView
	ViewWidth             float64
	startTime             time.Time
	currentTime           time.Time
	zoomStack             *zcontainer.StackView
	timeField             *ztext.TimeFieldView
	rightPole             *zcontainer.StackView
	leftPole              *zcontainer.StackView
	nowButton             *zshape.ImageButtonView
	GotoLockedButton      *zshape.ShapeView
	zoomLevels            []zoomLevel
	panDurations          []panDuration
	panDuration           time.Duration
	zoomIndex             int
	storeKey              string
	currentNowBlockIndex  int
	updateBlocks          map[int]time.Time
	scrollToNow           bool
	lastScrollToX         int
	updateNowRepeater     *ztimer.Repeater
	updatingBlock         bool
	timeAxisHeight        float64
	lastBlockUpdateTime   time.Time
	markerPole            *zcustom.CustomView
	moveStartX            float64
	markerTimes           []time.Time
	keepScrollingRepeater *ztimer.Repeater
	keepScrollingDir      float64
	startupGotoTime       time.Time
}

type Updater interface {
	Update()
}

type Lane struct {
	ID           int64
	Name         string
	Rows         []Row
	TextColor    zgeo.Color
	overlayViews []zview.View
	y            float64
	hasAxis      bool
}

type Row struct {
	ID            int64
	Name          string
	Height        float64
	y             float64 // accumulated from top
	ForGlobalLane bool
	overlayViews  []zview.View
}

type EventOptions struct {
	StoreKey             string
	BlocksIndexGetWidth  int
	BlockIndexCacheDelta int
	BlockDuration        time.Duration
	StartTime            time.Time
	ShowNowPole          bool
	GutterWidth          zmath.RangeF64
	TimeAxisHeight       float64
	BGColor              zgeo.Color
}

const (
	LockedItemStrokeWidth = 4.0

	zoomIndexKey              = ".horblock.Events.zoom"
	dividerHeight             = 2
	widthRatioToNowLine       = 0.8
	markerButtonTimesKey      = "MarkerButtonTimesKey"
	OverlayBackgroundViewName = "OverlayBackgroundViewName"
)

var BestRowIDForCentering int64

func NewEventsView(v *HorEventsView, opts EventOptions) *HorEventsView {
	if v == nil {
		v = &HorEventsView{}
	}
	v.StackView.Init(v, true, "hor-events")
	v.SetBGColor(opts.BGColor)
	v.SetCanTabFocus(true)
	v.Updater = v
	// v.TestMode = true
	v.GutterWidth = opts.GutterWidth
	v.timeAxisHeight = opts.TimeAxisHeight
	v.storeKey = opts.StoreKey
	v.zoomLevels = []zoomLevel{
		zoomLevel{"24 Hours", time.Hour * 24},
		zoomLevel{"1 Hour", time.Hour * 1},
		zoomLevel{"10 Min", time.Minute * 10},
		zoomLevel{"1 Min", time.Minute * 1},
		zoomLevel{"10 Secs", time.Second * 10},
	}
	v.panDurations = []panDuration{
		panDuration{"1d", zkeyboard.ModifierNone, time.Hour * 24, nil},
		panDuration{"1h", zkeyboard.ModifierCommand, time.Hour * 1, nil},
		panDuration{"10m", zkeyboard.ModifierAlt, time.Minute * 10, nil},
		panDuration{"1m", zkeyboard.ModifierShift, time.Minute, nil},
	}
	blockDuration := opts.BlockDuration
	v.zoomIndex = -1
	v.currentNowBlockIndex = zint.Undefined
	if blockDuration == 0 && zkeyvalue.DefaultStore != nil && opts.StoreKey != "" {
		n, got := zkeyvalue.DefaultStore.GetInt(opts.StoreKey+zoomIndexKey, 0)
		if got && n >= 0 && n < len(v.zoomLevels) {
			v.zoomIndex = n
			blockDuration = v.zoomLevels[n].duration
		}
	}
	if v.zoomIndex == -1 {
		for i, z := range v.zoomLevels {
			if z.duration == blockDuration {
				v.zoomIndex = i
				break
			}
		}
	}
	if v.zoomIndex == -1 {
		v.zoomIndex = 1
		blockDuration = v.zoomLevels[v.zoomIndex].duration
	}
	v.LastEventTimeForBlock = map[int]time.Time{}
	v.updateBlocks = map[int]time.Time{}
	v.BlockDuration = blockDuration // must be before calculating startTime
	start := opts.StartTime
	t := v.calcTimePosToShowTime(start)
	v.setStartTime(t)
	v.SetSpacing(0)
	v.Bar = zcontainer.StackViewHor("bar")
	v.Bar.SetBGColor(zgeo.ColorNewGray(0.4, 1))
	v.Bar.SetMarginS(zgeo.SizeD(8, 3))
	v.Bar.SetSpacing(8)
	v.Add(v.Bar, zgeo.TopLeft|zgeo.HorExpand)
	v.makeButtons()
	v.makeMarkerPole()
	v.horInfinite = NewHorBlocksView(opts.BlocksIndexGetWidth, opts.BlockIndexCacheDelta)
	v.horInfinite.GetViewFunc = v.makeBlockView
	v.horInfinite.RemovedViewFunc = v.handleBlockViewRemoved
	v.horInfinite.PanHandler = v
	v.horInfinite.SetBGColor(opts.BGColor)
	v.Add(v.horInfinite, zgeo.TopLeft|zgeo.Expand)
	v.horInfinite.IgnoreScroll = true
	v.MakeRowBackgroundViewFunc = defaultMakeBackgroundViewFunc

	v.leftPole = v.makeSidePole(zgeo.Left)
	v.rightPole = v.makeSidePole(zgeo.Right)

	v.horInfinite.CreateHeaderBlockViewFunc = func(blockIndex int, w float64) zview.View {
		box := v.makeAxisRow(blockIndex)
		return box
	}
	if opts.ShowNowPole {
		line := zcustom.NewView("now-pole")
		line.SetZIndex(11000)
		line.SetMinSize(zgeo.SizeD(20, 100))
		line.SetInteractive(false)
		line.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
			// zlog.Info("draw now pole:", rect)
			colors := []zgeo.Color{zgeo.ColorNew(0, 1, 0, 0), zgeo.ColorNew(0, 1, 0, 0.8)}
			path := zgeo.PathNewRect(rect, zgeo.SizeNull)
			canvas.DrawGradient(path, colors, rect.Min(), rect.TopRight(), nil)
		})
		// line.SetBGColor(zgeo.ColorNew(0, 1, 0, 0.5))
		v.nowLine = line
		v.Add(line, zgeo.AlignmentNone)
	}
	v.loadMarkerButtons()
	v.updateNowRepeater = ztimer.RepeaterNew()
	v.AddOnRemoveFunc(v.updateNowRepeater.Stop)
	v.horInfinite.SetMaxIndex(1)
	v.UpdateWidgets()
	lineRepeater := ztimer.RepeatForever(0.1, v.updateNowScrollAndPole)
	v.AddOnRemoveFunc(lineRepeater.Stop)
	updateRepeater := ztimer.RepeatForever(0.1, func() {
		v.updateCurrentBlockViews()
	})
	v.AddOnRemoveFunc(updateRepeater.Stop)
	return v
}

func (v *HorEventsView) setStartTime(t time.Time) {
	var start time.Time
	switch v.BlockDuration {
	case time.Hour * 24:
		start = ztime.OnThisDay(t, 0)
	case time.Hour:
		start = ztime.ChangedPartsOfTime(t, -1, 0, 0, 0)
	case time.Minute * 10:
		m := zmath.RoundToMod(t.Minute(), 10)
		start = ztime.ChangedPartsOfTime(t, -1, m, 0, 0)
	case time.Minute:
		start = ztime.ChangedPartsOfTime(t, -1, -1, 0, 0)
	case time.Second * 10:
		s := zmath.RoundToMod(t.Second(), 10)
		start = ztime.ChangedPartsOfTime(t, -1, -1, s, 0)
	default:
		zlog.Fatal("Bad Duration", v.BlockDuration)
	}
	v.startTime = start
	v.currentTime = start
	v.startupGotoTime = t
}

func (v *HorEventsView) loadMarkerButtons() {
	if zkeyvalue.DefaultStore != nil && v.storeKey != "" {
		var times []time.Time
		zkeyvalue.DefaultStore.GetObject(v.storeKey+markerButtonTimesKey, &times)
		for _, t := range times {
			v.makeMarkerButton(t)
		}
	}
}

func (v *HorEventsView) saveMarkerButtons() {
	if zkeyvalue.DefaultStore != nil && v.storeKey != "" {
		zkeyvalue.DefaultStore.SetObject(v.markerTimes, v.storeKey+markerButtonTimesKey, true)
	}
}

func (v *HorEventsView) makeMarkerButton(t time.Time) {
	button := zshape.NewView(zshape.TypeRoundRect, zgeo.SizeD(60, 24))
	stime := ztime.GetNice(t, false)
	button.SetText(stime)
	button.SetTextColor(zgeo.ColorBlack)
	button.SetColor(zgeo.ColorGreen)
	button.SetFont(zgeo.FontDefault(-3))
	button.SetMarginS(zgeo.SizeD(4, 0))
	v.markerTimes = append(v.markerTimes, t)
	button.SetTextColor(zgeo.ColorBlack)
	v.Bar.Add(button, zgeo.CenterLeft)
	v.Bar.ArrangeChildren()
	tip := "marker. Press to go to " + ztime.GetNice(t, true) + "\n"
	tip += "shift-press to remove"
	button.SetToolTip(tip)
	button.SetPressedHandler("goto", 0, func() {
		v.SetScrollToNowOn(false)
		v.GotoTime(t.Add(-v.BlockDuration / 2))
		v.showMarkerPoleAt(t)
		ztimer.StartIn(1, func() {
			v.markerPole.Show(false)
		})
	})
	button.SetPressedHandler("clear", zkeyboard.ModifierShift, func() {
		i := slices.Index(v.markerTimes, t)
		zslice.RemoveAt(&v.markerTimes, i)
		v.Bar.RemoveChild(button, true)
		v.Bar.ArrangeChildren()
		v.saveMarkerButtons()
	})
}

func (v *HorEventsView) makeMarkerPole() {
	v.markerPole = zcustom.NewView("marker")
	v.markerPole.SetMinSize(zgeo.SizeD(60, 40))
	v.markerPole.SetZIndex(12000)
	// v.markerPole.SetBGColor(zgeo.ColorBlue)
	v.markerPole.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, drawView zview.View) {
		x := v.markerPole.Rect().Size.W / 2
		canvas.SetColor(zgeo.ColorGreen)
		canvas.StrokeVertical(x, rect.Min().Y, rect.Max().Y, 1, zgeo.PathLineButt)
		ti := ztextinfo.New()
		ti.Alignment = zgeo.Center
		ti.Font = zgeo.FontDefault(-2)
		ti.Text = strings.Replace(v.markerPole.ObjectName(), " ", "\n", -1)
		ti.Color = zgeo.ColorBlack
		r := rect.Align(zgeo.SizeBoth(26), zgeo.CenterLeft|zgeo.HorExpand, zgeo.SizeNull)
		ti.Rect = r.ExpandedD(-2)
		path := zgeo.PathNewRect(r, zgeo.SizeBoth(4))
		canvas.FillPath(path)
		ti.Draw(canvas)
	})
	v.Add(v.markerPole, zgeo.AlignmentNone)
}

func (v *HorEventsView) makeSidePole(a zgeo.Alignment) *zcontainer.StackView {
	pole := zcontainer.StackViewVert(a.String() + "-pole")
	w := v.GutterWidth.Min
	if a == zgeo.Right {
		w = v.GutterWidth.Max
		w += zwindow.ScrollBarSizeForView(v)
	}
	pole.SetMinSize(zgeo.SizeD(w, 10))
	pole.SetBGColor(v.BGColor().WithOpacity(0.8))
	// pole.SetBGColor(zgeo.ColorBlue)
	pole.SetPressedHandler("pole-click", 0, func() {
		v.handlePolePress(false)
	})
	v.horInfinite.VertOverlay.Add(pole, zgeo.Top|a|zgeo.VertExpand, zgeo.SizeD(0, v.timeAxisHeight)).Free = true
	return pole
}

func (v *HorEventsView) handlePolePress(left bool) {
	// y := zview.LastPressedPos.Y
	// zlog.Info("handlePolePress:", left, y, v.horInfinite.VertStack.ContentOffset().Y)
}

// func rescaleGraph(rt *row, laneID int64, view zview.View) {
// 	zlog.Info("Rescale:", rt.rowType, laneID)
// }

func (v *HorEventsView) updateNowPole() {
	if v.nowLine == nil {
		return
	}
	x, show := v.TimeToXInHorEventView(time.Now())
	v.nowLine.Show(show)
	y := v.horInfinite.Rect().Pos.Y
	v.nowLine.SetZIndex(911000)
	v.nowLine.Expose()
	// zlog.Info("NowPole:", diff, x, now, show, nowBlockIndex, v.horInfinite.currentIndex, "y:", y, v.horInfinite.Rect().Size.H)
	v.nowLine.SetRect(zgeo.RectFromXYWH(x, y, 10, v.horInfinite.Rect().Size.H))
}

func (v *HorEventsView) lockCenterView() zview.View {
	// cx := v.horInfinite.IndexToOffset()
	cx := v.Rect().Center().X
	// ox := v.horInfinite.ScrollOffset()
	min := math.MaxFloat64
	var centerBlock *zcontainer.ContainerView
	for _, view := range v.horInfinite.BlockViews() {
		x := view.Native().AbsoluteRect().Center().X
		diff := math.Abs(x - cx)
		// zlog.Info("Center?:", x, cx, x-cx, view.ObjectName())
		if diff < min {
			min = diff
			centerBlock, _ = view.(*zcontainer.ContainerView)
		}
	}
	if centerBlock == nil {
		return nil
	}
	// zlog.Info("CenterBlock:", min, centerBlock.ObjectName(), cx)
	var centerChild zview.View
	for _, lane := range v.lanes {
		if lane.ID == 0 { // this is a hack, we don't know 0 is system row in timeline in this base class
			continue
		}
		for _, row := range lane.Rows {
			if row.y+row.Height > v.horInfinite.viewSize.H {
				break
			}
			if row.ID != BestRowIDForCentering {
				continue
			}
			name := fmt.Sprintf("%d-%v", lane.ID, row.ID)
			bgRow, _ := centerBlock.FindViewWithName(name, false)
			if bgRow == nil {
				return nil
			}
			// zlog.Info("Center:", centerBlock.ObjectName(), min, bestRow.Name, bestRow.y)
			min = math.MaxFloat64
			zcontainer.ViewRangeChildren(bgRow, false, false, func(child zview.View) bool {
				x := child.Native().AbsoluteRect().Center().X
				diff := math.Abs(x - cx)
				if diff < min {
					min = diff
					centerChild = child
				}
				return true
			})
		}
	}
	return centerChild
}

func (v *HorEventsView) SetBlockDuration(d time.Duration) {
	now := time.Now()
	_, shown := v.TimeToXInHorEventView(now)
	// zlog.Info("SetBlockDuration Pre:", x, shown, t, v.BlockDuration)
	var t, start time.Time
	if shown && v.LockedTime.IsZero() {
		t = now
		v.BlockDuration = d
		start = v.calcTimePosToShowTime(t) //.Add(time.Second * 3)
	} else {
		if v.LockedTime.IsZero() && v.LockChildViewFunc != nil {
			lockChild := v.lockCenterView()
			if lockChild != nil {
				v.LockChildViewFunc(lockChild)
				// zlog.Info("Child2Select:", lockChild.ObjectName(), blockIndex)
				ztimer.StartIn(1, func() {
					v.SetBlockDuration(d)
					ztimer.StartIn(0.8, func() {
						v.GotoLockedButton.Click("", false, zkeyboard.ModifierShift)
					})
				})
				return
			}
		}
		if t.IsZero() {
			t = v.LockedTime
			if t.IsZero() {
				t = v.currentTime.Add(v.BlockDuration / 2)
			}
		}
		t = t.Add(-d / 2)
		ztime.Minimize(&t, now)
		v.BlockDuration = d
		start = t
	}
	v.setStartTime(start)
	v.horInfinite.SetFloatingCurrentIndex(0)
	v.Updater.Update()
}

func (v *HorEventsView) Reset(updateBlocks bool) {
	v.LastEventTimeForBlock = map[int]time.Time{}
	v.updateBlocks = map[int]time.Time{}
	// zlog.Info("HEV.Reset")
	v.horInfinite.SetMaxIndex(1)
	v.horInfinite.Reset(updateBlocks && v.ViewWidth != 0)
}

func (v *HorEventsView) updateCurrentBlockViews() {
	if v.updatingBlock || v.horInfinite.Updating {
		return
	}
	if time.Since(v.lastBlockUpdateTime) < time.Millisecond*100 {
		return
	}
	v.updatingBlock = true
	bestDiff := math.MaxInt
	var bestIndex int
	var bestTime time.Time
	secs := min(5, max(1, ztime.DurSeconds(v.BlockDuration)/30))
	ci := v.horInfinite.CurrentIndex()
	for bi, t := range v.updateBlocks {
		if ztime.Since(t) < secs {
			continue
		}
		startOfBlock := v.IndexToTime(float64(bi))
		if time.Since(startOfBlock) < 0 {
			continue
		}
		diff := zint.Abs(bi - ci)
		if diff < bestDiff {
			bestDiff = diff
			bestIndex = bi
			bestTime = t
		}
	}
	if bestDiff == math.MaxInt {
		v.updatingBlock = false
		return
	}
	v.updateBlockView(bestIndex, bestTime.IsZero())
	ni := v.TimeToBlockIndex(time.Now())
	v.horInfinite.SetMaxIndex(ni + 1)
}

func (v *HorEventsView) updateBlockView(blockIndex int, isNew bool) {
	if v.ViewWidth == 0 || len(v.lanes) == 0 {
		delete(v.updateBlocks, blockIndex) // without this we loop forever
		v.updatingBlock = false
		return
	}
	endTimeOfBlock := v.IndexToTime(float64(blockIndex) + 1)
	if !isNew && time.Since(endTimeOfBlock) > time.Second*10 { // if it's an old block, and last events written to db and gotten, don't update anymore.
		delete(v.updateBlocks, blockIndex)
		v.updatingBlock = false
		return
	}
	// zlog.Info("updateBlock", blockIndex)
	// start := time.Now()
	v.updateBlocks[blockIndex] = time.Now()
	view, _ := v.horInfinite.FindViewForBlockIndex(blockIndex)
	// if blockIndex == 0 {
	// 	zlog.Info("updateBlock", zlog.Pointer(view), blockIndex, len(v.lanes))
	// }
	if view == nil {
		v.updatingBlock = false
		zlog.Info("updateBlockView no view:", blockIndex)
		return // we haven't created it yet, or it's now being updated with repeater, when that block is outside index window
	}
	// zlog.Info("updateBlock", blockIndex)
	blockView := view.(*zcontainer.ContainerView)
	// zlog.Info("updateBlockView", blockIndex, zlog.Pointer(blockView))
	// go below removed
	v.GetEventViewsFunc(blockIndex, isNew, func(childView zview.View, x int, cellBox zgeo.Size, laneID, rowType int64, blockDone bool) bool {
		if blockDone {
			// if isNew {
			// zlog.Info("BlockDone", blockIndex, zlog.Pointer(blockView))
			// }
			v.lastBlockUpdateTime = time.Now()
			v.updateBlocks[blockIndex] = time.Now() // set it to after event views gotten
			v.updatingBlock = false
			blockView.ArrangeChildren()
			return false
		}
		bView, _ := v.horInfinite.FindViewForBlockIndex(blockIndex) // in case black view has changed
		if bView != view {
			// zlog.Info("Block View ChangedDone", blockIndex, "to:", zlog.Pointer(blockView))
			v.updateBlocks[blockIndex] = time.Time{} // force new update for new block view
			v.lastBlockUpdateTime = time.Now()
			v.updatingBlock = false
			return false
		}
		_, _, row, _ := v.FindLaneAndRow(laneID, rowType)
		oname := fmt.Sprintf("%d-%d", laneID, rowType)
		rv, _ := blockView.FindViewWithName(oname, false)
		zlog.Assert(rv != nil, oname)
		rowBGView, _ := rv.(*zcontainer.ContainerView)
		zlog.Assert(row != nil, "FindLR:", laneID, rowType, len(v.lanes))
		posMarg := zgeo.SizeD(float64(x), 0)
		mg := childView.(zview.MinSizeGettable) // let's let this panic if not available, not sure what to do yet if so.
		size := mg.MinSize()
		size.H--
		cellRect := blockView.LocalRect().Align(cellBox, zgeo.TopLeft, posMarg)
		y := int(posMarg.H)
		pos := zgeo.PosI(x, y)
		for _, c := range rowBGView.Cells {
			if c.View.Rect().Pos == pos {
				// zlog.Info("Remove cell view in same spot", cellRect, zlog.Pointer(c.View))
				rowBGView.RemoveChild(c.View, true)
				break
			}
		}
		// if blockIndex == 0 {
		// zlog.Info("AddChildView:", cellRect)
		// }
		rowBGView.Add(childView, zgeo.TopLeft, cellRect.Pos.Size()).Free = true
		// if blockIndex == 0 && pressed {
		// 	zlog.Info("updateBlockView got view:", zlog.Pointer(blockView), x, cellBox, laneID, rowType, blockDone, blockView.CountChildren())
		// }
		childView.SetRect(cellRect)
		return true
	})
}

func defaultMakeBackgroundViewFunc(blockIndex int, laneID int64, row *Row, size zgeo.Size) *zcontainer.ContainerView {
	bg := zcontainer.New("")
	bg.SetMinSize(size)
	// bg.SetBGColor(zgeo.ColorBlue)
	return bg
}

func (v *HorEventsView) makeButtons() {
	v.timeField = ztext.TimeFieldNew("time", ztime.TimeFieldNotFutureIfAmbiguous|ztime.TimeFieldSecs)
	v.timeField.CallChangedOnTabPressed = true
	v.timeField.HandleValueChangedFunc = func() {
		// zlog.Info("Time Field changed")
		t, err := v.timeField.Value()
		if err != nil {
			return
		}
		v.SetScrollToNowOn(false)
		v.GotoTime(t)
	}
	v.timeField.SetToolTip("Enter a time/date and press return to jump to this time")
	v.Bar.Add(v.timeField, zgeo.CenterLeft)

	for i, pan := range v.panDurations {
		var leftKey, rightKey zkeyboard.KeyMod
		if pan.modifier != zkeyboard.ModifierNone {
			leftKey = zkeyboard.KeyMod{Key: zkeyboard.KeyLeftArrow, Modifier: pan.modifier}
			rightKey = zkeyboard.KeyMod{Key: zkeyboard.KeyRightArrow, Modifier: pan.modifier}
		}
		stack := makeButtonPairInStack(int(pan.duration), "zcore/triangle-left-light-gray", "go back in time by "+pan.name, leftKey, "zcore/triangle-right-light-gray", "go forward in time by "+pan.name, rightKey, pan.name, 0, 0, 3, v.panPressed)
		stack.SetCorner(3)
		v.panDurations[i].stack = stack
		v.Bar.Add(stack, zgeo.CenterLeft)
	}
	outKey := zkeyboard.KeyMod{Char: "-"}
	inKey := zkeyboard.KeyMod{Char: "+"}
	v.zoomStack = makeButtonPairInStack(-2, "zcore/zoom-out-gray", "zoom out", outKey, "zcore/zoom-in-gray", "zoom in", inKey, "filler", 4, 28, -1, v.zoomPressed)
	title, _ := v.zoomStack.FindViewWithName("title", true)
	title.SetColor(zgeo.ColorOrange)
	v.Bar.Add(v.zoomStack, zgeo.CenterLeft)

	v.nowButton = zshape.ImageButtonViewSimpleInsets("now", "lightGray")
	v.Bar.Add(v.nowButton, zgeo.CenterLeft)
	v.nowButton.SetPressedHandler("", 0, func() {
		v.SetScrollToNowOn(!v.scrollToNow)
	})
	v.timeField.SetValue(v.startTime) // must b

	v.GotoLockedButton = zshape.NewView(zshape.TypeRoundRect, zgeo.SizeD(24, 20))
	v.GotoLockedButton.Ratio = 0.2
	v.GotoLockedButton.StrokeColor = zgridlist.DefaultSelectColor
	v.GotoLockedButton.StrokeWidth = LockedItemStrokeWidth
	v.GotoLockedButton.SetPressedHandler("", 0, func() {
		v.SetScrollToNowOn(false)
		v.GotoTime(v.LockedTime.Add(-v.BlockDuration / 2))
	})
	v.Bar.Add(v.GotoLockedButton, zgeo.CenterLeft).Collapsed = true
}

func (v *HorEventsView) updateNowScrollAndPole() {
	x := int(v.TimeToXInCorrectBlock(time.Now()))
	if x != v.lastScrollToX {
		if v.scrollToNow {
			v.gotoNowScrollTime()
			v.lastScrollToX = x
		}
		v.updateNowPole()
	}
}

func (v *HorEventsView) SetScrollToNowOn(on bool) {
	v.scrollToNow = on
	name := "lightGray"
	col := zgeo.ColorBlack
	text := "now"
	if v.scrollToNow {
		name = "lightBlue"
		col = zgeo.ColorWhite
		text += " ⏲"
	}
	v.nowButton.SetText(text)
	v.nowButton.SetToolTip("press to jump forward to now-time and keep scrolling.\nPress again to disable.")
	v.nowButton.SetImageName(name, zgeo.Size{})
	v.nowButton.SetTextColor(col)
	if on {
		v.gotoNowScrollTime()
	}
	v.Bar.ArrangeChildren()
}

func (v *HorEventsView) calcDurationToNowLineRatio() time.Duration {
	return time.Duration(float64(v.BlockDuration) * widthRatioToNowLine)
}

func (v *HorEventsView) calcTimePosToShowTime(t time.Time) time.Time {
	return t.Add(-v.calcDurationToNowLineRatio())
}
func (v *HorEventsView) gotoNowScrollTime() {
	t := v.calcTimePosToShowTime(time.Now())
	// zlog.Info("GotoNow:", t)
	v.GotoTime(t)
}

func makeImageView(pathStub string, shortCut zkeyboard.KeyMod, left bool, id int, tip string, pressed func(left bool, id int)) *imageView {
	v := &imageView{}
	v.Init(v, true, nil, "images/"+pathStub+".png", zgeo.SizeBoth(20))
	v.KeyboardShortcut = shortCut
	v.left = left
	v.pressed = pressed
	v.id = id
	v.SetToolTip(tip)
	v.SetPressedHandler("", zkeyboard.ModifierNone, func() {
		v.pressed(left, v.id)
	})
	return v
}

func makeButtonPairInStack(id int, leftImageStub, leftTip string, leftKey zkeyboard.KeyMod, rightImageStub, rightTip string, rightKey zkeyboard.KeyMod, midTitle string, midFontInc, midMinWidth, modMarg float64, pressed func(left bool, id int)) *zcontainer.StackView {
	stack := zcontainer.StackViewHor(midTitle)
	stack.SetMarginS(zgeo.SizeBoth(2))
	stack.SetSpacing(0)

	vleft := makeImageView(leftImageStub, leftKey, true, id, leftTip, pressed)
	vleft.SetObjectName("left")
	vleft.SetMarginS(zgeo.SizeBoth(2))
	stack.Add(vleft, zgeo.CenterLeft)

	label := zlabel.New(midTitle)
	label.SetObjectName("title")
	label.SetTextAlignment(zgeo.Center)
	label.SetMinWidth(midMinWidth)
	label.SetColor(zgeo.ColorLightGray)
	label.SetFont(zgeo.FontNice(zgeo.FontDefaultSize+midFontInc, zgeo.FontStyleBold))
	stack.Add(label, zgeo.CenterLeft, zgeo.SizeD(modMarg, 0))

	vright := makeImageView(rightImageStub, rightKey, false, id, rightTip, pressed)
	vright.SetObjectName("right")
	vright.SetMarginS(zgeo.SizeBoth(2))
	stack.Add(vright, zgeo.CenterLeft)

	return stack
}

var pressed = false

func (v *HorEventsView) zoomPressed(left bool, id int) {
	pressed = true
	v.Bar.SetInteractive(false)
	defer v.Bar.SetInteractive(true)
	if left {
		v.zoomIndex--
	} else {
		v.zoomIndex++
	}
	dur := v.zoomLevels[v.zoomIndex].duration
	// zlog.Info("zoom:", dur, v.zoomIndex)
	v.SetBlockDuration(dur)
	v.UpdateWidgets()
	// v.Bar.ArrangeChildren()
	if v.storeKey != "" {
		zkeyvalue.DefaultStore.SetInt(v.zoomIndex, v.storeKey+zoomIndexKey, true)
	}
	v.Bar.ArrangeChildren()
}

func (v *HorEventsView) UpdateWidgets() {
	leftZoom, _ := v.zoomStack.FindViewWithName("left", true)
	leftZoom.SetUsable(v.zoomIndex > 0)

	rightZoom, _ := v.zoomStack.FindViewWithName("right", true)
	rightZoom.SetUsable(v.zoomIndex < len(v.zoomLevels)-1)
	title, _ := v.zoomStack.FindViewWithName("title", true)
	str := v.zoomLevels[v.zoomIndex].name
	title.(*zlabel.Label).SetText(str)

	for _, pd := range v.panDurations {
		if pd.stack == nil {
			continue
		}
		nearNow := (time.Since(v.currentTime.Add(pd.duration)) < 0)
		rightArrow, _ := pd.stack.FindViewWithName("right", true)
		rightArrow.SetUsable(!nearNow)
		if pd.duration == v.BlockDuration {
			pd.stack.SetBGColor(zgeo.ColorNewGray(0, 0.1))
		} else {
			pd.stack.SetBGColor(zgeo.ColorClear)
		}
	}
	v.Bar.CollapseChild(v.GotoLockedButton, v.LockedTime.IsZero(), true)
}

func (v *HorEventsView) HandleOutsideShortcut(sc zkeyboard.KeyMod) bool {
	if sc.Modifier != zkeyboard.ModifierNone {
		return false
	}
	var left bool
	if sc.Key == zkeyboard.KeyLeftArrow {
		left = true
	} else if sc.Key != zkeyboard.KeyRightArrow {
		return false
	}
	v.panPressed(left, -1)
	return true
}

func (v *HorEventsView) panPressed(left bool, id int) {
	v.SetScrollToNowOn(false)
	v.Bar.SetInteractive(false)
	defer v.Bar.SetInteractive(true)
	d := v.BlockDuration
	if id != -1 {
		d = time.Duration(id)
	}
	if left {
		d *= -1
	}
	t := v.currentTime.Add(d)
	v.GotoTime(t)
}

func (v *HorEventsView) GotoTime(t time.Time) {
	i := v.timeToFractionalBlockIndex(t)
	i = min(i, float64(v.horInfinite.maxIndex))
	v.currentTime = t
	v.horInfinite.SetFloatingCurrentIndex(i)
	v.UpdateWidgets()
}

func (v *HorEventsView) DurationToWidth(d time.Duration) float64 {
	return float64(d) / float64(v.BlockDuration) * v.ViewWidth
}

func (v *HorEventsView) WidthToDuration(w float64) time.Duration {
	return time.Duration(w / v.ViewWidth * float64(v.BlockDuration))
}

func (v *HorEventsView) TimeToBlockIndex(t time.Time) int {
	// zlog.Info("HE TimeToBlockIndex:", t, v.startTime)
	return int(t.Sub(v.startTime) / v.BlockDuration)
}

func (v *HorEventsView) timeToFractionalBlockIndex(t time.Time) float64 {
	diff := t.Sub(v.startTime)
	return float64(diff) / float64(v.BlockDuration)
}

func (v *HorEventsView) TimeToXInHorEventView(t time.Time) (x float64, visible bool) {
	blockIndex := v.TimeToBlockIndex(t)
	diff := blockIndex - int(v.horInfinite.currentIndex)
	x = v.TimeToXInCorrectBlock(t)
	ox := v.horInfinite.ScrollOffsetInBlock()
	x -= ox
	x += float64(diff) * v.ViewWidth
	x = math.Ceil(x)
	visible = (x >= 0 && x < v.ViewWidth)
	// zlog.Info("diff:", diff)
	// visible = (zint.Abs(diff) <= 1)
	return x, visible
}

func (v *HorEventsView) TimeToXInCorrectBlock(t time.Time) float64 {
	n := v.timeToFractionalBlockIndex(t)
	_, fract := math.Modf(n)
	// zlog.Info("TimeToXInCorrectBlock:", t, n, fract, v.ViewWidth)
	return v.ViewWidth * fract
}

func (v *HorEventsView) IndexToTime(i float64) time.Time {
	return v.startTime.Add(time.Duration(i * float64(v.BlockDuration)))
}

func (v *HorEventsView) XInBlockToTime(x float64, blockIndex int) time.Time {
	return v.IndexToTime(float64(blockIndex) + x/v.ViewWidth)
}

func (v *HorEventsView) Update() {
	v.SetLanes(v.lanes)
	v.Reset(true)
}

func (v *HorEventsView) SetLanes(lanes []Lane) {
	const laneTitleHeight = 20
	// zlog.Info("HV.SetLanes:", len(lanes))
	for _, lane := range v.lanes {
		for _, view := range lane.overlayViews {
			v.horInfinite.VertOverlay.RemoveChild(view, true)
		}
		for _, r := range lane.Rows {
			for _, view := range r.overlayViews {
				v.horInfinite.VertOverlay.RemoveChild(view, true)
			}
		}
	}
	v.lanes = make([]Lane, len(lanes))
	y := 0.0
	for i, lane := range lanes {
		v.lanes[i] = lane
		v.lanes[i].y = y
		// zlog.Info("SetLane:", lane.Name, y)
		if len(lane.Rows) == 0 {
			y += 22
		}
		v.lanes[i].Rows = make([]Row, len(lane.Rows))
		for j, r := range lane.Rows {
			r.y = y
			v.lanes[i].Rows[j] = r
			// zlog.Info("AddLaneRow:", lane.Name, r.Name, r.Height, y)
			y += r.Height
		}
		y += dividerHeight
	}
	// if v.IsPresented() {
	v.createLanes()
	// }
}

func (v *HorEventsView) ForLaneOverlayViews(each func(view zview.View, laneID, rowID int64)) {
	for _, lane := range v.lanes {
		for _, view := range lane.overlayViews {
			each(view, lane.ID, 0)
		}
		for _, r := range lane.Rows {
			for _, view := range r.overlayViews {
				each(view, lane.ID, r.ID)
			}
		}
	}
}

func (v *HorEventsView) createLanes() {
	// zlog.Info("He.createLanes", len(v.lanes))
	const laneTitleHeight = 20
	bgWidth := v.ViewWidth //+ v.PoleWidth*2
	var y float64
	for i, lane := range v.lanes {
		title := makeTextTitle(lane.Name, 2, lane.TextColor)
		titleSize, _ := title.CalculatedSize(v.Rect().Size)
		zslice.Add(&v.lanes[i].overlayViews, title)
		y = lane.y
		titlePos := zgeo.PosD(v.GutterWidth.Min+2, lane.y+v.timeAxisHeight)
		v.horInfinite.VertOverlay.Add(title, zgeo.TopLeft, titlePos.Size()).Free = true
		if v.MakeLaneActionIconFunc != nil {
			icon := v.MakeLaneActionIconFunc(lane.ID)
			icon.SetInteractive(true)
			zslice.Add(&v.lanes[i].overlayViews, icon.View)
			iconPos := titlePos //.PlusX(-2)
			iconPos.X -= icon.FitSize().W
			titleSize.W += icon.FitSize().W + 8
			v.horInfinite.VertOverlay.Add(icon, zgeo.TopLeft, iconPos.Size()).Free = true
		}
		// zlog.Info("SetLaneY:", lane.Name, lane.ID, y, len(lane.Rows))
		for j, r := range lane.Rows {
			y = r.y
			rowTitle := makeTextTitle(r.Name, 0, zgeo.ColorLightGray) // lane.TextColor.Mixed(zgeo.ColorBlue, 0.3))
			zslice.Add(&v.lanes[i].Rows[j].overlayViews, rowTitle)
			rowTitlePos := titlePos
			rowTitlePos.Y = r.y + v.timeAxisHeight
			if j == 0 {
				rowTitlePos.X += titleSize.W
			} else {
				rowTitlePos.Y += laneTitleHeight
			}
			v.horInfinite.VertOverlay.Add(rowTitle, zgeo.TopLeft, rowTitlePos.Size()).Free = true
			if v.MakeLaneRowActionIconFunc != nil {
				icon := v.MakeLaneRowActionIconFunc(lane.ID, r.ID)
				icon.SetZIndex(99999)
				icon.SetInteractive(true)
				zslice.Add(&v.lanes[i].overlayViews, icon.View)
				iconPos := rowTitlePos.PlusX(-16)
				v.horInfinite.VertOverlay.Add(icon, zgeo.TopLeft, iconPos.Size()).Free = true
			}
			h := r.Height
			if j == len(lane.Rows)-1 {
				h += dividerHeight
			}
			bgView := v.MakeRowBackgroundViewFunc(zint.Undefined, lane.ID, &r, zgeo.SizeD(bgWidth, h))
			if bgView != nil {
				bgView.SetObjectName(OverlayBackgroundViewName)
				bgView.Native().SetDimUsable(false)
				// bgView.Native().SetUsable(false)
				zslice.Add(&v.lanes[i].Rows[j].overlayViews, bgView.View)
				v.horInfinite.VertOverlay.Add(bgView, zgeo.TopLeft, zgeo.SizeD(0, r.y-1+v.timeAxisHeight)).Free = true
			}
			y = r.y + h
			// zlog.Info("SetLaneRowY:", lane.Name, r.Name, r.Height, y)
		}
	}
	h := max(v.Rect().Size.H-v.Bar.Rect().Size.H, y)
	h = y
	v.horInfinite.SetContentHeight(h)
	v.horInfinite.ArrangeChildren()
}

func (v *HorEventsView) SetRect(r zgeo.Rect) {
	// zlog.Info("HV SetRect", r)
	v.StackView.SetRect(r)
	oldWidth := v.ViewWidth
	v.ViewWidth = r.Size.W - zwindow.ScrollBarSizeForView(v)
	if oldWidth != 0 && v.IsPresented() {
		v.Updater.Update()
	}
	v.horInfinite.IgnoreScroll = false
}

func makeTextTitle(text string, fontAdd float64, col zgeo.Color) zview.View {
	label := zlabel.New(text)
	label.SetCorner(2)
	if !col.Valid {
		col = zgeo.ColorNewGray(0.9, 1)
	}
	label.SetColor(col)
	label.SetFont(zgeo.FontNice(14+fontAdd, zgeo.FontStyleBold))
	label.OutsideDropStroke(3, zgeo.ColorBlack)
	label.SetZIndex(5000)
	return label
}

func (v *HorEventsView) FindLaneAndRow(laneID, rowID int64) (*Lane, int, *Row, int) {
	for i, lane := range v.lanes {
		// zlog.Info("FindLaneAndRow", zlog.Full(lane))
		if lane.ID == laneID {
			if rowID == 0 {
				return &v.lanes[i], i, nil, -1
			}
			for j, r := range lane.Rows {
				if r.ID == rowID {
					return &v.lanes[i], i, &v.lanes[i].Rows[j], j
				}
			}
		}
	}
	return nil, -1, nil, -1
}

func (v *HorEventsView) makeAxisRow(blockIndex int) zview.View {
	start := v.IndexToTime(float64(blockIndex))
	end := start.Add(v.BlockDuration)
	axis := zcustom.NewView("axis")
	axis.SetMinSize(zgeo.SizeD(100, v.timeAxisHeight))
	axis.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, drawView zview.View) {
		// zlog.Info("Axis draw:", blockIndex, rect, start, end)
		beyond := true
		dark := true
		zdraw.DrawHorTimeAxis(canvas, rect, start, end, beyond, dark)
	})
	axis.ExposeIn(0.1)
	axis.SetPressedHandler("marker", 0, func() {
		if len(v.markerTimes) >= 6 {
			return
		}
		t := v.xInBlockToMarkerTime(zview.LastPressedPos.X, blockIndex)
		v.makeMarkerButton(t)
		v.saveMarkerButtons()
	})
	handleMoves := true
	axis.SetPointerEnterHandler(handleMoves, func(pos zgeo.Pos, inside zbool.BoolInd) {
		if inside.IsFalse() {
			v.markerPole.Show(false)
			return
		}
		if inside.IsTrue() {
			v.moveStartX = pos.X
			return
		}
		if zmath.Abs(pos.X-v.moveStartX) < 5 {
			return
		}
		t := v.xInBlockToMarkerTime(pos.X-1, blockIndex)
		v.showMarkerPoleAt(t)
		// zlog.Info("ScrollHeaderX:", blockIndex, pos.X, "->", x, t)
		//		do this in each individual bar-part? so we know blockid?
	})
	return axis
}

func (v *HorEventsView) showMarkerPoleAt(t time.Time) {
	v.markerPole.Show(true)
	v.markerPole.SetObjectName(ztime.GetNice(t, true))
	x, _ := v.TimeToXInHorEventView(t)
	r := zgeo.RectFromXY2(x-35, v.timeAxisHeight+v.Bar.Rect().Size.H, x+35, v.Rect().Size.H)
	v.markerPole.SetRect(r)
	v.markerPole.Expose()
}

func (v *HorEventsView) xInBlockToMarkerTime(x float64, blockIndex int) time.Time {
	t := v.XInBlockToTime(x-1, blockIndex)
	if v.BlockDuration >= time.Hour*8 {
		sec := 0
		if t.Second() > 30 {
			sec = 60
		}
		t = ztime.ChangedPartsOfTime(t, -1, -1, sec, 0)
	}
	return t
}

func (v *HorEventsView) handleBlockViewRemoved(blockIndex int) {
	// zlog.Info("handleBlockViewRemoved:", blockIndex)
	delete(v.updateBlocks, blockIndex)
}

func (v *HorEventsView) makeBlockView(blockIndex int) zview.View {
	v.updateBlocks[blockIndex] = time.Time{}
	blockView := zcontainer.New(strconv.Itoa(blockIndex))
	// if blockIndex == 0 {
	// fmt.Printf("MakeBlockView %d %p\n", blockIndex, blockView)
	// }
	// start := v.IndexToTime(float64(blockIndex))
	// end := start.Add(v.BlockDuration)
	col := v.BGColor()
	if v.TestMode {
		col = zgeo.ColorRed
		num := zlabel.New(fmt.Sprint(blockIndex, ": ", zlog.Pointer(blockView)))
		num.SetTextAlignment(zgeo.Center)
		num.SetColor(zgeo.ColorWhite)
		num.SetZIndex(999999)
		num.SetFont(zgeo.FontNice(25, zgeo.FontStyleBold))
		blockView.Add(num, zgeo.Center).Free = true
	}
	if zint.Abs(blockIndex)%2 == 1 && v.MixOddBlocksColor.Valid {
		col = col.Mixed(v.MixOddBlocksColor, 0.05)
	}
	blockView.SetBGColor(col)

	for i, lane := range v.lanes {
		for j, r := range lane.Rows {
			s := zgeo.SizeD(v.ViewWidth, r.Height)
			lastRow := (j == len(lane.Rows)-1)
			if lastRow {
				s.H += dividerHeight
			}
			// isOverlay := false
			bg := v.MakeRowBackgroundViewFunc(blockIndex, lane.ID, &r, s)
			if bg == nil {
				bg = defaultMakeBackgroundViewFunc(blockIndex, lane.ID, &r, s)
			}
			if i%2 == 1 && v.OddLaneColor.Valid {
				bg.SetBGColor(v.OddLaneColor)
			}
			oname := fmt.Sprintf("%d-%v", lane.ID, r.ID)
			bg.SetObjectName(oname)
			if lastRow {
				bg.SetStrokeSide(dividerHeight, zgeo.ColorBlack, zgeo.Bottom, true)
			} else {
				bg.SetStrokeSide(1, zgeo.ColorNewGray(0, 0.3), zgeo.Bottom, true)
			}
			// zlog.Info("addBG:", blockView != nil, bg != nil)
			blockView.Add(bg, zgeo.TopLeft|zgeo.HorExpand, zgeo.SizeD(0, r.y))
		}
	}
	// zlog.Info("MakeBlockView done", blockIndex)
	return blockView
}

func (v *HorEventsView) IsBlockInWindow(blockIndex int) bool {
	return v.horInfinite.IsBlockInWindow(blockIndex)
}

// updateCurrentBlock adds move events to the current block who's end might not be yet.
func (v *HorEventsView) updateCurrentBlock(blockIndex int) bool {
	return true // return false if we get past end
}

func (v *HorEventsView) releaseViewForIndex(blockIndex int) {
}

func (v *HorEventsView) HandlePan(blockIndex float64) {
	if !v.startupGotoTime.IsZero() {
		t := v.startupGotoTime
		v.startupGotoTime = time.Time{} // let's clear this before calling goto
		// zlog.Info("GOT New Time", t)
		v.GotoTime(t)
		return
	}
	t := v.IndexToTime(blockIndex)
	// zlog.Info("HandlePan", blockIndex, t)
	v.currentTime = t
	v.timeField.SetValue(t)
	v.updateNowPole()
	v.UpdateWidgets()
}

func (v *HorEventsView) BlockViews() []zview.View {
	return v.horInfinite.BlockViews()
}

func (v *HorEventsView) SpinActivity(blockIndex int, spin bool) {
	v.horInfinite.SpinActivity(blockIndex, spin)
}

func (v *HorEventsView) LaneCount() int {
	return len(v.lanes)
}

func (v *HorEventsView) LastUpdateOfBlock(blockIndex int) time.Time {
	return v.updateBlocks[blockIndex]
}

func (v *HorEventsView) ForAllBackgroundViews(got func(blockIndex int, cview zview.View, lane *Lane, row *Row)) {
	v.horInfinite.ForAllBlockViews(func(blockIndex int, bview zview.View) {
		zcontainer.ViewRangeChildren(bview, false, false, func(view zview.View) bool {
			var laneID, rowID int64
			if zstr.SplitN(view.ObjectName(), "-", &laneID, &rowID) {
				// if laneID != 0 && rowID != 0 {
				lane, _, row, _ := v.FindLaneAndRow(laneID, rowID)
				if lane != nil && row != nil {
					got(blockIndex, view, lane, row)
					return true
				}
				// }
			}
			// zlog.Error("Bad bg:", view.ObjectName())
			return true
		})
	})
}

func (v *HorEventsView) ExposeRowViews(rowID int64) {
	v.WithBackgroundViews(rowID, func(blockIndex int, view zview.View, laneID int64, isOverlay bool) {
		if isOverlay {
			view.(*zcontainer.ContainerView).Expose()
			return
		}
		zview.ExposeView(view)
		zcontainer.ViewRangeChildren(view, false, false, func(cview zview.View) bool {
			zview.ExposeView(cview)
			return true
		})
	})
}

func (v *HorEventsView) WithBackgroundViews(rowID int64, got func(blockIndex int, view zview.View, laneID int64, isOverlay bool)) {
	v.ForLaneOverlayViews(func(view zview.View, laneID, laneRowType int64) {
		if view.ObjectName() == OverlayBackgroundViewName && (rowID == 0 || rowID == laneRowType) {
			got(0, view, laneID, true)
		}
	})
	v.ForAllBackgroundViews(func(blockIndex int, cview zview.View, lane *Lane, row *Row) {
		if row.ID == rowID {
			// zlog.Info("All background", lane.ID, row.ID, rowID)
			got(blockIndex, cview, lane.ID, false)
		}
	})
}

func (v *HorEventsView) KeepScrollingToShowAllRows() {
	if v.keepScrollingRepeater == nil && v.horInfinite.Scroller.Rect().Size.H > v.horInfinite.VertStack.Rect().Size.H {
		v.keepScrollingDir = 2
		diff := v.horInfinite.Scroller.Rect().Size.H - v.horInfinite.VertStack.Rect().Size.H
		v.keepScrollingRepeater = ztimer.RepeatForever(0.05, func() {
			o := v.horInfinite.VertStack.ContentOffset().Y
			o += v.keepScrollingDir
			if v.keepScrollingDir > 0 && o >= diff || v.keepScrollingDir < 0 && o < 0 {
				v.keepScrollingDir *= -1
			}
			v.horInfinite.VertStack.SetYContentOffset(o)
		})
		v.AddOnRemoveFunc(v.keepScrollingRepeater.Stop)
	}
}

type panDuration struct {
	name     string
	modifier zkeyboard.Modifier
	duration time.Duration
	stack    *zcontainer.StackView
}

type zoomLevel struct {
	name     string
	duration time.Duration
}

type viewX struct {
	view zview.View
	x    int
}

type imageView struct {
	zimageview.ImageView
	left    bool
	pressed func(left bool, id int)
	id      int
}
