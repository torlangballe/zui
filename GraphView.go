// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

type GraphView struct {
}

func GraphViewNew(minSize zgeo.Size) *GraphView {
	v := &GraphView{}
	return v
}
