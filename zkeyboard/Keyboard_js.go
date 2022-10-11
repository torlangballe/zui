package zkeyboard

import (
	"github.com/torlangballe/zui/zdom"
	"syscall/js"
)

func GetKeyAndModsFromEvent(event js.Value) (key Key, mods Modifier) {
	gkey := event.Get("which")
	if !gkey.IsUndefined() {
		key = Key(gkey.Int())
	}
	if zdom.GetBoolIfDefined(event, "altKey") {
		mods |= ModifierAlt
	}
	if zdom.GetBoolIfDefined(event, "ctrlKey") {
		mods |= ModifierControl
	}
	if zdom.GetBoolIfDefined(event, "metaKey") || zdom.GetBoolIfDefined(event, "osKey") {
		mods |= ModifierCommand
	}
	if zdom.GetBoolIfDefined(event, "shiftKey") {
		mods |= ModifierShift
	}
	return
}
