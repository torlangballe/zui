//go:build !js && zui

package zcustom

import "github.com/torlangballe/zui/zview"

func (v *CustomView) drawSelf()                         {}
func (v *CustomView) makeCanvas()                       {}
func (v *CustomView) ReadyToShow(beforeWindow bool)     {}
func (v *CustomView) ExposeIn(secs float64)             {}
func (v *CustomView) Init(view zview.View, name string) { v.SetObjectName(name) }
func (v *CustomView) Expose()                           { v.ExposeIn(0.1) }
