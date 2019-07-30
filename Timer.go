package zgo

//  Created by Tor Langballe on /18/11/15.

import (
	"fmt"
	"runtime"
	"time"

	"github.com/torlangballe/zutil/ztime"
)

func init() {
	fmt.Println("zgo init()")
	runtime.LockOSThread()
}

// /*
// #include "bridge.h"
// */
// import "C"

var mainfunc = make(chan func())

func Dummy() {

}

func TimerInitBlockingDispatchFromMain() {
	runtime.LockOSThread()
	for f := range mainfunc {
		f()
	}
}

type Repeater struct {
	ticker *time.Ticker
}

func (r *Repeater) Set(secs float64, now, onMainThread bool, perform func() bool) {
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
	r.ticker.Stop()
}

type Timer struct {
	timer *time.Timer
}

func (t Timer) Set(secs float64, onMainThread bool, perform func()) {
	t.timer = time.NewTimer(ztime.SecondsDur(secs))
	go func() {
		<-t.timer.C
		perform()
	}()
}

func (t Timer) Stop() {
	t.timer.Stop()
}

func TimerPerformAfterDelay(delay float64, perform func()) {
	onMainThread := true
	Timer{}.Set(delay, onMainThread, perform)
}

func TimerMainQue() mainQue {
	return mainQue(1)
}

func (m mainQue) Async(do func()) {
	go func() {
		mainfunc <- do
	}()
}

func TimerBackgroundQue(name string, serial bool) backgroundQue {
	return backgroundQue(1)
}

func (m backgroundQue) Async(do func()) {
	go do()
}

type mainQue int
type backgroundQue int
