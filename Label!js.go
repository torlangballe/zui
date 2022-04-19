//go:build !js && zui
// +build !js,zui

package zui

import (
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zutil/zgeo"
)

func LabelNew(text string) *Label {
	return nil
}

func (l *Label) SetTextAlignment(a zgeo.Alignment) View {
	l.alignment = a
	return l
}

func (v *Label) SetWrap(wrap ztextinfo.WrapType) {}

func (v *Label) SetMargin(m zgeo.Rect) {
	v.margin = m
}

func (v *Label) SetPressedHandler(handler func()) {
}

func (v *Label) SetMaxLines(max int) View {
	v.maxLines = max
	return v
}
