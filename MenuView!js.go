// +build !js

package zui

func MenuViewNew(name string, items MenuItems, value interface{}, isStatic bool) *MenuView {
	return &MenuView{}
}

func (v *MenuView) IDAndValue() (id string, value interface{}) {
	return "", nil
}

func (v *MenuView) SetWithID(id string) *MenuView {
	return v
}

func (v *MenuView) ChangedHandler(handler func(id, name string, value interface{})) {}
func (v *MenuView) UpdateValues(items MenuItems)                                    {}

func menuViewGetHackedFontForSize(font *Font) *Font {
	return font
}

func (v *MenuView) SetFont(font *Font) View {
	return v
}
