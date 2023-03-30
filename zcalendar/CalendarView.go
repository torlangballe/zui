//go:build zui

package zcalendar

import (
	"fmt"
	"time"

	"github.com/torlangballe/zui/zanimation"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
)

type Flag int

const (
	WeekNumbers Flag = 1 << iota
	SundayFirst
)

var headerColor = zgeo.ColorNew(0.8, 0.4, 0.1, 1)

type CalendarView struct {
	zcontainer.StackView
	Flags          Flag
	SelectedTime   time.Time
	HandleSelected func(t time.Time)
	current        time.Time
	daysGrid       *zcontainer.GridView
	monthLabel     *zlabel.Label
	days           map[zgeo.IPos]time.Time
	lastTrans      zgeo.Pos
	gridRect       zgeo.Rect
	navigator      zcontainer.ChildFocusNavigator
}

func New(storeName string) *CalendarView {
	v := &CalendarView{}
	v.Init(v, storeName)
	return v
}

func setFocusColorsForView(v *CalendarView, view zview.View, focused bool) {
	box := view.(*zcontainer.ContainerView)
	cell, _ := v.daysGrid.FindCellWithView(box)
	zlog.Assert(cell != nil)
	t := cell.AnyInfo.(time.Time)
	v.setColors(box, false, focused, t)
}

func (v *CalendarView) Init(view zview.View, storeName string) {
	v.StackView.Init(view, true, "calendar")
	v.SetSpacing(0)
	v.SetMinSize(zgeo.SizeBoth(100))
	v.SetStroke(1, zgeo.ColorDarkGray, false)
	v.SetAboveParent(false)
	v.SetCanFocus(true)
	v.days = map[zgeo.IPos]time.Time{}
	v.navigator.HandleSelect = func(view zview.View) {
		if v.navigator.CurrentFocused != nil {
			setFocusColorsForView(v, v.navigator.CurrentFocused, false)
		}
		setFocusColorsForView(v, view, true)
		zlog.Info("Arrow2:", view.ObjectName(), v.navigator.CurrentFocused.ObjectName())
	}
	header := zcontainer.StackViewHor("header")
	header.SetZIndex(zview.BaseZIndex + 1)
	header.SetBGColor(headerColor)
	v.Add(header, zgeo.TopCenter|zgeo.HorExpand)

	v.monthLabel = makeHeaderLabel("", zgeo.HorCenter)
	header.Add(v.monthLabel, zgeo.HorCenter|zgeo.HorExpand) // for space to month-forward

	monthAdd := makeHeaderLabel("▼", zgeo.Right)
	header.Add(monthAdd, zgeo.CenterRight, zgeo.Size{30, 0}).Free = true
	monthAdd.SetPressedHandler(func() {
		v.Increase(1, 0)
	})
	yearAdd := makeHeaderLabel("⏵⏵", zgeo.Right)
	header.Add(yearAdd, zgeo.CenterRight, zgeo.Size{4, 0}).Free = true
	yearAdd.SetPressedHandler(func() {
		v.Increase(0, 1)
	})
	monthSub := makeHeaderLabel("▲", zgeo.Left)
	header.Add(monthSub, zgeo.CenterLeft, zgeo.Size{30, 0}).Free = true
	monthSub.SetPressedHandler(func() {
		v.Increase(-1, 0)
	})
	yearSub := makeHeaderLabel("⏴⏴", zgeo.Left)
	header.Add(yearSub, zgeo.CenterLeft, zgeo.Size{4, 0}).Free = true
	yearSub.SetPressedHandler(func() {
		v.Increase(0, -1)
	})
	v.SetKeyHandler(v.navigator.HandleKey)
}

func (v *CalendarView) SetTime(t time.Time) {
	v.updateWithTime(t, zgeo.Pos{})
}

func (v *CalendarView) Increase(monthInc int, yearInc int) {
	var dir zgeo.Pos
	t := v.current.AddDate(yearInc, monthInc, 0)
	dir.Y = -float64(monthInc)
	dir.X = -float64(yearInc)
	v.updateWithTime(t, dir)
}

func makeHeaderLabel(str string, a zgeo.Alignment) *zlabel.Label {
	label := zlabel.New(str)
	label.SetColor(zgeo.ColorWhite)
	label.SetTextAlignment(a | zgeo.VertCenter)
	label.SetFont(zgeo.FontNice(zgeo.FontDefaultSize+1, zgeo.FontStyleNormal))
	return label
}

func (v *CalendarView) setColors(box *zcontainer.ContainerView, hovering, focused bool, t time.Time) {
	day := t.Day()
	month := t.Month()
	today := time.Now().In(t.Location())
	isNowMonth := (month == today.Month() && t.Year() == today.Year())
	isToday := (isNowMonth && day == today.Day())
	isSelectedDay := (day == v.SelectedTime.Day())
	// isSelectedMonth := (!v.SelectedTime.IsZero() && v.SelectedTime.Month() == month && v.SelectedTime.Year() == year)
	// outsideMonth := t.Month() != month
	bg := zstyle.DefaultBGColor()
	fg := zstyle.DefaultFGColor()
	if focused {
		box.Native().SetStroke(2, zstyle.DefaultFocusColor, false)
	} else {
		box.Native().SetStroke(0, zgeo.ColorClear, 0)
		if hovering {
			bg = zstyle.DefaultHoverColor
			fg = zgeo.ColorWhite
		} else {
			if !isNowMonth {
				fg.SetOpacity(0.4)
			} else {
				if isSelectedDay {
					fg = zstyle.DefaultBGColor()
					bg = zstyle.DefaultFGColor()
				}
				if isToday {
					bg = headerColor
					if isSelectedDay {
						bg = bg.Mixed(zstyle.DefaultFGColor(), 0.5)
					}
					fg = bg.ContrastingGray()
				}
			}
		}
	}
	box.SetBGColor(bg)
	box.SetColor(fg)
}

