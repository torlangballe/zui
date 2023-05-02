//go:build zui

package ztext

import (
	"fmt"
	"strconv"
	"time"

	"github.com/torlangballe/zui/zcalendar"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zgeo"
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
	value              time.Time
	flags              TimeFieldFlags
	ampmLabel          **zlabel.Label
}

func TimeNew(name string, flags TimeFieldFlags) *TimeFieldView {
	v := &TimeFieldView{}
	v.flags = flags
	v.Init(v, false, name)
	v.SetSpacing(-1)
	v.SetMargin(zgeo.RectFromXY2(6, -2, -6, 0))
	v.SetCorner(6)
	v.SetStroke(1, zgeo.ColorNewGray(0.6, 1), true)
	v.SetBGColor(zgeo.ColorNewGray(0.8, 1))

	v.onChrome = (zdevice.WasmBrowser() == "chrome")
	if flags&TimeFieldDateOnly == 0 {
		v.hourText = addText(v, 2)
		v.minuteText = addText(v, 2)
		if flags&TimeFieldSecs != 0 {
			v.secondsText = addText(v, 2)
		}
		if flags&TimeFieldTimeOnly == 0 {
			spacing := zcustom.NewView("spacing")
			spacing.SetMinSize(zgeo.Size{10, 0})
			v.Add(spacing, zgeo.CenterLeft)
		}
	}
	if flags&TimeFieldTimeOnly == 0 {
		v.dayText = addText(v, 2)
		v.monthText = addText(v, 2)
		if flags&TimeFieldYears != 0 {
			cols := 2
			if flags&TimeFieldFullYear != 0 {
				cols = 4
			}
			v.yearText = addText(v, cols)
		}
		if flags&TimeFieldNoCalendar == 0 {
			label := zlabel.New("📅")
			// label.SetFont(zgeo.FontNice(-3, zgeo.FontStyleNormal))
			label.SetPressedHandler(v.popCalendar)
			v.Add(label, zgeo.CenterLeft)
		}
	}
	return v
}

func (v *TimeFieldView) popCalendar() {
	cal := zcalendar.New("")
	cal.SetTime(v.value)
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
		t := time.Date(ct.Year(), ct.Month(), ct.Day(), v.value.Hour(), v.value.Minute(), v.value.Second(), 0, v.value.Location())
		zpresent.Close(cal, true, nil)
		v.SetValue(t)
	}
	zpresent.PresentView(cal, att, func(win *zwindow.Window) {
		cal.Focus(true)
	}, func(dismissed bool) {})
}

func addText(v *TimeFieldView, columns int) *TextView {
	style := Style{KeyboardType: zkeyboard.TypeInteger}
	tv := NewView("", style, columns, 1)
	tv.UpdateSecs = 0
	// tv.SetCorners(5, zgeo.TopLeft|zgeo.Bottom)
	tv.SetZIndex(100)
	tv.SetChangedHandler(v.changed)
	tv.SetTextAlignment(zgeo.Right)
	tv.SetKeyHandler(func(km zkeyboard.KeyMod, down bool) bool {
		if km.Key.IsReturnish() && down && v.HandleValueChanged != nil {
			v.changed()
			v.HandleValueChanged(v.value)
			return true
		}
		return false
	})
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
	v.value = time.Time{}
	if v.HandleValueChanged != nil {
		v.HandleValueChanged(v.value)
	}
}

func getInt(v *TextView, i *int) bool {
	if v == nil {
		return true
	}
	n, err := strconv.Atoi(v.Text())
	if err != nil {
		return false
	}
	*i = n
	return true
}

func setInt(v *TextView, i int, format string) {
	if v == nil {
		return
	}
	str := fmt.Sprintf(format, i)
	v.SetText(str)
}

func (v *TimeFieldView) changed() {
	var hour, min, sec int
	var day, month int

	ok := true
	loc := v.value.Location()
	year := time.Now().In(loc).Year()

	ok = ok && getInt(v.hourText, &hour)
	ok = ok && getInt(v.minuteText, &min)
	getInt(v.secondsText, &sec)

	ok = ok && getInt(v.dayText, &day)
	ok = ok && getInt(v.monthText, &month)
	getInt(v.yearText, &year)

	if ok {
		t := time.Date(year, time.Month(month), day, hour, min, sec, 0, loc)
		v.SetValue(t)
	}
}

func (v *TimeFieldView) SetValue(t time.Time) {
	v.value = t
	setInt(v.hourText, t.Hour(), "%02d")
	setInt(v.minuteText, t.Minute(), "%02d")
	setInt(v.secondsText, t.Second(), "%02d")

	setInt(v.dayText, t.Day(), "%d")
	setInt(v.monthText, int(t.Month()), "%d")
	setInt(v.yearText, int(t.Year()%100), "%d")

}
