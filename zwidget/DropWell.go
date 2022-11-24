//go:build zui

package zwidget

import (
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

type DropWell struct {
	zcustom.CustomView
	placeHolder         string
	HandleDroppedFile   func(data []byte, name string)
	HandleDropPreflight func(name string) bool
	Styling             zstyle.Styling
}

func NewDropWell(filePath string, size zgeo.Size) *DropWell {
	v := &DropWell{}
	v.Init(v, "dropwell")
	v.SetMinSize(size)
	v.Styling.Corner = 7
	v.Styling.BGColor = zgeo.ColorNewGray(0.95, 1)
	v.Styling.FGColor = zgeo.ColorNewGray(0, 0.05)
	v.Styling.StrokeWidth = 1
	v.Styling.StrokeColor = zgeo.ColorNewGray(0.7, 1)
	v.Styling.DropShadow = zstyle.DropShadow{Delta: zgeo.Size{5, 5}, Blur: 5, Color: zgeo.ColorNewGray(0, 0.7)}
	v.DownsampleImages = true
	v.SetMinSize(size)
	v.SetCanFocus(true)
	v.SetPointerDropHandler(func(dtype zview.DragType, data []byte, name string, pos zgeo.Pos) bool {
		if v.HandleDropPreflight != nil && dtype == zview.DragDropFilePreflight {
			r := v.HandleDropPreflight(name)
			// zlog.Info("HandleDropPreflight:", name, r)
			return r
		}
		v.SetHighlighted(dtype == zview.DragEnter || dtype == zview.DragOver)
		switch dtype {
		case zview.DragDropFile:
			if v.HandleDroppedFile != nil {
				v.HandleDroppedFile(data, name)
			}
		}
		return true
	})
	v.SetDrawHandler(v.draw)
	return v
}

func (v *DropWell) draw(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	// zlog.Info("DropWell draw", rect)
	const corner = 7
	path := zgeo.PathNewRect(rect.ExpandedD(-1), zgeo.SizeBoth(v.Styling.Corner))
	canvas.SetColor(v.Styling.BGColor)
	fillCol := v.Styling.FGColor
	if v.IsHighlighted() {
		fillCol = zgeo.ColorYellow
	}
	canvas.SetColor(fillCol)
	path = zgeo.PathNewRect(rect.ExpandedD(-1), zgeo.SizeBoth(v.Styling.Corner))
	canvas.FillPath(path)
	canvas.PushState()
	canvas.ClipPath(path, false)
	path.AddRect(rect.ExpandedD(20), zgeo.Size{}) // add an outer box, so drop-shadow is an inset
	canvas.SetDropShadow(v.Styling.DropShadow)
	canvas.SetColor(v.Styling.StrokeColor)
	canvas.FillPathEO(path)
	canvas.PopState()

	canvas.SetColor(v.Styling.StrokeColor)
	canvas.StrokePath(path, v.Styling.StrokeWidth, zgeo.PathLineRound)

	if v.placeHolder != "" {
		ti := ztextinfo.New()
		ti.Font = zgeo.FontNice(zgeo.FontDefaultSize-2, zgeo.FontStyleNormal)
		ti.Rect = rect
		ti.Text = v.placeHolder
		ti.Alignment = zgeo.Center
		ti.Color = zgeo.ColorNewGray(0, 0.5)
		ti.Draw(canvas)
	}
}

func (v *DropWell) SetPlaceholder(str string) {
	v.placeHolder = str
	v.Expose()
}
