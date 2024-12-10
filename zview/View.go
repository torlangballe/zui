// The View interface is used to refer to any view. It is used by NativeViews to store the actual inherited type of NativeView they are.
// It has some minimum methods to perform on views, including *ObjectName*; a non-unique id used to identify it.
// The Native() method returns its *NativeView, which has a lot more methods.

//go:build zui

package zview

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zslice"
)

type View interface {
	CalculatedSize(total zgeo.Size) (s, max zgeo.Size)
	SetObjectName(name string)
	ObjectName() string
	SetUsable(usable bool)
	IsUsable() bool
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

type DownPressable interface {
	SetPressedDownHandler(id string, handler func())
}

type InteractiveSetter interface {
	SetInteractive(interactive bool)
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
	MinSize() zgeo.Size
}

type ToolTipAdder interface {
	GetToolTipAddition() string
}

type MaxSizeGettable interface {
	GetMaxSize() zgeo.Size
}

type ChildReplacer interface {
	ReplaceChild(child, with View)
}

type callback struct {
	id     int64
	delete func()
}

// viewCallbacks are arbitrary items associated with a view.
// They are given an id on add, and have a delete function.
// On view removal the function is called and the attachment removed.
// They are actually used to associate an event callback function to a view,
// removing it if the view or its window is deleted.

var viewCallbacks = map[View][]callback{}

func RegisterViewCallback(view View, id int64, del func()) {
	viewCallbacks[view] = append(viewCallbacks[view], callback{id, del})
	view.Native().AddOnRemoveFunc(func() {
		RemoveAViewCallback(view, id)
	})
}

func HasViewCallback(view View, id int64) bool {
	for _, c := range viewCallbacks[view] {
		if c.id == id {
			return true
		}
	}
	return false
}

func RemoveACallback(id int64) {
	for v := range viewCallbacks {
		if RemoveAViewCallback(v, id) {
			return
		}
	}
}

func RemoveAViewCallback(view View, id int64) bool {
	s, has := viewCallbacks[view]
	if has {
		for i, c := range viewCallbacks[view] {
			if c.id == id {
				if c.delete != nil {
					c.delete()
				}
				zslice.RemoveAt(&s, i)
				viewCallbacks[view] = s
				return true
			}
		}
	}
	return false
}

func RemoveViewCallbacks(view View) {
	for _, c := range viewCallbacks[view] {
		if c.delete != nil {
			c.delete()
		}
	}
	delete(viewCallbacks, view)
}

// ReplaceChild(child, with zview.View)

func ExposeView(v View) {
	et, _ := v.(ExposableType)
	if et != nil {
		et.Expose()
	}
}

func ObjectName(v View) string {
	if v == nil {
		return "<nil>"
	}
	return v.ObjectName()
}

type AnyValueSetter interface {
	SetValueWithAny(any)
}

type AnyValueGetter interface {
	ValueAsAny() any
}

type ValueHandler interface {
	SetValueHandler(id string, f func(edited bool))
}

type ValueHandlers struct {
	handlers map[string]func(edited bool)
}

func (vh *ValueHandlers) Add(id string, f func(edited bool)) {
	if f == nil {
		delete(vh.handlers, id)
	} else {
		if vh.handlers == nil {
			vh.handlers = map[string]func(edited bool){}
		}
		vh.handlers[id] = f
	}
}

func (vh ValueHandlers) CallAll(edited bool) {
	for _, f := range vh.handlers {
		f(edited)
	}
}

func (vh ValueHandlers) Count() int {
	return len(vh.handlers)
}
