package zkeyboard

import "github.com/torlangballe/zutil/zgeo"

type Key int
type Modifier int
type Type string
type AutoCapType string
type ReturnKeyType string

type KeyMod struct {
	Key      Key
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
	KeyControl    = 17
	KeyAlt        = 18
	KeyCommand    = 91
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

// ModifiersAtPress is set from  events before handlers are called.
// This is global to avoid passing mods in all pressed/longpressed handlers
var ModifiersAtPress Modifier

// Android: https://developer.android.com/reference/android/widget/TextView.html#attr_android:inputType

func KMod(k Key, m Modifier) KeyMod {
	return KeyMod{Key: k, Modifier: m}
}

func (k KeyMod) IsNull() bool {
	return k.Key == 0 && k.Modifier == 0
}

func GetModifiersString(m Modifier) string {
	var str string
	if m&ModifierShift != 0 {
		str += "⇧"
	}
	if m&ModifierControl != 0 {
		str += "^"
	}
	if m&ModifierAlt != 0 {
		str += "⎇"
	}
	if m&ModifierCommand != 0 {
		str += "⌘"
	}
	return str
}

func GetStringForKey(key Key) string {
	switch key {
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

func KeyIsReturnish(key Key) bool {
	return key == KeyReturn || key == KeyEnter
}
