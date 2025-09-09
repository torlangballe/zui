//go:build zui

package zwidgets

import (
	"fmt"
	"path"
	"strings"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zdocs"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zshape"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zweb"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

// https://apple.stackexchange.com/questions/365857/create-system-preferences-url-to-privacy-files-and-folders-in-10-15-catalina
// https://dillinger.io

type DocumentationIconView struct {
	zshape.ShapeView
	docPath               string
	cachedSearchableItems []zdocs.SearchableItem
	Modal                 bool
}

var (
	DocumentationPathPrefix       = "doc/"
	DocumentationDefaultIconColor = zstyle.GrayF(0.9, 0.5)
	DocumentationViewDefaultModal = false
	DocumentationShowInBrowser    bool
	DocumentationCookieMap        map[string]string
)

func DocumentationIconViewNew(docPath string) *DocumentationIconView {
	v := &DocumentationIconView{}
	v.ShapeView.Init(v, zshape.TypeCircle, zgeo.SizeD(22, 22), "documentation:"+docPath)
	v.MaxSize = v.MinSize()
	v.SetText("?")
	v.SetColor(DocumentationDefaultIconColor())
	m := v.Margin()
	m.Pos.X += 1
	// m.Pos.Y -= 0
	v.SetMargin(m)
	v.SetTextAlignment(zgeo.Center)
	v.SetFont(zgeo.FontNice(15, zgeo.FontStyleNormal))
	v.StrokeColor = zgeo.ColorNewGray(0.3, 1)
	v.StrokeWidth = 2
	v.Modal = DocumentationViewDefaultModal
	v.SetPressedHandler("", zkeyboard.ModifierNone, func() {
		// editor := CodeEditorViewNew("editor")
		// attr := PresentViewAttributes{}
		// PresentView(editor, attr, func(win *Window) {
		// }, nil)
		v.docPath = docPath
		DocumentationViewPresent(docPath, v.Modal) // go
	})
	return v
}

type DocumentationView struct {
	zcontainer.StackView
	WebView *zweb.WebView
	// OldContentHash int64 -- what is this?
}

func (v *DocumentationView) handleURLChange(surl, oldURL string) {
	var rest string
	if zstr.HasPrefix(surl, oldURL, &rest) {
		if zstr.FirstRuneAsString(rest) == "#" {
			return
		}
	}
	// This is done because jumping to a new page sometimes doesn't scroll to top
	// TODO: Should be general in WebView? Figure out why.
	v.SetYContentOffset(0)
}

func DocumentationViewNew(minSize zgeo.Size) *DocumentationView {
	v := &DocumentationView{}
	v.Init(v, true, "docview")
	v.SetSpacing(0)

	isFrame := true
	isMakeBar := true
	v.WebView = zweb.NewView(minSize, isFrame, isMakeBar)
	v.WebView.URLChangedFunc = v.handleURLChange
	v.Add(v.WebView.Bar, zgeo.TopLeft|zgeo.HorExpand)
	v.Add(v.WebView, zgeo.TopLeft|zgeo.Expand)

	if zui.DebugOwnerMode {
		edit := zimageview.NewWithCachedPath("images/zcore/edit-dark-gray.png", zgeo.SizeBoth(zweb.DefaultBarIconSize))
		edit.DownsampleImages = true
		// edit.SetPressedHandler(v.startEdit)
		v.WebView.Bar.Add(edit, zgeo.CenterLeft)
	}
	return v
}

// func (v *DocumentationView) startEdit() {
// 	zlog.Info("Edit")
// 	editor := zcode.NewEditorView("", 50, 50)
// 	hstack := v.NativeView.Child("hstack").(*zcontainer.StackView)
// 	hstack.AddAdvanced(editor, zgeo.TopLeft|zgeo.Expand, zgeo.SizeNull, zgeo.SizeNull, 0, false)
// 	v.ArrangeChildren()
// }

// func setCSSFile(win *Window, surl string) {
// 	var css string
// 	params := zhttp.MakeParameters()
// 	_, err := zhttp.Get(surl, params, &css)
// 	if zlog.OnError(err) {
// 		return
// 	}
// 	wdoc := win.element.Get("document")
// 	style := wdoc.Call("createElement", "style")
// 	style.Set("innerHTML", css)
// 	body := wdoc.Get("body")
// 	body.Call("insertBefore", style, body.Get("firstElementChild"))
// 	zlog.Info("DOCSTYLE:", style, len(css))

// }

func makeURL(docPath string) string {
	if !zhttp.StringStartsWithHTTPX(docPath) {
		docPath = DocumentationPathPrefix + docPath
	}
	if zui.DebugOwnerMode {
		docPath += "?zdev=1"
	}
	if path.Ext(docPath) == "" {
		docPath = docPath + ".md"
	}

	return docPath
}

func DocumentationViewPresent(path string, modal bool) error {
	if DocumentationShowInBrowser {
		opts := zwindow.Options{
			URL: zfile.JoinPathParts(DocumentationPathPrefix, path),
		}
		zwindow.Open(opts)
		return nil
	}
	opts := zwindow.Options{}
	opts.ID = "doc:" + path
	if zwindow.ExistsActivate(opts.ID) {
		return nil
	}
	v := DocumentationViewNew(zgeo.SizeD(980, 800))
	path = makeURL(path)
	//	isMarkdown := zstr.HasSuffix(title, ".md", &title)

	attr := zpresent.AttributesNew()
	attr.Options = opts
	attr.FocusView = v.WebView.Bar
	if modal {
		attr.ModalCloseOnOutsidePress = true
		attr.Modal = true
	}
	attr.PresentedFunc = func(win *zwindow.Window) {
		if win == nil {
			return
		}
		// zlog.Info("SetCookie", path, DocumentationCookieMap)
		v.WebView.SetCookies(DocumentationCookieMap)
		v.WebView.SetURL(path)
		if strings.Contains(path, "#") {
			ztimer.StartIn(0.5, func() {
				v.WebView.SetURL(path)
			})
		}
	}
	zpresent.PresentView(v, attr)
	return nil
}

func addItem(currentPath []zdocs.PathPart, title string, items *[]zdocs.SearchableItem, start *int, lines *[]string) {
	if len(*lines) == 0 {
		return
	}
	var text string
	for _, l := range *lines {
		text += l + "\n"
	}
	end := *start + len(text)
	pathStub := fmt.Sprintf("%d-%d", *start, end)
	si := zdocs.MakeSearchableItem(currentPath, zdocs.InlineDocumentation, title, pathStub, text)
	*items = append(*items, si)
	*lines = []string{}
	*start = end
}

func (v *DocumentationIconView) getMarkdownAsSearchableItems(currentPath []zdocs.PathPart) {
	params := zhttp.MakeParameters()
	surl := makeURL(v.docPath)
	var text string
	_, err := zhttp.Get(surl, params, &text)
	if zlog.OnError(err, surl, currentPath) {
		return
	}
	var lines []string
	var items []zdocs.SearchableItem
	var start int
	var title string
	zstr.RangeStringLines(text, false, func(s string) bool {
		var rest string
		pre := zstr.HeadUntilWithRest(s, " ", &rest)
		count := strings.Count(pre, "#")
		if count == len(pre) { // it's all ## to start
			title = rest
			addItem(currentPath, title, &items, &start, &lines)
			lines = append(lines, s)
		}
		return true
	})
	addItem(currentPath, title, &items, &start, &lines)
	v.cachedSearchableItems = items
}

func (v *DocumentationIconView) GetSearchableItems(currentPath []zdocs.PathPart) []zdocs.SearchableItem {
	if len(v.cachedSearchableItems) != 0 {
		return v.cachedSearchableItems
	}
	go v.getMarkdownAsSearchableItems()
	return nil // parts
}
