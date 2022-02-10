package zfields

import (
	"github.com/torlangballe/zui"
	"github.com/torlangballe/zutil/zgeo"
)

type StatusWidgeter struct{}

func init() {
	RegisterWigeter("zstatus", StatusWidgeter{})
}

func (s StatusWidgeter) Create(f *Field) zui.View {
	// zlog.Info("Status.Create", zlog.GetCallingStackString())
	size := f.Size
	if size.IsNull() {
		size = zgeo.SizeBoth(20)
	}
	v := zui.ImageViewNew(nil, "images/error.png", size)
	v.SetMinSize(size)
	//		image.Show(e.Error != nil)
	v.SetPressedHandler(func() {
		zui.AlertShow(v.ObjectName())
		v.SetObjectName("")
		v.Show(false)
	})
	return v
}

func (s StatusWidgeter) SetupField(f *Field) {
	f.MinWidth = 24
	f.Title = "⚠"
}

func (s StatusWidgeter) SetValue(view zui.View, val interface{}) {
	serr := val.(string)
	view.SetObjectName(serr)
	// zlog.Info("Status.SetVal", serr, zlog.GetCallingStackString())
	iv := view.(*zui.ImageView)
	iv.Expose()
	iv.SetToolTip(serr)
	iv.Show(serr != "")
}

func (s StatusWidgeter) GetValue(view zui.View) interface{} {
	return view.ObjectName()
}


