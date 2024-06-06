//go:build zui

package zwidgets

// SetImagesView uses a value that can give a String() as a|b|c, and shows it as a row of images,
// where each image is found in images/flags/<prefix>/a.png etc.
// It is typically used with zbits.NamedBit

import (
	"fmt"
	"strings"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
)

type SetImagesView struct {
	zcontainer.StackView
	prefix    string
	imageSize zgeo.Size
}

func NewSetImagesView(name, imagePathPrefix string, imageSize zgeo.Size, styling *zstyle.Styling) *SetImagesView {
	if imageSize.IsNull() {
		imageSize = zgeo.SizeBoth(16)
	}
	v := &SetImagesView{}
	v.Init(v, false, name)
	v.prefix = imagePathPrefix
	v.imageSize = imageSize

	spacing := 2.0
	if styling != nil && styling.Spacing != zgeo.UndefValue {
		spacing = styling.Spacing
	}
	v.SetSpacing(spacing)
	return v
}

func (v *SetImagesView) SetValueWithAny(bitset any) {
	v.RemoveAllChildren()
	stringer := bitset.(fmt.Stringer)
	parts := strings.Split(stringer.String(), "|")
	for _, part := range parts {
		path := zstr.Concat("/", "images/flags", v.prefix, part) + ".png"
		iv := zimageview.New(nil, true, path, v.imageSize)
		iv.DownsampleImages = true
		iv.SetToolTip(part)
		v.Add(iv, zgeo.CenterLeft)
	}
	if v.IsPresented() {
		v.ArrangeChildren()
	}
}
