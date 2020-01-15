package zui

import (
	"strings"

	"github.com/torlangballe/zutil/zgeo"
)

type View interface {
	GetCalculatedSize(total zgeo.Size) zgeo.Size

	SetObjectName(name string) View
	GetObjectName() string

	Usable(usable bool) View
	IsUsable() bool

	Alpha(alpha float32) View
	GetAlpha() float32

	SetColor(color zgeo.Color) View   // Color is the main color of a view. If it is stroked and filled, it is fill color
	SetBGColor(color zgeo.Color) View // BGColor is all color in background of view, not just fill color
	CornerRadius(radius float64) View
	Stroke(width float64, color zgeo.Color) View

	CanFocus(can bool) View
	Focus(focus bool) View
	IsFocused() bool

	Opaque(opaque bool) View

	SetRect(rect zgeo.Rect) View
	Rect() zgeo.Rect

	Show(show bool) View
	IsShown() bool

	AddChild(child View, index int) // -1 is add to end
	RemoveChild(child View)
}

type NativeViewOwner interface {
	GetNative() *NativeView
}

func ViewGetNative(view View) *NativeView {
	o := view.(NativeViewOwner)
	if o != nil {
		return o.GetNative()
	}
	return nil
}

type ViewDrawProtocol interface {
	Draw(rect zgeo.Rect, canvas Canvas, view View)
}

type ExposableType interface {
	drawIfExposed()
	Expose()
}

type ReadyToShowType interface {
	ReadyToShow()
}

type ContainerType interface {
	GetChildren() []View
	ArrangeChildren(onlyChild *View)
	WhenLoaded(done func())
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

func ViewChild(v View, path string) View {
	parts := strings.Split(path, "/")
	name := parts[0]
	ct := v.(ContainerType)
	if ct != nil {
		for _, ch := range ct.GetChildren() {
			if ch.GetObjectName() == name {
				if len(parts) == 1 {
					return ch
				}
				path = strings.Join(parts[1:], "/")
				return ViewChild(ch, path)
			}
		}
	}
	return nil
}
