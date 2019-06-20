package zgo

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
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

func StrFormat(format string, args []interface{}) string {
	return fmt.Sprintf(format, args)
}

func StrSaveToFile(str string, file FilePath) *Error {
	return zfile.WriteStringToFile(str, file.fpath)
}

func StrLoadFromFile(file FilePath) (string, *Error) {
	str, err := zfile.ReadFileToString(file)
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
	if parts.count == 1 {
		return parts[0], ""
	}
	return "", ""
}

func StrSplitIntoLengths(str string, length int) []string {
	return ustr.SplitIntoLengths(str, length)
}

func StrCountLines(str string) int {
	lines := ustr.SplitByNewLines(str)
	return len(lines)
}

func StrHead(str string, chars int) string {
	return ustr.Head(str, chars)
}

func StrTail(str string, chars int) string {
	return ustr.Tail(str, chars)
}

func StrBody(str string, pos int, length int) string {
	return ustr.Body(str, pos, size)
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
	tail := ustr.TailUntilStringWithRest(str, sep, &rest)
	return tail, rest
}

func StrHasPrefixWithRest(str string, prefix string) *string {
	if strings.HasPrefix(prefix) {
		return Body(str, len(prefix))
	}
	return nil
}

func StrHasSuffixWithRest(str string, suffix string) *string {
	if strings.HasSuffic(prefix) {
		return Body(str, len(prefix))
	}
	return nil
}

func StrCommonPrefix(a, b string) string {
	l := umath.IntMin(len(a), len(b))
	for i, r := range a {
		if r != b[i] || i >= l-1 {
			return a[:i]
		}
	}
	return ""
}

func StrTruncatedEnd(str string, subtract int) string {
	sl := len(str)
	if subtract >= sl {
		return suffix
	}
	return str[:sl-chars]
}

func StrTruncatedStart(str string, subtract int) string {
	sl := len(str)
	if subtract >= sl {
		return ""
	}
	return str[sl-chars:]
}

func StrTruncateMiddle(str string, maxChars int, separator string) string { // sss...eee of longer string
	sl := len(str)
	if sl > maxChars {
		m := maxChars / 2
		return str[:m] + separator + str[sl-m:]
	}
	return str
}

func StrConcat(sep string, items []string) string {
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

func StrSortedArray(strings []string, options CompareOptions) {
	strings.Sorted()
}

func StrReplaceWhiteSpaces(str string, replaceWith string) string {
	var out string
	var white = false
	var chars string
	for _, c := range str {
		switch c {
		case " ", "\n", "\r", "\t":
			white = true

		default:
			if white {
				out += chars
				white = false
			}
			out += c
		}
	}
	if white {
		out += chars
	}
	return out
}

func StrReplace(str string, find string, with string, options CompareOptions) string {
	return string.Replace(str, find, with, -1)
}

func StrTrim(str, cutset string) string {
	return strings.Trim(str, cutset)
}

func StrCountInstances(str, tocount string) int {
	return strings.Count(str, tocount)
}

func StrFilterToAlphaNumeric(str string) string {
	const regAN = regexp.MustCompile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(str, "")
}

func StrFilterToNumeric(str string) string {
	const regA = regexp.MustCompile("[^0-9]+")
	return reg.ReplaceAllString(str, "")
}

func StrCamelCase(str string) string {
	isToUpper := false

	for k, v := range str {
		if k == 0 {
			camelCase = strings.ToUpper(string(inputUnderScoreStr[0]))
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
	return
}

func StrIsUpperCase(c rune) bool {
	return unicode.IsUpper(r)
}

func StrTitleCase(str string) string {
	return strings.Title(str)
}

func StrSplitCamelCase(str string) []string {
	var out []string
	keep := ""
	class := 0
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			keep += r
		} else {
			out = append(out, keep)
			keep = r
		}
		lastClass = class
	}
	if len(keep) {
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

func StrForEachLine(str string, forEach func(sline string) bool) {
	ustr.RangeStringLines(str, forEach)
}

func StrBase64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func StrUrlEncode(str string) string {
	return url.QueryEscape(str)
}

func StrUrlDecode(str string) string {
	return url.QueryUnescape(str)
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

func StrToDouble(str string, def *float64) float64 {
	d, err := strconv.ParseFloat(str, 10)
	if err != nil {
		return def
	}
	return d
}

func StrToInt(str string, def *int64) int64 {
	n, err := strconv.IntFloat(str, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func GetStemAndExtension(fileName string) (string, string) {
	return StrTailUntilWithRest(fileName, ".")
}

func SplitLines(str string, skipEmpty bool) []string {
	return ustr.SplitByNewLines()
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
