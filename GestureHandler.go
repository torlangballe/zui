package zgo

type GestureHandlerType int
type GestureHandlerState int

const (
	GestureHandlerTap       = 1
	GestureHandlerLongpress = 2
	GestureHandlerPan       = 4
	GestureHandlerPinch     = 8
	GestureHandlerSwipe     = 16
	GestureHandlerRotation  = 32
)

const (
	GestureHandlerBegan    = 1
	GestureHandlerEnded    = 2
	GestureHandlerChanged  = 4
	GestureHandlerPossible = 8
	GestureHandlerCanceled
	GestureHandlerFailed = 32
)

type GestureHandler interface {
	AddGestureTo(view View, gtype GestureHandlerType, taps int, touches int, duration float32, movement float32, dir Alignment)
	HandleGestureType(gtype GestureHandlerType, view View, pos Pos, delta Pos, state GestureHandlerState, taps int, touches int, dir Alignment, velocity Pos, gvalue float32, name string) bool
}
