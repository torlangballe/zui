package zfields

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
)

func GetCommandArgsHelpForStructFields(s any) []zstr.KeyValue {
	var args []zstr.KeyValue
	ForEachField(s, FieldParameters{}, nil, func(each FieldInfo) bool {
		kind := zreflect.KindFromReflectKindAndType(each.ReflectValue.Kind(), each.ReflectValue.Type())
		isPointer := (kind == zreflect.KindPointer)
		name := strings.ToLower(each.Field.FieldName)
		if each.Field.Title != "" {
			name = each.Field.Title
		}
		var arg zstr.KeyValue
		if isPointer {
			eval := reflect.New(each.ReflectValue.Type().Elem()).Elem()
			// zlog.Info("GetCommandArgsHelpForStructFields:", eval, each.Field.Title, each.Field.Name)
			kind := zreflect.KindFromReflectKindAndType(eval.Kind(), eval.Type())
			if len(name) == 1 && kind == zreflect.KindBool {
				arg.Key = fmt.Sprintf("[-%s]", name)
			} else {
				arg.Key = fmt.Sprintf("[--%s=<%s>]", name, kind)
			}
		} else {
			arg.Key = fmt.Sprintf("%s", name)
			if each.Field.Flags&FlagAllowEmptyAsZero != 0 {
				arg.Key = "[" + arg.Key + "]"
			}
		}
		var info string
		info = each.Field.Description
		if info == "" {
			info = each.Field.Tooltip
		}
		if each.Field.HasFlag(FlagHasDefault) {
			info = zstr.Concat(".", info, " default: "+each.Field.Default)
			arg.Key = "[" + arg.Key + "]"
		}
		arg.Value = info
		args = append(args, arg)
		return true
	})
	return args
}

func ParseCommandArgsToStructFields(args []string, rval reflect.Value) error {
	var err error
	var hasAllowEmpty bool
	ForEachField(rval.Interface(), FieldParameters{}, nil, func(each FieldInfo) bool {
		// kind := zreflect.KindFromReflectKindAndType(each.ReflectValue.Kind(), each.ReflectValue.Type())
		kind := each.ReflectValue.Kind()
		isPointer := (kind == reflect.Pointer)
		name := strings.ToLower(each.Field.FieldName)
		if each.Field.Title != "" {
			name = each.Field.Title
		}
		// zlog.Info("ParseCommandArgsToStructFields:", name, isPointer)
		if isPointer {
			if len(args) == 0 {
				return false
			}
			for i, a := range args {
				var set string
				if kind == reflect.Bool && a == "-"+name {
					each.ReflectValue.SetBool(true)
					zslice.RemoveAt(&args, i)
					break
				}
				// zlog.Info("ParseCommandArgsToStructFields:", a, name)
				if zstr.HasPrefix(a, "--"+name+"=", &set) {
					each.ReflectValue.Set(reflect.New(each.ReflectValue.Type().Elem())) // we make it point to an new'ed instance
					err = setStrToRVal(set, each.Field, each.ReflectValue.Elem(), name)
					if err != nil {
						zlog.OnError(err)
						return false
					}
					zslice.RemoveAt(&args, i)
					break
				}
			}
			return true
		}
		var arg string
		if len(args) == 0 {
			if each.Field.HasFlag(FlagHasDefault) {
				arg = each.Field.Default
			} else {
				if each.Field.Flags&FlagAllowEmptyAsZero != 0 {
					if !hasAllowEmpty {
						return true
					}
					hasAllowEmpty = true
				}
				err = zlog.NewError("no argument for", name)
				return false
			}
		} else {
			arg = zstr.ExtractFirstString(&args)
		}
		// zlog.Info("setStr2Val", arg, kind, each.ReflectValue.Type(), name)
		err = setStrToRVal(arg, each.Field, each.ReflectValue, name)
		if err != nil {
			return false
		}
		return true
	})
	return err
}

func setStrToRVal(arg string, f *Field, rval reflect.Value, name string) error {
	kind := zreflect.KindFromReflectKindAndType(rval.Kind(), rval.Type())
	switch kind {
	case zreflect.KindString:
		rval.SetString(arg)
	case zreflect.KindBool:
		b, err := zbool.FromStringWithError(arg)
		if err != nil {
			return err
		}
		rval.SetBool(b)
	case zreflect.KindInt:
		n, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return err
		}
		rval.SetInt(n)
	case zreflect.KindFloat:
		r, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return err
		}
		rval.SetFloat(r)
	default:
		return zlog.NewError("unsupported arg type:", kind, "for", name, "argument")
	}
	return nil // never gets here
}
