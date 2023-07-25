//go:build zui

package zfields

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zutil/zgeo"
)

type AmountBarWidgeter struct{}
type AmountCircleWidgeter struct{}
type ActivityWidgeter struct{}
type SetImagesWidgeter struct{}

func init() {
	RegisterWidgeter("zamount-bar", AmountBarWidgeter{})
	RegisterWidgeter("zamount-circle", AmountCircleWidgeter{})
	RegisterWidgeter("zactivity", ActivityWidgeter{})
	RegisterWidgeter("set-images", SetImagesWidgeter{})
}

func (a AmountBarWidgeter) Create(f *Field) zview.View {
	min := f.MinWidth
	if min == 0 {
		min = 100
	}
	progress := zwidgets.AmountViewBarNew(min)
	if f.Styling.FGColor.Valid {
		col := f.Styling.FGColor
		if col.Valid {
			progress.SetColor(col)
		}
	}
	return progress
}

func (a AmountBarWidgeter) SetupField(f *Field) {
	f.Flags |= FlagIsStatic
}

func (a AmountCircleWidgeter) Create(f *Field) zview.View {
	if f.Size.IsNull() {
		f.Size = zgeo.SizeBoth(20)
	}
	view := zwidgets.AmountViewCircleNew()
	view.SetMinSize(f.Size)
	view.SetColor(zgeo.ColorNew(0, 0.8, 0, 1))
	for i, n := range []float64{0, 70, 90} {
		if len(f.Colors) > 1 {
			if i < len(f.Colors) {
				view.ColorsFromValue[n] = zgeo.ColorFromString(f.Colors[i])
			}
		} else if f.Styling.FGColor.Valid {
			view.ColorsFromValue[0] = f.Styling.FGColor
		}
	}
	return view
}

func (a AmountCircleWidgeter) SetupField(f *Field) {
	f.Flags |= FlagIsStatic
}

func (a ActivityWidgeter) Create(f *Field) zview.View {
	if f.Size.IsNull() {
		f.Size = zgeo.SizeBoth(20)
	}
	f.SetEdited = false
	av := zwidgets.NewActivityView(f.Size)
	av.AlwaysVisible = f.Visible
	return av
}

func (a ActivityWidgeter) SetupField(f *Field) {
	f.Flags |= FlagIsStatic
}

func (a SetImagesWidgeter) SetupField(f *Field) {
	f.Flags |= FlagIsStatic
}

func (a SetImagesWidgeter) Create(f *Field) zview.View {
	f.Flags |= FlagIsStatic
	v := zwidgets.NewSetImagesView(f.FieldName, f.ImageFixedPath, f.Size, &f.Styling)
	v.SetStyling(f.Styling)
	return v
}
