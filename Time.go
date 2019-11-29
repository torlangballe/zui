package zgo

import (
	"fmt"
	"strings"
	"time"

	"github.com/torlangballe/zutil/ztime"
)

//  Created by Tor Langballe on /31/10/15.

//typealias ZCalendarUnit = NSCalendar.Unit

type Time struct {
	Time time.Time
}
type TimeMonth int
type TimeWeekDay int

var TimeNull Time
var TimeDistantFuture = time.Unix(1<<63-62135596801, 999999999)

func TimeNow() Time {
	return Time{Time: time.Now()}
}

func (t Time) IsNull() bool {
	return t.Time.IsZero()
}

func (t Time) SecsSinceEpoc() float64 {
	return t.SecsSinceUnixEpoc()
}

func (t Time) SecsSinceUnixEpoc() float64 {
	return float64(t.Time.UnixNano()) / 1000000000
}

func (t Time) Since() float64 {
	return ztime.DurSeconds(time.Since(t.Time))
}

func TimeFromDate(year int, month TimeMonth, day, hour, min, second, nano int, timezone *TimeZone) Time {
	t := time.Date(year, time.Month(month), day, hour, min, second, nano, (*time.Location)(timezone))
	return Time{Time: t}
}

// Converts from ISO8601Z with msecs if contains '.'
func TimeFromIsoStr(isoStr string) (Time, error) {
	format := TimeIsoFormat
	if strings.Contains(isoStr, ".") {
		format = TimeIsoFormatWithMSecs
	}
	return FromFormat(format, isoStr, "", nil)
}

func FromFormat(format, timestring, locale string, timezone *TimeZone) (Time, error) {
	if timezone == nil {
		timezone, _ = TimeZoneNew("UTC")
	}
	t, err := time.Parse(format, timestring)
	return Time{Time: t}, err
}

func (t Time) GetGregorianParts(useAm bool, timezone *TimeZone) GregorianParts {
	var g GregorianParts
	g.Year = t.Time.Year()
	g.Month = TimeMonth(t.Time.Month())
	g.Day = t.Time.Day()
	g.Hour = t.Time.Hour()
	g.Minute = t.Time.Minute()
	g.Second = t.Time.Second()
	g.Nano = t.Time.Nanosecond()
	g.Weekday = TimeWeekDay(t.Time.Weekday())
	return g
}

func (t Time) GetGregorianDifferenceParts(toTime Time, timezone *TimeZone) GregorianParts {
	var g GregorianParts
	return g
}

func (t Time) IsToday() bool {
	return ztime.IsToday(t.Time)
}

func (t Time) GetString(format string, locale string, timezone *TimeZone) string {
	if timezone == nil {
		timezone, _ = TimeZoneNew("UTC")
	}
	return t.Time.In((*time.Location)(timezone)).Format(formatToGo(format))
}

func (t Time) PlusD(d float64) Time {
	return Time{Time: t.Time.Add(ztime.SecondsDur(d))}
}

func (t Time) Minus(p Time) float64 {
	return ztime.DurSeconds(t.Time.Sub(p.Time))
}

func (t Time) MinusD(d float64) Time {
	return Time{Time: t.Time.Add(-ztime.SecondsDur(d))}
}

func (t Time) IsLessThan(a Time) bool {
	return t.Minus(a) < 0
}

func (t Time) IsGreaterThan(a Time) bool {
	return t.Minus(a) > 0
}

func (t Time) Until() float64 {
	return -t.Since()
}

func IsAm(hour int) (bool, int) { // isam, 12-hour hour
	var h = hour
	var am = true
	if hour >= 12 {
		am = false
	}
	h %= 12
	if h == 0 {
		h = 12
	}
	return am, h
}

func Get24Hour(hour int, am bool) int {
	var h = hour
	if h == 12 {
		h = 0
	}
	if !am {
		h += 12
	}
	h %= 24
	return h
}

