package zui

//  Created by Tor Langballe on /3/12/15.

import (
	"strings"
	"time"

	"github.com/torlangballe/zutil/ustr"
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
	str := ustr.TailUntil(t.Name(), "/")
	return strings.Replace(str, "_", "", -1)
}

func (t *TimeZone) HoursFromUTC(at time.Time) float64 {
	_, offset := at.In((*time.Location)(t)).Zone()
	return float64(offset) / 3600
}

func (t *TimeZone) IsUTC() bool {
	return t == TimeZoneUTC
}
