package zkeyboard

type Key int
type Modifier int
type Type string
type AutoCapType string
type ReturnKeyType string

const (
	ModifierNone           = 0
	ModifierShift Modifier = 1 << iota
	ModifierControl
	ModifierAlt
	ModifierCommand
)

const (
	KeyReturn     = 13
	KeyEnter      = 131313 // not sure what it is elsewhere, doesn't exist in js/html
	KeyTab        = 9
	KeyBackspace  = 8
	KeyEscape     = 27
	KeyLeftArrow  = 37
	KeyRightArrow = 39
	KeyUpArrow    = 38
	KeyDownArrow  = 40
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
	TypeDecimalPad            Type = "decimal"
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
