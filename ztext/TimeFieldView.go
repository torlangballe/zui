//go:build zui

package ztext

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/torlangballe/zui/zcalendar"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/ztime"
)

// https://www.npmjs.com/package/js-datepicker

type TimeFieldFlags int

const (
	TimeFieldSecs TimeFieldFlags = 1 << iota
	TimeFieldYears
	TimeFieldDateOnly
	TimeFieldTimeOnly
	TimeFieldNoCalendar
	TimeFieldFullYear
)

type TimeFieldView struct {
	zcontainer.StackView
	hourText           *TextView
	minuteText         *TextView
	secondsText        *TextView
	dayText            *TextView
	monthText          *TextView
	yearText           *TextView
	UseYear            bool
	UseSeconds         bool
	HandleValueChanged func(t time.Time)
	onChrome           bool
	location           *time.Location
	flags              TimeFieldFlags
	ampmLabel          *zlabel.Label
	currentUse24Clock  bool
}

func TimeNew(name string, flags TimeFieldFlags) *TimeFieldView {
	v := &TimeFieldView{}
	v.flags = flags
	v.Init(v, false, name)
	v.SetSpacing(-1)
	v.SetMargin(zgeo.RectFromXY2(6, -2, -6, 0))
	v.SetCorner(6)
	// v.SetStroke(1, zgeo.ColorNewGray(0.6, 1), true)
	v.SetBGColor(zgeo.ColorNewGray(0.8, 1))

	v.onChrome = (zdevice.WasmBrowser() == "chrome")
	if flags&TimeFieldDateOnly == 0 {
		v.hourText = addText(v, 2, "H")
		v.minuteText = addText(v, 2, "M")
		if flags&TimeFieldSecs != 0 {
			v.secondsText = addText(v, 2, "S")
		}
		v.ampmLabel = zlabel.New("AM")
		v.ampmLabel.SetCanFocus(zview.FocusAllowTab)
		v.ampmLabel.View.SetObjectName("ampm")
		v.ampmLabel.SetFont(zgeo.FontNice(-2, zgeo.FontStyleBold))
		v.ampmLabel.SetColor(zgeo.ColorNewGray(0.5, 1))
		v.ampmLabel.SetPressedHandler(v.toggleAMPM)
		v.Add(v.ampmLabel, zgeo.CenterLeft)
		v.CollapseChild(v.ampmLabel, zlocale.IsUse24HourClock.Get(), false)
		zkeyvalue.SetOptionChangeHandler(v, func(key string) {
			changed := v.CollapseChild(v.ampmLabel, zlocale.IsUse24HourClock.Get(), false)
			hour, err := strconv.Atoi(v.hourText.Text())
			zlog.Info("Opt changed", changed, hour, err)
			if err == nil {
				var pm bool
				if v.currentUse24Clock != zlocale.IsUse24HourClock.Get() {
					if v.currentUse24Clock {
						hour, pm = convertFrom24Hour(v, hour)
					} else {
						hour, pm = get24Hour(v, hour)
					}
				}
				v.hourText.SetText(strconv.Itoa(hour))
				setPM(v, pm)
			}
			flipDayMonth(v, true)
			v.currentUse24Clock = zlocale.IsUse24HourClock.Get()
			if changed {
				zcontainer.ArrangeChildrenAtRootContainer(v)
			}
		})
		v.AddOnRemoveFunc(func() { zkeyvalue.SetOptionChangeHandler(v, nil) })
		if flags&TimeFieldTimeOnly == 0 {
			spacing := zcustom.NewView("spacing")
			spacing.SetMinSize(zgeo.Size{10, 0})
			v.Add(spacing, zgeo.CenterLeft)
		}
	}
	if flags&TimeFieldTimeOnly == 0 {
		v.dayText = addText(v, 2, "D")
		v.monthText = addText(v, 2, "M")
		v.monthText.SetColor(zgeo.ColorNew(0, 0, 0.8, 1))
		if flags&TimeFieldYears != 0 {
			cols := 2
			if flags&TimeFieldFullYear != 0 {
				cols = 4
			}
			v.yearText = addText(v, cols, "Y")
		}
		if flags&TimeFieldNoCalendar == 0 {
			label := zlabel.New("📅")
			// label.SetFont(zgeo.FontNice(-3, zgeo.FontStyleNormal))
			label.SetPressedHandler(v.popCalendar)
			v.Add(label, zgeo.CenterLeft)
		}
	}
	flipDayMonth(v, false)
	return v
}

func convertFrom24Hour(v *TimeFieldView, hour int) (int, bool) {
	switch hour {
	case 0:
		return 12, false
	case 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
		return hour, false
	case 12:
		return hour, true
	default:
		return hour - 12, true
	}
}

func setPM(v *TimeFieldView, pm bool) {
	set := "AM"
	if pm {
		set = "PM"
	}
	v.ampmLabel.SetText(set)
}

func flipDayMonth(v *TimeFieldView, arrange bool) {
	if v.flags&TimeFieldTimeOnly != 0 {
		return
	}
	_, di := v.FindCellWithView(v.dayText)
	_, mi := v.FindCellWithView(v.monthText)
	if mi < di != zlocale.IsShowMonthBeforeDay.Get() {
		// zlog.Info("FLIP!", di, mi, zlocale.IsShowMonthBeforeDay.Get(), arrange)
		min, max := zint.MinMax(di, mi)
		v.Cells[max].View.Native().InsertBefore(v.Cells[min].View.Native())
		zslice.Swap(v.Cells, di, mi)
		if arrange {
			v.ArrangeChildren()
		}
	}
}

