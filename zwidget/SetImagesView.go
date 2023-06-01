package zwidget

// SetImagesView uses a value that can give a String() as a|b|c, and shows it as a row of images,
// where each image is found in images/flags/<prefix>/a.png etc.
// It is typically used with zbits.NamedBit

import (
	"fmt"
	"strings"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
)

type SetImagesView struct {
	zcontainer.StackView
	prefix    string
	imageSize zgeo.Size
}

func NewSetImagesView(name, imagePathPrefix string, imageSize zgeo.Size) *SetImagesView {
	if imageSize.IsNull() {
		imageSize = zgeo.SizeBoth(16)
	}
	v := &SetImagesView{}
	v.Init(v, false, name)
	v.prefix = imagePathPrefix
	v.imageSize = imageSize
	return v
}

func (v *SetImagesView) SetValueWithAny(bitset any) {
	v.RemoveAllChildren()
	stringer := bitset.(fmt.Stringer)
	parts := strings.Split(stringer.String(), "|")
	for _, part := range parts {
		path := zstr.Concat("/", "images/flags", v.prefix, part) + ".png"
		iv := zimageview.New(nil, path, v.imageSize)
		iv.DownsampleImages = true
		iv.SetToolTip(part)
		v.Add(iv, zgeo.CenterLeft)
	}
}
