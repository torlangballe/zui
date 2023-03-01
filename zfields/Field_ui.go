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
// Widgeter is an interface to make a type create it's own view when build with zfields package.
// It is registered with the RegisterWigeter function, and specified with the zui:"widget:xxx" tag.
type Widgeter interface {
	IsStatic() bool
	Create(f *Field) zview.View
	SetValue(view zview.View, val any)
}

// ReadWidgeter is a Widgeter that also can return it's value
type ReadWidgeter interface {
	GetValue(view zview.View) any
}

// SetupWidgeter is a Widgeter that also can setup it's field before creation
type SetupWidgeter interface {
	SetupField(f *Field)
}

var widgeters = map[string]Widgeter{}

func RegisterWigeter(name string, w Widgeter) {
	widgeters[name] = w
}

func (f *Field) SetFont(view zview.View, from *zgeo.Font) {
	to := view.(ztext.LayoutOwner)
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