//go:build zui

package ztext

import (
	"fmt"
	"strconv"
	"time"

	"github.com/torlangballe/zui/zcalendar"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
)

type TimeFieldView struct {
	zcontainer.StackView
	UseYear                 bool
	UseSeconds              bool
	CallChangedOnTabPressed bool
	HandleValueChangedFunc  func()
	PreviousYearIfLessDays  int
	hourText                *TextView
	minuteText              *TextView
	secondsText             *TextView
	dayText                 *TextView
	monthText               *TextView
	yearText                *TextView
	calendar                *zimageview.ImageView
	location                *time.Location
	flags                   ztime.TimeFieldFlags
	ampmLabel               *zlabel.Label
	currentUse24Clock       bool
}

var (
	PopupViewFunc func(view, on zview.View)
	CloseViewFunc func(view zview.View, dismiss bool)
)

func TimeFieldNew(name string, flags ztime.TimeFieldFlags) *TimeFieldView {
	v := &TimeFieldView{}
	v.flags = flags
	v.location = time.Local
	v.Init(v, false, name)
	v.SetSpacing(0)
	v.SetMinSize(zgeo.SizeD(20, 30))
	v.SetMargin(zgeo.RectFromXY2(4, 2, -2, -2))
	v.SetCorner(6)
	v.SetBGColor(zgeo.ColorNewGray(0.7, 1))
	v.SetSearchable(false)
	if false && zdevice.WasmBrowser() == zdevice.Safari { // this is a hack because on safari, first number field's focus doesn't show when in popup
		style := Style{KeyboardType: zkeyboard.TypeInteger}
		tv := NewView("", style, 1, 1)
		tv.Show(false)
		v.Add(tv, zgeo.TopLeft, zgeo.SizeD(-15, 0))
	}
	if flags&ztime.TimeFieldDateOnly == 0 {
		hmax := 23
		if v.currentUse24Clock {
			hmax = 12
		}
		v.hourText = addText(v, 2, "H", "", hmax)
		v.minuteText = addText(v, 2, "M", ":", 59)
		if flags&ztime.TimeFieldSecs != 0 {
			v.secondsText = addText(v, 2, "S", ":", 59)
		}
		v.ampmLabel = zlabel.New("AM")
		v.ampmLabel.SetCanTabFocus(true)
		v.ampmLabel.View.SetObjectName("ampm")
		v.ampmLabel.SetFont(zgeo.FontNice(-2, zgeo.FontStyleBold))
		v.ampmLabel.SetColor(zgeo.ColorNew(0, 0, 0.1, 1))
		v.ampmLabel.SetPressedHandler("", zkeyboard.ModifierNone, v.toggleAMPM)
		v.Add(v.ampmLabel, zgeo.CenterLeft, zgeo.SizeD(2, 0))
		v.CollapseChild(v.ampmLabel, zlocale.IsUse24HourClock.Get(), false)
		zlocale.IsUse24HourClock.AddChangedHandler(func() {
			changed := v.CollapseChild(v.ampmLabel, zlocale.IsUse24HourClock.Get(), false)
			hour, err := strconv.Atoi(v.hourText.Text())
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
				zcontainer.ArrangeChildrenAtRootContainer(v, true)
			}
		})
		v.AddOnRemoveFunc(func() {
			zlocale.IsUse24HourClock.AddChangedHandler(nil)
		})
		if flags&ztime.TimeFieldTimeOnly == 0 {
			spacing := zcustom.NewView("spacing")
			spacing.SetMinSize(zgeo.SizeD(6, 6))
			v.Add(spacing, zgeo.CenterLeft)
		}
	}
	if flags&ztime.TimeFieldTimeOnly == 0 {
		v.dayText = addText(v, 2, "D", "", 31)
		v.dayText.SetMin(1)
		v.monthText = addText(v, 2, "M", "/", 12)
		v.monthText.SetMin(1)
		if flags&ztime.TimeFieldYears != 0 {
			cols := 4
			if flags&ztime.TimeFieldShortYear != 0 {
				cols = 2
			}
			v.yearText = addText(v, cols, "Y", "/", 2100)
		}
		if flags&ztime.TimeFieldNoCalendar == 0 {
			v.calendar = zimageview.NewWithCachedPath("images/zcore/calendar.png", zgeo.SizeD(16, 16))
			v.calendar.SetUsable(false)
			v.calendar.SetPressedHandler("", zkeyboard.ModifierNone, v.popCalendar)
			v.Add(v.calendar, zgeo.CenterLeft, zgeo.SizeD(11, 0))
		}
	}
	clear := zimageview.NewWithCachedPath("images/zcore/cross-circled.png", zgeo.SizeD(12, 14))
	clear.SetPressedHandler("", zkeyboard.ModifierNone, func() {
		v.Clear()
	})
	v.Add(clear, zgeo.CenterLeft, zgeo.SizeD(2, 0))
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

