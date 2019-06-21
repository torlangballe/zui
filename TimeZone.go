package zgo

//  Created by Tor Langballe on /3/12/15.

import (
	"time"
)

type TimeZone time.Location

var TimeZoneUTC = (*TimeZone)(time.UTC)
var TimeZoneGMT = TimeZoneUTC

func TimeZoneNew(id string) (*TimeZone, error) {
	t, err := time.LoadLocation(id)
	return (*TimeZone)(t), err
}

func TimeZoneForDevice() *TimeZone {
	return nil
}

func (t *TimeZone) Name() string {
	return (*time.Location)(t).String()
}

func (t *TimeZone) NiceName() string {
	str := StrTailUntil(t.Name(), "/")
	return StrReplace(str, "_", "", 0)
}

func (t *TimeZone) HoursFromUTC(at time.Time) float64 {
	_, offset := at.In((*time.Location)(t)).Zone()
	return float64(offset) / 3600
}

func (t *TimeZone) IsUTC() bool {
	return t == TimeZoneUTC
}
