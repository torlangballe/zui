package zgo

import "syscall/js"

func (a *Alert) Show(handle func(result AlertResult)) {
	r := true
	if a.CancelButton != "" {
		alert := js.Global().Get("confirm")
		r = alert.Invoke(a.Text).Bool()
	} else {
		alert := js.Global().Get("alert")
		alert.Invoke(a.Text)
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
