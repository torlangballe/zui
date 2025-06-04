package zwidgets

import (
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zview"
)

func init() {
	zpresent.DocumentationIconViewNewFunc = func(path string) zview.View {
		return DocumentationIconViewNew(path)
	}
}
