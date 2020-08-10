package zui

import "github.com/torlangballe/zutil/zgeo"

// WindowBarHeight is height of a normal window's title bar, can be different for each os
var WindowBarHeight = 21.0

type Window struct {
	windowNative
	HandleBeforeResized func(r zgeo.Rect) bool // HandleBeforeResize  is called before window re-arranges child view
	HandleAfterResized  func(r zgeo.Rect) bool // HandleAfterResize  is called after window re-arranges child view
}