func addText(v *TimeFieldView, columns int, placeholder string, pre string, max int) *TextView {
	style := Style{KeyboardType: zkeyboard.TypeInteger}
	tv := NewView("", style, columns, 1)
	tv.SetFont(zgeo.FontNice(14, zgeo.FontStyleNormal))
	tv.UpdateSecs = 0
	tv.SetPlaceholder(placeholder)
	tv.SetMin(0)
	if max != -1 {
		tv.SetMax(float64(max))
	}
	tv.SetZIndex(zview.BaseZIndex)
	tv.SetFocusHandler(func(focused bool) {
		index := zview.BaseZIndex
		if focused {
			index = zview.BaseZIndex + 2
		}
		tv.SetZIndex(index)
	})
	tv.SetValueHandler("", func(edited bool) {
		clearColorTexts(v.hourText, v.minuteText, v.secondsText, v.dayText, v.monthText, v.yearText)
		v.Value() // getting value will set error color
	})
	if v.CallChangedOnTabPressed {
		tv.SetKeyHandler(func(km zkeyboard.KeyMod, down bool) bool {
			if down {
				return false
			}
			ztimer.StartIn(0.1, func() {
				if v.GetFocusedChildView(false) == nil {
					v.HandleValueChangedFunc()
				}
			})
			return false
		})
	}
	tv.SetTextAlignment(zgeo.Right)
	v.Add(tv, zgeo.TopLeft, zgeo.SizeD(2, 2))
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

func getAMPMString(pm bool) string {
	if pm {
		return "PM"
	}
	return "AM"
}

func setPM(v *TimeFieldView, pm bool) {
	v.ampmLabel.SetText(getAMPMString(pm))
}

func flipDayMonth(v *TimeFieldView, arrange bool) {
	if v.flags&ztime.TimeFieldTimeOnly != 0 {
		return
	}
	_, di := v.FindCellWithView(v.dayText)
	_, mi := v.FindCellWithView(v.monthText)
	if mi < di != zlocale.IsShowMonthBeforeDay.Get() {
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
	// zlog.Info("popCalendar:", val, err)
	if err != nil {
		return
	}
	// zlog.Info("CalPop:", val, err)
	cal.SetValue(val)
	cal.HandleValueChangedFunc = func() {
		ct := cal.Value()
		t := time.Date(ct.Year(), ct.Month(), ct.Day(), val.Hour(), val.Minute(), val.Second(), 0, v.location)
		CloseViewFunc(cal, true)
		v.SetValue(t)
	}
	cal.JSSet("className", "znofocus")
	PopupViewFunc(cal, v)
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
	if v.HandleValueChangedFunc != nil {
		v.HandleValueChangedFunc()
	}
	// v.location = nil
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
	if v.flags&ztime.TimeFieldShortYear != 0 {
		year %= 100

	}
	setInt(v.yearText, year, "%d")
	v.calendar.SetUsable(true)
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

func getNumString(v *TextView, str *string, def string) {
	if v == nil {
		return
	}
	text := v.Text()
	if text == "" {
		*str = def
		return
	}
	*str = text
}

func (v *TimeFieldView) Value() (time.Time, error) {
	var shour, smin, ssec string
	var syear, smonth, sday string

	var err error
	getNumString(v.hourText, &shour, "0")
	getNumString(v.minuteText, &smin, "0")
	getNumString(v.secondsText, &ssec, "0")
	v.calendar.SetUsable(err == nil)
	getNumString(v.monthText, &smonth, "")
	getNumString(v.yearText, &syear, "")
	// days := ztime.DaysInMonth(time.Month(month), year)
	// if v.flags&TimeFieldYears != 0 {
	// 	getNumString(v.yearText, &year, 0, 0, &err, true, &set)
	// 	if year < 100 {
	// 		year += 2000
	// 	}
	// }
	getNumString(v.dayText, &sday, "")

	v.currentUse24Clock = zlocale.IsUse24HourClock.Get()
	stime := zstr.Concat(":", shour, smin, ssec)
	if !v.currentUse24Clock {
		stime += v.ampmLabel.Text()
	}
	sdate := zstr.Concat("-", sday, smonth, syear)
	str := zstr.Concat(" ", stime, sdate)
	// zlog.Info("PARSE:", str)
	fieldToView := map[ztime.TimeFieldFlags]zview.View{
		ztime.TimeFieldHours:  v.hourText,
		ztime.TimeFieldMins:   v.minuteText,
		ztime.TimeFieldSecs:   v.secondsText,
		ztime.TimeFieldDays:   v.dayText,
		ztime.TimeFieldMonths: v.monthText,
		ztime.TimeFieldYears:  v.yearText,
		ztime.TimeFieldAMPM:   v.ampmLabel,
	}
	flags := ztime.TimeFieldNotFutureIfAmbiguous
	if !v.currentUse24Clock {
		flags |= ztime.TimeFieldAMPM
	}
	t, faults, err := ztime.ParseDate(str, v.location, flags)
	// zlog.Info("ParseDate:", str, t, faults, err)
	pink := zgeo.ColorNew(1, 0.7, 0.7, 1)
	for _, f := range faults {
		v := fieldToView[f]
		if v != nil {
			v.SetBGColor(pink)
		} else {
			v.SetBGColor(zgeo.ColorBlack)
		}
	}
	return t, err
}
