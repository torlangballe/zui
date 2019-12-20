package zui

import (
	"sync"
	"time"

	"github.com/torlangballe/zutil/ztime"
)

//  Created by Tor Langballe on /11/12/16.

type Mutex sync.Mutex

type CountDownLatch struct {
	sync.WaitGroup
}

func CountDownLatchNew(count int) CountDownLatch {
	cdl := CountDownLatch{}
	cdl.Add(count)
	return cdl
}

func (cdl CountDownLatch) Wait(timeoutSecs float64) error {
	var err error
	timer := time.NewTimer(ztime.SecondsDur(timeoutSecs))
	go func() {
		<-timer.C
		err = ErrorNew("countdownlatch timeout")
		cdl.Done()
	}()
	cdl.WaitGroup.Wait()
	return err
}

func (cdl CountDownLatch) Leave() {
	cdl.Leave()
}
