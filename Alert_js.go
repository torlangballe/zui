package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zlog"
)

func (a *Alert) Show(handle func(result AlertResult)) {
	r := true
	str := a.Text
	if a.SubText != "" {
		str += "\n\n" + a.SubText
	}
	zlog.Info("alert:", str)
	if a.CancelButton != "" {
		alert := js.Global().Get("confirm")
		r = alert.Invoke(str).Bool()
	} else {
		alert := js.Global().Get("alert")
		alert.Invoke(str)
	}
	go func() {
		if handle != nil {
			if r {
				handle(AlertOK)
			} else {
				handle(AlertCancel)
			}
		}
	}()
}
