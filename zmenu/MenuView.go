//go:build zui

package zmenu

import (
	"fmt"
	"strings"

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
	RowFormat       string
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

func (v *MenuView) GetDump() string {
	return fmt.Sprintf("%+v", v.items)
}

func (v *MenuView) getNumberOfItemsString() string {
	return zwords.PluralWordWithCount("item", float64(v.items.Count()), "", "", 0)
}

func getRowString(format string, item zdict.Item) string {
	if format == "" {
		return item.Name
	}
	str := strings.Replace(format, "$N", item.Name, -1)
	str = strings.Replace(str, "$V", fmt.Sprint(item.Value), -1)
	return str
}

func (v *MenuView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	var maxStr string
	for _, item := range v.items {
		rowStr := getRowString(v.RowFormat, item)
		if len(rowStr) > len(maxStr) {
			maxStr = rowStr
		}
	}
	w := ztextinfo.WidthOfString(maxStr, v.Font())
	if v.maxWidth != 0 {
		zfloat.Minimize(&w, v.maxWidth)
	}
	w += 36
	if zdevice.OS() != zdevice.MacOSType {
		w += 5
	}
	s = zgeo.SizeD(w, menuViewHeight)
	s.H += 4 // 8
	max = s
	if v.maxWidth == 0 || len(v.items) == 0 {
		max.W = 0
	}
	// zlog.Info("MV.CalculatedSize:", s, max, maxStr, len(v.items))
	return s, max
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
		if v.SelectWithValue(a) {
			return
		}
	}
	if len(v.items) != 0 && v.currentValue == nil {
		v.SelectWithValue(v.items[0].Value)
	}
}
