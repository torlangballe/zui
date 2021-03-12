package zui

type KeyboardKey int
type KeyboardModifier int
type KeyboardType string
type KeyboardAutoCapType string
type KeyboardReturnKeyType string

const (
	KeyboardModifierNone                   = 0
	KeyboardModifierShift KeyboardModifier = 1 << iota
	KeyboardModifierControl
	KeyboardModifierAlt
	KeyboardModifierCommand
)

const (
	KeyboardKeyReturn    = 13
	KeyboardKeyEnter     = 131313 // not sure what it is elsewhere, doesn't exist in js/html
	KeyboardKeyTab       = 9
	KeyboardKeyBackspace = 8
	KeyboardKeyEscape    = 27
	KeyboardKeyUpArrow   = 38
	KeyboardKeyDownArrow = 40
)

const (
	KeyboardTypeDefault               KeyboardType = ""
	KeyboardTypeAsciiCapable          KeyboardType = "ascii"
	KeyboardTypeNumbersAndPunctuation KeyboardType = "numpunct"
	KeyboardTypeURL                   KeyboardType = "url"
	KeyboardTypeNumberPad             KeyboardType = "numpad"
	KeyboardTypePhonePad              KeyboardType = "phonepad"
	KeyboardTypeNamePhonePad          KeyboardType = "namephonepad"
	KeyboardTypeEmailAddress          KeyboardType = "email"
	KeyboardTypeDecimalPad            KeyboardType = "decimal"
	KeyboardTypeWebSearch             KeyboardType = "websearch"
	KeyboardTypeASCIICapableNumberPad KeyboardType = "ascinumber"
	KeyboardTypePassword              KeyboardType = "password" // set textfield.isSecureTextEntry on ios
)

const (
	KeyboardAutoCapNone          KeyboardAutoCapType = ""
	KeyboardAutoCapWords         KeyboardAutoCapType = "words"
	KeyboardAutoCapSentences     KeyboardAutoCapType = "sentences"
	KeyboardAutoCapAllCharacters KeyboardAutoCapType = "all"
)

const (
	KeyboardReturnKeyDefault  KeyboardReturnKeyType = ""
	KeyboardReturnKeyDone     KeyboardReturnKeyType = "done"
	KeyboardReturnKeyGo       KeyboardReturnKeyType = "go"
	KeyboardReturnKeyNext     KeyboardReturnKeyType = "next"
	KeyboardReturnKeySend     KeyboardReturnKeyType = "send"
	KeyboardReturnKeySearch   KeyboardReturnKeyType = "search"
	KeyboardReturnKeyContinue KeyboardReturnKeyType = "continue"
)

// KeyboardModifiersAtPress is set from keyboard events before handlers are called.
// This is global to avoid passing mods in all pressed/longpressed handlers
var KeyboardModifiersAtPress KeyboardModifier

// Android: https://developer.android.com/reference/android/widget/TextView.html#attr_android:inputType
