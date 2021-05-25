// +build zui

package zui

import (
	"math"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
	"github.com/torlangballe/zutil/zwords"
)

//  Created by Tor Langballe on /7/11/15.

type AlertResult int

const (
	AlertCancel AlertResult = iota
	AlertOK
	AlertDestructive
	AlertOther
)

type Alert struct {
	Text              string
	OKButton          string
	CancelButton      string
	OtherButton       string
	DestructiveButton string
	SubText           string
	BuildGUI          bool
}

func AlertNew(items ...interface{}) *Alert {
	a := &Alert{}
	str := zstr.SprintSpaced(items...)
	a.OKButton = zwords.OK()
	a.Text = str
	return a
}

func AlertNewWithCancel(items ...interface{}) *Alert {
	a := AlertNew(items...)
	a.SetCancel("")
	return a
}

func (a *Alert) SetCancel(text string) *Alert {
	if text == "" {
		text = zwords.Cancel()
	}
	a.CancelButton = text
	return a
}

func (a *Alert) SetOther(text string) *Alert {
	a.OtherButton = text
	return a
}

func (a *Alert) SetDestructive(text string) *Alert {
	a.DestructiveButton = text
	return a
}

func (a *Alert) SetSub(text string) *Alert {
	a.SubText = text
	return a
}

func AlertShow(items ...interface{}) {
	a := AlertNew(items...)
	a.Show(nil)
}

func AlertAsk(title string, handle func(ok bool)) {
	alert := AlertNewWithCancel(title)
	alert.Show(func(a AlertResult) {
		handle(a == AlertOK)
	})
}

func AlertShowError(err error, items ...interface{}) {
	str := zstr.SprintSpaced(items...)
	str = zstr.Concat("\n", str, err)
	a := AlertNew(str)
	a.Show(nil)
	zlog.Error(err, str)
}

func (a *Alert) ShowOK(handle func()) {
	a.Show(func(a AlertResult) {
		if handle != nil && a == AlertOK {
			handle()
		}
	})
}

var AlertSetStatus func(parts ...interface{}) = func(parts ...interface{}) {
	zlog.Info(parts...)
}
var statusTimer *ztimer.Timer

// AlertShowStatus shows an status/error in a label on gui, and can hide it after secs
func AlertShowStatus(secs float64, parts ...interface{}) {
	// zlog.Info("AlertShowStatus", len(parts))
	AlertSetStatus(parts...)
	if statusTimer != nil {
		statusTimer.Stop()
	}
	if secs != 0 {
		statusTimer = ztimer.StartIn(secs, func() {
			statusTimer = nil
			AlertSetStatus("")
		})
	}
}

func addButtonIfNotEmpty(stack, bar *StackView, text string, handle func(result AlertResult), result AlertResult) {
	if text != "" {
		button := ButtonNew(text)
		bar.AddAlertButton(button)
		button.SetPressedHandler(func() {
			zlog.Info("Button pressed!")
			PresentViewClose(stack, result == AlertCancel, func(dismissed bool) {
				if handle != nil {
					handle(result)
				}
			})
		})
	}
}

func (a *Alert) Show(handle func(result AlertResult)) {
	if !a.BuildGUI {
		a.showNative(handle)
		return
	}

	textWidth := math.Min(640, ScreenMain().Rect.Size.W/2)

	stack := StackViewVert("alert")
	stack.SetMargin(zgeo.RectFromXY2(20, 20, -20, -20))
	stack.SetBGColor(zgeo.ColorWhite)

	label := LabelNew(a.Text)
	label.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
	label.SetMaxLines(0)
	label.SetMaxWidth(textWidth)
	stack.Add(label, zgeo.TopCenter)
	if a.SubText != "" {
		subLabel := LabelNew(a.SubText)
		subLabel.SetFont(FontNice(FontDefaultSize-2, FontStyleNormal))
		// subLabel.SetMaxLines(4)
		stack.Add(subLabel, zgeo.TopCenter)
	}
	bar := StackViewHor("bar")
	stack.Add(bar, zgeo.TopCenter|zgeo.HorExpand, zgeo.Size{0, 10})

	addButtonIfNotEmpty(stack, bar, a.OKButton, handle, AlertOK)
	addButtonIfNotEmpty(stack, bar, a.CancelButton, handle, AlertCancel)
	addButtonIfNotEmpty(stack, bar, a.DestructiveButton, handle, AlertDestructive)
	addButtonIfNotEmpty(stack, bar, a.OtherButton, handle, AlertOther)

	att := PresentViewAttributesNew()
	att.Modal = true
	PresentView(stack, att, nil, nil)
}
