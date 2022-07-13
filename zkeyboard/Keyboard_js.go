package zkeyboard

import (
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zlog"
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

func SetKeyHandler(e js.Value, handler func(key Key, mods Modifier) bool) {
	//!!	v.keyPressed = handler
	e.Set("onkeyup", js.FuncOf(func(val js.Value, args []js.Value) interface{} {
		zlog.Info("KeyUp")
		if handler != nil {
			event := args[0]
			key, mods := GetKeyAndModsFromEvent(event)
			if handler(key, mods) {
				event.Call("stopPropagation")
			}
			// zlog.Info("KeyUp:", key, mods)
		}
		return nil
	}))
}
