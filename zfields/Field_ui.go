//go:build zui

package zfields

import (
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

func init() {
	callSetupWidgeter = setupWidgeter
}

// Widgeter is an interface to make a type create its own view when build with zfields package.
// It is registered with the RegisterWidgeter function, and specified with the zui:"widget:xxx" tag.
type Widgeter interface {
	Create(f *Field) zview.View
	// SetValue(view zview.View, val any)
}

// // ReadWidgeter is a Widgeter that also can return its value
// type ReadWidgeter interface {
// 	GetValue(view zview.View) any
// }

// SetupWidgeter is a Widgeter that also can setup its field before creation
type SetupWidgeter interface {
	SetupField(f *Field)
}

// type ChangedWidgeter interface {
// 	SetChangeHandler(func())
// }

var widgeters = map[string]Widgeter{}

func RegisterWidgeter(name string, w Widgeter) {
	widgeters[name] = w
}

func (f *Field) SetFont(view zview.View, from *zgeo.Font) {
	to := view.(ztext.TextOwner)
	size := f.Styling.Font.Size
	if size <= 0 {
		if from != nil {
			size = from.Size
		} else {
			size = zgeo.FontDefaultSize
		}
	}
	style := f.Styling.Font.Style
	if from != nil {
		style = from.Style
	}
	if style == zgeo.FontStyleUndef {
		style = zgeo.FontStyleNormal
	}
	var font *zgeo.Font
	if f.Styling.Font.Name != "" {
		font = zgeo.FontNew(f.Styling.Font.Name, size, style)
	} else if from != nil {
		font = new(zgeo.Font)
		*font = *from
	} else {
		font = zgeo.FontNice(size, style)
	}
	// zlog.Info("Field SetFont:", view.Native().Hierarchy(), *font)
	to.SetFont(font)
}

func setupWidgeter(f *Field) {
	w := widgeters[f.WidgetName]
	if w != nil {
		sw, _ := w.(SetupWidgeter)
		if sw != nil {
			sw.SetupField(f)
		}
	}
}
