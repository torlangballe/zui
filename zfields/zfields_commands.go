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
	ForEachField(s, FieldParameters{}, nil, func(index int, f *Field, val reflect.Value, sf reflect.StructField) bool {
		kind := zreflect.KindFromReflectKindAndType(val.Kind(), val.Type())
		isPointer := (kind == zreflect.KindPointer)
		name := strings.ToLower(f.FieldName)
		var arg zstr.KeyValue
		if isPointer {
			eval := reflect.New(val.Type().Elem()).Elem()
			zlog.Info("GetCommandArgsHelpForStructFields:", eval)
			kind := zreflect.KindFromReflectKindAndType(eval.Kind(), eval.Type())
			zlog.Info("GetCommandArgsHelpForStructFields2:", f.Name, kind)
			if len(name) == 1 && kind == zreflect.KindBool {
				arg.Key = fmt.Sprintf("[-%s]", name)
			} else {
				arg.Key = fmt.Sprintf("[%s=<%s>]", name, kind)
			}
		} else {
			arg.Key = fmt.Sprintf("%s", name)
			if f.Flags&FlagAllowEmptyAsZero != 0 {
				arg.Key = "[" + arg.Key + "]"
			}
		}
		arg.Value = f.Description
		if arg.Value == "" {
			arg.Value = f.Tooltip
		}
		args = append(args, arg)
		return true
	})
	return args
}

func ParseCommandArgsToStructFields(args []string, rval reflect.Value) error {
	var err error
	var hasAllowEmpty bool
	ForEachField(rval.Interface(), FieldParameters{}, nil, func(index int, f *Field, val reflect.Value, sf reflect.StructField) bool {
		kind := zreflect.KindFromReflectKindAndType(val.Kind(), val.Type())
		isPointer := (kind == zreflect.KindPointer)
		name := strings.ToLower(f.FieldName)
		zlog.Info("ParseCommandArgsToStructFields:", name, isPointer)
		if isPointer {
			val = reflect.New(val.Type().Elem()).Elem()
			if len(args) == 0 {
				return false
			}
			for i, a := range args {
				var set string
				if kind == zreflect.KindBool && a == "-"+name {
					val.SetBool(true)
					zslice.RemoveAt(&args, i)
					break
				}
				if zstr.HasPrefix(a, "--"+name+"=", &set) {
					err = setStrToRVal(set, kind, f, val, name)
					if err != nil {
						return false
					}
					zslice.RemoveAt(&args, i)
					break
				}
			}
			return true
		}
		if len(args) == 0 {
			if f.Flags&FlagAllowEmptyAsZero != 0 {
				if !hasAllowEmpty {
					return true
				}
				hasAllowEmpty = true
			}
			err = zlog.NewError("no argument for", name)
			return false
		}
		arg := zstr.ExtractFirstString(&args)
		zlog.Info("setStr2Val", arg, kind, val.Type(), name)
		err = setStrToRVal(arg, kind, f, val, name)
		if err != nil {
			return false
		}
		return true
	})
	return err
}

func setStrToRVal(arg string, kind zreflect.TypeKind, f *Field, rval reflect.Value, name string) error {
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
