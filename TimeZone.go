package zgo

import (
	"strings"
	"time"
)

//  Created by Tor Langballe on /3/12/15.

type TimeZone time.Location

var TimeZoneUTC = time.UTC
var TimeZoneGMT = time.UTC

func TimeZoneNew(id string) (*TimeZone, error) {
	t, err := time.LoadLocation()
	return TimeZone(t), err
}

func TimeZoneForDevice() *TimeZone {
	return nil
}

func (t *TimeZone) Name() string {
	return t.Str
}

func (t *TimeZone) NiceName() string {
	str := Str.TailUntil(t.Name(), "/")
	return strings.Replace(str, "_", "", -1)
}

func (t *TimeZone) HoursFromUTC() float64 {
	return float64(secondsFromGMT()) / 3600
}

func (t *TimeZone) CalculateOffsetHours(time Time, localDeltaHours *float64) float64 {
	// secs := secondsFromGMT(for: time.date)
	// lsecs := TimeZone.autoupdatingCurrent.secondsFromGMT(for: time.date)
	// localDeltaHours = float64(secs - lsecs) / 3600
	// return float64(secs) / 3600
}

func (t *TimeZone) IsUTC() bool {
	return t == TimeZoneUTC
}
