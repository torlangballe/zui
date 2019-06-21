package zgo

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/torlangballe/zutil/uinteger"
	"github.com/torlangballe/zutil/umath"
	"github.com/torlangballe/zutil/ustr"
	"github.com/torlangballe/zutil/zfile"
)

//  Created by Tor Langballe on /23/9/14.

func StrRemovedFirst(s []string) []string {
	if len(s) == 0 {
		return s
	}
	return s[1:]
}

func StrFormat(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args)
}

func StrSaveToFile(str string, file FilePath) *Error {
	err := zfile.WriteStringToFile(str, file.fpath)
	return ErrorFromErr(err)
}

func StrLoadFromFile(file FilePath) (string, *Error) {
	str, err := zfile.ReadFileToString(file.fpath)
	return str, ErrorFromErr(err)
}

func StrFindFirstOfChars(str string, charset string) int {
	return strings.IndexAny(str, charset)
}

func StrFindLastOfChars(str string, charset string) int {
	return strings.LastIndexAny(str, charset)
}

func StrJoin(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

func StrSplit(str string, sep string) []string {
	return strings.Split(str, sep)
}

func StrSplitByChars(str string, chars string) []string {
	return strings.FieldsFunc(str, func(r rune) bool {
		return strings.IndexRune(chars, r) != -1
	})

}

func StrSplitN(str string, sep string, n int) []string {
	return strings.SplitN(str, sep, n)
}

func StrSplitInTwo(str string, sep string) (string, string) {
	parts := strings.SplitN(str, sep, 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return "", ""
}

func StrSplitIntoLengths(str string, length int) []string {
	return ustr.SplitIntoLengths(str, length)
}

func StrCountLines(str string) int {
	skipEmpty := false
	lines := ustr.SplitByNewLines(str, skipEmpty)
	return len(lines)
}

func StrHead(str string, chars int) string {
	return ustr.Head(str, chars)
}

func StrTail(str string, chars int) string {
	return ustr.Tail(str, chars)
}

func StrBody(str string, pos int, length int) string {
	return ustr.Body(str, pos, length)
}

func StrHeadUntilWithRest(str string, sep string) (string, string) {
	var rest string
	head := ustr.HeadUntilStringWithRest(str, sep, &rest)
	return head, rest
}

func StrHeadUntil(str, sep string) string {
	return ustr.HeadUntilString(str, sep)
}

func StrTailUntil(str, sep string) string {
	return ustr.TailUntil(str, sep)
}

func StrTailUntilWithRest(str, sep string) (string, string) { // tail, then rest
	var rest string
	tail := ustr.TailUntilWithRest(str, sep, &rest)
	return tail, rest
}

func StrHasPrefixWithRest(str string, prefix string) (bool, string) {
	if strings.HasPrefix(str, prefix) {
		return true, StrBody(str, len(prefix), -1)
	}
	return false, ""
}

func StrHasSuffixWithRest(str string, suffix string) (bool, string) {
	if strings.HasSuffix(str, suffix) {
		return true, StrBody(str, len(suffix), -1)
	}
	return false, ""
}

func StrCommonPrefix(a, b string) string {
	l := umath.IntMin(len(a), len(b))
	br := []rune(b)
	for i, r := range a {
		if r != br[i] || i >= l-1 {
			return string(br[:i])
		}
	}
	return ""
}

func StrTruncatedEnd(str string, subtract int, suffix string) string {
	sl := len(str)
	if subtract >= sl {
		return suffix
	}
	return str[:sl-subtract] + suffix
}

func StrTruncatedStart(str string, subtract int, prefix string) string {
	sl := len(str)
	if subtract >= sl {
		return prefix
	}
	return prefix + str[sl-subtract:]
}

func StrTruncateMiddle(str string, maxChars int, separator string) string { // sss...eee of longer string
	sl := len(str)
	if sl > maxChars {
		m := maxChars / 2
		return str[:m] + separator + str[sl-m:]
	}
	return str
}

func StrConcat(sep string, items ...string) string {
	var str = ""
	var first = true
	for _, s := range items {
		if str != "" && s != "" {
			str += sep
		}
		str += s
	}
	return str
}

type CompareOptions int

const (
	CompareDefault                              = 0
	CompareReverse               CompareOptions = 1
	CompareCaseless                             = 2
	CompareAlphaFirst                           = 4
	CompareWithoutSimplePrefixes                = 8
)

func StrCompare(a string, b string, options CompareOptions) bool {
	return strings.Compare(a, b) > 0
}

func StrSortedArray(strs []string, options CompareOptions) []string {
	// TODO: options
	s := make([]string, len(strs))
	copy(s, strs)
	sort.StringSlice(s).Sort()

	return s
}

func StrMap(str string, convert func(i int, r rune) string) string {
	return ustr.Map(str, convert)
}

func StrReplaceWhiteSpaces(str string, replaceWith string) string {
	// TOD: check if works
	var reg = regexp.MustCompile(`(\s+)`)
	return reg.ReplaceAllString(str, str)
}

func StrReplace(str string, find string, with string, options CompareOptions) string {
	return strings.Replace(str, find, with, -1)
}

func StrTrim(str, cutset string) string {
	return strings.Trim(str, cutset)
}

func StrCountInstances(str, tocount string) int {
	return strings.Count(str, tocount)
}

var regAN = regexp.MustCompile("[^a-zA-Z0-9]+")

func StrFilterToAlphaNumeric(str string) string {
	return regAN.ReplaceAllString(str, "")
}

var regA = regexp.MustCompile("[^0-9]+")

func StrFilterToNumeric(str string) string {
	return regA.ReplaceAllString(str, "")
}

func StrCamelCase(str string) string {
	isToUpper := false
	var camelCase string
	for k, v := range str {
		if k == 0 {
			camelCase = strings.ToUpper(string(v))
		} else {
			if isToUpper {
				camelCase += strings.ToUpper(string(v))
				isToUpper = false
			} else {
				if v == ' ' || v == '_' {
					isToUpper = true
				} else {
					camelCase += string(v)
				}
			}
		}
	}
	return camelCase
}

func StrToUpper(str string) string {
	return strings.ToUpper(str)
}

func StrToLower(str string) string {
	return strings.ToLower(str)
}

func StrIsUpperCase(r rune) bool {
	return unicode.IsUpper(r)
}

func StrTitleCase(str string) string {
	return strings.Title(str)
}

func StrSplitCamelCase(str string) []string {
	var out []string
	keep := ""
	lastClass := 0
	for _, r := range str {
		class := 4
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		}
		if class == lastClass {
			keep += string(r)
		} else {
			out = append(out, keep)
			keep = string(r)
		}
		lastClass = class
	}
	if keep != "" {
		out = append(out, keep)
	}
	return out
}
func StrHashToU64(str string) uint64 {
	return uinteger.HashTo64(str)
}

func StrMakeHashTagWord(str string) string {
	// let split = ZStr.SplitByChars(str, chars:" .-,/()_")
	// var words = []string]()
	// for s in split {
	//     words += ZStr.SplitCamelCase(ZStr.FilterToAlphaNumeric(s))
	// }
	// let flat = words.reduce("") { $0 + $1.capitalized }
	// return flat
	return ""
}

// func Unescape(str string)  string {
//         var vstr = str.replacingOccurrences(of: "\\n", with:"\n")
//         vstr = vstr.replacingOccurrences(of: "\\r", with:"\r")
//         vstr = vstr.replacingOccurrences(of: "\\t", with:"\t")
//         vstr = vstr.replacingOccurrences(of: "\\\"", with:"\"")
//         vstr = vstr.replacingOccurrences(of: "\\'", with:"'")
//         vstr = vstr.replacingOccurrences(of: "\\\\", with:"\\")

//         return vstr
//     }

func StrForEachLine(str string, skipEmpty bool, forEach func(sline string)) {
	ustr.RangeStringLines(str, skipEmpty, forEach)
}

func StrBase64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func StrUrlEncode(str string) string {
	return url.QueryEscape(str)
}

func StrUrlDecode(str string) string {
	s, _ := url.QueryUnescape(str)
	return s
}

func StrMatchsWildcard(str string, wild string) bool {
	matched, _ := filepath.Match(wild, str)
	return matched

}

// func CopyToCCharArray(carray:UnsafeMutablePointer<int8>, str string) { // this requires pointer to FIRST tuple using .0
//     let c = str.utf8.count + 1
//     = strlcpy(carray, str, c) // str is coerced to c-string amazingly enough
// }

// func StrFromCCharArray(carray:UnsafeMutablePointer<int8>?)  string { // this requires pointer to FIRST tuple using .0 if it's char[256] etc
//     if carray == nil {
//         return ""
//     }
//     return string(cstring:carray!)
// }

// func CopyStrToAllocedCStr(str string, len:int)  UnsafeMutablePointer<int8> {
//     let pointer = UnsafeMutablePointer<int8>.allocate(capacity:len)
//     strcpy(pointer, str)
//     return pointer
// }

func StrNiceDouble(d float64, maxSig int, separator string) string {
	str := ustr.NiceFloat(d, maxSig)
	return str
}

func StrToDouble(str string, def float64) (float64, bool) {
	d, err := strconv.ParseFloat(str, 10)
	if err != nil {
		return def, false
	}
	return d, true
}

func StrToInt(str string, def int64) (int64, bool) {
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return def, false
	}
	return n, true
}

func GetStemAndExtension(fileName string) (string, string) {
	return StrTailUntilWithRest(fileName, ".")
}

func SplitLines(str string, skipEmpty bool) []string {
	return ustr.SplitByNewLines(str, skipEmpty)
	// handle empty...
}

func Base64CharToNumber(char int) int {
	if char >= 'A' && char <= 'Z' {
		return char - 'A'
	}
	if char >= 'a' && char <= 'z' {
		return char - 'a' + 26
	}
	if char >= '0' && char <= '9' {
		return char - '0' + 26 + 26
	}
	if char == '+' {
		return 62
	}
	if char == '/' {
		return 63
	}
	return -1
}

func NumberToBase64string(num int) string {
	return string(rune(NumberToBase64Char(num)))
}

func NumberToBase64Char(num int) int {
	if num >= 0 && num <= 25 {
		return 'A' + num
	}
	if num >= 26 && num <= 51 {
		return 'a' + num - 26
	}

	if num >= 52 && num <= 62 {
		return '0' + num - 26 - 26
	}
	if num == 62 {
		return '+'
	}
	if num == 63 {
		return '/'
	}
	return -1
}
