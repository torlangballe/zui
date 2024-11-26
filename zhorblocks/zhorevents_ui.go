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
	GetEventViewsFunc         func(blockIndex int, got func(childView zview.View, x int, cellBox zgeo.Size, laneID, rowType int64))
	MakeRowBackgroundViewFunc func(laneID int64, row *Row, size zgeo.Size) zview.View
	ResetFunc                 func()
	MakeLaneActionIconFunc    func(laneID int64) zview.View
	TimeAxisHeight            float64
	WidthChanger              WidthChanger
	TestMode                  bool

	nowBar                zview.View
	blockStack            *zcontainer.StackView
	lanes                 []Lane
	horInfinite           *HorBlocksView
	ViewWidth             float64
	startTime             time.Time
	lanesSet              bool
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
	scrollToNow           bool
	lastScrollToX         int
	updateNowRepeater     *ztimer.Repeater
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
	ID     int64
	Name   string
	Height float64
	y      float64 // accumulated from top
	views  []zview.View
}

type Options struct {
	StoreKey             string
	BlocksIndexGetWidth  int
	BlockIndexCacheDelta int
	BlockDuration        time.Duration
	StartTime            time.Time
	ShowNowPole          bool
}

const (
	zoomIndexKey = ".horblock.Events.zoom"
	poleWidth    = 26
)

var BGColor = zgeo.ColorNewGray(0.3, 1)

func NewEventsView(v *HorEventsView, opts Options) *HorEventsView {
	if v == nil {
		v = &HorEventsView{}
	}
	// v.TestMode = true
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
	v.BlockDuration = blockDuration // must be before calculating startTime
	v.startTime = v.calcTimePosToShowTime(opts.StartTime)
	v.StackView.Init(v, true, "hor-events")
	v.SetSpacing(0)
	v.Bar = zcontainer.StackViewHor("bar")
	v.Bar.SetBGColor(zgeo.ColorNewGray(0.4, 1))
	v.Bar.SetMarginS(zgeo.SizeD(8, 3))
	v.Bar.SetSpacing(12)
	v.Add(v.Bar, zgeo.TopLeft|zgeo.HorExpand)
	v.makeButtons()

	vertScroll := zscrollview.New()
	vertScroll.SetBGColor(BGColor)
	vertScroll.SetObjectName("block-vstack")
	vertScroll.JSSet("className", "zdarkscroll")
	v.Add(vertScroll, zgeo.TopLeft|zgeo.Expand)

	v.blockStack = zcontainer.StackViewHor("block-stack")
	v.blockStack.SetSpacing(0)
	// v.Add(v.blockStack, zgeo.TopLeft|zgeo.Expand)
	vertScroll.AddChild(v.blockStack, nil)

	v.leftPole = zcontainer.StackViewVert("left-pole")
	v.leftPole.SetMinSize(zgeo.SizeD(poleWidth, 10))
	v.leftPole.SetBGColor(BGColor)
	v.blockStack.Add(v.leftPole, zgeo.TopLeft|zgeo.VertExpand)

	v.horInfinite = NewHorBlocksView(opts.BlocksIndexGetWidth, opts.BlockIndexCacheDelta)
	v.horInfinite.GetViewFunc = v.getBlockViewForIndex
	v.horInfinite.WidthChanger = v
	v.horInfinite.PanHandler = v
	v.horInfinite.SetBGColor(BGColor)
	v.blockStack.Add(v.horInfinite, zgeo.TopLeft|zgeo.Expand)

	v.rightPole = zcontainer.StackViewVert("right-pole")
	v.rightPole.SetMinSize(zgeo.SizeD(poleWidth, 10))
	v.rightPole.SetBGColor(BGColor)
	v.blockStack.Add(v.rightPole, zgeo.TopRight|zgeo.VertExpand)

	v.SetCanTabFocus(true)

	v.SetKeyHandler(func(km zkeyboard.KeyMod, down bool) bool {
		if !down {
			return false
		}
		return zcontainer.HandleOutsideShortcutRecursively(v, km)
	})
	v.currentTime = v.startTime
	v.timeField.SetValue(v.startTime)

	if opts.ShowNowPole {
		pole := zcustom.NewView("now-pole")
		pole.SetZIndex(6000)
		pole.SetMinSize(zgeo.SizeD(2, 100))
		pole.SetBGColor(zgeo.ColorNew(0, 1, 0, 0.5))
		v.blockStack.Add(pole, zgeo.AlignmentNone)
		v.nowBar = pole
	}
	v.updateNowRepeater = ztimer.RepeaterNew()
	v.AddOnRemoveFunc(v.updateNowRepeater.Stop)
	v.updateWidgets()
	poleRepeater := ztimer.RepeatForever(0.1, v.updateNowScrollAndPole)
	v.AddOnRemoveFunc(poleRepeater.Stop)
	return v
}

