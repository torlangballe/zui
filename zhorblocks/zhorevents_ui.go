//go:build zui && js

package zhorblocks

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zscrollview"
	"github.com/torlangballe/zui/zshape"
	"github.com/torlangballe/zui/zshortcuts"
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zdraw"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmath"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
)

type HorEventsView struct {
	zcontainer.StackView
	BlockDuration             time.Duration
	Bar                       *zcontainer.StackView
	GetEventViewsFunc         func(blockIndex int, isNewView bool, got func(childView zview.View, x int, cellBox zgeo.Size, laneID, rowType int64, blockDone bool))
	MakeRowBackgroundViewFunc func(laneID int64, row *Row, size zgeo.Size) zview.View
	MakeLaneActionIconFunc    func(laneID int64) zview.View
	TestMode                  bool
	Updater                   Updater

	nowLine               zview.View
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
	zoomLevels            []zoomLevel
	panDurations          []panDuration
	panDuration           time.Duration
	zoomIndex             int
	storeKey              string
	currentNowBlockIndex  int
	LastEventTimeForBlock map[int]time.Time
	updateBlocks          map[int]time.Time
	scrollToNow           bool
	lastScrollToX         int
	updateNowRepeater     *ztimer.Repeater
	updatingBlock         bool
	GutterWidth           float64
	timeAxisHeight        float64
}

type Updater interface {
	Update()
}

type Lane struct {
	ID        int64
	Name      string
	Rows      []Row
	TextColor zgeo.Color
	views     []zview.View
	y         float64
	hasAxis   bool
}

type Row struct {
	ID            int64
	Name          string
	Height        float64
	y             float64 // accumulated from top
	ForGlobalLane bool
	views         []zview.View
}

type EventOptions struct {
	StoreKey             string
	BlocksIndexGetWidth  int
	BlockIndexCacheDelta int
	BlockDuration        time.Duration
	StartTime            time.Time
	ShowNowPole          bool
	GutterWidth          float64
	TimeAxisHeight       float64
	BGColor              zgeo.Color
	SelectedEventID      int64
}

const (
	zoomIndexKey        = ".horblock.Events.zoom"
	dividerHeight       = 2
	widthRatioToNowLine = 0.8
)

