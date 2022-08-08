//go:build !js && zui

package zview

import (
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

type baseNativeView struct {
}

type LongPresser struct{}

func (v *NativeView) GetView() *NativeView                                                        { return v }
func (v *NativeView) Child(path string) View                                                      { return nil }
func (v *NativeView) Rect() zgeo.Rect                                                             { return zgeo.Rect{} }
func (v *NativeView) LocalRect() zgeo.Rect                                                        { return zgeo.Rect{} }
func (v *NativeView) Parent() *NativeView                                                         { return nil }
func (v *NativeView) CalculatedSize(total zgeo.Size) zgeo.Size                                    { return zgeo.Size{10, 10} }
func (v *NativeView) ObjectName() string                                                          { return "" }
func (v *NativeView) Hierarchy() string                                                           { return "" }
func (v *NativeView) SetLocalRect(rect zgeo.Rect)                                                 {}
func (v *NativeView) SetObjectName(name string)                                                   {}
func (v *NativeView) SetColor(c zgeo.Color)                                                       {}
func (v *NativeView) SetAlpha(alpha float32)                                                      {}
func (v *NativeView) SetBGColor(c zgeo.Color)                                                     {}
func (v *NativeView) SetCorner(radius float64)                                                    {}
func (v *NativeView) SetCorners(radius float64, align zgeo.Alignment)                             {}
func (v *NativeView) SetStroke(width float64, c zgeo.Color, inset bool)                           {}
func (v *NativeView) Scale(scale float64)                                                         {}
func (v *NativeView) Rotate(deg float64)                                                          {}
func (v *NativeView) Focus(focus bool)                                                            {}
func (v *NativeView) SetCanFocus(can bool)                                                        {}
func (v *NativeView) SetOpaque(opaque bool)                                                       {}
func (v *NativeView) DumpTree()                                                                   {}
func (v *NativeView) SetFont(font *zgeo.Font)                                                     {}
func (v *NativeView) Color() zgeo.Color                                                           { return zgeo.Color{} }
func (v *NativeView) BGColor() zgeo.Color                                                         { return zgeo.Color{} }
func (v *NativeView) SetCursor(cursor CursorType)                                                 {}
func (v *NativeView) Alpha() float32                                                              { return 1 }
func (v *NativeView) GetScale() float64                                                           { return 1 }
func (v *NativeView) Show(show bool)                                                              {}
func (v *NativeView) IsShown() bool                                                               { return true }
func (v *NativeView) SetUsable(usable bool)                                                       {}
func (v *NativeView) Usable() bool                                                                { return true }
func (v *NativeView) SetInteractive(interactive bool)                                             {}
func (v *NativeView) Interactive() bool                                                           { return true }
func (v *NativeView) IsFocused() bool                                                             { return true }
func (v *NativeView) GetChild(path string) *NativeView                                            { return nil }
func (v *NativeView) RemoveFromParent()                                                           {}
func (v *NativeView) Font() *zgeo.Font                                                            { return nil }
func (v *NativeView) SetText(text string)                                                         {}
func (v *NativeView) Text() string                                                                { return "" }
func (v *NativeView) AddChild(child View, index int)                                              {}
func (v *NativeView) RemoveChild(child View)                                                      {}
func (v *NativeView) SetDropShadow(shadow zstyle.DropShadow)                                      {}
func (v *NativeView) SetToolTip(str string)                                                       {}
func (v *NativeView) SetAboveParent(above bool)                                                   {}
func NativeViewAddToRoot(v View)                                                                  {}
func (v *NativeView) SetScrollHandler(handler func(pos zgeo.Pos))                                 {}
func (v *NativeView) JSSet(property string, value interface{})                                    {}
func (v *NativeView) SetDraggable(getData func() (data string, mime string))                      {}
func (v *NativeView) SetUploader(got func(data []byte, name string))                              {}
func (v *NativeView) SetOnInputHandler(handler func())                                            {}
func (v *NativeView) SetKeyHandler(handler func(key zkeyboard.Key, mods zkeyboard.Modifier) bool) {}
func (v *NativeView) SetRect(rect zgeo.Rect)                                                      {}
func (v *NativeView) WrapInLink(surl, name string) *zcontainer.StackView                          { return nil }
func (v *NativeView) ReplaceChild(child, with View) error                                         { return nil }
func (v *NativeView) AllParents() (all []*NativeView)                                             { return }
func (v *NativeView) SetZIndex(index int)                                                         {}
func (v *NativeView) GetFocusedView() *NativeView                                                 { return nil }
func (v *NativeView) GetWindow() *Window                                                          { return nil }
func (v *NativeView) AbsoluteRect() zgeo.Rect                                                     { return zgeo.Rect{} }
func (v *NativeView) SetStyle(key, value string)                                                  {}
func (v *NativeView) SetSwipeHandler(handler func(pos, dir zgeo.Pos))                             {}
func (v *NativeView) SetOnPointerMoved(handler func(pos zgeo.Pos))                                {}
func (v *NativeView) SetHandleExposed(handle func(intersects bool))                               {}
func (v *NativeView) MakeLink(surl, name string)                                                  {}
func (v *NativeView) SetStrokeSide(width float64, c zgeo.Color, a zgeo.Alignment)                 {}
func (v *NativeView) HasSize() bool                                                               { return false }

func (v *NativeView) SetPointerDropHandler(handler func(dtype DragType, data []byte, name string, pos zgeo.Pos) bool) {
}
func (v *NativeView) SetPointerEnterHandler(moves bool, handler func(pos zgeo.Pos, inside zbool.BoolInd)) {
}
func (v *NativeView) SetPressUpDownMovedHandler(handler func(pos zgeo.Pos, down zbool.BoolInd) bool) {
}
func (v *NativeView) SetPressedDownHandler(handler func()) {}
func (v *NativeView) HasPressedDownHandler() bool          { return false }
func (v *NativeView) Native() *NativeView                  { return v }
func (v *NativeView) SetStyling(style zstyle.Styling)      {}
