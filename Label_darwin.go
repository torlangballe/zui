package zui

import "github.com/torlangballe/zutil/zgeo"

func LabelNew(text string) *Label {
	return nil
}

func (l *Label) SetTextAlignment(a zgeo.Alignment) View {
	l.alignment = a
	return l
}

func (v *Label) SetMargin(m zgeo.Rect) *Label {
	v.margin = m
	return v
}

func (v *Label) SetPressedHandler(handler func()) {
}
