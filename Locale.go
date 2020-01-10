package zui

import (
	"strings"

	"github.com/torlangballe/zutil/zstr"
)

//  Created by Tor Langballe on /1/11/15.

// https://github.com/RadhiFadlillah/sysloc
func LocaleGetDeviceLanguageCode() string {
	return "en"
}

func LocaleGetLangCodeAndCountryFromLocaleId(bcp string, forceNo bool) (string, string) { // lang, country-code
	lang, ccode := zstr.SplitInTwo(bcp, "-")
	if ccode == "" {
		_, ccode := zstr.SplitInTwo(bcp, "_")
		if ccode == "" {
			parts := strings.Split(bcp, "-")
			if len(parts) > 2 {
				return parts[0], parts[len(parts)-1]
			}
			return bcp, ""
		}
	}
	if lang == "nb" {
		lang = "no"
	}
	return lang, ccode
}

func LocaleUsesMetric() bool {
	return true
}

func LocaleUsesCelsius() bool {
	return true
}

func LocaleUses24Hour() bool {
	return true
}
