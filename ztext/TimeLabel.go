//go:build zui

package ztext

import (
	"time"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/ztime"
)

type TimeLabel struct {
	zlabel.Label
	flags ztime.TimeFieldFlags
	value time.Time
}

func TimeLabelNew(name string, flags ztime.TimeFieldFlags) *TimeLabel {
	tl := &TimeLabel{}
	tl.flags = flags
	tl.Init(tl, name)
	col := zstyle.DefaultFGColor()
	if flags&ztime.TimeFieldStatic == 0 {
		col = zgeo.ColorBlue
		tl.SetPressedHandler("", zkeyboard.ModifierNone, func() {
			tf := TimeFieldNew(name, flags)
			tf.HandleValueChangedFunc = func() {
				t, err := tf.Value()
				if err == nil {
					tl.SetValue(t)
					zpresent.Close(tf, false, nil)
				} else {
					tl.SetToolTip(err.Error())
				}
			}
			tf.SetValue(tl.value)
			zpresent.PopupView(tf, tl, zpresent.AttributesNew())
		})
	}
	tl.SetColor(col)
	zlocale.IsUse24HourClock.AddChangedHandler(func() {
		tl.SetValue(tl.value)
		zcontainer.ArrangeChildrenAtRootContainer(tl)
	})
	return tl
}

func (tl *TimeLabel) SetValue(t time.Time) {
	tl.value = t
	var format string
	if tl.flags&ztime.TimeFieldDateOnly == 0 {
		if zlocale.IsUse24HourClock.Get() {
			format = "15"
		} else {
			format = "3"
		}
		format += ":04"
		if tl.flags&ztime.TimeFieldSecs != 0 {
			format += ":05"
		}
		if !zlocale.IsUse24HourClock.Get() {
			format += "pm"
		}
		if tl.flags&ztime.TimeFieldTimeOnly == 0 {
			format += " "
		}
	}
	if tl.flags&ztime.TimeFieldTimeOnly == 0 {
		if zlocale.IsShowMonthBeforeDay.Get() {
			format += "Jan-02"
		} else {
			format += "02-Jan"
		}
		if tl.flags&ztime.TimeFieldYears != 0 {
			if tl.flags&ztime.TimeFieldShortYear != 0 {
				format += "-06"
			} else {
				format += "-2006"
			}
		}
	}
	str := tl.value.Format(format)
	tl.SetText(str)
}
