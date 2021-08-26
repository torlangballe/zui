// +build zui

package zui

import (
	"math"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zscreen"
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
	AlertUpload
)

type Alert struct {
	Text              string
	OKButton          string
	CancelButton      string
	OtherButton       string
	UploadButton      string
	DestructiveButton string
	SubText           string
	BuildGUI          bool
	HandleUpload      func(data []byte, filename string)
	DialogView        View
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

func (a *Alert) SetCancel(text string) {
	if text == "" {
		text = zwords.Cancel()
	}
	a.CancelButton = text
}

func (a *Alert) SetOther(text string) {
	a.OtherButton = text
}

func (a *Alert) SetDestructive(text string) {
	a.DestructiveButton = text
}

func (a *Alert) SetSub(text string) {
	a.SubText = text
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
	if err != nil {
		str = zstr.Concat("\n", err, str)
	}
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

func alertMakeUploadButton() *ShapeView {
	v := ShapeViewNew(ShapeViewTypeRoundRect, zgeo.Size{68, 22})
	v.SetColor(zgeo.ColorWhite)
	v.StrokeColor = zgeo.ColorNew(0, 0.6, 0, 1)
	v.StrokeWidth = 2
	v.Ratio = 0.3
	v.SetBGColor(zgeo.ColorClear)
	v.SetText("Upload")
	return v
}

func (a *Alert) addButtonIfNotEmpty(stack, bar *StackView, text string, handle func(result AlertResult), result AlertResult) {
	if text != "" {
		if result == AlertUpload {
			zlog.Assert(a.HandleUpload != nil)
			button := alertMakeUploadButton()
			button.SetUploader(func(data []byte, filename string) {
				a.HandleUpload(data, filename)
				PresentViewClose(a.DialogView, false, nil)
			})
			bar.AddAlertButton(button)
		} else {
			button := ButtonNew(text)
			bar.AddAlertButton(button)
			button.SetPressedHandler(func() {
				// zlog.Info("Button pressed!")
				PresentViewClose(stack, result == AlertCancel, func(dismissed bool) {
					if handle != nil {
						handle(result)
					}
				})
			})
		}
	}
}

func (a *Alert) Show(handle func(result AlertResult)) {
	if a.UploadButton != "" || a.OKButton != zwords.OK() {
		a.BuildGUI = true
	}
	if !a.BuildGUI {
		a.showNative(handle)
		return
	}

	textWidth := math.Min(640, zscreen.GetMain().Rect.Size.W/2)

	stack := StackViewVert("alert")
	stack.SetMargin(zgeo.RectFromXY2(20, 20, -20, -20))
	stack.SetBGColor(zgeo.ColorWhite)

	label := LabelNew(a.Text)
	label.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
	label.SetMaxLines(0)
	label.SetMaxWidth(textWidth)
	stack.Add(label, zgeo.TopCenter|zgeo.HorExpand)
	if a.SubText != "" {
		subLabel := LabelNew(a.SubText)
		subLabel.SetMaxLines(0)
		// subLabel.SetMinSize(zgeo.Size{100, 100})
		subLabel.SetFont(FontNice(FontDefaultSize-2, FontStyleNormal))
		// subLabel.SetMaxLines(4)
		stack.Add(subLabel, zgeo.TopCenter|zgeo.HorExpand)
	}
	bar := StackViewHor("bar")
	stack.Add(bar, zgeo.TopCenter|zgeo.HorExpand, zgeo.Size{0, 10})

	a.addButtonIfNotEmpty(stack, bar, a.CancelButton, handle, AlertCancel)
	a.addButtonIfNotEmpty(stack, bar, a.DestructiveButton, handle, AlertDestructive)
	a.addButtonIfNotEmpty(stack, bar, a.OtherButton, handle, AlertOther)
	a.addButtonIfNotEmpty(stack, bar, a.UploadButton, handle, AlertUpload)
	a.addButtonIfNotEmpty(stack, bar, a.OKButton, handle, AlertOK)

	a.DialogView = stack
	att := PresentViewAttributesNew()
	att.Modal = true
	PresentView(stack, att, nil, nil)
}

func addButton(view View, bar *StackView, title string, ok bool, done func(ok bool) bool) {
	button := ButtonNew(title)
	button.SetMinWidth(80)
	bar.AddAlertButton(button)
	button.SetPressedHandler(func() {
		parent := ViewGetNative(view).Parent()
		close := done(ok)
		if close {
			PresentViewClose(parent, !ok, nil)
		}
	})
}

func PresentOKCanceledView(view View, title string, done func(ok bool) bool) { // move this to PresentView?
	stack := StackViewVert("alert")
	stack.SetBGColor(zgeo.ColorWhite)
	stack.SetMargin(zgeo.RectFromXY2(20, 20, -20, -20))

	stack.Add(view, zgeo.TopCenter|zgeo.Expand)
	bar := StackViewHor("bar")
	stack.Add(bar, zgeo.TopCenter|zgeo.HorExpand, zgeo.Size{0, 10})

	addButton(stack, bar, "Cancel", false, done)
	addButton(stack, bar, "OK", true, done)

	att := PresentViewAttributesNew()
	att.Modal = true
	if title != "" {
		PresentTitledView(stack, title, att, nil, nil, nil, nil)
	} else {
		PresentView(stack, att, nil, nil)
	}
}