func addLabel(v *CalendarView, grid *zcontainer.GridView, a any) (box *zcontainer.ContainerView, cell *zcontainer.Cell) {
	str := fmt.Sprint(a)
	box = zcontainer.New(str)
	box.SetMinSize(zgeo.Size{24, 22})
	box.SetCorner(3)

	label := zlabel.New(str)
	label.SetFont(zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal))

	box.Add(label, zgeo.Center, zgeo.Size{4, 4})
	cell = grid.Add(box, zgeo.TopLeft)
	// label.SetMargin(zgeo.RectFromXY2(2, 2, -4, -2))
	return
}

func addDayLabel(v *CalendarView, grid *zcontainer.GridView, t time.Time, a any) {
	labelBox, cell := addLabel(v, grid, a)
	labelBox.SetAboveParent(true)
	cell.AnyInfo = t
	v.navigator.AddChild(labelBox)
	v.setColors(labelBox, false, false, t)
	labelBox.SetPointerEnterHandler(false, func(pos zgeo.Pos, inside zbool.BoolInd) {
		v.setColors(labelBox, inside.IsTrue(), false, t)
	})
	if v.HandleSelected != nil {
		labelBox.SetPressedHandler(func() {
			if v.SelectedTime.IsZero() {
				v.HandleSelected(t)
				return
			}
			v.SelectedTime = t
			v.updateWithTime(v.current, zgeo.Pos{})
			ztimer.StartIn(0.3, func() {
				v.HandleSelected(t)
			})
		})
	}
}

func makeDate(day int, month time.Month, year int, loc *time.Location) time.Time {
	return time.Date(year, month, day, 12, 0, 0, 0, loc)
}

func (v *CalendarView) updateWithTime(t time.Time, dir zgeo.Pos) {
	v.current = t
	loc := t.Location()

	v.navigator.Clear()
	v.navigator.Focus()
	cols := 7
	if v.Flags&WeekNumbers != 0 {
		cols += 2
	}
	grid := zcontainer.GridViewNew("days", cols)
	grid.SetAboveParent(false)
	grid.SetSpacing(zgeo.Size{0, 0})

	month := t.Month()
	year := t.Year()
	// day := t.Day()

	str := fmt.Sprintf("%s %d", month, year)
	v.monthLabel.SetText(str)

	timeOnFirst := makeDate(1, month, year, loc)
	weekDayOnFirst := timeOnFirst.Weekday()
	if v.Flags&WeekNumbers != 0 {
		weeks, _ := addLabel(v, grid, "# ")
		weeks.SetColor(zgeo.ColorGray)
		grid.Add(zlabel.New("   "), zgeo.TopLeft)
	}
	wd := ztime.Weekdays
	if v.Flags&SundayFirst != 0 {
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
		label, _ := addLabel(v, grid, d.String()[:1])
		label.SetFont(label.Font().NewWithStyle(zgeo.FontStyleBold))
	}
	day := -skips + 1
	for row := 0; row < 6; row++ {
		t := makeDate(day, month, year, loc)
		if v.Flags&WeekNumbers != 0 {
			_, weekNo := t.ISOWeek()
			wlabel, _ := addLabel(v, grid, weekNo)
			wlabel.SetColor(zgeo.ColorNew(0.5, 0.3, 0, 0.8))
			grid.Add(nil, zgeo.TopLeft)
		}
		for d := 0; d < 7; d++ {
			t := makeDate(day, month, year, loc)
			addDayLabel(v, grid, t, t.Day())
			day++
		}
	}
	if v.daysGrid == nil {
		v.Add(grid, zgeo.Center, zgeo.Size{4, 4})
	} else {
		tranformGrid(v, grid, dir)
	}
	v.daysGrid = grid
}

func tranformGrid(v *CalendarView, grid *zcontainer.GridView, dir zgeo.Pos) {
	grid.SetAlpha(0.1)
	r := v.gridRect
	v.AddChild(grid, -1) // needs to preserve index, which isn't really supported in AddChild yet anyway
	move := dir.Times(r.Size.Pos())
	grid.SetRect(r.Translated(move))
	r.Pos.Subtract(move)
	// zlog.Info("SetGrid1:", dir, r, move)
	grid.SetRect(r)
	grid.SetAlpha(1)
	zanimation.Transform(&grid.NativeView, move, 0.4, false)
	ddd := move.Plus(v.lastTrans)
	zanimation.Transform(&v.daysGrid.NativeView, ddd, 0.4, true)
	v.lastTrans = move
}

func (v *CalendarView) ArrangeChildren() {
	v.StackView.ArrangeChildren()
	if v.daysGrid != nil {
		v.gridRect = v.daysGrid.Rect()
	}
}
