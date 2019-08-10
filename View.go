package zgo

type View interface {
	GetView() *ViewNative

	GetCalculatedSize(total Size) Size

	ObjectName(name string) View
	GetObjectName() string

	Usable(usable bool) View
	IsUsable() bool

	Alpha(alpha float32) View
	GetAlpha() float32

	Color(color Color) View
	BGColor(color Color) View
	CornerRadius(radius float64) View
	Stroke(width float64, color Color) View

	Focus(focus bool) View
	IsFocused() bool

	Opaque(opaque bool) View

	Scale(scale float64) View
	GetScale() float64

	Show(show bool) View
	IsShown() bool
}

type AnyView interface {
	GetView() *ViewNative
}

type ViewValueChangedProtocol interface {
}

type ViewPressedProtocol interface {
	PressedInPos(pos Pos)
}

type ViewDrawProtocol interface {
	Draw(rect Rect, canvas Canvas, view View)
}

type ViewLayoutProtocol interface {
	HandleBeforeLayout()
	HandleAfterLayout()
	HandleTransitionedToSize()
	HandleClosing()
	HandleRotation()
	HandleBackButton() // only android has hardware back button...
	RefreshAccessibility()
}
