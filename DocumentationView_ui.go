// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
)

// https://apple.stackexchange.com/questions/365857/create-system-preferences-url-to-privacy-files-and-folders-in-10-15-catalina
// https://dillinger.io

type DocumentationIconView struct {
	ShapeView
}

var DocumentationPathPrefix = "doc/"
var DocumentationDefaultIconColor = zgeo.ColorNewGray(0.5, 1)

func DocumentationIconViewNew(path string) *DocumentationIconView {
	v := &DocumentationIconView{}
	v.ShapeView.Init(v, ShapeViewTypeCircle, zgeo.Size{22, 22}, "documentation:"+path)
	v.textInfo.Text = "?"
	v.SetColor(DocumentationDefaultIconColor)
	v.SetTextAlignment(zgeo.Center)
	v.SetFont(FontNice(16, FontStyleBold))
	v.StrokeColor = DocumentationDefaultIconColor
	v.StrokeWidth = 2
	v.SetColor(zgeo.ColorNew(0.9, 0.9, 1, 1))
	v.SetPressedHandler(func() {
		// editor := CodeEditorViewNew("editor")
		// attr := PresentViewAttributes{}
		// PresentView(editor, attr, func(win *Window) {
		// }, nil)
		go DocumentationViewPresent(path)
	})
	return v
}

type DocumentationView struct {
	StackView
	WebView *WebView
	// OldContentHash int64 -- what is this?
}

func DocumentationViewNew(minSize zgeo.Size) *DocumentationView {
	v := &DocumentationView{}
	v.Init(v, true, "docview")
	v.SetSpacing(0)

	hstack := StackViewHor("hstack")
	hstack.SetSpacing(0)

	isFrame := true
	isMakeBar := true
	v.WebView = WebViewNew(minSize, isFrame, isMakeBar)
	v.Add(v.WebView.Bar, zgeo.TopLeft|zgeo.HorExpand)
	v.Add(hstack, zgeo.TopLeft|zgeo.Expand)
	hstack.Add(v.WebView, zgeo.TopLeft|zgeo.Expand)

	if DebugOwnerMode {
		edit := ImageViewNew(nil, "images/edit.png", zgeo.SizeBoth(WebViewDefaultBarIconSize))
		edit.DownsampleImages = true
		edit.SetPressedHandler(v.startEdit)
		v.WebView.Bar.Add(edit, zgeo.CenterLeft)
	}
	return v
}

func (v *DocumentationView) startEdit() {
	zlog.Info("Edit")
	editor := CodeEditorViewNew("", 50, 50)
	hstack := v.NativeView.Child("hstack").(*StackView)
	hstack.Add(editor, zgeo.TopLeft|zgeo.Expand, 0)
	v.ArrangeChildren(nil)
}

func DocumentationViewPresent(path string) error {
	opts := WindowOptions{}
	opts.ID = "doc:" + path
	if WindowExistsActivate(opts.ID) {
		return nil
	}
	v := DocumentationViewNew(zgeo.Size{980, 800})
	filepath := path
	if !zhttp.StringStartsWithHTTPX(path) {
		filepath = DocumentationPathPrefix + path
	}
	zlog.Info("SetDocPath:", filepath)
	v.WebView.SetURL(filepath)
	//	isMarkdown := zstr.HasSuffix(title, ".md", &title)

	attr := PresentViewAttributes{
		WindowOptions: opts,
	}
	PresentView(v, attr, nil, nil)
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