var toGoFormatReplacer = strings.NewReplacer(
	"yyyy", "2006",
	"yy", "06",
	"MM", "02",
	"MMMM", "January",
	"MMM", "Jan",
	"dd", "02",
	"d", "2",
	"EEE", "Mon",
	"EEEE", "Monday",
	"HH", "15",
	"mm", "04",
	"ss", "05",
	"KK", "03",
	"K", "3",
	"a", "PM",
)

func formatToGo(str string) string {
	return toGoFormatReplacer.Replace(str)
}

func (t Time) GetNicestring(locale string, timezone *TimeZone) string {
	if locale == "" {
		locale = TimeLocaleEngUsPosix
	}
	if t.IsToday() {
		return WordsGetToday() + " " + t.GetString(formatToGo("HH:mm"), locale, timezone)
	}
	return t.GetString(TimeNiceFormat, locale, timezone)
}

func (t Time) GetNiceDaysSince(locale string, timezone *TimeZone) string {
	if locale == "" {
		locale = TimeLocaleEngUsPosix
	}
	now := TimeNow()
	isPast := now.IsGreaterThan(t)
	g := t.GetGregorianDifferenceParts(now, timezone)
	var preposition = TS("ago") // generic word for 5 days ago
	if !isPast {
		preposition = TS("until") // generic word for 5 days until
	}
	switch g.Day {
	case 0:
		return WordsGetToday()
	case 1:
		if isPast {
			return WordsGetYesterday()
		}
		return WordsGetTomorrow()
	case 2, 3, 4, 5, 6, 7:
		return fmt.Sprintf("%d %s %s", g.Day, WordsGetDay(true), preposition)
	default:
		return t.GetString("MMM dd", locale, timezone)
	}
}

func (t Time) GetIsoString(format string, useNull bool) string {
	if format == "" {
		format = TimeIsoFormat
	}
	if useNull && t.IsNull() {
		return "null"
	}
	return t.GetString(format, "", nil)
}

func GetDurationSecsAsHMSstring(dur float64) string {
	var str = ""
	h := int(dur) / 3600
	var m = int(dur) / 60
	if h > 0 {
		m %= 60
		str = fmt.Sprintf("%d:", h)
	}
	s := int(dur) % 60
	str += fmt.Sprintf("%02d:%02d", m, s)
	return str
}

const TimeLocaleEngUsPosix = "en_US_POSIX"

const (
	TimeNoMonth TimeMonth = iota
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

const (
	TimeNoWeekDay TimeWeekDay = iota
	TimeMon
	TimeTue
	TimeWed
	TimeThu
	TimeFri
	TimeSat
	TimeSun
)

type GregorianParts struct {
	Year    int
	Month   TimeMonth
	Day     int
	Hour    int
	Minute  int
	Second  int
	Nano    int
	Weekday TimeWeekDay
}

const (
	TimeIsoFormat                  = "yyyy-MM-dd'T'HH:mm:ss'Z'" // UploadFileToBucket
	TimeIsoFormatWithMSecs         = "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'"
	TimeIsoFormatCompact           = "yyyyMMdd'T'HHmmss'Z'"
	TimeIsoFormatWithZone          = "yyyy-MM-dd'T'HH:mm:ssZZZZZ" // UploadFileToBucket
	TimeIsoFormatWithMSecsWithZone = "yyyy-MM-dd'T'HH:mm:ss.SSSZZZZZ"
	TimeIsoFormatCompactWithZone   = "yyyyMMdd'T'HHmmssZZZZZ"
	TimeCompactFormat              = "yyyy-MM-dd' 'HH:mm:ss"
	TimeNiceFormat                 = "dd-MMM-yyyy' 'HH:mm"
	TimeHTTPHeaderDateFormat       = "EEEE, dd LLL yyyy HH:mm:ss zzz"
	TimeMinute                     = 60.0
	TimeHour                       = 3600.0
	TimeDay                        = 86400.0
)
