// +build zui

package zui

import (
	"strconv"
	"time"

	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
)

// https://www.npmjs.com/package/js-datepicker

type TimeFieldView struct {
}

type TimeTextView struct {
	StackView
	TimeText    *TextView
	DateText    *TextView
	ParsedLabel *Label
	UseYear     bool
	Handle      func(t time.Time)
	onChrome    bool
}

func TimeTextViewNew() *TimeTextView {
	v := &TimeTextView{}
	v.Init(v, false, "timestack")
	v.SetSpacing(0)

	v.onChrome = (zdevice.WasmBrowser() == "chrome")
	style := TextViewStyle{}
	v.TimeText = TextViewNew("", style, 7, 1)
	v.TimeText.UpdateSecs = 0
	v.TimeText.SetPlaceholder("mm:hh")
	v.TimeText.SetToolTip("type minute:hour")
	v.TimeText.SetCorners(5, zgeo.TopLeft|zgeo.Bottom)
	v.TimeText.SetZIndex(100)
	v.Add(v.TimeText, zgeo.TopLeft)

	//	style = TextViewStyle{}//Type: TextViewDate}
	v.DateText = TextViewNew("", style, 7, 1)
	v.DateText.UpdateSecs = 0
	v.DateText.SetPlaceholder("dd-mm")
	v.DateText.SetToolTip("type date-month")
	v.DateText.SetCorners(5, zgeo.TopRight|zgeo.Bottom)
	v.Add(v.DateText, zgeo.CenterLeft, zgeo.Size{-1, 0})
	// if !v.UseYear {
	// 	year := strconv.Itoa(time.Now().Year())
	// 	v.DateText.setjs("min", year+"-01-01")
	// 	v.DateText.setjs("max", year+"-12-31")

	// }
	v.ParsedLabel = LabelNew("")
	v.ParsedLabel.SetMinWidth(80)
	v.ParsedLabel.SetColor(zgeo.ColorBlue)
	v.Add(v.ParsedLabel, zgeo.TopLeft, zgeo.Size{3, 0})

	changed := func() {
		var str string
		t := v.Parse()
		if !t.IsZero() {
			str = ztime.GetNice(t, false)
		}
		v.setLabel(str)
	}
	v.TimeText.SetChangedHandler(changed)
	v.DateText.SetChangedHandler(changed)
	keyHandler := func(key KeyboardKey, mods KeyboardModifier) bool {
		// zlog.Info("key:", key)
		if key == KeyboardKeyReturn {
			if v.Handle != nil {
				t := v.Parse()
				v.Handle(t)
			}
			return true
		}
		return false
	}
	v.TimeText.SetKeyHandler(keyHandler)
	v.DateText.SetKeyHandler(keyHandler)
	v.setLabel("")
	return v
}

func (v *TimeTextView) setLabel(str string) {
	s := zgeo.FontDefaultSize
	marg := 0.0
	if str == "" {
		str = "📅"
		if !v.onChrome {
			marg = 2
		}
	}
	c, _ := v.FindCellWithView(v.ParsedLabel)
	c.Margin.H = marg
	v.ParsedLabel.SetText(str)
	v.ParsedLabel.SetFont(zgeo.FontNice(s, zgeo.FontStyleNormal))
	v.ArrangeChildren()
}

func (v *TimeTextView) Clear() {
	v.TimeText.SetText("")
	v.DateText.SetText("")
	v.setLabel("")
	if v.Handle != nil {
		v.Handle(time.Time{})
	}
}

func (v *TimeTextView) Parse() time.Time {
	var shour, smin, sday, smonth string
	var err error
	var min, hour int

	stime := v.TimeText.Text()
	sdate := v.DateText.Text()
	now := time.Now().Local()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	if !zstr.SplitN(stime, ":", &shour, &smin) {
		return time.Time{}
	}
	if sdate != "" {
		if !zstr.SplitN(sdate, "-", &sday, &smonth) {
			sday = sdate
		} else {
			m, _ := strconv.Atoi(smonth)
			if m == 0 {
				m, got := ztime.MonthFromString(smonth)
				if !got {
					return time.Time{}
				}
				month = m
			} else {
				if m <= 0 || m > 12 {
					return time.Time{}
				}
				month = time.Month(m)
			}
		}
		day, err = strconv.Atoi(sday)
		if err != nil {
			return time.Time{}
		}
	}
	hour, err = strconv.Atoi(shour)
	if err != nil {
		return time.Time{}
	}
	min, err = strconv.Atoi(smin)
	if err != nil {
		return time.Time{}
	}
	t := time.Date(year, month, day, hour, min, 0, 0, time.Local)
	if time.Since(t) > 0 {
		return t
	}
	if sdate == "" {
		t = t.Add(-ztime.Day)
		if time.Since(t) > 0 {
			return t
		}
	}
	if smonth == "" && month != 0 {
		month--
		if month == 0 {
			month = 12
		}
		t = time.Date(year, month, day, hour, min, 0, 0, time.Local)
		if time.Since(t) > 0 {
			return t
		}
	}
	return time.Date(year-1, month, day, hour, min, 0, 0, time.Local)
}
