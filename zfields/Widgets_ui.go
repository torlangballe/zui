//go:build zui

package zfields

import (
	"reflect"

	"github.com/torlangballe/zui/zcolor"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zutil/zerrors"
	"github.com/torlangballe/zutil/zgeo"
)

type AmountBarWidgeter struct{}
type AmountCircleWidgeter struct{}
type ActivityWidgeter struct{}
type ImagesSetWidgeter struct{}
type ColorWidgeter struct{}
type ScreensViewWidgeter struct{}

func init() {
	RegisterWidgeter("zamount-bar", AmountBarWidgeter{})
	RegisterWidgeter("zamount-circle", AmountCircleWidgeter{})
	RegisterWidgeter("zactivity", ActivityWidgeter{})
	RegisterWidgeter("images-set", ImagesSetWidgeter{})
	RegisterWidgeter("zcolor", ColorWidgeter{})
	RegisterWidgeter("zscreens", ScreensViewWidgeter{})
	RegisterCreator("zerrors.ContextError", buildContextError)
}

func (a AmountBarWidgeter) Create(fv *FieldView, f *Field) zview.View {
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

func (a AmountCircleWidgeter) Create(fv *FieldView, f *Field) zview.View {
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

func (a ActivityWidgeter) Create(fv *FieldView, f *Field) zview.View {
	if f.Size.IsNull() {
		f.Size = zgeo.SizeBoth(20)
	}
	f.SetEdited = false
	av := zwidgets.NewActivityView(f.Size, f.Styling.FGColor)
	av.AlwaysVisible = f.Visible
	return av
}

func (a ActivityWidgeter) SetupField(f *Field) {
	f.Flags |= FlagIsStatic
}

func (a ImagesSetWidgeter) SetupField(f *Field) {
	f.Flags |= FlagIsStatic
}

func (a ImagesSetWidgeter) Create(fv *FieldView, f *Field) zview.View {
	f.Flags |= FlagIsStatic
	s := f.Size
	if s.IsNull() {
		s = zgeo.SizeBoth(14)
	}
	v := zwidgets.NewImagesSetView(f.FieldName, f.ImageFixedPath, f.Size, &f.Styling)
	v.SetMinSize(zgeo.SizeD(f.MinWidth, s.H))
	v.SetStyling(f.Styling)
	return v
}

func (a ColorWidgeter) Create(fv *FieldView, f *Field) zview.View {
	return zcolor.New(zgeo.ColorClear)
}

func (ScreensViewWidgeter) Create(fv *FieldView, f *Field) zview.View {
	minSize := zgeo.SizeD(120, 90)
	if !f.Size.IsNull() {
		minSize = f.Size
	}
	return zwidgets.NewScreensView(minSize)
}

func buildContextError(in *FieldView, f *Field, val any) zview.View {
	e := val.(zerrors.ContextError)
	// zlog.Info("buildContextError:", f.Name, e.Title, len(e.KeyValues), e.SubContextError != nil)
	frame := in.BuildMapList(reflect.ValueOf(e.KeyValues), f, e.Title)
	if e.SubContextError != nil {
		// zlog.Info("buildContextError sub:", f.Name, e.SubContextError.Title)
		sub := buildContextError(in, f, *e.SubContextError)
		adder := frame.(zcontainer.AdvancedAdder)
		adder.AddAdvanced(sub, zgeo.TopLeft|zgeo.Expand, zgeo.RectNull, zgeo.Size{}, -1, false)
	}
	return frame
}
