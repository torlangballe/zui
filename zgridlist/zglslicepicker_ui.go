//go:build zui

package zgridlist

import (
	"strconv"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

func PresentSlicePicker[S zstr.TitleOwner](title string, slice []S, keepPicking bool, lastPickedClientForAssetType int64, got func(row S, closed bool)) {
	grid := NewView("picker")
	grid.Spacing = zgeo.SizeNull
	grid.MakeFullSize = true
	// grid.SetMinSize(zgeo.SizeD(340, 400))
	// v.grid.MinRowsForFullSize = 5
	// v.grid.MaxRowsForFullSize = 20
	grid.CellCountFunc = func() int {
		return len(slice)
	}
	grid.CreateCellFunc = func(grid *GridListView, id string) zview.View {
		h := zcontainer.StackViewHor("h")
		h.SetMarginS(zgeo.SizeD(6, 2))
		i := grid.IndexOfID(id)
		str := slice[i].GetTitle()
		label := zlabel.New(str)
		label.SetMaxWidth(400)
		h.Add(label, zgeo.TopLeft)
		return h
	}
	grid.MaxColumns = 1
	grid.HandleRowPressedFunc = func(id string) bool {
		i := grid.IndexOfID(id)
		if !keepPicking {
			zpresent.Close(grid, false, nil)
			zlog.Info("Here")
			// return true //!!
		}
		ztimer.StartIn(0.1, func() {
			got(slice[i], false)
		})
		return true
	}
	att := zpresent.ModalDialogAttributes
	att.ClosedFunc = func(dismissed bool) {
		var s S
		got(s, true)
	}
	if lastPickedClientForAssetType != 0 {
		att.PresentedFunc = func(win *zwindow.Window) {
			sid := strconv.FormatInt(lastPickedClientForAssetType, 10)
			grid.HandleRowPressedFunc(sid)
		}
	}
	zpresent.PresentTitledView(grid, title, att, nil, nil)
}
