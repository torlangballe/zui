package zapp

import (
	"github.com/torlangballe/zutil/zdict"
)

type Notification struct {
}

type AppHandler interface {
	HandleAppNotification(notification Notification, action string)
	HandlePushNotificationWithDictionary(dict zdict.Dict, fromStartup bool, whileActive bool)
	HandleLocationRegionCross(regionId string, enter bool, fromAdd bool)
	HandleMemoryNearFull()
	HandleAudiointerrupted()
	HandleAudioResume()
	HandleAudioRouteChanged(reason int)
	HandleAudioRemote(command AudioRemoteCommand)
	HandleRemoteAudioSeekTo(posSecs float64)
	HandleVoiceOverStatusChanged()
	HandleBackgrounded(background bool)
	HandleActivated(activated bool)
	HandleOpenedFiles(files []string, modifiers int)
	ShowDebugText(str string)
	HandleGotPushToken(token string)
	HandleLanguageBCPChanged(bcp string)
	HandleAppWillTerminate()
	HandleShake()
	HandleExit()
	HandleOpenUrl(url string, showMessage bool, done *func()) bool
}
