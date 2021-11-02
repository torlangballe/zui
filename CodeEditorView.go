// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

type CodeEditorView struct {
	TextView
	// NativeView
	// nativeCodeEditorView
	// MinSize zgeo.Size
}

func CodeEditorViewNew(text string, cols, rows int) *CodeEditorView {
	v := &CodeEditorView{}
	v.TextView.Init(v, text, TextViewStyle{}, cols, rows)
	v.SetBGColor(zgeo.ColorNewGray(0.3, 1))
	v.SetColor(zgeo.ColorNewGray(0.8, 1))
	v.SetObjectName("editor2")
	v.SetMargin(zgeo.RectFromXY2(10, 10, -10, -10))
	font := zgeo.FontNew("Lucida Console, Monaco, monospace", 14, zgeo.FontStyleNormal)
	v.SetFont(font)
	return v
}

// func (v *CodeEditorView) CalculatedSize(total zgeo.Size) zgeo.Size {
// 	return v.MinSize
// }
