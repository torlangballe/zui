//go:build zui

package zimageview

import (
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /20/10/15.

type XImageView struct {
	zcontainer.StackView
	minSize   zgeo.Size
	alignment zgeo.Alignment
}

func XNew(name, imagePath string, fitSize zgeo.Size) *XImageView {
	v := &XImageView{}
	v.Init(v, name, imagePath, fitSize)
	return v
}

func (v *XImageView) Init(view zview.View, name string, imagePath string, fitSize zgeo.Size) {
	v.StackView.Init(view, false, name+"#type:img")
	v.minSize = fitSize
	v.SetCanTabFocus(false)
	v.alignment = zgeo.Center | zgeo.Proportional
	v.SetURL(imagePath)
}

func (v *XImageView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	// zlog.Info("IV CS", v.Hierarchy(), s, p, v.image != nil, zlog.GetCallingStackString())
	// margSize := v.Margin().Size
	if !v.minSize.IsNull() {
		s = v.minSize
		// s = s.ShrunkInto(v.fitSize.Plus(margSize))
	}
	// if !v.minSize.IsNull() {
	// 	ms := v.minSize.Plus(margSize)
	// 	if s.IsNull() || v.alignment&zgeo.Proportional == 0 {
	// 		s = ms
	// 	} else {
	// 		s = s.ExpandedInto(ms)
	// 	}
	// }
	// s.Add(margSize.Negative())
	s.Maximize(zgeo.SizeD(2, 2))
	return s, s
}

func (v *XImageView) MinSize() zgeo.Size {
	return v.minSize
}

func (v *XImageView) SetMinSize(s zgeo.Size) {
	v.minSize = s
}

// func (v *XImageView) Alignment() zgeo.Alignment {
// 	return v.alignment
// }

// func (v *XImageView) SetAlignment(a zgeo.Alignment) {
// 	v.alignment = a
// }

func (v *XImageView) SetURL(path string) {
	// zlog.Info("IV SetImage", path, v.JSGet("id").String(), v.Rect(), v.image != nil)
	v.JSSet("src", path)
}

func (v *XImageView) SetStroke(width float64, c zgeo.Color, inset bool) {
	v.SetNativePadding(zgeo.RectFromMarginSize(zgeo.SizeBoth(width)))
	v.NativeView.SetStroke(width, c, inset)
	// d := zstyle.MakeDropShadow(0, 0, 0, c)
	// d.Inset = true
	// d.Spread = width
	// v.SetDropShadow(d)
	// str := fmt.Sprintf("0px 0px 0px %dpx %s", int(width), c.Hex())
	// // str := fmt.Sprintf("%dpx solid %s", int(width), c.Hex())
	// if inset {
	// 	str += " inset"
	// }
	// v.SetJSStyle("boxShadow", str)
}
