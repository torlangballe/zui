//go:build zui

package zalert

import (
	"math"

	"github.com/torlangballe/zui/zbutton"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zscreen"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/zwords"
)

//  Created by Tor Langballe on /7/11/15.

type Alert struct {
	Text                       string
	OKButton                   string
	CancelButton               string
	OtherButton                string
	UploadButton               string
	DestructiveButton          string
	SubText                    string
	BuildGUI                   bool
	HandleUpload               func(data []byte, filename string)
	DialogView                 zview.View
	MimesOrExtensionsForUpload []string
}

type StatusSetter interface {
	Clear(id string)
	ClearAll()
	SetError(id, str string, secs int, details ...any)
	SetOKText(id, str string, secs int, details ...any)
}

type Result int

const (
	Cancel Result = iota
	OK
	Destructive
	Other
	Upload
	borderMargin = 10
)

func init() {
	zpresent.ShowErrorFunc = func(title, subTitle string) {
		a := New(title)
		a.SubText = subTitle
		a.Show(nil)
	}
}

func New(items ...interface{}) *Alert {
	a := &Alert{}
	str := zstr.Spaced(items...)
	a.OKButton = zwords.OK()
	a.Text = str
	return a
}

func NewWithCancel(items ...interface{}) *Alert {
	a := New(items...)
	a.SetCancel("")
	return a
}

func (a *Alert) SetCancel(text string) {
	if text == "" {
		text = zwords.Cancel()
	}
	a.CancelButton = text
}

// func (a *Alert) SetOther(text string) {
// 	a.OtherButton = text
// }

// func (a *Alert) SetDestructive(text string) {
// 	a.DestructiveButton = text
// }

// func (a *Alert) SetSub(text string) {
// 	a.SubText = text
// }

func Show(items ...interface{}) {
	a := New(items...)
	a.Show(nil)
}

func Ask(title string, handle func(ok bool)) {
	alert := NewWithCancel(title)
	alert.Show(func(a Result) {
		handle(a == OK)
	})
}

func ShowError(err error, items ...interface{}) {
	str := zstr.Spaced(items...)
	if err != nil {
		str = zstr.Concat("\n", err, str)
	}
	a := New(str)
	a.Show(nil)
	zlog.Error(str, err)
}

func (a *Alert) ShowOK(handle func()) {
	a.Show(func(a Result) {
		if handle != nil && a == OK {
			handle()
		}
	})
}

// var SetStatus func(parts ...interface{}) = func(parts ...interface{}) {
// 	zlog.Info(parts...)
// }

// var StatusLabel *widgets.StatusLabel

// var statusTimer *ztimer.Timer

// ShowStatus shows an status/error in a label on gui, and can hide it after secs
// func ShowStatus(secs float64, parts ...interface{}) {
// 	// zlog.Info("ShowStatus", len(parts))
// 	SetStatus(parts...)
// 	if statusTimer != nil {
// 		statusTimer.Stop()
// 	}
// 	if secs != 0 {
// 		statusTimer = ztimer.StartIn(secs, func() {
// 			statusTimer = nil
// 			SetStatus("")
// 		})
// 	}
// }

func (a *Alert) addButtonIfNotEmpty(stack, bar *zcontainer.StackView, text string, handle func(result Result), result Result) {
	if text != "" {
		if result == Upload {
			zlog.Assert(a.HandleUpload != nil)
			button := zbutton.New(text)
			button.SetUploader(a.MimesOrExtensionsForUpload, func(data []byte, filename string) {
				a.HandleUpload(data, filename)
				zpresent.Close(a.DialogView, false, nil)
			}, nil, nil)
			bar.Add(button, zgeo.CenterRight)
		} else {
			button := zbutton.New(text)
			bar.Add(button, zgeo.CenterRight)
			if result == Cancel {
				button.MakeEscapeCanceler()
			}
			button.SetPressedHandler("", zkeyboard.ModifierNone, func() {
				zpresent.Close(stack, result == Cancel, func(dismissed bool) {
					if handle != nil {
						handle(result)
					}
				})
			})
			// 	button.SetKeyHandler(func(km zkeyboard.KeyMod, down bool) bool {
			// 		zlog.Info("PRess:", km.Key.IsReturnish(), km.Modifier)
			// 		if !down {
			// 			return false
			// 		}
			// 		if km.Key.IsReturnish() && km.Modifier == 0 {
			// 			zpresent.Close(stack, result == Cancel, func(dismissed bool) {
			// 				if handle != nil {
			// 					handle(result)
			// 				}
			// 			})
			// 			return true
			// 		}
			// 		return false
			// 	})
		}
	}
}

