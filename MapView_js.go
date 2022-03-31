package zui

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zgeo"
)

// https://developers.google.com/maps/documentation/javascript/maxzoom
// https://stackoverflow.com/questions/44659884/load-google-map-without-using-callback-method

type baseMapView struct {
	MapJS js.Value
}

func makeLatLingJS(pos zgeo.Pos) js.Value {
	m := map[string]interface{}{
		"lat": pos.Y,
		"lng": pos.X,
	}
	return js.ValueOf(m)
}

func (v *MapView) Init(view View, center zgeo.Pos, zoom int) {
	v.Element = zdom.DocumentJS.Call("createElement", "div")
	v.Element.Set("style", "position:absolute")
	v.View = view
	v.SetObjectName("map")
	mapConstructor := zdom.CreateDotSeparatedObject("google.maps.Map")
	opts := map[string]interface{}{
		"zoom": zoom,
	}
	if !center.IsNull() {
		opts["center"] = makeLatLingJS(center)
	}
	v.MapJS = mapConstructor.New(v.Element, opts)
	v.minSize = zgeo.Size{300, 200}
}
