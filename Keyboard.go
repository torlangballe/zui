package zui

type KeyboardKey int
type KeyboardModifier int
type KeyboardType string
type KeyboardAutoCapType string
type KeyboardReturnKeyType string

const (
	KeyboardModifierShift KeyboardModifier = 1 << iota
	KeyboardModifierControl
	KeyboardModifierAlt
	KeyboardModifierMeta
)

const (
	KeyboardKeyReturn    = 13
	KeyboardKeyTab       = 9
	KeyboardKeyBackspace = 8
	KeyboardKeyEscape = 27
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

// Android: https://developer.android.com/reference/android/widget/TextView.html#attr_android:inputType
