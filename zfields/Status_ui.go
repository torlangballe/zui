package zfields

import (
	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

type StatusWidgeter struct{}

func init() {
	RegisterWigeter("zstatus", StatusWidgeter{})
}

func (s StatusWidgeter) Create(f *Field) zview.View {
	// zlog.Info("Status.Create", zlog.GetCallingStackString())
	size := f.Size
	if size.IsNull() {
		size = zgeo.SizeBoth(20)
	}
	v := zimageview.New(nil, "images/error.png", size)
	v.SetMinSize(size)
	//		image.Show(e.Error != nil)
	v.SetPressedHandler(func() {
		zalert.Show(v.ObjectName())
		v.SetObjectName("")
		v.Show(false)
	})
	return v
}

func (s StatusWidgeter) SetupField(f *Field) {
	f.MinWidth = 24
	f.Title = "⚠"
}

func (s StatusWidgeter) SetValue(view zview.View, val interface{}) {
	serr := val.(string)
	view.SetObjectName(serr)
	// zlog.Info("Status.SetVal", serr, zlog.GetCallingStackString())
	iv := view.(*zimageview.ImageView)
	iv.Expose()
	iv.SetToolTip(serr)
	iv.Show(serr != "")
}

func (s StatusWidgeter) GetValue(view zview.View) interface{} {
	return view.ObjectName()
}