func (v *HorEventsView) updateNowPole() {
	if v.nowBar == nil {
		return
	}
	now := time.Now()
	nowBlockIndex := v.TimeToBlockIndex(now)
	diff := nowBlockIndex - v.horInfinite.currentIndex
	show := (diff == 0 || diff == 1)
	v.nowBar.Show(show)
	x := v.TimeToXInCorrectBlock(now)
	ox := v.horInfinite.ScrollOffsetFromCurrent()
	x += poleWidth
	x -= ox
	x += float64(diff) * v.ViewWidth
	x = math.Ceil(x)
	// zlog.Info("NowPole:", x, now, show)
	v.nowBar.Native().SetRect(zgeo.RectFromXYWH(x, 0, 2, v.horInfinite.Rect().Size.H))
}

func (v *HorEventsView) calculateUpdateNowSecs() float64 {
	return min(5, max(1, ztime.DurSeconds(v.BlockDuration)/25))
}

func (v *HorEventsView) SetBlockDuration(d time.Duration) {
	t := v.calcTimePosToShowTime(time.Now())
	v.BlockDuration = d
	v.Reset()
	t = v.calcTimePosToShowTime(time.Now())
	v.GotoTime(t)
}

func (v *HorEventsView) Reset() {
	if v.ResetFunc != nil {
		v.ResetFunc()
	}
	v.LastEventTimeForBlock = map[int]time.Time{}
	v.horInfinite.Reset(true)
	v.updateNowRepeater.Set(v.calculateUpdateNowSecs(), false, func() bool {
		v.updateCurrentBlockViews()
		return true
	})
}

func (v *HorEventsView) updateCurrentBlockViews() {
	// zlog.Info("UpdateNow")
	i := v.TimeToBlockIndex(time.Now())
	if i != v.currentNowBlockIndex {
		v.horInfinite.SetMaxIndex(i + 1)
		if v.currentNowBlockIndex != zint.Undefined {
			old := v.currentNowBlockIndex
			v.updateBlockView(old, nil)
			ztimer.StartIn(v.calculateUpdateNowSecs(), func() { //
				v.updateBlockView(old, nil) // do outgoing one one last time in a bit, when we hope last events are in
			})
		}
		v.currentNowBlockIndex = i
	}
	go v.updateBlockView(v.currentNowBlockIndex, nil)
}

