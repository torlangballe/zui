package zui

import (
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
)

type MenuView struct {
	NativeView
	maxWidth float64
	changed  func(item zdict.Item)
	keyVals  zdict.Items
	oldValue *zdict.Item

	IsStatic bool // if static, user can't set a different value, but can press and see them
}

func (v *MenuView) GetCalculatedSize(total zgeo.Size) zgeo.Size {
	maxString := ""
	for _, di := range v.keyVals {
		if len(di.Name) > len(maxString) {
			maxString = di.Name
		}
		if v.IsStatic {
			break
		}
	}
	maxString += "m"
	s := TextLayoutCalculateSize(zgeo.Left, v.Font(), maxString, 1, v.maxWidth)
	// fmt.Println("MenuView calcedsize:", s)
	s.Add(zgeo.Size{20, 4})
	return s
}

func (v *MenuView) GetMaxWidth() float64 {
	return v.maxWidth
}

func (v *MenuView) SetMaxWidth(max float64) View {
	v.maxWidth = max
	return v
}
