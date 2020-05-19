// +build !js

package zui

func ScreenMain() Screen {
	s := Screen{}
	s.SoftScale = 1
	s.Scale = 1
	return s
}
