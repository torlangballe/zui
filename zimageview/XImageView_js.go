//go:build zui

package zimageview

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /20/10/15.

type XImageView struct {
	zview.NativeView
	fitSize   zgeo.Size
	alignment zgeo.Alignment
	// imageCorner float64
	// strokeWidth float64
	// strokeColor zgeo.Color
	// strokeInset bool
	// EmptyColor  zgeo.Color
}

func XNew(name, imagePath string, fitSize zgeo.Size) *XImageView {
	v := &XImageView{}
	v.Init(v, name, imagePath, fitSize)
	return v
}

func (v *XImageView) Init(view zview.View, name string, imagePath string, fitSize zgeo.Size) {
	v.MakeJSElement(view, "img")
	v.SetCanTabFocus(false)
	v.alignment = zgeo.Center | zgeo.Proportional
}

// func (v *ImageView) SetPressToShowImage(on bool) {
// 	if on {
// 		v.SetPressedHandler(func() {
// 			if v.image != nil {
// 				path := v.image.Path
// 				if !zhttp.StringStartsWithHTTPX(path) {
// 					path = zstr.Concat("/", zrest.AppURLPrefix, path)
// 				}
// 				nv := New(v.image, v.image.Path, zgeo.SizeNull)
// 				att := zpresent.AttributesNew()
// 				att.Modal = true
// 				att.ModalCloseOnOutsidePress = true
// 				zpresent.PresentTitledView(nv, path, att, nil, nil)
// 			}
// 		})
// 	} else {
// 		v.SetPressedHandler(nil)
// 	}
// }

// func (v *ImageView) SetStroke(width float64, c zgeo.Color, inset bool) {
// 	v.strokeWidth = width
// 	v.strokeColor = c
// 	v.strokeInset = inset
// 	v.Expose()
// }

// func (v *ImageView) SetImageCorner(radius float64) {
// 	v.imageCorner = radius
// 	v.Expose()
// }

func (v *XImageView) CalculatedSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	// zlog.Info("IV CS", v.Hierarchy(), s, p, v.image != nil, zlog.GetCallingStackString())
	// margSize := v.Margin().Size
	if !v.fitSize.IsNull() {
		s = v.fitSize
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
	return s
}

func (v *XImageView) FitSize() zgeo.Size {
	return v.fitSize
}

func (v *XImageView) SetFitSize(s zgeo.Size) {
	v.fitSize = s
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
