package zui

import (
	"github.com/lxn/walk"
	"github.com/torlangballe/zutil/zlog"
)

// TODO: Implement this

func PasteboardSetString(str string) {
	c := walk.Clipboard()
	c.SetText(str)
}

func PasteboardGetString() string {
	c := walk.Clipboard()
	str, err := c.Text()
	zlog.OnError(err, "get text")
	return str
}
