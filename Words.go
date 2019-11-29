package zgo

import (
	"fmt"
	"math"
	"strconv"
)

//  Created by Tor Langballe on /11/8/18.

func WordsPluralize(word string, count float64, langCode string, pluralWord string) string {
	var lang = LocaleGetDeviceLanguageCode()
	if langCode != "" {
		lang = langCode
	}
	if pluralWord != "" {
		if count == 1 {
			return word
		}
		return pluralWord
	}
	if lang == "no" {
		if count == 1 {
			return word
		}
		return word + "er"
	}
	if lang == "de" {
		if count == 1 {
			return word
		}
		return word + "e"
	}
	if lang == "ja" {
		return word
	}
	// english
	if count == 1.0 {
		return word
	}
	if StrTail(word, 1) == "s" {
		return word + "es"
	}
	return word + "s"
}

func WordsGetLogin() string {
	return TS("Log in") // generic name for login button etc
}

func WordsGetLogout() string {
	return TS("Log out") // generic name for login button etc
}

func WordsGetAnd() string {
	return TS("And") // generic name for and, i.e  cats and dogs
}

func WordsGetHour(plural bool) string {
	if plural {
		return TS("hours") // generic name for hours plural
	}
	return TS("hour") // generic name for hour singular
}

func WordsGetToday() string {
	return TS("Today") // generic name for today
}

func WordsGetYesterday() string {
	return TS("Yesterday") // generic name for yesterday
}

func WordsGetTomorrow() string {
	return TS("Tomorrow") // generic name for tomorrow
}

// these three functions insert day/month/year symbol after date in picker, only needed for ja so far.
func WordsGetDateInsertDaySymbol() string {
	if LocaleGetDeviceLanguageCode() == "ja" {
		return "日"
	}
	return ""
}

func WordsGetDateInsertMonthSymbol() string {
	if LocaleGetDeviceLanguageCode() == "ja" {
		return "月"
	}
	return ""
}

func WordsGetDateInsertYearSymbol() string {
	if LocaleGetDeviceLanguageCode() == "ja" {
		return "年"
	}
	return ""
}

func WordsGetMinute(plural bool) string {
	if plural {
		return TS("minutes") // generic name for minutes plural
	}
	return TS("minute") // generic name for minute singular
}

func WordsGetMeter(plural bool, langCode string) string {
	if plural {
		return TSL("meters", langCode) // generic name for meters plural
	}
	return TSL("meter", langCode) // generic name for meter singular
}

func WordsGetKiloMeter(plural bool, langCode string) string {
	if plural {
		return TSL("kilometers", langCode) // generic name for kilometers plural
	}
	return TSL("kilometer", langCode) // generic name for kilometer singular
}

func WordsGetMile(plural bool, langCode string) string {
	if plural {
		return TSL("miles", langCode) // generic name for miles plural
	}
	return TSL("mile", langCode) // generic name for mile singular
}

func WordsGetYard(plural bool, langCode string) string {
	if plural {
		return TSL("yards", langCode) // generic name for yards plural
	}
	return TSL("yard", langCode) // generic name for yard singular
}

func WordsGetInch(plural bool, langCode string) string {
	if plural {
		return TSL("inches", langCode) // generic name for inch plural
	}
	return TSL("inch", langCode) // generic name for inches singular
}

func WordsGetDayPeriod() string     { return TS("am/pm") }      // generic name for am/pm part of day when used as a column title etc
func WordsGetOK() string            { return TS("OK") }         // generic name for OK in button etc
func WordsGetSet() string           { return TS("Set") }        // generic name for Set in button, i.e set value
func WordsGetOff() string           { return TS("Off") }        // generic name for Off in button, i.e value/switch is off. this is RMEOVED by VO in value
func WordsGetOpen() string          { return TS("Open") }       // generic name for button to open a window or something
func WordsGetBack() string          { return TS("Back") }       // generic name for back button in navigation bar
func WordsGetCancel() string        { return TS("Cancel") }     // generic name for Cancel in button etc
func WordsGetClose() string         { return TS("Close") }      // generic name for Close in button etc
func WordsGetPlay() string          { return TS("Play") }       // generic name for Play in button etc
func WordsGetPost() string          { return TS("Post") }       // generic name for Post in button etc, post a message to social media etc
func WordsGetEdit() string          { return TS("Edit") }       // generic name for Edit in button etc, to start an edit action
func WordsGetReset() string         { return TS("Reset") }      // generic name for Reset in button etc, to reset/restart something
func WordsGetPause() string         { return TS("Pause") }      // generic name for Pause in button etc
func WordsGetSave() string          { return TS("Save") }       // generic name for Save in button etc
func WordsGetAdd() string           { return TS("Add") }        // generic name for Add in button etc
func WordsGetDelete() string        { return TS("Delete") }     // generic name for Delete in button etc
func WordsGetExit() string          { return TS("Exit") }       // generic name for Exit in button etc. i.e  You have unsaved changes. [Save] [Exit]
func WordsGetRetryQuestion() string { return TS("Retry?") }     // generic name for Retry? in button etc, must be formulated like a question
func WordsGetFahrenheit() string    { return TS("fahrenheit") } // generic name for fahrenheit, used in buttons etc.
func WordsGetCelsius() string       { return TS("celsius") }    // generic name for celsius, used in buttons etc.
func WordsGetSettings() string      { return TS("settings") }   // generic name for settings, used in buttons / title etc

func WordsGetDayOfMonth() string { return TS("Day") }   // generic name for the day of a month i.e 23rd of July
func WordsGetMonth() string      { return TS("Month") } // generic name for month.
func WordsGetYear() string       { return TS("Year") }  // generic name for year.