func (v *HorEventsView) makeButtons() {
	v.timeField = ztext.TimeFieldNew("time", ztime.TimeFieldNotFutureIfAmbiguous|ztime.TimeFieldSecs)
	v.timeField.CallChangedOnTabPressed = true
	v.timeField.HandleValueChangedFunc = func() {
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
	v.zoomStack = makeButtonPairInStack(-2, "zcore/zoom-out-gray", "zoom out", leftKey, "zcore/zoom-in-gray", "zoom in", rightKey, "xxx2xxx", 4, 28, -1, v.zoomPressed)
	v.Bar.Add(v.zoomStack, zgeo.CenterLeft)

	v.nowButton = zshape.ImageButtonViewSimpleInsets("now", "lightGray")
	v.Bar.Add(v.nowButton, zgeo.CenterLeft)
	v.nowButton.SetPressedHandler("", 0, func() {
		v.setScrollToNowOn(!v.scrollToNow)
	})
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

func (v *HorEventsView) calcTimePosToShowTime(t time.Time) time.Time {
	return t.Add(-(v.BlockDuration * 7) / 10)
}
func (v *HorEventsView) gotoNowScrollTime() {
	t := v.calcTimePosToShowTime(time.Now())
	zlog.Info("GotoNow:", t)
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

func (v *HorEventsView) zoomPressed(left bool, id int) {
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
	v.Bar.ArrangeChildren()
	if v.storeKey != "" {
		zkeyvalue.DefaultStore.SetInt(v.zoomIndex, v.storeKey+zoomIndexKey, true)
	}
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
		nearNow := (time.Since(v.currentTime.Add(v.BlockDuration)) < 0)
		rightArrow, _ := pd.stack.FindViewWithName("right", true)
		zlog.Info("near?:", pd.duration, nearNow, v.currentTime, v.currentTime.Add(pd.duration))
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
	x := v.horInfinite.ContentOffset().X
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
	zlog.Info("GotoTime:", t, "index:", i, v.BlockDuration)
	v.horInfinite.SetCurrentIndex(i)
}

func (v *HorEventsView) DurationToWidth(d time.Duration) float64 {
	return float64(d) / float64(v.BlockDuration) * v.ViewWidth
}

func (v *HorEventsView) WidthToDuration(w float64) time.Duration {
	return time.Duration(w / v.ViewWidth * float64(v.BlockDuration))
}

func (v *HorEventsView) TimeToBlockIndex(t time.Time) int {
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

func (v *HorEventsView) SetLanes(lanes []Lane) {
	if v.IsPresented() {
		for _, lane := range v.lanes {
			for _, view := range lane.views {
				v.blockStack.RemoveChild(view, true)
			}
			for _, r := range lane.Rows {
				for _, view := range r.views {
					v.blockStack.RemoveChild(view, true)
				}
			}
		}
		v.setLanes(lanes)
	} else {
		v.lanes = lanes
	}
}

func (v *HorEventsView) setLanes(lanes []Lane) {
	const dividerHeight = 2
	// zlog.Info("HEV setLanes:", len(lanes), v.blockStack.Rect())
	v.lanesSet = true
	v.lanes = make([]Lane, len(lanes))

	// bs, _ := v.Bar.CalculatedSize(v.Rect().Size)
	lastAxisY := -999999.0
	y := 0.0
	// zlog.Info("SetLanes:", len(lanes), y, bs.H, v.Bar.Rect().Max().Y, v.TimeAxisHeight)
	bgWidth := v.ViewWidth + poleWidth*2
	for i, lane := range lanes {
		v.lanes[i] = lane
		v.lanes[i].y = y
		if y-lastAxisY > v.LocalRect().Size.H*0.7 {
			lastAxisY = y
			v.lanes[i].hasAxis = true
			y += v.TimeAxisHeight
		}
		v.lanes[i].Rows = make([]Row, len(lane.Rows))
		title := makeTextTitle(lane.Name, 2, lane.TextColor)
		zslice.Add(&v.lanes[i].views, title)
		v.blockStack.Add(title, zgeo.TopLeft, zgeo.SizeD(poleWidth+2, y)).Free = true
		if v.MakeLaneActionIconFunc != nil {
			view := v.MakeLaneActionIconFunc(lane.ID)
			zslice.Add(&v.lanes[i].views, view)
			v.blockStack.Add(view, zgeo.TopLeft, zgeo.SizeD(2, y+2)).Free = true
		}
		// zlog.Info("SetLaneY:", lane.Name, lane.ID, y, len(lane.Rows))
		if len(lane.Rows) == 0 {
			y += 22
		}
		for j, r := range lane.Rows {
			v.lanes[i].Rows[j] = r
			v.lanes[i].Rows[j].y = y
			rowTitle := makeTextTitle(r.Name, 0, zgeo.Color{})
			zslice.Add(&v.lanes[i].Rows[j].views, rowTitle)
			if v.MakeRowBackgroundViewFunc != nil {
				bgView := v.MakeRowBackgroundViewFunc(lane.ID, &r, zgeo.SizeD(bgWidth, r.Height))
				if bgView != nil {
					bgView.Native().SetDimUsable(false)
					bgView.Native().SetUsable(false)
					zslice.Add(&v.lanes[i].Rows[j].views, bgView)
					v.blockStack.Add(bgView, zgeo.TopRight, zgeo.SizeD(0, y)).Free = true
				}
			}
			v.blockStack.Add(rowTitle, zgeo.TopRight, zgeo.SizeD(poleWidth+2, y)).Free = true
			// zlog.Info("SetLaneRowY:", lane.Name, r.Name, r.Height, y)
			y += r.Height
		}
		div := zcustom.NewView("divider")
		div.SetBGColor(zgeo.ColorNewGray(0.1, 1))
		div.SetMinSize(zgeo.SizeD(bgWidth, dividerHeight))
		v.blockStack.Add(div, zgeo.TopRight, zgeo.SizeD(0, y)).Free = true
		zslice.Add(&v.lanes[i].views, div.View)
		y += dividerHeight
		v.blockStack.SetMinSize(zgeo.SizeD(0, y))
	}
	freeOnly := true
	h := max(v.Rect().Size.H-v.Bar.Rect().Size.H, y+zscrollview.DefaultBarSize)
	v.blockStack.SetMinSize(zgeo.SizeD(100, h))
	// zlog.Info("SetLaneY:", v.blockStack.MinSize(), zdebug.CallingStackString())
	v.blockStack.ArrangeAdvanced(freeOnly)
}

func (v *HorEventsView) SetRect(r zgeo.Rect) {
	v.NativeView.SetRect(r) // native just sets v's rect
	v.Reset()
	if !v.lanesSet {
		v.SetLanes(v.lanes)
	}
	v.ContainerView.SetRect(r) // now we set rect as container, so lanes set in setLanes above are placed.
}

func makeTextTitle(text string, fontAdd float64, col zgeo.Color) zview.View {
	label := zlabel.New(text)
	label.SetCorner(2)
	if !col.Valid {
		col = zgeo.ColorNewGray(0.9, 1)
	}
	label.SetColor(col)
	label.SetFont(zgeo.FontNice(14+fontAdd, zgeo.FontStyleBold))
	label.OutsideDropStroke(1, zgeo.ColorBlack)
	label.SetZIndex(5000)
	return label
}

func (v *HorEventsView) FindLaneAndRow(laneID, rowID int64) (*Lane, *Row) {
	for i, lane := range v.lanes {
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
	axis.SetMinSize(zgeo.SizeD(100, v.TimeAxisHeight))
	axis.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, drawView zview.View) {
		beyond := true
		dark := true
		zdraw.DrawHorTimeAxis(canvas, rect, start, end, false, beyond, dark)
	})
	return axis
}

func (v *HorEventsView) makeBlockView(blockIndex int) zview.View {
	// if blockIndex == 0 {
	// 	zlog.Info("MakeBlockView", v.blockStack.Rect())
	// }
	blockView := zcontainer.New(strconv.Itoa(blockIndex))
	start := v.IndexToTime(float64(blockIndex))
	end := start.Add(v.BlockDuration)
	col := BGColor
	if v.TestMode {
		col = zgeo.ColorGreen
		if zint.Abs(blockIndex)%2 == 1 {
			col = zgeo.ColorRed
		}
		num := zlabel.New(fmt.Sprint(blockIndex, ": ", ztime.GetNiceSubSecs(start, 3), "-", end.Sub(start), 3))
		num.SetTextAlignment(zgeo.Center)
		num.SetColor(zgeo.ColorWhite)
		num.SetFont(zgeo.FontNice(25, zgeo.FontStyleBold))
		blockView.Add(num, zgeo.Center).Free = true
	}
	blockView.SetBGColor(col)
	for _, lane := range v.lanes {
		if lane.hasAxis {
			axis := v.makeAxisRow(blockIndex)
			pos := zgeo.PosD(0, lane.y)
			blockView.Add(axis, zgeo.TopLeft|zgeo.HorExpand, pos.Size()).Free = true
			axis.Native().SetPos(pos)
		}
	}
	v.updateBlockView(blockIndex, blockView)
	return blockView
}

func (v *HorEventsView) updateBlockView(blockIndex int, blockView *zcontainer.ContainerView) {
	// zlog.Info("updateBlockView:", blockIndex, v.GetEventViewsFunc != nil, zdebug.CallingStackString())
	if blockView == nil {
		view, _ := v.horInfinite.FindViewForIndex(blockIndex)
		if view == nil {
			zlog.Info("updateBlockView no view:", blockIndex)
			return // we haven't created it yet, so just quit
		}
		blockView = view.(*zcontainer.ContainerView)
	}
	go v.GetEventViewsFunc(blockIndex, func(childView zview.View, x int, cellBox zgeo.Size, laneID, rowType int64) {
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
				blockView.RemoveChild(c.View, true)
				break
			}
		}
		blockView.Add(childView, zgeo.TopLeft, cellRect.Pos.Size()).Free = true
		childView.SetRect(cellRect)
	})
}

// updateCurrentBlock adds move events to the current block who's end might not be yet.
func (v *HorEventsView) updateCurrentBlock(blockIndex int) bool {
	return true // return false if we get past end
}

func (v *HorEventsView) getBlockViewForIndex(blockIndex int) zview.View {
	view := v.makeBlockView(blockIndex)
	return view
}

func (v *HorEventsView) releaseViewForIndex(blockIndex int) {
}

func (v *HorEventsView) HandleWidthChanged(w float64) {
	// zlog.Info("HandleWidthChanged:", w, v.LocalRect().Size.W)
	v.ViewWidth = w
	if v.WidthChanger != nil {
		v.WidthChanger.HandleWidthChanged(w)
	}
}

func (v *HorEventsView) HandlePan(blockIndex float64) {
	t := v.IndexToTime(blockIndex)
	v.currentTime = t
	v.timeField.SetValue(t)
	v.updateNowPole()
	v.updateWidgets()
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
