// +build !js,zui

package zui

import "github.com/torlangballe/zutil/zgeo"

type baseNativeView struct {
}

type LongPresser struct{}

func (v *NativeView) GetView() *NativeView                                                    { return v }
func (v *NativeView) Child(path string) View                                                  { return nil }
func (v *NativeView) Rect() zgeo.Rect                                                         { return zgeo.Rect{} }
func (v *NativeView) LocalRect() zgeo.Rect                                                    { return zgeo.Rect{} }
func (v *NativeView) Parent() *NativeView                                                     { return nil }
func (v *NativeView) CalculatedSize(total zgeo.Size) zgeo.Size                                { return zgeo.Size{10, 10} }
func (v *NativeView) ObjectName() string                                                      { return "" }
func (v *NativeView) Hierarchy() string                                                       { return "" }
func (v *NativeView) SetLocalRect(rect zgeo.Rect)                                             {}
func (v *NativeView) SetObjectName(name string)                                               {}
func (v *NativeView) SetColor(c zgeo.Color)                                                   {}
func (v *NativeView) SetAlpha(alpha float32)                                                  {}
func (v *NativeView) SetBGColor(c zgeo.Color)                                                 {}
func (v *NativeView) SetCorner(radius float64)                                                {}
func (v *NativeView) SetStroke(width float64, c zgeo.Color)                                   {}
func (v *NativeView) Scale(scale float64)                                                     {}
func (v *NativeView) Rotate(deg float64)                                                      {}
func (v *NativeView) Focus(focus bool)                                                        {}
func (v *NativeView) SetCanFocus(can bool)                                                    {}
func (v *NativeView) SetOpaque(opaque bool)                                                   {}
func (v *NativeView) DumpTree()                                                               {}
func (v *NativeView) SetFont(font *Font)                                                      {}
func (v *NativeView) Color() zgeo.Color                                                       { return zgeo.Color{} }
func (v *NativeView) BGColor() zgeo.Color                                                     { return zgeo.Color{} }
func (v *NativeView) Alpha() float32                                                          { return 1 }
func (v *NativeView) GetScale() float64                                                       { return 1 }
func (v *NativeView) Show(show bool)                                                          {}
func (v *NativeView) IsShown() bool                                                           { return true }
func (v *NativeView) SetUsable(usable bool)                                                   {}
func (v *NativeView) Usable() bool                                                            { return true }
func (v *NativeView) IsFocused() bool                                                         { return true }
func (v *NativeView) GetChild(path string) *NativeView                                        { return nil }
func (v *NativeView) RemoveFromParent()                                                       { v.StopStoppers() }
func (v *NativeView) Font() *Font                                                             { return nil }
func (v *NativeView) SetText(text string)                                                     {}
func (v *NativeView) Text() string                                                            { return "" }
func (v *NativeView) AddChild(child View, index int)                                          {}
func (v *NativeView) RemoveChild(child View)                                                  {}
func (v *NativeView) SetDropShadow(shadow zgeo.DropShadow)                                    {}
func (v *NativeView) SetToolTip(str string)                                                   {}
func (v *NativeView) SetAboveParent(above bool)                                               {}
func NativeViewAddToRoot(v View)                                                              {}
func (v *NativeView) SetScrollHandler(handler func(pos zgeo.Pos))                             {}
func (v *NativeView) setjs(property string, value interface{})                                {}
func (v *NativeView) SetPointerEnterHandler(handler func(pos zgeo.Pos, inside bool))          {}
func (v *NativeView) SetDraggable(getData func() (data string, mime string))                  {}
func (v *NativeView) SetUploader(got func(data []byte, name string))                          {}
func (v *NativeView) SetOnInputHandler(handler func())                                        {}
func (v *NativeView) SetKeyHandler(handler func(key KeyboardKey, mods KeyboardModifier) bool) {}
func (v *NativeView) SetRect(rect zgeo.Rect)                                                  {}
func (v *NativeView) WrapInLink(surl, name string) *StackView                                 { return nil }
func (v *NativeView) ReplaceChild(child, with View) error                                     { return nil }
func (v *NativeView) AllParents() (all []*NativeView)                                         { return }
func (v *NativeView) SetZIndex(index int)                                                     {}
func (v *NativeView) GetFocusedView() *NativeView                                             { return nil }
func (v *NativeView) GetWindow() *Window                                                      { return nil }
func (v *NativeView) AbsoluteRect() zgeo.Rect                                                 { return zgeo.Rect{} }
func (v *NativeView) SetStyle(key, value string)                                              {}
func (v *NativeView) SetSwipeHandler(handler func(pos, dir zgeo.Pos))                         {}
func (v *NativeView) SetOnPointerMoved(handler func(pos zgeo.Pos))                            {}

func (v *NativeView) SetPointerDropHandler(handler func(dtype DragType, data []byte, name string, pos zgeo.Pos) bool) {
}
