//go:build !js && zui

package zlabel

import (
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

func (label *Label) Init(view zview.View, text string)             {}
func (label *Label) InitAsLink(view zview.View, name, surl string) {}
func (l *Label) SetTextAlignment(a zgeo.Alignment)                 { l.alignment = a }
func (v *Label) SetMargin(m zgeo.Rect)                             { v.margin = m }
func (v *Label) SetMaxLines(max int)                               { v.maxLines = max }
func (v *Label) SetWrap(wrap ztextinfo.WrapType)                   {}
func (v *Label) SetPressedHandler(handler func())                  {}
func (v *Label) SetPressedDownHandler(handler func())              {}
func (v *Label) SetLongPressedHandler(handler func())              {}
func (label *Label) SetURL(surl string, newWindow bool)            {}
