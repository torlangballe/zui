package zui

import (
	"math"
	"net/url"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zstr"
)

// WindowBarHeight is height of a normal window's title bar, can be different for each os
var WindowBarHeight = 21.0

var windows = map[*Window]bool{}

type Window struct {
	windowNative
	HandleClosed        func()
	HandleBeforeResized func(r zgeo.Rect) bool // HandleBeforeResize  is called before window re-arranges child view
	HandleAfterResized  func(r zgeo.Rect) bool // HandleAfterResize  is called after window re-arranges child view
	ID                  string
	ProgrammaticView    View // this is set if the window has zui views added to it. If from URL, it is nil
	dismissed           bool // this stores if window is dismissed or closed for other reasons, used by present close functions
	keyHandlers         map[View]func(KeyboardKey, KeyboardModifier)
}

type WindowOptions struct {
	URL       string
	ID        string
	Pos       *zgeo.Pos
	Size      zgeo.Size
	Alignment zgeo.Alignment
}

func WindowNew() *Window {
	w := &Window{}
	w.keyHandlers = map[View]func(KeyboardKey, KeyboardModifier){}
	return w
}

// if WindowExistsActivate finds an open window in windows with id == winID it activates it and returns true
// This can be used to decide if to create a window or not if it already exists
func WindowExistsActivate(winID string) bool {
	for w, _ := range windows {
		if w.ID == winID {
			w.Activate()
			return true
		}
	}
	return false
}

func (win *Window) GetURLWithNewPathAndArgs(spath string, args zdict.Dict) string {
	u, err := url.Parse(win.GetURL())
	zlog.OnError(err)
	// spath = zstr.Concat("/", zrest.AppURLPrefix, spath)
	u.Path = spath //path.Join(zrest.AppURLPrefix, spath)
	// zlog.Info("GetURLWithNewPathAndArgs:", spath, args, u)
	q := args.ToURLValues()
	u.RawQuery = q.Encode()
	return u.String()
}

func (win *Window) SetPathAndArgs(path string, args zdict.Dict) {
	surl := win.GetURLWithNewPathAndArgs(path, args)
	win.SetLocation(surl)
}

func (win *Window) GetScreen() {

}
func getRectFromOptions(o WindowOptions) (rect zgeo.Rect, gotPos, gotSize bool) {
	size := o.Size
	if o.Alignment != zgeo.AlignmentNone {
		zlog.Assert(!o.Size.IsNull())
		wrects := []zgeo.Rect{WindowGetMain().Rect()}
		srect := ScreenMain().Rect
		for w := range windows {
			wrects = append(wrects, w.Rect())
		}
		// orects = append(orects, srect.PlusPos(zgeo.Pos{0, srect.Size.H}))
		// orects = append(orects, srect.PlusPos(zgeo.Pos{0, -srect.Size.H}))
		// orects = append(orects, srect.PlusPos(zgeo.Pos{srect.Size.W, 0}))
		// orects = append(orects, srect.PlusPos(zgeo.Pos{-srect.Size.W, 0}))
		// zlog.Info("getRectFromOptions:", len(windows))
		var minSum float64
		for _, ai := range o.Alignment.SplitIntoIndividual() {
			for _, wr := range wrects {
				b4 := wr.Align(size, ai, zgeo.Size{}, zgeo.Size{})
				// zlog.Info("RECT:", wr, ai, b4)
				r := b4.MovedInto(srect)
				var sumArea float64
				for _, or := range wrects {
					s := math.Max(0, or.Intersected(r).Size.Area())
					sumArea += s
				}
				if rect.IsNull() || sumArea < minSum {
					minSum = sumArea
					rect = r
				}
				if sumArea <= 0 {
					break
				}
			}
		}
		gotPos = true
		gotSize = true
	} else {
		if o.Pos != nil {
			rect.Pos = *o.Pos
		}
		rect.Size = o.Size
		gotPos = (o.Pos != nil)
		gotSize = !o.Size.IsNull()
	}
	return
}

func (win *Window) SetAddressBarPathAndArgs(spath string, args zdict.Dict) {
	// zlog.Info("WindowSetAddressBarPathAndArgs:", path, args)
	spath = zstr.Concat("/", zrest.AppURLPrefix, spath)
	surl := win.GetURLWithNewPathAndArgs(spath, args)
	win.SetAddressBarURL(surl)
}
