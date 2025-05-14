//go:build zui

package zwidgets

import (
	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zshape"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zweb"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zstr"
)

// https://apple.stackexchange.com/questions/365857/create-system-preferences-url-to-privacy-files-and-folders-in-10-15-catalina
// https://dillinger.io

type DocumentationIconView struct {
	zshape.ShapeView
	Modal bool
}

var (
	DocumentationPathPrefix       = "doc/"
	DocumentationDefaultIconColor = zstyle.GrayF(0.9, 0.5)
	DocumentationViewDefaultModal = false
	DocumentationCookieMap        map[string]string
)

func DocumentationIconViewNew(path string) *DocumentationIconView {
	v := &DocumentationIconView{}
	v.ShapeView.Init(v, zshape.TypeCircle, zgeo.SizeD(22, 22), "documentation:"+path)
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
		DocumentationViewPresent(path+".md", v.Modal) // go
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

func DocumentationViewPresent(path string, modal bool) error {
	opts := zwindow.Options{}
	opts.ID = "doc:" + path
	if zwindow.ExistsActivate(opts.ID) {
		return nil
	}
	v := DocumentationViewNew(zgeo.SizeD(980, 800))
	if !zhttp.StringStartsWithHTTPX(path) {
		path = DocumentationPathPrefix + path
	}
	if zui.DebugOwnerMode {
		path += "?zdev=1"
	}
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
	}
	zpresent.PresentView(v, attr)
	return nil
}

/*

x-apple.systempreferences:

Accessibility Preference Pane
Main    x-apple.systempreferences:com.apple.preference.universalaccess
Display x-apple.systempreferences:com.apple.preference.universalaccess?Seeing_Display
Zoom    x-apple.systempreferences:com.apple.preference.universalaccess?Seeing_Zoom
VoiceOver   x-apple.systempreferences:com.apple.preference.universalaccess?Seeing_VoiceOver
Descriptions    x-apple.systempreferences:com.apple.preference.universalaccess?Media_Descriptions
Captions    x-apple.systempreferences:com.apple.preference.universalaccess?Captioning
Audio   x-apple.systempreferences:com.apple.preference.universalaccess?Hearing
Keyboard    x-apple.systempreferences:com.apple.preference.universalaccess?Keyboard
Mouse & Trackpad    x-apple.systempreferences:com.apple.preference.universalaccess?Mouse
Switch Control  x-apple.systempreferences:com.apple.preference.universalaccess?Switch
Dictation   x-apple.systempreferences:com.apple.preference.universalaccess?SpeakableItems

Security & Privacy Preference Pane
Main    x-apple.systempreferences:com.apple.preference.security
General x-apple.systempreferences:com.apple.preference.security?General
FileVault   x-apple.systempreferences:com.apple.preference.security?FDE
Firewall    x-apple.systempreferences:com.apple.preference.security?Firewall
Advanced    x-apple.systempreferences:com.apple.preference.security?Advanced
Privacy x-apple.systempreferences:com.apple.preference.security?Privacy
Privacy-Camera x-apple.systempreferences:com.apple.preference.security?Privacy_Camera
Privacy-Microphone  x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone
Privacy-Automation  x-apple.systempreferences:com.apple.preference.security?Privacy_Automation
Privacy-AllFiles    x-apple.systempreferences:com.apple.preference.security?Privacy_AllFiles
Privacy-Accessibility   x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility
Privacy-Assistive   x-apple.systempreferences:com.apple.preference.security?Privacy_Assistive
Privacy-Location Services   x-apple.systempreferences:com.apple.preference.security?Privacy_LocationServices
Privacy-SystemServices  x-apple.systempreferences:com.apple.preference.security?Privacy_SystemServices
Privacy-Advertising x-apple.systempreferences:com.apple.preference.security?Privacy_Advertising
Privacy-Contacts    x-apple.systempreferences:com.apple.preference.security?Privacy_Contacts
Privacy-Diagnostics & Usage x-apple.systempreferences:com.apple.preference.security?Privacy_Diagnostics
Privacy-Calendars   x-apple.systempreferences:com.apple.preference.security?Privacy_Calendars
Privacy-Reminders   x-apple.systempreferences:com.apple.preference.security?Privacy_Reminders
Privacy-Facebook    x-apple.systempreferences:com.apple.preference.security?Privacy_Facebook
Privacy-LinkedIn    x-apple.systempreferences:com.apple.preference.security?Privacy_LinkedIn
Privacy-Twitter x-apple.systempreferences:com.apple.preference.security?Privacy_Twitter
Privacy-Weibo   x-apple.systempreferences:com.apple.preference.security?Privacy_Weibo
Privacy-Tencent Weibo   x-apple.systempreferences:com.apple.preference.security?Privacy_TencentWeibo

macOS Catalina 10.15:
Privacy-ScreenCapture   x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture
Privacy-DevTools    x-apple.systempreferences:com.apple.preference.security?Privacy_DevTools
Privacy-InputMonitoring x-apple.systempreferences:com.apple.preference.security?Privacy_ListenEvent
Privacy-DesktopFolder   x-apple.systempreferences:com.apple.preference.security?Privacy_DesktopFolder
Privacy-DocumentsFolder x-apple.systempreferences:com.apple.preference.security?Privacy_DocumentsFolder
Privacy-DownloadsFolder x-apple.systempreferences:com.apple.preference.security?Privacy_DownloadsFolder
Privacy-NetworkVolume   x-apple.systempreferences:com.apple.preference.security?Privacy_NetworkVolume
Privacy-RemovableVolume x-apple.systempreferences:com.apple.preference.security?Privacy_RemovableVolume
Privacy-SpeechRecognition   x-apple.systempreferences:com.apple.preference.security?Privacy_SpeechRecognition

Dictation & Speech Preference Pane
Dictation   x-apple.systempreferences:com.apple.preference.speech?Dictation
Text to Speech  x-apple.systempreferences:com.apple.preference.speech?TTS
Sharing Preference Pane
Main    x-apple.systempreferences:com.apple.preferences.sharing
Screen Sharing  x-apple.systempreferences:com.apple.preferences.sharing?Services_ScreenSharing
File Sharing    x-apple.systempreferences:com.apple.preferences.sharing?Services_PersonalFileSharing
Printer Sharing x-apple.systempreferences:com.apple.preferences.sharing?Services_PrinterSharing
Remote Login    x-apple.systempreferences:com.apple.preferences.sharing?Services_RemoteLogin
Remote Management   x-apple.systempreferences:com.apple.preferences.sharing?Services_ARDService
Remote Apple Events x-apple.systempreferences:com.apple.preferences.sharing?Services_RemoteAppleEvent
Internet Sharing    x-apple.systempreferences:com.apple.preferences.sharing?Internet
Bluetooth Sharing   x-apple.systempreferences:com.apple.preferences.sharing?Services_BluetoothSharing

Software update x-apple.systempreferences:com.apple.preferences.softwareupdate?client=softwareupdateapp

*/
