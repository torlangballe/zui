//go:build !js && zui

package zanimation

import "github.com/torlangballe/zui/zview"

func Animate(view zview.View, secs float64, handler func(secsPos float64) bool) {}
