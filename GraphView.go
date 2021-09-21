// +build zui

package zui

import (
	"bytes"

	"github.com/go-echarts/go-echarts/v2/render"
	"github.com/torlangballe/zutil/zgeo"
)

type GraphView struct {
	WebView
	Renderer render.Renderer
}

func GraphViewNew(minSize zgeo.Size) *GraphView {
	v := &GraphView{}
	isFrame := true
	v.WebView.init(minSize, isFrame)
	return v
}

func (v *GraphView) Render() error {
	var buffer bytes.Buffer
	err := v.Renderer.Render(&buffer)
	if err != nil {
		return err
	}
	v.SetHTMLContent(string(buffer.Bytes()))
	return nil
}
