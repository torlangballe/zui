//go:build zui

package zdocs

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zlog"
)

type GUIPartOpener interface {
	OpenGUIFromPathParts(parts []PathPart) bool
}

var PartOpener GUIPartOpener

/*
import "github.com/torlangballe/zui/zcontainer"

type LinkPartsView struct {
	zcontainer.StackView
	Path        []PathPart
	currentPart int
}

func NewLinkPartsView(address string) *LinkPartsView {
	v := &LinkPartsView{}
	v.StackView.Init(v, true, "link")
	return v
}
*/

func GetSearchableItems(root zview.View) []SearchableItem {
	IsGettingSearchItems = true
	sig, _ := root.(SearchableItemsGetter)
	zlog.Assert(sig != nil)
	items := sig.GetSearchableItems(nil) // add root here?
	IsGettingSearchItems = false
	return items
}
