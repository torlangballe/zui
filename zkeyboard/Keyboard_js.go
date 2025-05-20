package zkeyboard

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zlog"
)

func GetKeyModFromEvent(event js.Value) KeyMod {
	var km KeyMod
	which := event.Get("which")
	if !which.IsUndefined() {
		km.Key = Key(which.Int())
	}
	key := event.Get("key")
	if !key.IsUndefined() {
		km.Char = key.String()
		if km.Key >= 'A' && km.Key <= 'Z' || km.Key >= 'a' && km.Key <= 'z' {
			km.Char = string(rune(km.Key))
			km.Key = 0
		}
	}
	switch km.Char {
	case "Control":
		km.Modifier = ModifierControl
		km.Key = KeyNone
		km.Char = ""
	case "Shift":
		km.Modifier = ModifierShift
		km.Key = KeyNone
		km.Char = ""
	case "Meta":
		km.Modifier = ModifierCommand
		km.Key = KeyNone
		km.Char = ""
	case "Alt":
		km.Modifier = ModifierAlt
		km.Key = KeyNone
		km.Char = ""
	}
	if zdom.GetBoolIfDefined(event, "altKey") {
		km.Modifier |= ModifierAlt
	}
	if zdom.GetBoolIfDefined(event, "ctrlKey") {
		km.Modifier |= ModifierControl
	}
	zlog.Info("KM:", km.Char, "key:", km.Key, "mod1:", km.Modifier, "ascii:", string(rune(km.Key)))
	if zdom.GetBoolIfDefined(event, "metaKey") || zdom.GetBoolIfDefined(event, "osKey") {
		km.Modifier |= MetaModifier
	}
	if zdom.GetBoolIfDefined(event, "shiftKey") {
		km.Modifier |= ModifierShift
	}
	return km
}
