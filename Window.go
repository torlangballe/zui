package zui

import (
	"github.com/torlangballe/zutil/zgeo"
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
}

type WindowOptions struct {
	URL  string
	ID   string
	Pos  *zgeo.Pos
	Size zgeo.Size
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
