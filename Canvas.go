package zgo

//  Created by Tor Langballe on /21/10/15.

type Canvas struct {
	canvasNative
	currentMatrix Matrix // is currentTransform...
}

type CanvasBlendMode int

const (
	CanvasBlendModeNormal CanvasBlendMode = iota
	CanvasBlendModeMultiply
	CanvasBlendModeScreen
	CanvasBlendModeOverlay
	CanvasBlendModeDarken
	CanvasBlendModeLighten
	CanvasBlendModeColorDodge
	CanvasBlendModeColorBurn
	CanvasBlendModeSoftLight
	CanvasBlendModeHardLight
	CanvasBlendModeDifference
	CanvasBlendModeExclusion
	CanvasBlendModeHue
	CanvasBlendModeSaturation
	CanvasBlendModeColor
	CanvasBlendModeLuminosity
)
