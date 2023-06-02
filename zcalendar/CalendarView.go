//go:build zui

package zcalendar

import (
	"fmt"
	"strconv"
	"time"

	"github.com/torlangballe/zui/zanimation"
	"github.com/torlangballe/zui/zcheckbox"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zwindow"

	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
)

var HeaderColor = zgeo.ColorNew(0.8, 0.4, 0.1, 1)

type CalendarView struct {
	zcontainer.StackView
	value                  time.Time
	HandleValueChangedFunc func()
	currentShowing         time.Time
	daysGrid               *zcontainer.GridView
	monthLabel             *zlabel.Label
	days                   map[zgeo.IPos]time.Time
	daysSlider             zanimation.Swapper
	settingsSlider         zanimation.Swapper
	lastTrans              zgeo.Pos
	navigator              zcontainer.ChildFocusNavigator
	header                 *zcontainer.StackView
	settingsGear           *zimageview.ImageView
	updateAfterSettings    bool
	firstShow              bool
}

func New(storeName string) *CalendarView {
	v := &CalendarView{}
	v.Init(v)
	return v
}

func (v *CalendarView) Init(view zview.View) {
	v.StackView.Init(view, true, "calendar")
	v.firstShow = true
	v.SetSpacing(0)
	v.SetMinSize(zgeo.SizeBoth(100))
	v.SetStroke(1, zgeo.ColorDarkGray, false)
	v.SetCanTabFocus(true)
	v.SetBGColor(zgeo.ColorWhite)
	v.days = map[zgeo.IPos]time.Time{}
	v.navigator.HandleSelect = v.handleArrowMove
	v.header = zcontainer.StackViewHor("header")
	v.header.SetMargin(zgeo.RectFromXY2(0, 0, 0, -1))
	v.header.SetZIndex(zview.BaseZIndex + 1)
	v.header.SetBGColor(HeaderColor)
	v.Add(v.header, zgeo.TopCenter|zgeo.HorExpand)

	v.monthLabel = makeHeaderLabel("", zgeo.HorCenter)
	v.monthLabel.SetObjectName("header-title")
	v.header.Add(v.monthLabel, zgeo.HorCenter|zgeo.HorExpand) // for space to month-forward
	v.monthLabel.SetPressedHandler(func() {
		t := time.Now()
		zlog.Info("pressed", zkeyboard.ModifiersAtPress, v.value)
		if zkeyboard.ModifiersAtPress == zkeyboard.ModifierShift && !v.value.IsZero() {
			t = v.value
		}
		v.updateShowMonth(t, zgeo.AlignmentNone)
	})
	v.monthLabel.SetToolTip("Press to go to today. Shift press to go to selected month.")
	monthAdd := makeHeaderLabel("▼", zgeo.Right)
	// monthAdd.SetObjectName("month-add")
	v.header.Add(monthAdd, zgeo.CenterRight, zgeo.Size{30, 0}).Free = true
	monthAdd.KeyboardShortcut = zkeyboard.KMod(zkeyboard.KeyDownArrow, zkeyboard.ModifierShift)
	monthAdd.SetPressedHandler(func() {
		v.Increase(1, 0)
	})
	yearAdd := makeHeaderLabel("⏵⏵", zgeo.Right)
	yearAdd.KeyboardShortcut = zkeyboard.KMod(zkeyboard.KeyRightArrow, zkeyboard.ModifierShift)
	v.header.Add(yearAdd, zgeo.CenterRight, zgeo.Size{4, 0}).Free = true
	yearAdd.SetPressedHandler(func() {
		v.Increase(0, 1)
	})
	monthSub := makeHeaderLabel("▲", zgeo.Left)
	monthSub.KeyboardShortcut = zkeyboard.KMod(zkeyboard.KeyUpArrow, zkeyboard.ModifierShift)
	v.header.Add(monthSub, zgeo.CenterLeft, zgeo.Size{30, 0}).Free = true
	monthSub.SetPressedHandler(func() {
		v.Increase(-1, 0)
	})
	yearSub := makeHeaderLabel("⏴⏴", zgeo.Left)
	yearSub.KeyboardShortcut = zkeyboard.KMod(zkeyboard.KeyLeftArrow, zkeyboard.ModifierShift)
	v.header.Add(yearSub, zgeo.CenterLeft, zgeo.Size{4, 0}).Free = true
	yearSub.SetPressedHandler(func() {
		v.Increase(0, -1)
	})
	v.settingsGear = zimageview.New(nil, "images/zcore/gear.png", zgeo.Size{18, 18})
	v.settingsGear.SetZIndex(zview.BaseZIndex + 2)
	v.settingsGear.SetAlpha(0)
	v.settingsGear.SetPressedHandler(v.handleSettingsPressed)
	v.Add(v.settingsGear, zgeo.BottomRight, zgeo.Size{6, 6}).Free = true

	v.SetKeyHandler(func(km zkeyboard.KeyMod, down bool) bool {
		if !down {
			return false
		}
		if km.Key.IsReturnish() {
			if v.navigator.CurrentFocused != nil {
				cell, _ := v.daysGrid.FindCellWithView(v.navigator.CurrentFocused)
				t := cell.AnyInfo.(time.Time)
				handleSelect(v, t)
				return true
			}
		}
		if v.navigator.HandleKey(km, down) {
			return true
		}
		return zcontainer.HandleOutsideShortcutRecursively(v, km)
	})
	zwindow.FromNativeView(&v.NativeView).AddKeypressHandler(v, func(km zkeyboard.KeyMod, down bool) bool {
		if km.Key == zkeyboard.KeyCommand {
			showSettings(v, v.settingsGear, down)
		}
		return false
	})
}

