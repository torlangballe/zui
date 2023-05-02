package zkeyboard

import (
	"github.com/torlangballe/zui/zdom"
	"syscall/js"
)

func GetKeyModFromEvent(event js.Value) KeyMod {
	var km KeyMod
	gkey := event.Get("which")
	if !gkey.IsUndefined() {
		km.Key = Key(gkey.Int())
	}
	if zdom.GetBoolIfDefined(event, "altKey") {
		km.Modifier |= ModifierAlt
	}
	if zdom.GetBoolIfDefined(event, "ctrlKey") {
		km.Modifier |= ModifierControl
	}
	if zdom.GetBoolIfDefined(event, "metaKey") || zdom.GetBoolIfDefined(event, "osKey") {
		km.Modifier |= ModifierCommand
	}
	if zdom.GetBoolIfDefined(event, "shiftKey") {
		km.Modifier |= ModifierShift
	}
	return km
}