func WordsGetDay(plural bool) string {
	if plural {
		return TS("Days") // generic name for the plural of a number of days since/until etc
	}
	return TS("Day") // generic name for a days since/until etc
}

func WordsGetSelected(on bool) string {
	if on {
		return TS("Selected") // generic name for selected in button/title/switch, i.e something is selected/on
	} else {
		return TS("unselected") // generic name for unselected in button/title/switch, i.e something is unselected/off
	}
}

func WordsGetMonthFromNumber(m, chars int) string {
	var str = ""
	switch m {
	case 1:
		str = TS("January") // name of month
	case 2:
		str = TS("February") // name of month
	case 3:
		str = TS("March") // name of month
	case 4:
		str = TS("April") // name of month
	case 5:
		str = TS("May") // name of month
	case 6:
		str = TS("June") // name of month
	case 7:
		str = TS("July") // name of month
	case 8:
		str = TS("August") // name of month
	case 9:
		str = TS("September") // name of month
	case 10:
		str = TS("October") // name of month
	case 11:
		str = TS("November") // name of month
	case 12:
		str = TS("December") // name of month
	default:
		break
	}
	if chars != -1 {
		str = StrHead(str, chars)
	}
	return str
} // generic name for year.

func WordsGetNameOfLanguageCode(langCode, inLanguage string) string {
	if inLanguage == "" {
		inLanguage = "en"
	}
	switch StrToLower(langCode) {
	case "en":
		return TS("English") // name of english language
	case "de":
		return TS("German") // name of german language
	case "ja", "jp":
		return TS("Japanese") // name of english language
	case "no", "nb", "nn":
		return TS("Norwegian") // name of norwegian language
	case "us":
		return TS("American") // name of american language/person
	case "ca":
		return TS("Canadian") // name of canadian language/person
	case "nz":
		return TS("New Zealander") // name of canadian language/person
	case "at":
		return TS("Austrian") // name of austrian language/person
	case "ch":
		return TS("Swiss") // name of swiss language/person
	case "in":
		return TS("Indian") // name of indian language/person
	case "gb", "uk":
		return TS("British") // name of british language/person
	case "za":
		return TS("South African") // name of south african language/person
	case "ae":
		return TS("United Arab Emirati") // name of UAE language/person
	case "id":
		return TS("Indonesian") // name of indonesian language/person
	case "sa":
		return TS("Saudi Arabian") // name of saudi language/person
	case "au":
		return TS("Australian") // name of australian language/person
	case "ph":
		return TS("Filipino") // name of filipino language/person
	case "sg":
		return TS("Singaporean") // name of singaporean language/person
	case "ie":
		return TS("Irish") // name of irish language/person
	default:
		return ""
	}
}

func WordsGetDistance(meters float64, metric bool, langCode string, round bool) string {
	Meter := 1
	Km := 2
	Mile := 3
	Yard := 4

	var dtype = Meter
	var d = meters
	var distance = ""
	var word = ""

	if metric {
		if d >= 1000 {
			dtype = Km
			d /= 1000
		}
	} else {
		d /= 1.0936133
		if d >= 1760 {
			dtype = Mile
			d /= 1760
			distance = fmt.Sprintf("%.1lf", d)
		} else {
			dtype = Yard
			d = math.Floor(d)
			distance = strconv.Itoa(int(d))
		}
	}
	switch dtype {
	case Meter:
		word = WordsGetMeter(true, langCode)

	case Km:
		word = WordsGetKiloMeter(true, langCode)

	case Mile:
		word = WordsGetMile(true, langCode)

	case Yard:
		word = WordsGetYard(true, langCode)
	}
	if dtype == Meter || dtype == Yard && round {
		d = math.Ceil(((math.Ceil(d) + 9) / 10) * 10)
		distance = strconv.Itoa(int(d))
	} else if round && d > 50 {
		distance = fmt.Sprintf("%d", int(d))
	} else {
		distance = fmt.Sprintf("%.1lf", d)
	}
	return distance + " " + word
}

func WordsMemorySizeAsstring(b int64, langCode string, maxSignificant int, isBits bool) string {
	kiloByte := 1024.0
	megaByte := kiloByte * 1024
	gigaByte := megaByte * 1024
	terraByte := gigaByte * 1024
	var word = "T"
	var n = float64(b) / terraByte
	d := float64(b)
	if d < kiloByte {
		word = ""
		n = float64(b)
	} else if d < megaByte {
		word = "K"
		n = float64(b) / kiloByte
	} else if d < gigaByte {
		word = "M"
		n = float64(b) / megaByte
	} else if d < terraByte {
		word = "G"
		n = float64(b) / gigaByte
	}
	if isBits {
		word += "b"
	} else {
		word += "B"
	}
	str := StrNiceDouble(n, maxSignificant, "") + " " + word
	return str
}

func WordsGetHemisphereDirectionsFromGeoAlignment(alignment Alignment, separator, langCode string) string {
	var str = ""
	if alignment&AlignmentTop != 0 {
		str = TSL("North", langCode) // General name for north as in north-east wind etc
	}
	if alignment&AlignmentBottom != 0 {
		str = StrConcat(separator, str, TSL("South", langCode)) // General name for south as in south-east wind etc
	}
	if alignment&AlignmentLeft != 0 {
		str = StrConcat(separator, str, TSL("West", langCode)) // General name for west as in north-west wind etc
	}
	if alignment&AlignmentRight != 0 {
		str = StrConcat(separator, str, TSL("East", langCode)) // General name for north as in north-east wind etc
	}
	return str
}