func (v *CalendarView) CalculatedSize(total zgeo.Size) zgeo.Size {
	marg := zgeo.Size{6, 6}
	s := v.StackView.CalculatedSize(total)
	return s.Plus(marg)
}

func (v *CalendarView) SetRect(r zgeo.Rect) {
	v.StackView.SetRect(r.ExpandedD(-3))
}

func (v *CalendarView) handleArrowMove(view zview.View, dir zgeo.Alignment) {
	if view != nil {
		moveFocus(v, view)
		return
	}
	if dir == zgeo.AlignmentNone {
		return
	}
	if dir&zgeo.Horizontal == 0 {
		pos := dir.Vector().Swapped()
		v.Increase(int(pos.X), int(pos.Y))
		return
	}
	x, y, got := v.daysGrid.GetViewXY(v.navigator.CurrentFocused)
	if !got {
		return
	}
	if dir == zgeo.Left {
		y--
	} else {
		y++
	}
	if y < 0 || y >= v.daysGrid.RowCount() {
		return
	}
	var focus zview.View
	for x = 0; x < v.daysGrid.Columns; x++ {
		cell := v.daysGrid.GetCell(x, y)
		if cell == nil {
			continue
		}
		if !v.navigator.HasChild(cell.View) {
			continue
		}
		if dir == zgeo.Right {
			moveFocus(v, cell.View)
			return
		}
		if dir == zgeo.Left {
			focus = cell.View
		}
	}
	if focus != nil {
		moveFocus(v, focus)
	}
}

func moveFocus(v *CalendarView, focusView zview.View) {
	if v.navigator.CurrentFocused != nil {
		cur := v.navigator.CurrentFocused
		v.navigator.CurrentFocused = nil
		setColorsForView(v, cur)
	}
	v.navigator.CurrentFocused = focusView
	setColorsForView(v, focusView)
}

func showSettings(v *CalendarView, settings *zimageview.ImageView, show bool) {
	alpha := 0.0
	if show {
		alpha = 1
	}
	zanimation.SetAlpha(settings, alpha, 0.4, func() {
		//		settings.Show(show)
	})
	bottom := v.daysGrid.GetCell(v.daysGrid.Columns-1, v.daysGrid.RowCount()-1).View
	bottom.Show(!show)
}

