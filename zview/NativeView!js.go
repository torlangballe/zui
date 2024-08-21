//go:build !js && zui

package zview

import (
	"github.com/torlangballe/zui/zcursor"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

type baseNativeView struct {
}

type LongPresser struct{}

// func NativeViewAddToRoot(v View)                                                                  {}
func (v *NativeView) GetView() *NativeView                                   { return v }
func (v *NativeView) Child(path string) View                                 { return nil }
func (v *NativeView) Rect() zgeo.Rect                                        { return zgeo.Rect{} }
func (v *NativeView) LocalRect() zgeo.Rect                                   { return zgeo.Rect{} }
func (v *NativeView) Parent() *NativeView                                    { return nil }
func (v *NativeView) ObjectName() string                                     { return "" }
func (v *NativeView) Hierarchy() string                                      { return "" }
func (v *NativeView) SetLocalRect(rect zgeo.Rect)                            {}
func (v *NativeView) SetObjectName(name string)                              {}
func (v *NativeView) SetColor(c zgeo.Color)                                  {}
func (v *NativeView) SetAlpha(alpha float32)                                 {}
func (v *NativeView) SetBGColor(c zgeo.Color)                                {}
func (v *NativeView) SetCorner(radius float64)                               {}
func (v *NativeView) SetCorners(radius float64, align zgeo.Alignment)        {}
func (v *NativeView) SetStroke(width float64, c zgeo.Color, inset bool)      {}
func (v *NativeView) Scale(scale float64)                                    {}
func (v *NativeView) Rotate(deg float64)                                     {}
func (v *NativeView) Focus(focus bool)                                       {}
func (v *NativeView) SetCanTabFocus(can bool)                                {}
func (v *NativeView) CanTabFocus() bool                                      { return false }
func (v *NativeView) SetOpaque(opaque bool)                                  {}
func (v *NativeView) DumpTree()                                              {}
func (v *NativeView) SetFont(font *zgeo.Font)                                {}
func (v *NativeView) Color() zgeo.Color                                      { return zgeo.Color{} }
func (v *NativeView) BGColor() zgeo.Color                                    { return zgeo.Color{} }
func (v *NativeView) SetCursor(cursor zcursor.Type)                          {}
func (v *NativeView) Alpha() float32                                         { return 1 }
func (v *NativeView) GetScale() float64                                      { return 1 }
func (v *NativeView) Show(show bool)                                         {}
func (v *NativeView) IsShown() bool                                          { return true }
func (v *NativeView) SetUsable(usable bool)                                  {}
func (v *NativeView) Usable() bool                                           { return true }
func (v *NativeView) SetInteractive(interactive bool)                        {}
func (v *NativeView) Interactive() bool                                      { return true }
func (v *NativeView) IsFocused() bool                                        { return true }
func (v *NativeView) GetChild(path string) *NativeView                       { return nil }
func (v *NativeView) RemoveFromParent()                                      {}
func (v *NativeView) Font() *zgeo.Font                                       { return nil }
func (v *NativeView) SetText(text string)                                    {}
func (v *NativeView) Text() string                                           { return "" }
func (v *NativeView) InsertBefore(before View)                               {}
func (v *NativeView) AddChild(child View, index int)                         {}
func (v *NativeView) RemoveChild(child View)                                 {}
func (v *NativeView) SetDropShadow(shadow zstyle.DropShadow)                 {}
func (v *NativeView) SetToolTip(str string)                                  {}
func (v *NativeView) SetAboveParent(above bool)                              {}
func (v *NativeView) SetScrollHandler(handler func(pos zgeo.Pos))            {}
func (v *NativeView) JSSet(property string, value interface{})               {}
func (v *NativeView) JSGet(property string) any                              { return nil }
func (v *NativeView) JSCall(method string, args ...interface{}) any          { return nil }
func (v *NativeView) SetDraggable(getData func() (data string, mime string)) {}
func (v *NativeView) SetUploader(got func(data []byte, name string), use func(name string) bool, progress func(p float64)) {
}
func (v *NativeView) ReplaceChild(child, with View) error                             { return nil }
func (v *NativeView) AllParents() (all []*NativeView)                                 { return }
func (v *NativeView) GetFocusedChildView(andSelf bool) View                           { return nil }
func (v *NativeView) GetPathOfChild(child View) string                                { return "" }
func (v *NativeView) AbsoluteRect() zgeo.Rect                                         { return zgeo.Rect{} }
func (v *NativeView) SetOnInputHandler(handler func())                                {}
func (v *NativeView) SetKeyHandler(handler func(km zkeyboard.KeyMod, down bool) bool) {}
func (v *NativeView) SetRect(rect zgeo.Rect)                                          {}

// func (v *NativeView) SetStyle(kSetKeyHandlerey, value string)                                              {}
func (v *NativeView) SetZIndex(index int)                                                     {}
func (v *NativeView) SetSwipeHandler(handler func(pos, dir zgeo.Pos))                         {}
func (v *NativeView) SetOnPointerMoved(handler func(pos zgeo.Pos))                            {}
func (v *NativeView) SetHandleExposed(handle func(intersects bool))                           {}
func (v *NativeView) MakeLink(surl, name string)                                              {}
func (v *NativeView) SetStrokeSide(width float64, c zgeo.Color, a zgeo.Alignment, inset bool) {}
func (v *NativeView) HasSize() bool                                                           { return false }

func (v *NativeView) SetPointerDropHandler(handler func(dtype DragType, data []byte, name string, pos zgeo.Pos) bool) {
}
func (v *NativeView) SetPointerEnterHandler(moves bool, handler func(pos zgeo.Pos, inside zbool.BoolInd)) {
}
func (v *NativeView) SetPressUpDownMovedHandler(handler func(pos zgeo.Pos, down zbool.BoolInd) bool) {
}
func (v *NativeView) SetPressedDownHandler(handler func())                                   {}
func (v *NativeView) SetDoublePressedHandler(handler func())                                 {}
func (v *NativeView) HasPressedDownHandler() bool                                            { return false }
func (v *NativeView) Native() *NativeView                                                    { return v }
func (v *NativeView) SetStyling(style zstyle.Styling)                                        {}
func (v *NativeView) SetSelectable(on bool)                                                  {}
func (v *NativeView) SetJSStyle(key, value string)                                           {}
func (v *NativeView) SetOutline(width float64, c zgeo.Color, offset float64)                 {}
func (nv *NativeView) SetNativePadding(m zgeo.Rect)                                          {}
func (nv *NativeView) SetNativeMargin(m zgeo.Rect)                                           {}
func (nv *NativeView) ShowBackface(visible bool)                                             {}
func (v *NativeView) SetTilePath(spath string)                                               {}
func (v *NativeView) Click()                                                                 {}
func (v *NativeView) RootParent() *NativeView                                                { return nil }
func (v *NativeView) SetFocusHandler(focused func(focus bool))                               {}
func (*NativeView) HandleFocusInChildren(in, out bool, handle func(view View, focused bool)) {}

func DownloadURI(uri, name string) {}

func (v *NativeView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	return zgeo.SizeD(10, 10), zgeo.Size{}
}
