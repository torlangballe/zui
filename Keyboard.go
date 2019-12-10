package zgo

type KeyboardKey int
type KeyboardModifier int

const (
	KeyboardModifierShift KeyboardModifier = 1 << iota
	KeyboardModifierControl
	KeyboardModifierAlt
	KeyboardModifierCommand
)

const (
	KeyboardKeyReturn    = 13
	KeyboardKeyTab       = 9
	KeyboardKeyBackspace = 8
)