func (v *CalendarView) addSettingsCheck(s *zcontainer.StackView, title string, option *zkeyvalue.Option[bool], updateAfter bool) *zcheckbox.CheckBox {
	check, _, stack := zcheckbox.NewWithLabel(false, title, "")
	s.Add(stack, zgeo.CenterLeft)
	if option != nil {
		check.SetOn(option.Get())
		check.SetValueHandler(func() {
			v.updateAfterSettings = updateAfter
			option.Set(check.On())
		})
	}
	return check
}

func (v *CalendarView) handleSettingsPressed() {
	// secs := 0.6
	v.settingsGear.Show(false)
	// zlog.Info("handleSettingsPressed")
	s := zcontainer.StackViewVert("v1")
	s.SetMargin(zgeo.RectFromXY2(10, 10, -10, -10))
	// s.SetTilePath("images/tile.png")
	s.SetBGColor(zgeo.ColorNewGray(0.85, 0.9))
	v.daysGrid.SetJSStyle("filter", "blur(5px)")
	v.header.SetUsable(false)
	//!!!	backdrop-filter: url(filters.svg#filter) blur(4px) saturate(150%);
	// https://developer.mozilla.org/en-US/docs/Web/CSS/backdrop-filter

	v.addSettingsCheck(s, "Week Starts on Monday", &zlocale.IsMondayFirstInWeek, true)
	v.addSettingsCheck(s, "Show Week Numbers", &zlocale.IsShowWeekNumbersInCalendars, true)
	v.addSettingsCheck(s, "Use 24-hour Clock", &zlocale.IsUse24HourClock, false)
	v.addSettingsCheck(s, "Show Month before Day", &zlocale.IsShowMonthBeforeDay, false)
	close := zimageview.New(nil, "images/zcore/cross-circled.png", zgeo.Size{20, 20})
	s.Add(close, zgeo.BottomRight, zgeo.Size{4, 4})
	close.SetPressedHandler(func() {
		v.daysGrid.SetJSStyle("filter", "none")
		zanimation.Translate(s, zgeo.Pos{0, -v.settingsSlider.OriginalRect.Size.H}, 0.5, func() {
			v.settingsGear.Show(true)
			v.RemoveChild(s)
			v.header.SetUsable(true)
			if v.updateAfterSettings {
				v.updateAfterSettings = false
				v.updateShowMonth(v.currentShowing, zgeo.AlignmentNone)
			}
		})
	})
	// ztimer.StartIn(secs/2, func() {
	// 	v.daysGrid.SetJSStyle("filter", "blur(5px) brightness(150%)")
	// })
	v.settingsSlider.SlideViewInOverOld(v, v.daysGrid, s, zgeo.Bottom, 0.5, nil)
}

func (v *CalendarView) Value() time.Time {
	return v.value
}

func (v *CalendarView) SetValue(t time.Time) {
	v.value = makeDate(t.Day(), t.Month(), t.Year(), t.Location())
	v.updateShowMonth(v.value, zgeo.AlignmentNone)
}

func (v *CalendarView) Increase(monthInc int, yearInc int) {
	var dir zgeo.Alignment
	var x, y int
	var gotPos bool

	cur := v.navigator.CurrentFocused
	if cur != nil {
		x, y, gotPos = v.daysGrid.GetViewXY(cur)
		v.navigator.CurrentFocused = nil
		setColorsForView(v, cur)
	}
	t := ztime.AddMonthAndYearToTime(v.currentShowing, monthInc, yearInc)
	switch {
	case monthInc == 1:
		dir = zgeo.Top
		y = 1
	case monthInc == -1:
		dir = zgeo.Bottom
		y = v.daysGrid.RowCount() - 1
	case yearInc == 1:
		dir = zgeo.Left
		x = 2
	case yearInc == -1:
		dir = zgeo.Right
		x = v.daysGrid.Columns - 1
	}
	v.updateShowMonth(t, dir)
	if gotPos {
		cell := v.daysGrid.GetCell(x, y)
		zlog.Assert(cell != nil)
		v.navigator.CurrentFocused = cell.View
		setColorsForView(v, v.navigator.CurrentFocused)
	}
}

