//go:build zui

package zmap

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

// https://developers.google.com/maps/documentation/javascript/maxzoom
// https://stackoverflow.com/questions/load-google-map-without-using-callback-method
// Add somewhere: <script src="https://maps.googleapis.com/maps/api/js?key=xxxx&callback=googleMapCreated"></script>

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

func (v *MapView) Init(view zview.View, center zgeo.Pos, zoom int) {
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
