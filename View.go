// +build zui

package zui

import (
	"strings"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type View interface {
	CalculatedSize(total zgeo.Size) zgeo.Size

	SetObjectName(name string) View
	ObjectName() string

	SetUsable(usable bool) View
	Usable() bool

	SetAlpha(alpha float32) View
	Alpha() float32

	SetColor(color zgeo.Color) View   // Color is the main color of a view. If it is stroked and filled, it is fill color
	SetBGColor(color zgeo.Color) View // BGColor is all color in background of view, not just fill color
	SetCorner(radius float64) View
	SetStroke(width float64, color zgeo.Color) View

	SetCanFocus(can bool) View
	Focus(focus bool) View
	IsFocused() bool

	Opaque(opaque bool) View

	SetRect(rect zgeo.Rect)
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
	o, _ := view.(NativeViewOwner)
	if o != nil {
		return o.GetNative()
	}
	zlog.Error(nil, "no view", view != nil, zlog.GetCallingStackString())
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
	ReadyToShow(beforeWindow bool)
}

type Pressable interface {
	SetPressedHandler(handler func())
	SetLongPressedHandler(handler func())
	PressedHandler() func()
	LongPressedHandler() func()
}

// ViewLayoutProtocol ...not really using this yet left over from iOS stuff
type ViewLayoutProtocol interface {
	HandleBeforeLayout()
	HandleAfterLayout()
	HandleTransitionedToSize()
	HandleClosing()
	HandleRotation()
	HandleBackButton() // only android has hardware back button...
	RefreshAccessibility()
}

func ExposeView(v View) {
	et, _ := v.(ExposableType)
	if et != nil {
		et.Expose()
	}
}

func ViewChild(v View, path string) View {
	if path == "" {
		return v
	}
	parts := strings.Split(path, "/")
	// zlog.Info("ViewChild:", v.ObjectName(), path)
	name := parts[0]
	if name == ".." {
		path = strings.Join(parts[1:], "/")
		parent := ViewGetNative(v).Parent()
		if parent != nil {
			return ViewChild(parent, path)
		}
		return nil
	}
	ct := v.(ContainerType)
	if ct != nil {
		for _, ch := range ct.GetChildren() {
			// zlog.Info("Childs:", name, "'"+ch.ObjectName()+"'")
			if name == "*" || ch.ObjectName() == name {
				if len(parts) == 1 {
					return ch
				}
				path = strings.Join(parts[1:], "/")
				gotView := ViewChild(ch, path)
				if gotView != nil || name != "*" {
					return gotView
				}
			}
		}
	}
	return nil
}
