//go:build zui

package zeditabletexticon

import (
	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zcode"
	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type widgeter struct{}
type IconView struct {
	zlabel.Label
	editText   string
	field      zfields.Field
	fv         *zfields.FieldView
	IconEmojii string
}

func init() {
	zfields.RegisterWidgeter("zeditable-text", widgeter{})
}

func (widgeter) Create(fv *zfields.FieldView, f *zfields.Field) zview.View {
	v := NewIconView(f)
	v.fv = fv
	return v
}

func NewIconView(f *zfields.Field) *IconView {
	v := &IconView{}
	v.IconEmojii = "📜"
	v.Init(v, "   ")
	v.field = *f
	f.Styling.BGColor = zgeo.Color{}
	f.Styling.FGColor = zgeo.Color{}
	v.SetPressedHandler("$edit", 0, v.popEditor)
	v.update()
	return v
}

func (v *IconView) popEditor() {
	cols := 80
	if v.field.Columns != 0 {
		cols = v.field.Columns
	}
	rows := 20
	if v.field.Rows != 0 {
		rows = v.field.Rows
	}
	// zlog.Info("NewIconEditorView", cols, rows)
	ev := zcode.NewEditorView(v.editText, cols, rows)
	att := zpresent.ModalConfirmAttributes
	zalert.PresentOKCanceledView(ev, "Edit "+v.field.Name, att, nil, func(ok bool) bool {
		if ok {
			v.editText = ev.Text()
			if v.fv != nil {
				v.fv.ToData(true)
			}
			v.update()
		}
		return true
	})
}

func (v *IconView) SetValueWithAny(a any) {
	v.editText = a.(string)
	zlog.Info("IV SetValueWithAny", v.editText)
	v.update()
}

func (v *IconView) ValueAsAny() any {
	return v.editText
}

func (v *IconView) update() {
	str := v.IconEmojii
	if v.editText != "" {
		str += "✓"
	} else {
		str += "  "
	}
	v.SetText(str)
}
