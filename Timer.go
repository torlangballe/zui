package zui

/*
//  Created by Tor Langballe on /18/11/15.

import (
	"runtime"
	"time"

	"github.com/torlangballe/zutil/ztime"
)

func init() {
	DebugPrint("zui init()")
	runtime.LockOSThread()
}

var mainfunc = make(chan func())

func TimerInitBlockingDispatchFromMain() {
	runtime.LockOSThread()
	for f := range mainfunc {
		f()
	}
}

type Repeater struct {
	ticker *time.Ticker
}

func RepeaterNew() *Repeater {
	return &Repeater{}
}

func RepeaterSet(secs float64, now, onMainThread bool, perform func() bool) *Repeater {
	r := RepeaterNew()
	r.Set(secs, now, onMainThread, perform)
	return r
}

func (r *Repeater) Set(secs float64, now, onMainThread bool, perform func() bool) {
	r.Stop()
	if now {
		if !perform() {
			return
		}
	}
	r.ticker = time.NewTicker(ztime.SecondsDur(secs))
	go func() {
		for range r.ticker.C {
			if !perform() {
				r.ticker.Stop()
				break
			}
		}
	}()
}

func (r *Repeater) Stop() {
	if r.ticker != nil {
		r.ticker.Stop()
	}
}

// Timer

type Timer struct {
	timer *time.Timer
}

func TimerNew() *Timer {
	return &Timer{}
}

func TimerSet(secs float64, onMainThread bool, perform func()) *Timer {
	t := TimerNew()
	t.Set(secs, onMainThread, perform)
	return t
}

func (t *Timer) Set(secs float64, onMainThread bool, perform func()) *Timer {
	// fmt.Println("timer set:", secs, zlog.GetCallingStackString(3, "\n"))
	t.Stop()
	t.timer = time.NewTimer(ztime.SecondsDur(secs))
	go func() {
		<-t.timer.C
		perform()
	}()
	return t
}

func (t *Timer) Stop() {
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
}

*/
