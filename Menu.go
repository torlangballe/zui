package zgo

import "github.com/torlangballe/zutil/zgeo"

type MenuView struct {
	NativeView
	maxWidth float64
	changed  func(key string, val interface{})
	keyVals  Dictionary
}

func (v *MenuView) GetCalculatedSize(total zgeo.Size) zgeo.Size {
	maxString := ""
	for s := range v.keyVals {
		if len(s) > len(maxString) {
			maxString = s
		}
	}
	s := TextLayoutCalculateSize(zgeo.AlignmentLeft, v.GetFont(), maxString, 1, v.maxWidth)
	// fmt.Println("MenuView calcedsize:", s)
	s.Add(zgeo.Size{20, 4})
	return s
}

func (v *MenuView) GetMaxWidth() float64 {
	return v.maxWidth
}

func (v *MenuView) MaxWidth(max float64) View {
	v.maxWidth = max
	return v
}
