//go:build zui

package ztext

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/torlangballe/zui/zcalendar"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/ztime"
)

type TimeFieldFlags int

const (
	TimeFieldNone TimeFieldFlags = 1 << iota
	TimeFieldSecs
	TimeFieldYears
	TimeFieldDateOnly
	TimeFieldTimeOnly
	TimeFieldNoCalendar
	TimeFieldShortYear
	TimeFieldStatic
	TimeFieldPreviousYear30 // If !TimeFieldYears, use previous year date is then less than 30 days from now
)

type TimeFieldView struct {
	zcontainer.StackView
	UseYear                bool
	UseSeconds             bool
	HandleValueChangedFunc func()
	PreviousYearIfLessDays int
	hourText               *TextView
	minuteText             *TextView
	secondsText            *TextView
	dayText                *TextView
	monthText              *TextView
	yearText               *TextView
	location               *time.Location
	flags                  TimeFieldFlags
	ampmLabel              *zlabel.Label
	currentUse24Clock      bool
}

func TimeFieldNew(name string, flags TimeFieldFlags) *TimeFieldView {
	v := &TimeFieldView{}
	v.flags = flags
	v.location = time.Local
	v.Init(v, false, name)
	v.SetSpacing(-8)
	v.SetMargin(zgeo.RectFromXY2(6, -4, -8, 8))
	v.SetCorner(6)
	v.SetBGColor(zgeo.ColorNewGray(0.8, 1))
	if zdevice.WasmBrowser() == zdevice.Safari { // this is a hack because on safari, first number field's focus doesn't show when in popup
		style := Style{KeyboardType: zkeyboard.TypeInteger}
		tv := NewView("", style, 1, 1)
		tv.Show(false)
		v.Add(tv, zgeo.TopLeft, zgeo.Size{-15, 0})
	}
	if flags&TimeFieldDateOnly == 0 {
		v.hourText = addText(v, 2, "H", "")
		v.minuteText = addText(v, 2, "M", ":")
		if flags&TimeFieldSecs != 0 {
			v.secondsText = addText(v, 2, "S", ":")
		}
		v.ampmLabel = zlabel.New("AM")
		v.ampmLabel.SetCanTabFocus(true)
		v.ampmLabel.View.SetObjectName("ampm")
		v.ampmLabel.SetFont(zgeo.FontNice(-2, zgeo.FontStyleBold))
		v.ampmLabel.SetColor(zgeo.ColorNewGray(0.5, 1))
		v.ampmLabel.SetPressedHandler(v.toggleAMPM)
		v.Add(v.ampmLabel, zgeo.CenterLeft, zgeo.Size{-8, 0})
		v.CollapseChild(v.ampmLabel, zlocale.IsUse24HourClock.Get(), false)
		zkeyvalue.SetOptionChangeHandler(v, func(key string) {
			changed := v.CollapseChild(v.ampmLabel, zlocale.IsUse24HourClock.Get(), false)
			hour, err := strconv.Atoi(v.hourText.Text())
			// zlog.Info("Opt changed", changed, hour, err)
			if err == nil {
				var pm bool
				if v.currentUse24Clock != zlocale.IsUse24HourClock.Get() {
					if v.currentUse24Clock {
						hour, pm = convertFrom24Hour(v, hour)
						setInt(v.hourText, hour, "%02d")
					} else {
						hour, pm = get24Hour(v, hour)
						setInt(v.hourText, hour, "%d")
					}
				}
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
			spacing.SetMinSize(zgeo.Size{23, 6})
			v.Add(spacing, zgeo.CenterLeft)
		}
	}
	if flags&TimeFieldTimeOnly == 0 {
		v.dayText = addText(v, 2, "D", "")
		v.monthText = addText(v, 2, "M", "/")
		v.monthText.SetColor(zgeo.ColorNew(0, 0, 0.8, 1))
		if flags&TimeFieldYears != 0 {
			cols := 4
			if flags&TimeFieldShortYear != 0 {
				cols = 2
			}
			v.yearText = addText(v, cols, "Y", "/")
		}
		if flags&TimeFieldNoCalendar == 0 {
			cal := zimageview.New(nil, "images/zcore/calendar.png", zgeo.Size{16, 16})
			cal.SetPressedHandler(v.popCalendar)
			v.Add(cal, zgeo.CenterLeft, zgeo.Size{1, 0})
		}
	}
	flipDayMonth(v, false)
	return v
}

func (v *TimeFieldView) handleReturn(km zkeyboard.KeyMod, down bool) bool {
	if km.Key.IsReturnish() && km.Modifier == 0 && down && v.HandleValueChangedFunc != nil {
		// zlog.Info("HER KEY1?", km.Key, km.Key.IsReturnish(), km.Modifier, down, v.HandleValueChangedFunc, err)
		// zlog.Info("HER KEY2?", km.Key, km.Key.IsReturnish(), km.Modifier, down, v.HandleValueChangedFunc, err)
		v.HandleValueChangedFunc()
		return true
	}
	return false
}

func addText(v *TimeFieldView, columns int, placeholder string, pre string) *TextView {
	style := Style{KeyboardType: zkeyboard.TypeInteger}
	tv := NewView("", style, columns, 1)
	tv.UpdateSecs = 0
	tv.SetPlaceholder(placeholder)
	tv.SetZIndex(zview.BaseZIndex)
	tv.SetFocusHandler(func(focused bool) {
		index := zview.BaseZIndex
		if focused {
			index = zview.BaseZIndex + 2
		}
		tv.SetZIndex(index)
	})
	tv.SetChangedHandler(func() {
		clearColorTexts(v.hourText, v.minuteText, v.secondsText, v.dayText, v.monthText, v.yearText)
		v.Value() // getting value will set error color
	})
	tv.SetTextAlignment(zgeo.Right)
	v.Add(tv, zgeo.TopLeft, zgeo.Size{-2, 2})
	tv.SetKeyHandler(v.handleReturn)
	return tv
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
	cal.SetValue(val)
	cal.HandleValueChangedFunc = func() {
		ct := cal.Value()
		t := time.Date(ct.Year(), ct.Month(), ct.Day(), val.Hour(), val.Minute(), val.Second(), 0, v.location)
		zpresent.Close(cal, true, nil)
		v.SetValue(t)
	}
	cal.JSSet("className", "znofocus")

	zpresent.PopupView(cal, v, zgeo.TopRight|zgeo.HorOut, zgeo.Size{-8, -4})
}

func clearColorTexts(texts ...*TextView) {
	for _, t := range texts {
		if t != nil {
			t.SetBGColor(DefaultBGColor())
		}
	}
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

func getInt(v *TextView, i *int, min, max int, err *error, ignoreEmpty bool) {
	if v == nil {
		return
	}
	text := v.Text()
	if ignoreEmpty && text == "" {
		return
	}
	n, cerr := strconv.Atoi(text)
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
	if v.flags&TimeFieldShortYear != 0 {
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

	var err error
	now := time.Now().In(v.location)
	month := int(now.Month())
	year := now.Year()
	day := now.Day()

	maxHour := 12
	minHour := 1
	if v.currentUse24Clock {
		maxHour = 23
		minHour = 0
	}
	getInt(v.hourText, &hour, minHour, maxHour, &err, false)
	hour, _ = get24Hour(v, hour)
	getInt(v.minuteText, &min, 0, 60, &err, false)
	getInt(v.secondsText, &sec, 0, 60, &err, false)
	getInt(v.monthText, &month, 1, 12, &err, true)
	days := ztime.DaysInMonth(time.Month(month), year)
	if v.flags&TimeFieldYears != 0 {
		getInt(v.yearText, &year, 0, 0, &err, true)
		if year < 100 {
			year += 2000
		}
	}
	getInt(v.dayText, &day, 1, days, &err, true)
	v.currentUse24Clock = zlocale.IsUse24HourClock.Get()
	if err != nil {
		return time.Time{}, err
	}
	t := time.Date(year, time.Month(month), day, hour, min, sec, 0, v.location)
	if v.flags&TimeFieldYears == 0 && v.flags&TimeFieldPreviousYear30 != 0 {
		prev := time.Date(year-1, time.Month(month), day, hour, min, sec, 0, v.location)
		psince := time.Since(prev)
		if psince < ztime.Day*30 && math.Abs(ztime.DurSeconds(psince)) < math.Abs(ztime.Since(t)) {
			t = prev
		}
	}
	// zlog.Info("VAL:", t.Year())
	return t, nil
}
