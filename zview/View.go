// The View interface is used to refer to any view. It is used by NativeViews to store the actual inherited type of NativeView they are.
// It has some minimum methods to perform on views, including *ObjectName*; a non-unique id used to identify it.
// The Native() method returns its *NativeView, which has a lot more methods.

//go:build zui

package zview

import (
	"github.com/torlangballe/zutil/zgeo"
)

type View interface {
	CalculatedSize(total zgeo.Size) zgeo.Size
	SetObjectName(name string)
	ObjectName() string
	SetUsable(usable bool)
	Usable() bool
	SetColor(color zgeo.Color)   // Color is the main color of a view. If it is stroked and filled, it is fill color
	SetBGColor(color zgeo.Color) // BGColor is all color in background of view, not just fill color
	SetRect(rect zgeo.Rect)
	Rect() zgeo.Rect
	Show(show bool)
	Native() *NativeView
}

// ExposableType is an interface for exposing views.
type ExposableType interface {
	Expose()
}

// A Marginalizer allows setting a margin Rect. Although all views that inherit from NativeView get a
// SetMargin method, they might have their own method, so always convert a View to Marginalizer to set margins.
type Marginalizer interface {
	SetMargin(m zgeo.Rect)
}

// A MarginOwner has a margin Rect.
type MarginOwner interface {
	Margin() zgeo.Rect
}

type ReadyToShowType interface {
	ReadyToShow(beforeWindow bool)
}

type Pressable interface {
	SetPressedHandler(handler func())
	SetLongPressedHandler(handler func())
	PressedHandler() func()
	LongPressedHandler() func()
}

type DownPressable interface {
	SetPressedDownHandler(handler func())
}

// Layouter ...not really using this yet left over from iOS stuff
type Layouter interface {
	HandleBeforeLayout()
	HandleAfterLayout()
	HandleTransitionedToSize()
	HandleClosing()
	HandleRotation()
	HandleBackButton() // only android has hardware back button...
	RefreshAccessibility()
}

type MinSizeGettable interface {
	GetMinSize(s zgeo.Size)
}

type MaxSizeGettable interface {
	GetMaxSize() zgeo.Size
}

func ExposeView(v View) {
	et, _ := v.(ExposableType)
	if et != nil {
		et.Expose()
	}
}
