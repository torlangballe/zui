// +build !js

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

func (v *WebView) init(minSize zgeo.Size, isFrame bool) {
	v.minSize = minSize
}

func (v *WebView) SetURL(surl string)         {}
func (v *WebView) SetHTMLContent(html string) {}
