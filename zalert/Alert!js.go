//go:build !js && zui

package zalert

func (a *Alert) showNative(handle func(result Result))              {}
func PromptForText(title, defaultText string, got func(str string)) {}
