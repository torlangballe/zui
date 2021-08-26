// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zwords"
)

const MenuSeparatorID = "$sep"

type MenuType interface {
	UpdateItems(items zdict.Items, values []interface{})
	SetSelectedHandler(handler func())
}

type MenuView struct {
	NativeView
	maxWidth        float64
	selectedHandler func()
	items           zdict.Items
	currentValue    interface{}
}

var menuViewHeight = 21.0

func (v *MenuView) CurrentValue() interface{} {
	return v.currentValue
}

func (v *MenuView) Dump() {
	zlog.Info("DumpMenu:", v.ObjectName())
	zdict.DumpNamedValues(v.items)
}

func (v *MenuView) getNumberOfItemsString() string {
	return zwords.PluralWordWithCount("item", float64(v.items.Count()), "", "", 0)
}

func (v *MenuView) CalculatedSize(total zgeo.Size) zgeo.Size {
	var max string
	for _, item := range v.items {
		if len(item.Name) > len(max) {
			max = item.Name
		}
	}
	w := TextInfoWidthOfString(max, v.Font())
	if v.maxWidth != 0 {
		zfloat.Minimize(&w, v.maxWidth)
	}
	w += 25
	s := zgeo.Size{w, menuViewHeight}
	return s
}

func (v *MenuView) MaxWidth() float64 {
	return v.maxWidth
}

func (v *MenuView) SetMaxWidth(max float64) {
	v.maxWidth = max
}
