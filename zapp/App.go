package zapp

// Copyright 2022 Tor Langballe. All rights reserved.
// Created by Tor Langballe on /15/11/15.

// Package app implements an Application.
// This can be the gui served to a browser,
// A simple client app, or a server, possibly serving
// a gui-app counterpart of itself.

import (
	"os"
	"time"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zprocess"
	"github.com/torlangballe/zutil/ztime"
)

type App struct {
	nativeApp                 // nativeApp adds platform-specific fields to App
	activationTime time.Time  // activationTime is when app was last "active"
	StartTime      time.Time  // StartTime is when app started
	handler        AppHandler // settable interface that handle lots of app-specific callbacks
}

type LocationTimeInfo struct {
	ZoneName          string
	ZoneOffsetSeconds int
	JSISOTimeString   string
}

var (
	AppMain                    *App   // AppMain is the main instance of app. Likely the ONLY one
	DownloadPathPrefix         string // DownloadPathPrefix is the prefix to create a url to download something from the app
	DocumentationValues        = zdict.Dict{}
	AddDocumentationValuesFunc func() zdict.Dict
)

// SetHandler sets the handler for the app (see handler) above
func (a *App) SetHandler(handler AppHandler) {
	a.handler = handler
}

// IsActive returns true if app is currently active
func (a *App) IsActive() bool {
	return !a.activationTime.IsZero()
}

// Version returns full string, version with comma, and build string.
// This will typically be hard-coded for servers/clients, and gotten from some API for native apps in future.
func Version() (string, float32, int) { // version string, version with comma 1.2, build
	return "", 0, 0
}

func Quit() {
	os.Exit(-1)
}

func (a *App) GetRuntimeSecs() float64 {
	return ztime.DurSeconds(time.Since(a.activationTime))
}

// New creates an app
func New() *App {
	a := &App{}
	a.activationTime = time.Now()
	a.StartTime = a.activationTime
	AppMain = a
	return a
}

// GetProcessID returns the process id of the app.
func GetProcessID() int {
	return os.Getpid()
}

func GetDocumentationValues() zdict.Dict {
	m := zdict.Dict{}
	for k, v := range DocumentationValues {
		m[k] = v
	}
	if AddDocumentationValuesFunc != nil {
		for k, v := range AddDocumentationValuesFunc() {
			m[k] = v
		}
	}
	return m
}

// MemoryUsed returns the memstats System useage TODO: Use combination av values?
func MemoryUsed() int64 {
	return zprocess.MemoryBytesUsedByProcess(int64(os.Getpid()))
}
