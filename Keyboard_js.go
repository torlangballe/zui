package zui

import "syscall/js"

func getKeyAndModsFromEvent(event js.Value) (key KeyboardKey, mods KeyboardModifier) {
	gkey := event.Get("which")
	if !gkey.IsUndefined() {
		key = KeyboardKey(gkey.Int())
	}
	if jsGetBoolIfDefined(event, "altKey") {
		mods |= KeyboardModifierAlt
	}
	if jsGetBoolIfDefined(event, "ctrlKey") {
		mods |= KeyboardModifierControl
	}
	if jsGetBoolIfDefined(event, "metaKey") || jsGetBoolIfDefined(event, "osKey") {
		mods |= KeyboardModifierCommand
	}
	if jsGetBoolIfDefined(event, "shiftKey") {
		mods |= KeyboardModifierShift
	}
	return
}

// func GetModifierKeysPressed() (mod KeyboardModifier) {

// 	if KeyboardEvent.getModifierState("Alt") {
// 		mod |= KeyboardModifierAlt
// 	}
// 	if event.getModifierState("Control") {
// 		mod |= KeyboardModifierControl
// 	}
// 	if event.getModifierState("Shift") {
// 		mod |= KeyboardModifierAlt
// 	}
// 	if event.getModifierState("Meta") || event.getModifierState("OS") {
// 		mod |= KeyboardModifierCommand
// 	}
// 	return
// }

// This should be done by functionality that exists in Window
// func SetGlobalKeyHandler(handler func(key KeyboardKey, mods KeyboardModifier) bool) {
// 	body := DocumentJS.Get("body")
// 	//!!	v.keyPressed = handler
// 	jsSetKeyHandler(body, handler)
// }
