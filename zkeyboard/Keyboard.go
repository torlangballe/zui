package zkeyboard

import (
	"strings"

	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zgeo"
)

type Key int
type Modifier int
type Type string
type AutoCapType string
type ReturnKeyType string

type KeyMod struct {
	Key      Key
	Char     string
	Modifier Modifier
}

type ShortcutHandler interface {
	HandleOutsideShortcut(sc KeyMod) bool
}

type KeyConsumer interface {
	ConsumesKey(sc KeyMod) bool
}

const (
	ModifierNone  Modifier = 0
	ModifierShift Modifier = 1 << iota
	ModifierControl
	ModifierAlt
	ModifierCommand
)

const (
	KeyNone       = 0
	KeyReturn     = 13
	KeyEnter      = 131313 // not sure what it is elsewhere, doesn't exist in js/html
	KeyTab        = 9
	KeyBackspace  = 8
	KeyDelete     = 127
	KeyEscape     = 27
	KeyLeftArrow  = 37
	KeyRightArrow = 39
	KeyUpArrow    = 38
	KeyDownArrow  = 40
	KeyShiftKey   = 16
	KeyControlKey = 17
	KeyAltKey     = 18
	KeyCommandKey = 91
	KeyPageUp     = 33
	KeyPageDown   = 34
	KeyEnd        = 35
	KeyHome       = 36
)

const (
	TypeDefault               Type = ""
	TypeAsciiCapable          Type = "ascii"
	TypeNumbersAndPunctuation Type = "numpunct"
	TypeURL                   Type = "url"
	TypeNumberPad             Type = "numpad"
	TypePhonePad              Type = "phonepad"
	TypeNamePhonePad          Type = "namephonepad"
	TypeEmailAddress          Type = "email"
	TypeInteger               Type = "integer"
	TypeFloat                 Type = "float"
	TypeWebSearch             Type = "websearch"
	TypeASCIICapableNumberPad Type = "ascinumber"
	TypePassword              Type = "password" // set textfield.isSecureTextEntry on ios
)

const (
	AutoCapNone          AutoCapType = ""
	AutoCapWords         AutoCapType = "words"
	AutoCapSentences     AutoCapType = "sentences"
	AutoCapAllCharacters AutoCapType = "all"
)

const (
	ReturnKeyDefault  ReturnKeyType = ""
	ReturnKeyDone     ReturnKeyType = "done"
	ReturnKeyGo       ReturnKeyType = "go"
	ReturnKeyNext     ReturnKeyType = "next"
	ReturnKeySend     ReturnKeyType = "send"
	ReturnKeySearch   ReturnKeyType = "search"
	ReturnKeyContinue ReturnKeyType = "continue"
)

var (
	ModifiersAtPress    Modifier // ModifiersAtPress is set from  events before handlers are called. This is global to avoid passing mods in all pressed/longpressed handlers
	MetaModifier        = ModifierControl
	AltModifierName     = "Alt"
	CommandModifierName = "Meta"
	CurrentKeyDown      KeyMod
)

func init() {
	if zdevice.OS() == zdevice.MacOSType || zdevice.OS() == zdevice.IOSType {
		MetaModifier = ModifierCommand
		AltModifierName = "Option"
		CommandModifierName = "Command"
	}
}

// type Modifiers struct {
// 	AltName    string
// 	AltSymbol  string
// 	MetaName   string
// 	MetaSymbol string
// }
// func ModifiersForOS(os zdevice.OSType) Modifiers {
// 	if os == zdevice.MacOSType || os == zdevice.IOSType {
// 		return Modifiers{
// 			AltName:    "Option",
// 			AltSymbol:  "⎇",
// 			MetaName:   "Command",
// 			MetaSymbol: "⌘",
// 		}
// 	}
// 	return Modifiers{
// 		AltName:    "Alt",
// 		AltSymbol:  "alt-",
// 		MetaName:   "Control",
// 		MetaSymbol: "ctrl-",
// 	}
// }

// Android: https://developer.android.com/reference/android/widget/TextView.html#attr_android:inputType

func KMod(k Key, m Modifier) KeyMod {
	return KeyMod{Key: k, Modifier: m}
}

func (k KeyMod) IsNull() bool {
	return k.Key == 0 && k.Modifier == 0
}

func (k KeyMod) Matches(m KeyMod) bool {
	if k.Char != "" {
		return k.Char == m.Char
	}
	if k.Modifier != m.Modifier {
		return false
	}
	return k.Key == m.Key
}

func (m Modifier) IsNull() bool {
	return m == ModifierNone
}

func (m Modifier) String() string {
	var parts []string
	if m&ModifierShift != 0 {
		parts = append(parts, "shift")
	}
	if m&ModifierControl != 0 {
		parts = append(parts, "control")
	}
	if m&ModifierAlt != 0 {
		parts = append(parts, "alt")
	}
	if m&ModifierCommand != 0 {
		parts = append(parts, "command")
	}
	return strings.Join(parts, "|")
}

func (m Modifier) HumanString() string {
	switch m {
	case ModifierAlt:
		return AltModifierName
	case ModifierShift:
		return "Shift"
	case ModifierControl:
		return "Control"
	case ModifierCommand:
		return CommandModifierName
	}
	return ""
}

func (m Modifier) AsSymbols() string {
	var str string
	if m&ModifierShift != 0 {
		str += "⇧"
	}
	if m&ModifierControl != 0 {
		if zdevice.OS() == zdevice.MacOSType {
			str += "^"
		} else {
			str += "ctrl-"
		}
	}
	if m&ModifierAlt != 0 {
		if zdevice.OS() == zdevice.MacOSType {
			str += "⎇"
		} else {
			str += "alt-"
		}
	}
	if m&ModifierCommand != 0 {
		if zdevice.OS() == zdevice.MacOSType {
			str += "⌘"
		} else {
			str += "ctrl-"
		}
	}
	return str
}

func (key Key) AsString() string {
	switch key {
	case ' ':
		return "space"
	case KeyReturn, KeyEnter:
		return "⏎"
	case KeyEscape:
		return "␛"
	case KeyDelete:
		return "␐"
	case KeyBackspace:
		return "⌫"
	case KeyLeftArrow:
		return "←"
	case KeyRightArrow:
		return "→"
	case KeyUpArrow:
		return "↑"
	case KeyDownArrow:
		return "↓"
	case KeyTab:
		return "⇥"
	}
	return string(rune(key))
}

func ArrowKeyToDirection(key Key) zgeo.Alignment {
	switch key {
	case KeyLeftArrow:
		return zgeo.Left
	case KeyRightArrow:
		return zgeo.Right
	case KeyUpArrow:
		return zgeo.Top
	case KeyDownArrow:
		return zgeo.Bottom
	}
	return zgeo.AlignmentNone
}

func (k Key) IsReturnish() bool {
	return k == KeyReturn || k == KeyEnter
}

func (km KeyMod) AsString() string {
	str := km.Modifier.AsSymbols()
	if km.Key != 0 {
		str += km.Key.AsString()
	} else {
		str += km.Char
	}
	return str
}
