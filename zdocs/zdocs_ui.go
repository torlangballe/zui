package zdocs

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
