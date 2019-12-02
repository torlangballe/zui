package zgo

import "github.com/torlangballe/zutil/zgeo"

func LabelNew(text string) *Label {
	return nil
}

func (l *Label) TextAlignment(a zgeo.Alignment) View {
	l.alignment = a
	return l
}
