//go:build zui

package zcursor

type Type string

const (
	Alias       Type = "alias"
	AllScroll   Type = "all-scroll"
	Auto        Type = "auto"
	Cell        Type = "cell"
	ContextMenu Type = "context-menu"
	ColResize   Type = "col-resize"
	Copy        Type = "copy"
	Crosshair   Type = "crosshair"
	Default     Type = "default"
	EResize     Type = "e-resize"
	Grab        Type = "grab"
	Grabbing    Type = "grabbing"
	Help        Type = "help"
	Move        Type = "move"
	NResize     Type = "n-resize"
	NsResize    Type = "ns-resize"
	NoDrop      Type = "no-drop"
	None        Type = "none"
	NotAllowed  Type = "not-allowed"
	Pointer     Type = "pointer"
	Progress    Type = "progress"
	RowResize   Type = "row-resize"
	SResize     Type = "s-resize"
	Text        Type = "text"
	WResize     Type = "w-resize"
	Wait        Type = "wait"
	ZoomIn      Type = "zoom-in"
	ZoomOut     Type = "zoom-out"
	// Url = "url(myBall.cur),auto"
)
