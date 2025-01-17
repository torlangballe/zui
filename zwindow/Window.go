//go:build zui

package zwindow

import (
	"net/url"

	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

type Window struct {
	windowNative
	HandleClosed        func()
	HandleBeforeResized func(r zgeo.Rect) // HandleBeforeResize  is called before window re-arranges child view
	HandleAfterResized  func(r zgeo.Rect) // HandleAfterResize  is called after window re-arranges child view
	ID                  string
	ProgrammaticView    zview.View // this is set if the window has zui views added to it. If from URL, it is nil
	ViewsStack          []zview.View
	Scale               float64

	ResizeHandlingView zview.View
	resizeTimer        *ztimer.Timer
	dismissed          bool // this stores if window is dismissed or closed for other reasons, used by present close functions
	callbackIDs        []int64
}

type Options struct {
	URL          string
	ID           string
	Pos          *zgeo.Pos
	Size         zgeo.Size
	Alignment    zgeo.Alignment
	FullScreenID int64 // screen id to go full screen on. -1 is use main. 0 is ignore.
}

var (
	windows                          = map[*Window]bool{}
	PresentedViewCurrentIsParentFunc func(v zview.View) bool
	barHeight                        = 28.0
	winMain                          *Window
	barCalculated                    bool
)

func GetMain() *Window {
	return winMain
}

// BarHeight is height of a normal window's title bar, can be different for each os
func BarHeight() float64 {
	return barHeight
}

func New() *Window {
	w := &Window{}
	return w
}

// if ExistsActivate finds an open window in windows with id == winID it activates it and returns true
// This can be used to decide if to create a window or not if it already exists
func ExistsActivate(winID string) bool {
	for w, _ := range windows {
		// zlog.Info("ExistsActivate:", w.ID, "==", winID)
		if w.ID == winID {
			w.Activate(true)
			return true
		}
	}
	return false
}

func FindWithID(winID string) *Window {
	for w := range windows {
		// zlog.Info("ExistsActivate:", w.ID, "==", winID)
		if w.ID == winID {
			return w
		}
	}
	return nil
}

func setValues(v url.Values, add url.Values) {
	for k, ss := range add {
		for _, s := range ss {
			v.Set(k, s)
		}
	}
}

func (win *Window) GetURLWithNewPathAndArgs(spath string, args zdict.Dict) string {
	uBase, err := url.Parse(win.GetURL())
	if zlog.OnError(err) {
		return ""
	}
	// spath = zstr.Concat("/", zrest.AppURLPrefix, spath)
	// u.Path = spath //path.Join(zrest.AppURLPrefix, spath)
	uAdd, err := url.Parse(spath)
	if zlog.OnError(err) {
		return ""
	}
	uBase.Path = uAdd.Path
	vals := url.Values{}
	setValues(vals, uAdd.Query())
	setValues(vals, args.ToURLValues())
	uBase.RawQuery = vals.Encode()
	// zlog.Info("GetURLWithNewPathAndArgs:", uBase)
	return uBase.String()
}

func (win *Window) SetPathAndArgs(path string, args zdict.Dict) {
	surl := win.GetURLWithNewPathAndArgs(path, args)
	win.SetLocation(surl)
}

func (win *Window) SetAddressBarPathAndArgs(spath string, args zdict.Dict) {
	// zlog.Info("WindowSetAddressBarPathAndArgs:", path, args)
	spath = zstr.Concat("/", zrest.AppURLPrefix, spath)
	surl := win.GetURLWithNewPathAndArgs(spath, args)
	win.SetAddressBarURL(surl)
}

func TopView(win *Window) zview.View {
	if win == nil {
		win = Current()
	}
	return win.ViewsStack[len(win.ViewsStack)-1]
}
