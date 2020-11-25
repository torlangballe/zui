package zui

import (
	"math"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
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
	a.OKButton = WordsGetOK()
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
		text = WordsGetCancel()
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

func AlertShowError(text string, err error) {
	a := AlertNew(text + "\n" + err.Error())
	a.Show(nil)
	zlog.Error(err, text)
}

func (a *Alert) ShowOK(handle func()) {
	a.Show(func(a AlertResult) {
		if handle != nil && a == AlertOK {
			handle()
		}
	})
}

// AlertStatusLabel is a global Label you can point to somewhere in your visible gui, Alert.ShowStatus sets it's text for a limited time
var AlertSetStatus func(str string)
var statusTimer *ztimer.Timer

// AlertShowStatus shows an status/error in a label on gui, and can hide it after secs
func AlertShowStatus(text string, secs float64) {
	// zlog.Info("AlertShowStatus", text, secs)
	if AlertSetStatus != nil {
		AlertSetStatus(text)
		if statusTimer != nil {
			statusTimer.Stop()
		}
		if secs != 0 {
			statusTimer = ztimer.StartIn(secs, func() {
				if AlertSetStatus != nil { // in case it is nil'ed
					statusTimer = nil
					AlertSetStatus("")
				}
			})
		}
	}
}

func addButtonIfNotEmpty(stack *StackView, text string, handle func(result AlertResult), result AlertResult) {
	if text != "" {
		button := ButtonNew(text, "gray", zgeo.Size{10, 28}, zgeo.Size{})
		stack.AddAlertButton(button)
		button.SetPressedHandler(func() {
			PresentViewPop(stack, func() {
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
	label := LabelNew(a.Text)
	label.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
	label.SetMaxLines(0)
	label.SetMaxWidth(textWidth)
	stack.Add(zgeo.TopCenter, label)
	if a.SubText != "" {
		subLabel := LabelNew(a.SubText)
		subLabel.SetFont(FontNice(FontDefaultSize-2, FontStyleNormal))
		// subLabel.SetMaxLines(4)
		stack.Add(zgeo.TopCenter, subLabel)
	}
	bar := StackViewHor("bar")
	stack.Add(zgeo.TopCenter|zgeo.HorExpand, bar)

	addButtonIfNotEmpty(stack, a.OKButton, handle, AlertOK)
	addButtonIfNotEmpty(stack, a.CancelButton, handle, AlertCancel)
	addButtonIfNotEmpty(stack, a.DestructiveButton, handle, AlertDestructive)
	addButtonIfNotEmpty(stack, a.OtherButton, handle, AlertOther)

	att := PresentViewAttributesNew()
	att.Modal = true
	PresentView(stack, att, nil, nil)
}