func makeHeaderLabel(str string, a zgeo.Alignment) *zlabel.Label {
	label := zlabel.New(str)
	label.SetColor(zgeo.ColorWhite)
	label.SetTextAlignment(a | zgeo.VertCenter)
	label.SetFont(zgeo.FontNice(zgeo.FontDefaultSize+1, zgeo.FontStyleNormal))
	return label
}

func setColorsForView(v *CalendarView, view zview.View) {
	box := view.(*zcontainer.ContainerView)
	cell, _ := v.daysGrid.FindCellWithView(box)
	// zlog.Info("setColorsForView", view != nil, cell != nil, box != nil)
	if cell != nil {
		zlog.Assert(cell != nil && cell.View != nil)
		zlog.Assert(cell.AnyInfo != nil, view.ObjectName())
		t := cell.AnyInfo.(time.Time)
		v.setColors(box, box.GetChildren(true)[0].(*zlabel.Label), t)
	}
}

func (v *CalendarView) setColors(box *zcontainer.ContainerView, label *zlabel.Label, t time.Time) {
	year := t.Year()
	month := t.Month()
	day := t.Day()
	showingMonth := v.currentShowing.Month()
	today := time.Now().In(t.Location())
	isToday := (year == today.Year() && month == today.Month() && day == today.Day())
	isSelectedDay := (year == v.value.Year() && month == v.value.Month() && day == v.value.Day())
	bg := zgeo.ColorClear
	fg := zstyle.DefaultFGColor()
	width := 0.0
	// zlog.Info("setColors", zlog.Pointer(v.navigator.CurrentFocused), zlog.Pointer(box), label.Text())
	if box == v.navigator.CurrentFocused {
		width = 4
	}
	box.Native().SetStroke(width, zstyle.DefaultFocusColor, true)
	// box.Native().SetJSStyle("boxShadow", "0px 0px 0px 4px #4AA inset")

	if isSelectedDay {
		fg = zstyle.DefaultBGColor()
		bg = zstyle.DefaultFGColor()
	}
	if isToday {
		fg = zstyle.DefaultBGColor()
		bg = HeaderColor
		if isSelectedDay {
			bg = bg.Mixed(zstyle.DefaultFGColor(), 0.5)
		}
	}
	if t.Month() != showingMonth {
		// bg.SetOpacity(0.4)
		fg.SetOpacity(0.4)
	}
	box.SetBGColor(bg)
	label.SetColor(fg)
}

func addLabel(v *CalendarView, grid *zcontainer.GridView, a any) (label *zlabel.Label, box *zcontainer.ContainerView, cell *zcontainer.Cell) {
	str := fmt.Sprint(a)
	box = zcontainer.New(str)
	box.SetMinSize(zgeo.Size{28, 26})
	box.SetMargin(zgeo.RectFromXY2(2, 4, -6, -2))
	box.SetCorner(3)

	label = zlabel.New(str)
	label.SetFont(zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal))
	label.SetTextAlignment(zgeo.CenterRight)

	box.Add(label, zgeo.CenterLeft|zgeo.Expand)
	box.SetAboveParent(true)
	cell = grid.Add(box, zgeo.TopLeft)
	// label.SetMargin(zgeo.RectFromXY2(2, 2, -4, -2))
	return
}

func addDayLabel(v *CalendarView, grid *zcontainer.GridView, t time.Time, a any) {
	label, box, cell := addLabel(v, grid, a)
	label.Native().SetInteractive(false)
	cell.AnyInfo = t
	v.navigator.AddChild(box)
	v.setColors(box, label, t)
	box.SetPointerEnterHandler(false, func(pos zgeo.Pos, inside zbool.BoolInd) {
		cur := v.navigator.CurrentFocused
		if inside.IsTrue() {
			v.navigator.CurrentFocused = box
		} else {
			v.navigator.CurrentFocused = nil
		}
		if cur != nil {
			setColorsForView(v, cur)
		}
		v.setColors(box, label, t)
	})
	box.SetPressedDownHandler(func() {
		if !v.CanTabFocus() || v.IsFocused() {
			handleSelect(v, t)
		}
		v.navigator.CurrentFocused = box
	})
}

