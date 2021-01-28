package zui

import "syscall/js"

func getKeyAndModsFromEvent(event js.Value) (key KeyboardKey, mods KeyboardModifier) {
	key = KeyboardKey(event.Get("which").Int())
	if event.Get("altKey").Bool() {
		mods |= KeyboardModifierAlt
	}
	if event.Get("ctrlKey").Bool() {
		mods |= KeyboardModifierControl
	}
	if event.Get("metaKey").Bool() {
		mods |= KeyboardModifierMeta
	}
	if event.Get("shiftKey").Bool() {
		mods |= KeyboardModifierShift
	}
	return
}

func SetGlobalKeyHandler(handler func(key KeyboardKey, mods KeyboardModifier)) {
	body := DocumentJS.Get("body")
	//!!	v.keyPressed = handler
	jsSetKeyHandler(body, handler)
}
