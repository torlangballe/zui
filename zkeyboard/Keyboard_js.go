package zkeyboard

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zstr"
)

func GetKeyModFromEvent(event js.Value) KeyMod {
	var km KeyMod
	gkey := event.Get("which")
	km.Char = event.Get("key").String()
	if !gkey.IsUndefined() {
		km.Key = Key(gkey.Int())
	}
	if zstr.StringsContain([]string{"Control", "Shift", "Meta", "Alt"}, km.Char) {
		km.Key = KeyNone
		km.Char = ""
	}
	if zdom.GetBoolIfDefined(event, "altKey") {
		km.Modifier |= ModifierAlt
	}
	if zdom.GetBoolIfDefined(event, "ctrlKey") {
		km.Modifier |= ModifierControl
	}
	if zdom.GetBoolIfDefined(event, "metaKey") || zdom.GetBoolIfDefined(event, "osKey") {
		km.Modifier |= MetaModifier
	}
	if zdom.GetBoolIfDefined(event, "shiftKey") {
		km.Modifier |= ModifierShift
	}
	return km
}