func handleSelect(v *CalendarView, t time.Time) {
	oldVal := v.value
	v.value = t
	if !oldVal.IsZero() {
		for _, cell := range v.daysGrid.Cells {
			if cell.View != nil && cell.AnyInfo != nil {
			}
			if cell.AnyInfo != nil && cell.AnyInfo.(time.Time) == oldVal {
				setColorsForView(v, cell.View)
				break
			}
		}
	}
	setColorsForView(v, v.navigator.CurrentFocused)
	if v.HandleValueChangedFunc != nil {
		ztimer.StartIn(0.3, func() {
			v.HandleValueChangedFunc()
		})
	}
}

func makeDate(day int, month time.Month, year int, loc *time.Location) time.Time {
	return time.Date(year, month, day, 12, 0, 0, 0, loc)
}

func (v *CalendarView) updateShowMonth(t time.Time, dir zgeo.Alignment) {
	v.currentShowing = t
	loc := t.Location()

	v.navigator.Clear()
	cols := 9
	grid := zcontainer.GridViewNew("days", cols)
	grid.SetMargin(zgeo.RectFromXY2(3, 3, -3, -3))
	grid.Spacing = zgeo.Size{0, 0}

	month := t.Month()
	year := t.Year()

	str := fmt.Sprintf("%s %d", month, year)
	v.monthLabel.SetText(str)

	timeOnFirst := makeDate(1, month, year, loc)
	weekDayOnFirst := timeOnFirst.Weekday()
	str = ""
	if zlocale.IsShowWeekNumbersInCalendars.Get() {
		str = "# "
	}
	_, weeks, _ := addLabel(v, grid, str)
	weeks.SetColor(zgeo.ColorGray)
	grid.Add(zlabel.New("   "), zgeo.TopLeft)
	wd := ztime.Weekdays
	if !zlocale.IsMondayFirstInWeek.Get() {
		wd = ztime.SundayFirstWeekdays
	}
	var started bool
	var skips int
	for _, d := range wd {
		// zlog.Info("WF:", d, weekDayOnFirst)
		if d == weekDayOnFirst {
			started = true
		}
		if !started {
			skips++
		}
		label, _, _ := addLabel(v, grid, d.String()[:1])
		label.SetFont(label.Font().NewWithStyle(zgeo.FontStyleBold))
	}
	day := -skips + 1
	for row := 0; row < 6; row++ {
		t := makeDate(day, month, year, loc)
		var weekNo string
		if zlocale.IsShowWeekNumbersInCalendars.Get() {
			_, wn := t.ISOWeek()
			weekNo = strconv.Itoa(wn)
		}
		wlabel, _, _ := addLabel(v, grid, weekNo)
		wlabel.SetColor(zgeo.ColorNew(0, 0, 0.5, 0.4))
		wlabel.SetFont(wlabel.Font().NewWithSize(-2))
		grid.Add(nil, zgeo.TopLeft)
		for d := 0; d < 7; d++ {
			t := makeDate(day, month, year, loc)
			addDayLabel(v, grid, t, t.Day())
			day++
		}
	}
	if v.daysGrid == nil {
		v.Add(grid, zgeo.Center)
	} else {
		old := v.daysGrid
		v.daysSlider.SlideViewInOldOut(v, v.daysGrid, grid, dir, 0.4, func() {
			v.RemoveChild(old)
		})
		//		tranformGrid(v, grid, dir)
	}
	v.daysGrid = grid
	if v.firstShow {
		v.navigator.FocusNext()
		setColorsForView(v, v.navigator.CurrentFocused)
		v.firstShow = false
	}
}

func (v *CalendarView) ArrangeChildren() {
	v.StackView.ArrangeChildren()
	if v.daysGrid != nil {
		v.daysSlider.OriginalRect = v.daysGrid.Rect()
		v.settingsSlider.OriginalRect = v.daysGrid.Rect()
	}
}
