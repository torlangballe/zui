// +build !js,zui

package zui

// #cgo LDFLAGS: -framework CoreVideo
// #cgo LDFLAGS: -framework Foundation
// #cgo LDFLAGS: -framework AppKit
// #include <CoreGraphics/CoreGraphics.h>
// typedef struct ScreenInfo {
//     CGRect frame, visibleFrame;
//     int scale;
//     int ismain;
//     long sid;
// } ScreenInfo;
// int GetAllScreens(struct ScreenInfo *sis, int max);
// void SetMainScreenResolutionWithinWidths(long minw, long minh, long maxw, long maxh);
import "C"

import (
	"unsafe"

	"github.com/torlangballe/zutil/zgeo"
)

/*
func ScreenGetAll() (screens []Screen) {
	var count C.uint32_t = 0
	C.CGGetActiveDisplayList(0, nil, &count)
	mainID := C.CGMainDisplayID()
	ids := make([]C.CGDirectDisplayID, count)
	if C.CGGetActiveDisplayList(C.uint32_t(count), (*C.CGDirectDisplayID)(unsafe.Pointer(&ids[0])), nil) != C.kCGErrorSuccess {
		zlog.Fatal(nil, "getting display list")
	}
	for _, id := range ids {
		var s Screen
		bounds := C.CGDisplayBounds(id)
		w := C.CGDisplayPixelsWide(id)
		zlog.Info("SCREEN:", id, bounds.size.width, int(w))
		s.Rect = zgeo.RectFromXYWH(float64(bounds.origin.x), float64(bounds.origin.y), float64(bounds.size.width), float64(bounds.size.height))
		s.IsMain = (id == mainID)
		screens = append(screens, s)
	}
	return
}
*/

func ScreenGetAll() (screens []Screen) {
	var count C.uint32_t = 0
	C.CGGetActiveDisplayList(0, nil, &count)
	if count == 0 {
		return
	}
	cscreens := make([]C.ScreenInfo, count)
	p := (*C.ScreenInfo)(unsafe.Pointer(&cscreens[0]))
	c := int(C.GetAllScreens(p, C.int(count)))
	for i := 0; i < c; i++ {
		var s Screen
		si := cscreens[i]
		s.ID = int64(si.sid)
		s.Rect = zgeo.RectFromXYWH(float64(si.frame.origin.x), float64(si.frame.origin.y), float64(si.frame.size.width), float64(si.frame.size.height))
		s.UsableRect = zgeo.RectFromXYWH(float64(si.visibleFrame.origin.x), float64(si.visibleFrame.origin.y), float64(si.visibleFrame.size.width), float64(si.visibleFrame.size.height))
		s.Scale = float64(si.scale)
		s.IsMain = (si.ismain == 1)
		s.SoftScale = 1
		screens = append(screens, s)
	}
	return
}

// SetMainScreenResolutionMin goes through the display modes of the main screen, and finds the smallest width
// one that is as big as minWidth, and sets that.
func SetMainScreenResolutionWithinWidths(min, max zgeo.Size) {
	ms := ScreenMain().Rect.Size
	if max.IsNull() {
		max = min
	}
	if ms.Contains(min) && max.Contains(ms) {
		return
	}
	C.SetMainScreenResolutionWithinWidths(C.long(min.W), C.long(min.H), C.long(max.W), C.long(max.H))
}
