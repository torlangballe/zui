package zui

import (
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

const separatorID = "$sep"

type MenuType interface {
	//	UpdateAndSelect(items zdict.Items, value interface{})
	UpdateItems(items zdict.Items, values []interface{})
	//	SelectWithValue(value interface{}, set bool)
	SetSelectedHandler(handler func())
}

type MenuView struct {
	NativeView
	maxWidth        float64
	selectedHandler func()
	items           zdict.Items
	currentValue    interface{}
}

var menuViewHeight = 22.0

func (v *MenuView) CurrentValue() interface{} {
	return v.currentValue
}

func (v *MenuView) Dump() {
	zlog.Info("DumpMenu:", v.ObjectName())
	zdict.DumpNamedValues(v.items)
}

func (v *MenuView) getNumberOfItemsString() string {
	return WordsPluralizeString("%d %s", "en", float64(v.items.Count()), "item")
}

// func MenuViewCalcItemsWidth(items zdict.Items, font *Font) float64 {
// 	var maxString string
// 	// zlog.Info("MV Calc size:", v)
// 	for _, item := range items {
// 		if len(item.Name) > len(maxString) {
// 			maxString = item.Name
// 		}
// 	}
// 	ti := TextInfoNew()
// 	ti.Alignment = zgeo.Left
// 	ti.Text = maxString
// 	ti.IsMinimumOneLineHight = true
// 	ti.Font = font
// 	ti.MaxLines = 1
// 	s, _, _ := ti.GetBounds()
// 	return s.W
// }

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
	// zlog.Info("MenuView calcedsize:", v.Font().Size, v.ObjectName(), maxString, s)
	return zgeo.Size{w + 38, menuViewHeight}
}

func (v *MenuView) MaxWidth() float64 {
	return v.maxWidth
}

func (v *MenuView) SetMaxWidth(max float64) {
	v.maxWidth = max
}