func (v *TimeFieldView) toggleAMPM() {
	pm := (v.ampmLabel.Text() == "PM")
	setPM(v, !pm)
}

func (v *TimeFieldView) popCalendar() {
	cal := zcalendar.New("")
	val, err := v.Value()
	if err != nil {
		return
	}
	cal.SetTime(val)
	cal.JSSet("className", "znofocus")
	att := zpresent.AttributesNew()
	att.Modal = true
	att.ModalDimBackground = false
	att.ModalCloseOnOutsidePress = true
	att.ModalDropShadow.Delta = zgeo.SizeBoth(1)
	att.ModalDropShadow.Blur = 2
	att.ModalDismissOnEscapeKey = true
	pos := v.AbsoluteRect().Pos
	pos.X += v.Rect().Size.W - 20
	att.Pos = &pos
	cal.HandleValueChanged = func() {
		ct := cal.Value()
		t := time.Date(ct.Year(), ct.Month(), ct.Day(), val.Hour(), val.Minute(), val.Second(), 0, v.location)
		zpresent.Close(cal, true, nil)
		v.SetValue(t)
	}
	zpresent.PresentView(cal, att, func(win *zwindow.Window) {
		cal.Focus(true)
	}, func(dismissed bool) {})
}

func clearColorTexts(texts ...*TextView) {
	for _, t := range texts {
		if t != nil {
			t.SetBGColor(DefaultBGColor())
		}
	}
}

func addText(v *TimeFieldView, columns int, placeholder string) *TextView {
	style := Style{KeyboardType: zkeyboard.TypeInteger}
	tv := NewView("", style, columns, 1)
	tv.UpdateSecs = 0
	tv.SetMargin(zgeo.RectFromXY2(0, 2, -10, -9))
	tv.SetPlaceholder(placeholder)
	tv.SetZIndex(100)
	tv.SetChangedHandler(func() {
		clearColorTexts(v.hourText, v.minuteText, v.secondsText, v.dayText, v.monthText, v.yearText)
		v.Value() // getting value will set error color
	})
	tv.SetTextAlignment(zgeo.Right)
	tv.SetJSStyle("className", "znofocus")
	v.Add(tv, zgeo.CenterLeft)
	return tv
}

func clearField(v *TextView) {
	if v != nil {
		v.SetText("")
	}
}

func (v *TimeFieldView) Clear() {
	clearField(v.hourText)
	clearField(v.minuteText)
	clearField(v.secondsText)
	clearField(v.dayText)
	clearField(v.monthText)
	clearField(v.yearText)
	v.location = nil
}

func getInt(v *TextView, i *int, min, max int, err *error) {
	if v == nil {
		return
	}
	n, cerr := strconv.Atoi(v.Text())
	if cerr != nil {
		*err = cerr
		return
	}
	pink := zgeo.ColorNew(1, 0.7, 0.7, 1)
	// zlog.Info("getInt:", v.ObjectName(), n, max)
	if max != 0 && n > max {
		// zlog.Info("MAX>", v.ObjectName(), n, max)
		v.SetBGColor(pink)
		*err = errors.New("big")
		return
	}
	if n < min {
		*err = errors.New("small")
		v.SetBGColor(pink)
		return
	}
	*i = n
}

func setInt(v *TextView, i int, format string) {
	if v == nil {
		return
	}
	str := fmt.Sprintf(format, i)
	v.SetText(str)
}

func (v *TimeFieldView) SetValue(t time.Time) {
	v.location = t.Location()
	v.currentUse24Clock = zlocale.IsUse24HourClock.Get()
	if v.ampmLabel != nil && !v.currentUse24Clock {
		hour, am := ztime.GetHourAndAM(t)
		setInt(v.hourText, hour, "%d")
		setPM(v, !am)
	} else {
		setInt(v.hourText, t.Hour(), "%02d")
	}
	setInt(v.minuteText, t.Minute(), "%02d")
	setInt(v.secondsText, t.Second(), "%02d")

	setInt(v.dayText, t.Day(), "%d")
	setInt(v.monthText, int(t.Month()), "%d")
	year := t.Year()
	if v.flags&TimeFieldFullYear == 0 {
		year %= 100

	}
	setInt(v.yearText, year, "%d")
}

func get24Hour(v *TimeFieldView, hour int) (h int, pm bool) {
	if v.ampmLabel != nil {
		pm = v.ampmLabel.Text() == "PM"
		if !v.currentUse24Clock {
			if pm {
				if hour != 12 {
					hour += 12
				}
			} else if hour == 12 {
				hour = 0
			}
		}
	}
	return hour, pm
}

func (v *TimeFieldView) Value() (time.Time, error) {
	var hour, min, sec int
	var day int

	var err error
	now := time.Now().In(v.location)
	month := int(now.Month())
	year := now.Year()

	maxHour := 12
	minHour := 1
	if v.currentUse24Clock {
		maxHour = 23
		minHour = 0
	}
	getInt(v.hourText, &hour, minHour, maxHour, &err)
	hour, _ = get24Hour(v, hour)
	getInt(v.minuteText, &min, 0, 60, &err)
	getInt(v.secondsText, &sec, 0, 60, &err)
	getInt(v.monthText, &month, 1, 12, &err)
	getInt(v.yearText, &year, 0, 0, &err)
	days := ztime.DaysInMonth(time.Month(month), year)
	getInt(v.dayText, &day, 1, days, &err)
	v.currentUse24Clock = zlocale.IsUse24HourClock.Get()
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(year, time.Month(month), day, hour, min, sec, 0, v.location), nil
}
