package zui

import (
	"github.com/lxn/walk"
	"github.com/torlangballe/zutil/zlog"
)

// TODO: Implement this

func ClipboardSetString(str string) {
	c := walk.Clipboard()
	c.SetText(str)
}

func ClipboardGetString() string {
	c := walk.Clipboard()
	str, err := c.Text()
	zlog.OnError(err, "get text")
	return str
}