func (a *Alert) Show(handle func(result Result)) {
	if a.UploadButton != "" || a.OKButton != zwords.OK() {
		a.BuildGUI = true
	}
	if !a.BuildGUI {
		a.showNative(handle)
		return
	}

	textWidth := math.Min(640, zscreen.GetMain().Rect.Size.W/2)

	stack := zcontainer.StackViewVert("alert")
	stack.SetMargin(zgeo.RectFromXY2(borderMargin, borderMargin, -borderMargin, -borderMargin))
	stack.SetBGColor(zgeo.ColorWhite)

	label := zlabel.New(a.Text)
	label.SetFont(zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal))
	label.SetMaxLines(0)
	label.SetMaxWidth(textWidth)
	stack.Add(label, zgeo.TopCenter|zgeo.HorExpand)
	if a.SubText != "" {
		subLabel := zlabel.New(a.SubText)
		subLabel.SetMaxLines(0)
		subLabel.SetFont(zgeo.FontNice(zgeo.FontDefaultSize-2, zgeo.FontStyleNormal))
		// subLabel.SetMaxLines(4)
		stack.Add(subLabel, zgeo.TopCenter|zgeo.HorExpand)
	}
	bar := zcontainer.StackViewHor("bar")
	stack.Add(bar, zgeo.TopRight|zgeo.HorExpand, zgeo.SizeD(0, 10))

	a.addButtonIfNotEmpty(stack, bar, a.CancelButton, handle, Cancel)
	a.addButtonIfNotEmpty(stack, bar, a.DestructiveButton, handle, Destructive)
	a.addButtonIfNotEmpty(stack, bar, a.OtherButton, handle, Other)
	a.addButtonIfNotEmpty(stack, bar, a.UploadButton, handle, Upload)
	a.addButtonIfNotEmpty(stack, bar, a.OKButton, handle, OK)

	a.DialogView = stack
	att := zpresent.AttributesDefault()
	att.Modal = true
	zpresent.PresentView(stack, att)
}

func addButton(bar *zcontainer.StackView, view zview.View, title string, isOKButton bool, done func(isOKButton bool) (close bool)) *zbutton.Button {
	button := zbutton.New(title)
	button.SetMinWidth(80)
	bar.Add(button, zgeo.TopRight)
	button.SetPressedHandler("", zkeyboard.ModifierNone, func() {
		parent := view.Native().Parent()
		go func() {
			close := done(isOKButton)
			if close {
				zpresent.Close(parent, !isOKButton, nil)
			}
		}()
	})
	return button
}

const DisableOKCancelKeyDefaults = 1

func PresentOKCanceledView(view zview.View, title string, att zpresent.Attributes, barViews []zview.View, done func(ok bool) (close bool)) {
	stack := zcontainer.StackViewVert("alert")
	stack.SetBGColor(zstyle.DefaultBGColor())
	stack.SetMargin(zgeo.RectFromXY2(borderMargin, borderMargin, -borderMargin, -borderMargin))

	stack.Add(view, zgeo.TopCenter|zgeo.Expand)
	bar := zcontainer.StackViewHor("bar")
	bar.SetMargin(zgeo.RectFromXY2(5, 5, -5, -5))
	stack.Add(bar, zgeo.TopRight|zgeo.HorExpand, zgeo.SizeD(0, 2))

	cancelButton := addButton(bar, stack, "Cancel", false, done)
	okButton := addButton(bar, stack, "OK", true, done)
	if att.CustomFlags&DisableOKCancelKeyDefaults == 0 {
		okButton.MakeReturnKeyDefault()
		cancelButton.MakeEscapeCanceler()
	}
	for _, v := range barViews {
		bar.Add(v, zgeo.TopLeft)
	}
	att.Modal = true
	if title != "" {
		zpresent.PresentTitledView(stack, title, att, nil, nil)
	} else {
		zpresent.PresentView(stack, att)
	}
}
