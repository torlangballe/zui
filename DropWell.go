// +build zui

package zui

import (
	"bytes"
	"image"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

const wellBorderPath = "images/imagewell.png"

type DropWell struct {
	ImageView
	AcceptExtensions  []string
	HandleDroppedFile func(data []byte, name string) bool
	SetIconOnDrop     bool
}

func DropWellNew(filePath string, size zgeo.Size) *DropWell {
	v := &DropWell{}
	imagePath := filePath
	if !ImageExtensionInName(imagePath) {
		imagePath = wellBorderPath
	}
	v.Init(v, nil, "", size)
	v.DownsampleImages = true
	v.SetImageFromURL(imagePath)
	v.SetMinSize(size)
	v.SetCanFocus(true)
	v.SetPointerDropHandler(func(dtype DragType, data []byte, name string, pos zgeo.Pos) bool {
		v.SetHighlighted(dtype == DragEnter || dtype == DragOver)
		switch dtype {
		case DragDropFile:
			if v.HandleDroppedFile(data, name) {
				if v.SetIconOnDrop && ImageExtensionInName(name) {
					go v.SetIconFromBytes(data, name)
				}
			}
		}
		return true
	})
	v.SetDrawHandler(func(rect zgeo.Rect, canvas *Canvas, view View) {
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
		img = GoImageShrunkInto(img, size, true)
		ImageFromGo(img, func(zi *Image) {
			v.SetImage(zi, "", nil)
		})
	}
	return nil
}

func (v *DropWell) SetImageFromURL(surl string) {
	v.SetImage(nil, surl, func(img *Image) {
		if img == nil {
			v.SetImage(nil, wellBorderPath, nil)
		}
	})
}

func (v *DropWell) Clear() {
	v.SetImage(nil, wellBorderPath, nil)
}
