package zalert

import (
	"syscall/js"

	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zdebug"
)

func (a *Alert) showNative(handle func(result Result)) {
	r := true
	str := a.Text
	if a.SubText != "" {
		str += "\n\n" + a.SubText
	}
	e := zwindow.Current().Element
	// e := js.Global()
	// zlog.Info("alert:", str)
	if a.CancelButton != "" {
		alert := e.Get("confirm")
		r = alert.Invoke(str).Bool()
	} else {
		alert := e.Get("alert")
		alert.Invoke(str)
	}
	go func() {
		defer zdebug.RecoverFromPanic(true)
		if handle != nil {
			if r {
				handle(OK)
			} else {
				handle(Cancel)
			}
		}
	}()
}

func PromptForText(title, defaultText string, got func(str string)) {
	prompt := js.Global().Get("prompt")
	e := prompt.Invoke(title, defaultText)
	if !e.IsNull() {
		go got(e.String())
	}
}
