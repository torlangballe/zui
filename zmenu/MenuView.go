//go:build zui

package zmenu

import (
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zwords"
)

const MenuSeparatorID = "$sep"

type MenuType interface {
	UpdateItems(items zdict.Items, valuesSlicePtr any, isAction bool)
}

type MenuView struct {
	zview.NativeView
	StoreKey        string
	maxWidth        float64
	selectedHandler func()
	items           zdict.Items
	currentValue    interface{}
	storeKey        string
}

var menuViewHeight = 21.0

func (v *MenuView) CurrentValue() interface{} {
	return v.currentValue
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
	w := ztextinfo.WidthOfString(max, v.Font())
	if v.maxWidth != 0 {
		zfloat.Minimize(&w, v.maxWidth)
	}
	w += 36
	if zdevice.OS() != zdevice.MacOSType {
		w += 5
	}
	s := zgeo.Size{w, menuViewHeight}
	s.H += 4 // 8
	return s
}

func (v *MenuView) MaxWidth() float64 {
	return v.maxWidth
}

func (v *MenuView) SetMaxWidth(max float64) {
	v.maxWidth = max
}

func (v *MenuView) SetStoreKey(key string) {
	v.storeKey = key
	a, got := zkeyvalue.DefaultStore.GetItemAsAny(v.storeKey)
	if got {
		v.SelectWithValue(a)
	}
}
