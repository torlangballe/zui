package zkeyboard

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zstr"
)

func ModifierFromEvent(event js.Value) Modifier {
	var m Modifier
	// which := event.Get("which")
	// if !which.IsUndefined() {
	// 	km.Key = Key(which.Int())
	// }
	if zdom.GetBoolIfDefined(event, "altKey") {
		m |= ModifierAlt
	}
	if zdom.GetBoolIfDefined(event, "ctrlKey") {
		m |= ModifierControl
	}
	if zdom.GetBoolIfDefined(event, "metaKey") || zdom.GetBoolIfDefined(event, "osKey") {
		if zdevice.OS() == zdevice.MacOSType || zdevice.OS() == zdevice.IOSType {
			m |= ModifierCommand
		} else {
			m |= ModifierMeta
		}
	}
	if zdom.GetBoolIfDefined(event, "shiftKey") {
		m |= ModifierShift
	}
	return m
}

func GetKeyModFromEvent(event js.Value) KeyMod {
	var km KeyMod
	var c string

	km.Modifier = ModifierFromEvent(event)
	// zlog.Info("KEY:", km.Modifier, ModifierMenu, ModifierMeta, ModifierCommand)
	key := event.Get("key").String()
	switch key {
	case "+":
		km.Key = KeyPlus
		km.Modifier = 0
	case "-":
		km.Key = KeyMinus
		km.Modifier = 0
	}
	if km.Key != 0 {
		return km
	}
	scode := event.Get("code").String()
	for _, pre := range []string{"Key", "Digit"} {
		if zstr.HasPrefix(scode, pre, &c) && len(c) == 1 {
			km.Key = Key(c[0])
			return km
		}
	}
	switch scode {
	case "ControlLeft":
		km.Modifier = ModifierControl
		km.Key = KeyControlLeft
	case "ControlRight":
		km.Modifier = ModifierControl
		km.Key = KeyControlRight
	case "ShiftLeft":
		km.Modifier = ModifierShift
		km.Key = KeyShiftLeft
	case "ShiftRight":
		km.Modifier = ModifierShift
		km.Key = KeyShiftRight
	case "MetaLeft":
		km.Modifier = ModifierCommand
		km.Key = KeyCommandLeft
	case "MetaRight":
		km.Modifier = ModifierCommand
		km.Key = KeyCommandRight
	case "AltLeft":
		km.Modifier = ModifierAlt
		km.Key = KeyAltLeft
	case "AltRight":
		km.Modifier = ModifierAlt
		km.Key = KeyAltRight
	case "ArrowLeft":
		km.Key = KeyLeftArrow
	case "ArrowRight":
		km.Key = KeyRightArrow
	case "ArrowUp":
		km.Key = KeyUpArrow
	case "ArrowDown":
		km.Key = KeyDownArrow
	case "Return":
		km.Key = KeyReturn
	case "Enter":
		km.Key = KeyEnter
	case "Tab":
		km.Key = KeyTab
	case "Backspace":
		km.Key = KeyBackspace
	case "Space":
		km.Key = KeySpace
	case "Delete":
		km.Key = KeyDelete
	case "Escape":
		km.Key = KeyEscape
	case "PageUp":
		km.Key = KeyPageUp
	case "PageDown":
		km.Key = KeyPageDown
	case "End":
		km.Key = KeyEnd
	case "Home":
		km.Key = KeyHome
	}
	return km
}