func NewEventsView(v *HorEventsView, opts EventOptions) *HorEventsView {
	if v == nil {
		v = &HorEventsView{}
	}
	v.StackView.Init(v, true, "hor-events")
	v.SetBGColor(opts.BGColor)
	v.SetCanTabFocus(true)
	v.Updater = v
	v.TestMode = true
	v.GutterWidth = opts.GutterWidth
	v.timeAxisHeight = opts.TimeAxisHeight
	v.storeKey = opts.StoreKey
	v.zoomLevels = []zoomLevel{
		zoomLevel{"1d", time.Hour * 24},
		zoomLevel{"1h", time.Hour * 1},
		zoomLevel{"10m", time.Minute * 10},
		zoomLevel{"1m", time.Minute * 1},
		zoomLevel{"10s", time.Second * 10},
	}
	v.panDurations = []panDuration{
		panDuration{"1d", zkeyboard.ModifierNone, time.Hour * 24, nil},
		panDuration{"1h", zkeyboard.ModifierCommand, time.Hour * 1, nil},
		panDuration{"10m", zkeyboard.ModifierAlt, time.Minute * 10, nil},
		panDuration{"1m", zkeyboard.ModifierCommand | zkeyboard.ModifierAlt, time.Minute, nil},
	}
	blockDuration := opts.BlockDuration
	v.zoomIndex = -1
	v.currentNowBlockIndex = zint.Undefined
	if zkeyvalue.DefaultStore != nil && opts.StoreKey != "" {
		n, got := zkeyvalue.DefaultStore.GetInt(opts.StoreKey+zoomIndexKey, 0)
		if got && n > 0 && n < len(v.zoomLevels) {
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
		v.zoomIndex = 0
		blockDuration = v.zoomLevels[0].duration
	}
	v.LastEventTimeForBlock = map[int]time.Time{}
	v.updateBlocks = map[int]time.Time{}
	v.BlockDuration = blockDuration // must be before calculating startTime
	v.startTime = v.calcTimePosToShowTime(opts.StartTime)
	v.currentTime = v.startTime
	v.SetSpacing(0)
	v.Bar = zcontainer.StackViewHor("bar")
	v.Bar.SetBGColor(zgeo.ColorNewGray(0.4, 1))
	v.Bar.SetMarginS(zgeo.SizeD(8, 3))
	v.Bar.SetSpacing(12)
	v.Add(v.Bar, zgeo.TopLeft|zgeo.HorExpand)
	v.makeButtons()

	v.horInfinite = NewHorBlocksView(opts.BlocksIndexGetWidth, opts.BlockIndexCacheDelta)
	v.horInfinite.GetViewFunc = v.makeBlockView
	v.horInfinite.RemovedViewFunc = v.handleBlockViewRemoved
	v.horInfinite.PanHandler = v
	v.horInfinite.SetBGColor(opts.BGColor)
	v.Add(v.horInfinite, zgeo.TopLeft|zgeo.Expand)
	v.horInfinite.IgnoreScroll = true

	v.leftPole = v.makeSidePole(zgeo.Left)
	v.rightPole = v.makeSidePole(zgeo.Right)

	v.horInfinite.CreateHeaderBlockView = func(blockIndex int, w float64) zview.View {
		box := v.makeAxisRow(blockIndex)
		return box
	}

	v.SetKeyHandler(func(km zkeyboard.KeyMod, down bool) bool {
		if !down {
			return false
		}
		return zshortcuts.HandleOutsideShortcutRecursively(v, km)
	})

	if opts.ShowNowPole {
		line := zcustom.NewView("now-pole")
		line.SetZIndex(11000)
		line.SetMinSize(zgeo.SizeD(20, 100))
		line.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
			colors := []zgeo.Color{zgeo.ColorNew(0, 1, 0, 0), zgeo.ColorNew(0, 1, 0, 0.8)}
			path := zgeo.PathNewRect(rect, zgeo.SizeNull)
			canvas.DrawGradient(path, colors, rect.Min(), rect.TopRight(), nil)
		})
		//		line.SetBGColor(zgeo.ColorNew(0, 1, 0, 0.5))
		v.nowLine = line
		v.Add(line, zgeo.AlignmentNone)
	}
	v.updateNowRepeater = ztimer.RepeaterNew()
	v.AddOnRemoveFunc(v.updateNowRepeater.Stop)
	v.horInfinite.SetMaxIndex(1)
	v.updateWidgets()
	lineRepeater := ztimer.RepeatForever(0.1, v.updateNowScrollAndPole)
	v.AddOnRemoveFunc(lineRepeater.Stop)
	updateRepeater := ztimer.RepeatForever(0.01, func() {
		v.updateCurrentBlockViews()
	})
	v.AddOnRemoveFunc(updateRepeater.Stop)
	return v
}

func (v *HorEventsView) makeSidePole(a zgeo.Alignment) *zcontainer.StackView {
	pole := zcontainer.StackViewVert(a.String() + "-pole")
	pole.SetMinSize(zgeo.SizeD(v.GutterWidth, 10))
	pole.SetBGColor(v.BGColor().WithOpacity(0.8))
	// pole.SetBGColor(zgeo.ColorBlue)
	pole.SetPressedHandler("pole-click", 0, func() {
		v.handlePolePress(false)
	})
	v.horInfinite.VertOverlay.Add(pole, zgeo.Top|a|zgeo.VertExpand, zgeo.SizeD(0, v.timeAxisHeight)).Free = true
	return pole
}

func (v *HorEventsView) handlePolePress(left bool) {
	y := zview.LastPressedPos.Y
	zlog.Info("handlePolePress:", left, y, v.horInfinite.VertStack.ContentOffset().Y)

}

// func rescaleGraph(rt *row, laneID int64, view zview.View) {
// 	zlog.Info("Rescale:", rt.rowType, laneID)
// }

