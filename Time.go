package zgo

import (
	"strings"
	"time"

	"github.com/torlangballe/zutil/ztime"
)

//  Created by Tor Langballe on /31/10/15.

//typealias ZCalendarUnit = NSCalendar.Unit

type Time struct {
	Time time.Time
}

type Month int

const (
	TimeNoMonth Month = 0
	TimeJan
	TimeFeb
	TimeMar
	TimeApr
	TimeMay
	TimeJun
	TimeJul
	TimeAug
	TimeSep
	TimeOct
	TimeNov
	TimeDec
)

type WeekDay int

const (
	TimeNoWeekDay WeekDay = 0
	TimeMon
	TimeTue
	TimeWed
	TimeThu
	TimeFri
	TimeSat
	TimeSun
)

type GregorianParts struct {
	year    int
	month   Month
	day     int
	hour    int
	minute  int
	second  int
	nano    int
	weekday WeekDay
}

var ZTimeNull Time
var ZTimeDistantFuture = time.Unix(1<<63-62135596801, 999999999)

func TimeNow() Time {
	return Time{Time: time.Now()}
}

func (t Time) IsNull() bool {
	return t.Time.IsZero()
}

func SecsSinceEpoc() float64 {
	return SecsSinceUnixEpoc()
}

func (t Time) SecsSinceUnixEpoc() float64 {
	s, n := t.Time.Unix()
}

func (t Time) Since() float64 {
	return ztime.Seconds(time.Since(t.Time))
}

func TimeFromDate(year int, month Month, day, hour, min, second, nano int, timezone *TimeZone) Time {
	t := time.Date(year, month, day, hour, min, second, nano)
	return Time{Time: t}
}

func TimeFromIsoStr(iso8601Z string) Time {
	format := TimeIsoFormat
	if strings.Contains(iso8601Z, ".") {
		format = TimeIsoFormatWithMSecs
	}
	zone := TimeZoneNew("UTC")
}

func FromFormat(format, timeString, locale string, timezone *TimeZone) Time {
	// todo
	return Time{}
}

func (t Time) GetGregorianTimeParts(useAm bool, timezone *TimeZone) (int, int, int, int, bool) { // hour, min, sec, nano, isam

}

func GetGregorianDateParts(timezone *TimeZone) GregorianParts {
	var g GregorianParts
	return g
}

func GetGregorianDateDifferenceParts(toTime Time, timezone *TimeZone) GregorianParts {
	var g GregorianParts
	return g
}

func GetGregorianTimeDifferenceParts(toTime Time, timezone *TimeZone) (int, int, int, int) { // returns day, hour, minute, secs
}

func (t Time) IsToday() bool {
	return ztime.IsToday(t.Time)
}

func (t Time) GetString(format string, locale string, timezone *TimeZone) string {
}

func (t Time) Plus(p Time) Time {
	return Time{Time: t.Time.Add(p.Time)}
}

func (t Time) Minus(p Time) float64 {
	return ztime.Seconds(t.Time.Sub(p.Time))
}

func (t Time) PlusD(d Time) Time {
	return Time{Time: t.Time.Add(ztime.Second(d))}
}

func (t Time) MinusD(d float64) Time {
	return Time{Time: t.Time.Sub(ztime.Second(d))}
}

func (t Time) IsLessThan(a Time) bool {
	return t.Minus(a) < 0
}

func (t Time) IsGreaterThan(a Time) bool {
	return t.Minus(a) > 0
}
