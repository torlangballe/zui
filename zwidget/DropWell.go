//go:build zui

package zwidget

import (
	"bytes"
	"image"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

const wellBorderPath = "images/imagewell.png"

type DropWell struct {
	zimageview.ImageView
	AcceptExtensions  []string
	HandleDroppedFile func(data []byte, name string) bool
	SetIconOnDrop     bool
}

func DropWellNew(filePath string, size zgeo.Size) *DropWell {
	v := &DropWell{}
	imagePath := filePath
	if !zimage.IsImageExtensionInName(imagePath) {
		imagePath = wellBorderPath
	}
	v.Init(v, nil, "", size)
	v.DownsampleImages = true
	v.SetImageFromURL(imagePath)
	v.SetMinSize(size)
	v.SetCanFocus(true)
	v.SetPointerDropHandler(func(dtype zview.DragType, data []byte, name string, pos zgeo.Pos) bool {
		v.SetHighlighted(dtype == zview.DragEnter || dtype == zview.DragOver)
		switch dtype {
		case zview.DragDropFile:
			if v.HandleDroppedFile(data, name) {
				if v.SetIconOnDrop && zimage.IsImageExtensionInName(name) {
					go v.SetIconFromBytes(data, name)
				}
			}
		}
		return true
	})
	v.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
		v.Draw(rect, canvas, view)
		if v.IsHighlighted() {
			r := v.GetImageRect(rect)
			r.Add(zgeo.RectFromXY2(3, 3, -2, -2))
			path := zgeo.PathNewRect(r, zgeo.SizeBoth(2))
			canvas.SetColor(zgeo.ColorYellow)
			canvas.FillPath(path)
		}
	})
	// v.SetBGColor(zgeo.ColorGreen)
	return v
}

func (v *DropWell) SetIconFromBytes(data []byte, name string) error {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return zlog.Error(err, "decode dragged image", name)
	}
	if img != nil {
		size := v.Rect().Size
		img = zimage.GoImageShrunkInto(img, size, true)
		zimage.FromGo(img, func(zi *zimage.Image) {
			v.SetImage(zi, "", nil)
		})
	}
	return nil
}

func (v *DropWell) SetImageFromURL(surl string) {
	v.SetImage(nil, surl, func(img *zimage.Image) {
		if img == nil {
			v.SetImage(nil, wellBorderPath, nil)
		}
	})
}

func (v *DropWell) Clear() {
	v.SetImage(nil, wellBorderPath, nil)
}