func (v *HorEventsView) updateNowPole() {
	if v.nowLine == nil {
		return
	}
	now := time.Now()
	nowBlockIndex := v.TimeToBlockIndex(now)
	diff := nowBlockIndex - int(v.horInfinite.currentIndex)
	show := (zint.Abs(diff) <= 1)
	v.nowLine.Show(show)
	x := v.TimeToXInCorrectBlock(now)
	ox := v.horInfinite.ScrollOffsetInBlock()
	// x += v.GutterWidth
	x -= ox
	x += float64(diff) * v.ViewWidth
	x = math.Ceil(x)
	y := v.horInfinite.Rect().Pos.Y
	v.nowLine.Native().SetZIndex(911000)
	// zlog.Info("NowPole:", diff, x, now, show, nowBlockIndex, v.horInfinite.currentIndex, "y:", y, v.horInfinite.Rect().Size.H)
	v.nowLine.Native().SetRect(zgeo.RectFromXYWH(x, y, 10, v.horInfinite.Rect().Size.H))
}

func (v *HorEventsView) SetBlockDuration(d time.Duration) {
	t := v.currentTime.Add(v.calcDurationToNowLineRatio())
	// ztime.Minimize(&nowLineRatioTime, time.Now())
	// zlog.Info("*************** SetBlockDuration 1:", v.currentTime, "->", t, v.horInfinite.DebugPrintList())
	v.BlockDuration = d
	// t := nowLineRatioTime.Add(-v.calcDurationToNowLineRatio())
	ztime.Minimize(&t, time.Now())
	v.startTime = v.calcTimePosToShowTime(t).Add(time.Second * 3)
	v.currentTime = v.startTime
	v.horInfinite.SetFloatingCurrentIndex(0)
	v.Updater.Update()
	// zlog.Info("*************** SetBlockDuration 2:", v.currentTime, "->", t)
	// v.GotoTime(t)
}

func (v *HorEventsView) Reset() {
	v.LastEventTimeForBlock = map[int]time.Time{}
	v.updateBlocks = map[int]time.Time{}
	// zlog.Info("HEV.Reset")
	v.horInfinite.SetMaxIndex(1)
	v.horInfinite.Reset(v.ViewWidth != 0)
}

