//go:build zui

package zcursor

type Type string

const (
	Alias       Type = "alias"
	AllScroll   Type = "all-scroll"
	Auto        Type = "auto"
	Cell        Type = "cell"
	ContextMenu Type = "context-menu"
	Copy        Type = "copy"
	Crosshair   Type = "crosshair"
	Default     Type = "default"
	Grab        Type = "grab"
	Grabbing    Type = "grabbing"
	Help        Type = "help"
	Move        Type = "move"

	ResizeRight       Type = "e-resize"
	ResizeTop         Type = "n-resize"
	ResizeLeft        Type = "w-resize"
	ResizeBottom      Type = "s-resize"
	ResizeTopLeft     Type = "nw-resize"
	ResizeTopRight    Type = "ne-resize"
	ResizeBottomLeft  Type = "sw-resize"
	ResizeBottomRight Type = "se-resize"

	TopBottomResize Type = "ns-resize"
	RowResize       Type = "row-resize"
	ColResize       Type = "col-resize"

	NoDrop     Type = "no-drop"
	None       Type = "none"
	NotAllowed Type = "not-allowed"
	Pointer    Type = "pointer"
	Progress   Type = "progress"
	Text       Type = "text"
	Wait       Type = "wait"
	ZoomIn     Type = "zoom-in"
	ZoomOut    Type = "zoom-out"
	// Url = "url(myBall.cur),auto"
)
