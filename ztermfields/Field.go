package ztermfields

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zutil/zcommands"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/zwords"
)

func sliceValueString(val reflect.Value, maxChars int) string {
	slice, is := val.Interface().([]float64)
	if is && len(slice) <= 6 {
		var parts []string
		for _, f := range slice {
			parts = append(parts, zwords.NiceFloat(f, 1))
		}
		return strings.Join(parts, " ")
	}
	slen := val.Len()
	return zwords.Pluralize("item", slen)
}

func getValueString(val reflect.Value, f *zfields.Field, sf reflect.StructField, maxChars int, setStructString bool) (str string, skip bool) {
	kind := zreflect.KindFromReflectKindAndType(val.Kind(), val.Type())
	zuis, _ := val.Interface().(zfields.UIStringer)
	// zlog.Info("getValueString", f.Name, zuis != nil)
	if zuis != nil {
		return zuis.ZUIString(), false
	}
	var sss struct{}
	if kind == zreflect.KindStruct {
		if sf.Type == reflect.TypeOf(sss) {
			return "", true
		}
		if !setStructString {
			return
		}
	}
	istr := fmt.Sprint(val.Interface())
	if f.Enum != "" {
		enum := zfields.GetEnum(f.Enum)
		for _, e := range enum {
			if val.Equal(reflect.ValueOf(e.Value)) {
				return e.Name, false
			}
		}
		return "[enum]", false
	}

	if kind == zreflect.KindTime {
		t := val.Interface().(time.Time)
		str = ztime.GetNice(t, true)
	} else if kind == zreflect.KindSlice {
		str = sliceValueString(val, maxChars)
	} else {
		str = istr
	}
	return
}

func doRepeatEditIndex(c *zcommands.CommandInfo, what, prompt string, length int, do func(n int) (quit bool)) {
	c.Session.TermSession.Writeln("Press ["+zstr.EscYellow+"return"+zstr.EscNoColor+"] to quit", what)
	for {
		c.Session.TermSession.Write(zstr.EscYellow + prompt + zstr.EscNoColor + " ")
		sval, _ := c.Session.TermSession.ReadValueLine()
		if sval == "" {
			break
		}
		n, _ := strconv.Atoi(sval)
		if n <= 0 || n > length { // n is 1 -- x
			c.Session.TermSession.Writeln("Field number outside range")
			continue
		}
		if do(n - 1) {
			break
		}
	}
}
