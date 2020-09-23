package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

type WebView struct {
	ContainerView
	url     string
	minSize zgeo.Size
	// Padding zgeo.Size
}

func (v *WebView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return v.minSize
}
