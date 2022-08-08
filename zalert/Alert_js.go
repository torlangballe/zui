package zalert

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zlog"
)

func (a *Alert) showNative(handle func(result Result)) {
	r := true
	str := a.Text
	if a.SubText != "" {
		str += "\n\n" + a.SubText
	}
	// zlog.Info("alert:", str)
	if a.CancelButton != "" {
		alert := js.Global().Get("confirm")
		r = alert.Invoke(str).Bool()
	} else {
		alert := js.Global().Get("alert")
		alert.Invoke(str)
	}
	go func() {
		defer zlog.HandlePanic(true)
		if handle != nil {
			if r {
				handle(OK)
			} else {
				handle(Cancel)
			}
		}
	}()
}

func PromptForText(title string, got func(str string)) {
	prompt := js.Global().Get("prompt")
	e := prompt.Invoke(title)
	if !e.IsNull() {
		go got(e.String())
	}
}
