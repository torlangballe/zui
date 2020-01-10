package zui

import (
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /7/11/15.

type AlertResult int

const (
	AlertOK          = 1
	AlertCancel      = 2
	AlertDestructive = 3
	AlertOther       = 4
)

type Alert struct {
	Text              string
	OKButton          string
	CancelButton      string
	OtherButton       string
	DestructiveButton string
	SubText           string
}

func AlertNew(items ...interface{}) *Alert {
	a := &Alert{}
	str := zstr.SprintSpaced(items...)
	a.OKButton = WordsGetOK()
	a.Text = str
	return a
}

func (a *Alert) Cancel(text string) *Alert {
	if text == "" {
		text = WordsGetCancel()
	}
	a.CancelButton = text
	return a
}

func (a *Alert) Other(text string) *Alert {
	a.OtherButton = text
	return a
}

func (a *Alert) Destructive(text string) *Alert {
	a.DestructiveButton = text
	return a
}

func (a *Alert) Sub(text string) *Alert {
	a.SubText = text
	return a
}

func AlertShowError(text string, err error) {
	a := AlertNew(text + "\n" + err.Error())
	a.Show(nil)
	zlog.Error(err, text)
}

func (a *Alert) ShowOK(handle func()) {
	a.Show(func(a AlertResult) {
		if a == AlertOK {
			handle()
		}
	})
}
