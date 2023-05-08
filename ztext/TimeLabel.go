//go:build zui

package ztext

import (
	"time"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlocale"
)

type TimeLabel struct {
	zlabel.Label
	flags TimeFieldFlags
	value time.Time
}

func TimeLabelNew(name string, flags TimeFieldFlags) *TimeLabel {
	tl := &TimeLabel{}
	tl.flags = flags
	tl.Init(tl, name)
	col := zgeo.ColorBlue
	if flags&TimeFieldStatic == 0 {
		col = zstyle.DefaultFGColor()
		tl.SetPressedHandler(func() {
			tf := TimeFieldNew(name, flags)
			tf.HandleValueChanged = func(t time.Time) {
				tl.SetValue(t)
			}
			tf.SetValue(tl.value)
			zpresent.PopupView(tf, tl)
		})
	}
	tl.SetColor(col)
	zkeyvalue.SetOptionChangeHandler(tl, func(key string) {
		tl.SetValue(tl.value)
		zcontainer.ArrangeChildrenAtRootContainer(tl)
	})
	return tl
}

func (tl *TimeLabel) SetValue(t time.Time) {
	tl.value = t
	var format string
	if tl.flags&TimeFieldDateOnly == 0 {
		if zlocale.IsUse24HourClock.Get() {
			format = "15"
		} else {
			format = "3"
		}
		format += ":04"
		if tl.flags&TimeFieldSecs != 0 {
			format += ":05"
		}
		if !zlocale.IsUse24HourClock.Get() {
			format += "pm"
		}
		if tl.flags&TimeFieldTimeOnly == 0 {
			format += " "
		}
	}
	if tl.flags&TimeFieldTimeOnly == 0 {
		if zlocale.IsShowMonthBeforeDay.Get() {
			format += "Jan-02"
		} else {
			format += "02-Jan"
		}
		if tl.flags&TimeFieldYears != 0 {
			if tl.flags&TimeFieldShortYear != 0 {
				format += "-06"
			} else {
				format += "-2006"
			}
		}
	}
	str := tl.value.Format(format)
	tl.SetText(str)
}
