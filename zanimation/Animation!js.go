//go:build !js && zui

package zanimation

func Animate(view zui.View, secs float64, handler func(secsPos float64) bool) {}