func (v *HorEventsView) updateCurrentBlockViews() {
	if v.updatingBlock || v.horInfinite.Updating {
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
		v.updatingBlock = false
		return
	}
	endTimeOfBlock := v.IndexToTime(float64(blockIndex) + 1)
	if !isNew && time.Since(endTimeOfBlock) > time.Second*10 { // if it's an old block, and last events written to db and gotten, don't update anymore.
		// zlog.Info("updateBlockView1 delete:", blockIndex)
		delete(v.updateBlocks, blockIndex)
		v.updatingBlock = false
		return
	}
	// zlog.Info("updateBlock", blockIndex)
	// start := time.Now()
	v.updateBlocks[blockIndex] = time.Now()
	view, _ := v.horInfinite.FindViewForIndex(blockIndex)
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
	go v.GetEventViewsFunc(blockIndex, isNew, func(childView zview.View, x int, cellBox zgeo.Size, laneID, rowType int64, blockDone bool) {
		if blockDone {
			v.updatingBlock = false
			return
		}
		_, row := v.FindLaneAndRow(laneID, rowType)
		zlog.Assert(row != nil, "FindLR:", laneID, rowType, len(v.lanes))
		posMarg := zgeo.SizeD(float64(x), row.y)
		mg := childView.(zview.MinSizeGettable) // let's let this panic if not available, not sure what to do yet if so.
		size := mg.MinSize()
		size.H--
		cellRect := blockView.LocalRect().Align(cellBox, zgeo.TopLeft, posMarg)
		y := int(posMarg.H)
		pos := zgeo.PosI(x, y)
		for _, c := range blockView.Cells {
			if c.View.Rect().Pos == pos {
				// zlog.Info("Remove cell view in same spot")
				blockView.RemoveChild(c.View, true)
				break
			}
		}
		// if blockIndex == 0 {
		// 	zlog.Info("updateBlock AddChildView[0]:", zlog.Pointer(blockView), cellRect, laneID, rowType)
		// }
		blockView.Add(childView, zgeo.TopLeft, cellRect.Pos.Size()).Free = true
		// if blockIndex == 0 && pressed {
		// 	zlog.Info("updateBlockView got view:", zlog.Pointer(blockView), x, cellBox, laneID, rowType, blockDone, blockView.CountChildren())
		// }
		childView.SetRect(cellRect)
		v.updateBlocks[blockIndex] = time.Now() // set it to after event view gotten
	})
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
		v.setScrollToNowOn(false)
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
	leftKey := zkeyboard.KeyMod{Key: zkeyboard.KeyLeftArrow, Modifier: zkeyboard.ModifierShift}
	rightKey := zkeyboard.KeyMod{Key: zkeyboard.KeyRightArrow, Modifier: zkeyboard.ModifierShift}
	v.zoomStack = makeButtonPairInStack(-2, "zcore/zoom-out-gray", "zoom out", leftKey, "zcore/zoom-in-gray", "zoom in", rightKey, "filler", 4, 28, -1, v.zoomPressed)
	v.Bar.Add(v.zoomStack, zgeo.CenterLeft)

	v.nowButton = zshape.ImageButtonViewSimpleInsets("now", "lightGray")
	v.Bar.Add(v.nowButton, zgeo.CenterLeft)
	v.nowButton.SetPressedHandler("", 0, func() {
		v.setScrollToNowOn(!v.scrollToNow)
	})
	v.timeField.SetValue(v.startTime) // must b
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

func (v *HorEventsView) setScrollToNowOn(on bool) {
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
	// zlog.Info("zoomPressed")
	pressed = true
	v.Bar.SetInteractive(false)
	defer v.Bar.SetInteractive(true)
	if left {
		v.zoomIndex--
	} else {
		v.zoomIndex++
	}
	dur := v.zoomLevels[v.zoomIndex].duration
	v.SetBlockDuration(dur)
	v.updateWidgets()
	// v.Bar.ArrangeChildren()
	if v.storeKey != "" {
		zkeyvalue.DefaultStore.SetInt(v.zoomIndex, v.storeKey+zoomIndexKey, true)
	}
	v.Bar.ArrangeChildren()
}

func (v *HorEventsView) updateWidgets() {
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
}

func (v *HorEventsView) HandleOutsideShortcut(sc zkeyboard.KeyMod) bool {
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
	v.setScrollToNowOn(false)
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
	// zlog.Info("Pan:", v.currentTime, "+", d, "->", t)
	if zmath.Abs(d) > v.BlockDuration {
		v.GotoTime(t)
		return
	}
	w := v.ViewWidth * (float64(d) / float64(v.BlockDuration))
	x := v.horInfinite.VertStack.ContentOffset().X
	x += w
	// zlog.Info("pan", d, w, v.BlockDuration, x)
	// v.horInfinite.IgnoreScroll = true
	// v.horInfinite.SetXContentOffsetAnimated(x, func() {
	v.GotoTime(t)
	// v.horInfinite.IgnoreScroll = false
	// })
}

func (v *HorEventsView) GotoTime(t time.Time) {
	i := v.timeToFractionalBlockIndex(t)
	i = min(i, float64(v.horInfinite.maxIndex))
	// zlog.Info("GotoTime:", t, "index:", i, v.BlockDuration, v.horInfinite.DebugPrintList(), "cur:", v.horInfinite.CurrentIndex())
	v.currentTime = t
	v.horInfinite.SetFloatingCurrentIndex(i)
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

func (v *HorEventsView) TimeToXInCorrectBlock(t time.Time) float64 {
	n := v.timeToFractionalBlockIndex(t)
	_, fract := math.Modf(n)
	// zlog.Info("TimeToXInCorrectBlock:", t, n, fract, v.ViewWidth)
	return v.ViewWidth * fract
}

func (v *HorEventsView) IndexToTime(i float64) time.Time {
	return v.startTime.Add(time.Duration(i * float64(v.BlockDuration)))
}

func (v *HorEventsView) Update() {
	v.SetLanes(v.lanes)
	v.Reset()
}

func (v *HorEventsView) SetLanes(lanes []Lane) {
	// zlog.Info("HV.SetLanes:", len(lanes))
	for _, lane := range v.lanes {
		for _, view := range lane.views {
			v.horInfinite.VertOverlay.RemoveChild(view, true)
		}
		for _, r := range lane.Rows {
			for _, view := range r.views {
				v.horInfinite.VertOverlay.RemoveChild(view, true)
			}
		}
	}
	v.lanes = make([]Lane, len(lanes))
	y := 0.0
	for i, lane := range lanes {
		v.lanes[i] = lane
		v.lanes[i].y = y
		if len(lane.Rows) == 0 {
			y += 22
		}
		v.lanes[i].Rows = make([]Row, len(lane.Rows))
		for j, r := range lane.Rows {
			// zlog.Info("SetLaneRow:", lane.Name, r.Name, r.ID, r.Height)
			r.y = y
			v.lanes[i].Rows[j] = r
			y += r.Height
		}
		y += dividerHeight
	}
	if v.IsPresented() {
		v.createLanes()
	}
}

func (v *HorEventsView) ForLaneViews(each func(view zview.View, laneID, rowID int64)) {
	for _, lane := range v.lanes {
		for _, view := range lane.views {
			each(view, lane.ID, 0)
		}
		for _, r := range lane.Rows {
			for _, view := range r.views {
				each(view, lane.ID, r.ID)
			}
		}

	}
}

func (v *HorEventsView) createLanes() {
	// zlog.Info("HEV setLanes:", len(lanes), v.blockStack.Rect())

	// bs, _ := v.Bar.CalculatedSize(v.Rect().Size)
	// zlog.Info("SetLanes:", len(v.lanes), v.Bar.Rect().Max().Y, v.timeAxisHeight)
	bgWidth := v.ViewWidth //+ v.PoleWidth*2
	var y float64
	for i, lane := range v.lanes {
		title := makeTextTitle(lane.Name, 2, lane.TextColor)
		zslice.Add(&v.lanes[i].views, title)
		v.horInfinite.VertOverlay.Add(title, zgeo.TopLeft, zgeo.SizeD(v.GutterWidth+2, lane.y)).Free = true
		if v.MakeLaneActionIconFunc != nil {
			view := v.MakeLaneActionIconFunc(lane.ID)
			zslice.Add(&v.lanes[i].views, view)
			v.horInfinite.VertOverlay.Add(view, zgeo.TopLeft, zgeo.SizeD(2, lane.y+2)).Free = true
		}
		// zlog.Info("SetLaneY:", lane.Name, lane.ID, y, len(lane.Rows))
		for j, r := range lane.Rows {
			rowTitle := makeTextTitle(r.Name, 0, zgeo.Color{})
			zslice.Add(&v.lanes[i].Rows[j].views, rowTitle)
			if v.MakeRowBackgroundViewFunc != nil {
				bgView := v.MakeRowBackgroundViewFunc(lane.ID, &r, zgeo.SizeD(bgWidth, r.Height))
				if bgView != nil {
					bgView.Native().SetDimUsable(false)
					// bgView.Native().SetUsable(false)
					zslice.Add(&v.lanes[i].Rows[j].views, bgView)
					v.horInfinite.VertOverlay.Add(bgView, zgeo.TopLeft, zgeo.SizeD(0, r.y+1)).Free = true
				}
			}
			v.horInfinite.VertOverlay.Add(rowTitle, zgeo.TopRight, zgeo.SizeD(v.GutterWidth+2, r.y)).Free = true
			y = r.y + r.Height
			// zlog.Info("SetLaneRowY:", lane.Name, r.Name, r.Height, y)
		}
		div := zcustom.NewView("divider")
		div.SetBGColor(zgeo.ColorNewGray(0.1, 1))
		div.SetMinSize(zgeo.SizeD(bgWidth, dividerHeight))
		y += dividerHeight
		v.horInfinite.VertOverlay.Add(div, zgeo.TopRight, zgeo.SizeD(0, y)).Free = true
		zslice.Add(&v.lanes[i].views, div.View)
	}
	h := max(v.Rect().Size.H-v.Bar.Rect().Size.H, y) // +zscrollview.DefaultBarSize
	v.horInfinite.SetContentHeight(h)
	// zlog.Info("SetBlockH:", v.Rect().Size.H, v.Bar.Rect().Size.H, v.horInfinite.MinSize(), v.ViewWidth)
	//	freeOnly := true
	//	v.blockStack.ArrangeAdvanced(freeOnly)
	v.horInfinite.ArrangeChildren()
}

func (v *HorEventsView) SetRect(r zgeo.Rect) {
	// zlog.Info("HV SetRect", r.Size, v.ViewWidth)
	v.StackView.SetRect(r)
	oldWidth := v.ViewWidth
	v.ViewWidth = r.Size.W - zscrollview.DefaultBarSize
	if oldWidth != 0 && v.IsPresented() {
		// zlog.Info("HV SetRect Update", r.Size, v.ViewWidth)
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

func (v *HorEventsView) FindLaneAndRow(laneID, rowID int64) (*Lane, *Row) {
	for i, lane := range v.lanes {
		// zlog.Info("FindLaneAndRow", zlog.Full(lane))
		if lane.ID == laneID {
			for j, r := range lane.Rows {
				if r.ID == rowID {
					return &v.lanes[i], &v.lanes[i].Rows[j]
				}
			}
		}
	}
	return nil, nil
}

func (v *HorEventsView) makeAxisRow(blockIndex int) zview.View {
	start := v.IndexToTime(float64(blockIndex))
	end := start.Add(v.BlockDuration)
	axis := zcustom.NewView("axis")
	axis.SetMinSize(zgeo.SizeD(100, v.timeAxisHeight))
	axis.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, drawView zview.View) {
		// zlog.Info("Axis draw:", blockIndex)
		beyond := true
		dark := true
		zdraw.DrawHorTimeAxis(canvas, rect, start, end, beyond, dark)
	})
	axis.ExposeIn(0.1)
	return axis
}

func (v *HorEventsView) handleBlockViewRemoved(blockIndex int) {
	// zlog.Info("handleBlockViewRemoved:", blockIndex)
	delete(v.updateBlocks, blockIndex)
}

func (v *HorEventsView) makeBlockView(blockIndex int) zview.View {
	v.updateBlocks[blockIndex] = time.Time{}
	blockView := zcontainer.New(strconv.Itoa(blockIndex))
	// if blockIndex == 0 {
	// 	fmt.Printf("MakeBlockView %d %p\n", blockIndex, blockView)
	// }
	// start := v.IndexToTime(float64(blockIndex))
	// end := start.Add(v.BlockDuration)
	col := v.BGColor()
	if v.TestMode {
		// col = zgeo.ColorGreen
		// if zint.Abs(blockIndex)%2 == 1 {
		// 	col = zgeo.ColorRed
		// }
		// num := zlabel.New(fmt.Sprint(blockIndex, ": ", ztime.GetNiceSubSecs(start, 3), "-", end.Sub(start), 3))
		num := zlabel.New(fmt.Sprint(blockIndex, ": ", zlog.Pointer(blockView)))
		num.SetTextAlignment(zgeo.Center)
		num.SetColor(zgeo.ColorWhite)
		num.SetFont(zgeo.FontNice(25, zgeo.FontStyleBold))
		blockView.Add(num, zgeo.Center).Free = true
	}
	blockView.SetBGColor(col)
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
	t := v.IndexToTime(blockIndex)
	// zlog.Info("HandlePan", blockIndex, t)
	v.currentTime = t
	v.timeField.SetValue(t)
	v.updateNowPole()
	v.updateWidgets()
}

func (v *HorEventsView) BlockViews() []zview.View {
	return v.horInfinite.BlockViews()
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
