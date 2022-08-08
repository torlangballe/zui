//go:build zui

package zlabel

import (
	"github.com/torlangballe/zui/zcheckbox"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /2/11/15.

type Label struct {
	zview.NativeView
	zview.LongPresser

	minWidth  float64
	maxWidth  float64
	maxLines  int
	margin    zgeo.Rect
	alignment zgeo.Alignment

	pressed     func()
	longPressed func()

	Columns int
}

func (v *Label) GetTextInfo() ztextinfo.Info {
	t := ztextinfo.New()
	t.Alignment = v.alignment
	t.Font = v.Font()
	if v.Columns == 0 {
		t.Text = v.Text()
	}
	if v.maxWidth != 0 {
		t.SetWidthFreeHight(v.maxWidth)
	}
	t.MaxLines = v.maxLines
	return *t
}

func (v *Label) CalculatedSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	to := v.View.(ztextinfo.Owner)
	ti := to.GetTextInfo()
	if v.Columns != 0 {
		s = ti.GetColumnsSize(v.Columns)
	} else {
		s, _, _ = ti.GetBounds()
	}
	s.Add(v.margin.Size.Negative())
	zfloat.Maximize(&s.W, v.minWidth)
	// zlog.Info("L CS:", v.ObjectName(), s, v.minWidth, v.maxWidth)
	if v.maxWidth != 0 {
		zfloat.Minimize(&s.W, v.maxWidth)
	}
	s = s.Ceil()
	s.W += 1
	return s
}

func (v *Label) IsMinimumOneLineHight() bool {
	return v.maxLines > 0
}

func (v *Label) TextAlignment() zgeo.Alignment {
	return v.alignment
}

func (v *Label) MinWidth() float64 {
	return v.minWidth
}

func (v *Label) MaxWidth() float64 {
	return v.maxWidth
}

func (v *Label) MaxLines() int {
	return v.maxLines
}

func (v *Label) SetMinWidth(min float64) {
	v.minWidth = min
}

func (v *Label) SetMaxWidth(max float64) {
	v.maxWidth = max
}

func Labelize(view zview.View, prefix string, minWidth float64, alignment zgeo.Alignment) (label *Label, stack *zcontainer.StackView, viewCell *zcontainer.Cell) {
	font := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleBold)
	to, _ := view.(ztextinfo.Owner)
	if to != nil {
		ti := to.GetTextInfo()
		font = ti.Font
		font.Style = zgeo.FontStyleBold
	}
	title := prefix
	checkBox, isCheck := view.(*zcheckbox.CheckBox)
	if checkBox != nil && alignment&zgeo.Right != 0 {
		title = ""
		_, cstack := LabelizeCheckbox(checkBox, prefix)
		view = cstack
		alignment = alignment.FlippedHorizontal()
	}
	label = New(title)
	label.SetObjectName("$labelize.label " + prefix)
	label.SetTextAlignment(zgeo.Right)
	label.SetFont(font)
	label.SetColor(zstyle.DefaultFGColor().WithOpacity(0.7))
	stack = zcontainer.StackViewHor("$labelize." + prefix) // give it special name so not easy to mis-search for in recursive search

	stack.AddView(label, zgeo.CenterLeft).MinSize.W = minWidth
	marg := zgeo.Size{}
	if isCheck {
		marg.W = -6 // in html cell has a box around it of 20 pixels
	}
	// zlog.Info("Labelize view:", view.ObjectName(), alignment, marg)
	viewCell = stack.Add(view, alignment, marg)
	return
}

func (v *Label) PressedHandler() func() {
	return v.pressed
}

func (v *Label) LongPressedHandler() func() {
	return v.longPressed
}

func LabelizeCheckbox(c *zcheckbox.CheckBox, title string) (*Label, *zcontainer.StackView) {
	label := New(title)
	label.SetObjectName("$checkBoxLabel:[" + title + "]")
	stack := zcontainer.StackViewHor("$labledCheckBoxStack.[" + title + "]")
	stack.SetSpacing(0)
	stack.Add(c, zgeo.Left|zgeo.VertCenter, zgeo.Size{0, -4})
	stack.Add(label, zgeo.Left|zgeo.VertCenter, zgeo.Size{6, 0})

	return label, stack
}
