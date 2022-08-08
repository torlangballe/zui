//go:build !js && zui && !catalyst

package zweb

import (
	"github.com/torlangballe/zutil/zgeo"
)

type nativeWebView struct {
}

func (v *WebView) init(minSize zgeo.Size, isFrame bool) {
	v.minSize = minSize
}

func (v *WebView) SetURL(surl string)         {}
func (v *WebView) SetHTMLContent(html string) {}
