package app

// Created by Tor Langballe on /15/11/15.

import (
	"net/url"
	"os"
	"time"

	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/znet"
	"github.com/torlangballe/zutil/ztime"
)

type App struct {
	//    static var appFile  ZFileUrl? = nil
	nativeApp
	activationTime time.Time
	backgroundTime time.Time // IsNull if not in background
	StartTime      time.Time
	startedCount   int
	oldVersion     float32
	handler        AppHandler
}

var (
	AppMain            *App
	DownloadPathPrefix string
)

func (a *App) SetHandler(handler AppHandler) {
	a.handler = handler
}

func (a *App) IsActive() bool {
	return !a.activationTime.IsZero()
}

func (a *App) IsBackgrounded() bool {
	return !a.backgroundTime.IsZero()
}

func Version() (string, float32, int) { // version string, version with comma 1.2, build
	return "", 0, 0
}

func Id() string {
	return ""
}

func Quit() {
	os.Exit(-1)
}

func (a *App) GetRuntimeSecs() float64 {
	return ztime.DurSeconds(time.Since(a.activationTime))
}

func (a *App) GetbackgroundTimeSecs() float64 {
	return ztime.DurSeconds(time.Since(a.backgroundTime))
}

func New() *App {
	a := &App{}
	now := time.Now()
	a.activationTime = now
	a.StartTime = now
	AppMain = a
	return a
}

func (a *App) setVersions() { // this needs to be called by inheriting class, or strange stuff happens if called by ZApp
	// let (_, ver, _) = ZApp.Version
	// oldVersion = ZKeyValueStore.float64ForKey("ZVerson")
	// ZKeyValueStore.Setfloat64(float64(ver), key "ZVerson")
}

type AudioRemoteCommand int

func (a *App) EnableAudioRemote(command AudioRemoteCommand, on bool) {
}

func GetProcessId() int64 {
	return 0
}

func Host() (host string, port int) {
	u, err := url.Parse(URL())
	zlog.AssertNotError(err)
	host, port = znet.GetHostAndPort(u)
	return
}
