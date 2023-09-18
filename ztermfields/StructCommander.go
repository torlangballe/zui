package ztermfields

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zutil/zcommands"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zstr"
)

type StructCommander struct {
	Parameters       zfields.FieldParameters
	StructurePointer any
	UpdateFunc       func(structPtr any)
	lastEdits        []editField
}

type editField struct {
	value reflect.Value
	field *zfields.Field
	key   string
}

const checkedString = zstr.EscMagenta + " [√]" + zstr.EscNoColor

func (s *StructCommander) callUpdate() {
	if s.UpdateFunc != nil {
		s.UpdateFunc(s.StructurePointer)
	}
}

func (s *StructCommander) Show(c *zcommands.CommandInfo) string {
	if c.Type == zcommands.CommandHelp {
		return "Show all fields in the hierarchy, indexing editable ones.\nUse edit command to alter a field."
	}
	if c.Type == zcommands.CommandExpand {
		return ""
	}
	var edits []editField
	outputFields(s, c, "", s.StructurePointer, 0, 0, false, &edits)
	s.lastEdits = edits
	return ""
}

func (s *StructCommander) Edit(c *zcommands.CommandInfo, a struct{ Index int }) string {
	if c.Type == zcommands.CommandHelp {
		return "Edit a row's fields using an index from show command."
	}
	if c.Type == zcommands.CommandExpand {
		return ""
	}
	if len(s.lastEdits) == 0 {
		c.Session.TermSession.Writeln("No field hierarchy to edit specific field of")
		return ""
	}
	if a.Index <= 0 || a.Index > len(s.lastEdits) {
		c.Session.TermSession.Writeln("Field out of range:", a.Index)
		return ""
	}
	editIndex(s, c, s.lastEdits, a.Index-1)
	return ""
}

func editIndex(s *StructCommander, c *zcommands.CommandInfo, edits []editField, n int) {
	edit := edits[n]
	kind := zreflect.KindFromReflectKindAndType(edit.value.Kind(), edit.value.Type())
	// zlog.Info("editIndex:", kind, edit.value.Interface(), zlog.Full(*edit.field))
	if kind == zreflect.KindPointer {
		edit.value = reflect.New(edit.value.Type().Elem()).Elem()
	}

	if kind == zreflect.KindSlice {
		editSliceIndicator(s, c, edit)
		return
	}
	if edit.field.Enum != "" {
		editEnumIndicator(s, c, edit)
		return
	}
	if edit.field.LocalEnum != "" {
		editLocalEnumIndicator(s, c, edit)
		return
	}
	c.Session.TermSession.Write(zstr.EscGreen+edit.field.Name+zstr.EscNoColor, ": ")
	sval, err := c.Session.TermSession.ReadValueLine()
	if err == io.EOF {
		return
	}
	switch kind {
	case zreflect.KindString:
		edit.value.SetString(sval)
		s.callUpdate()
	case zreflect.KindInt, zreflect.KindFloat:
		n, err := strconv.ParseFloat(sval, 64)
		if err != nil {
			c.Session.TermSession.Writeln(err)
			return
		}
		if kind == zreflect.KindInt {
			edit.value.SetInt(int64(n))
		} else {
			edit.value.SetFloat(n)
		}
		s.callUpdate()
	}
}

func editEnumIndicator(s *StructCommander, c *zcommands.CommandInfo, edit editField) {
	enum := zfields.GetEnum(edit.field.Enum)
	// selectedIndex := -1
	for i, e := range enum {
		c.Session.TermSession.Write(i+1, ") ", e.Name)
		if edit.value.Equal(reflect.ValueOf(e.Value)) {
			c.Session.TermSession.Write(" " + checkedString)
			// selectedIndex = i
			// break
		}
		c.Session.TermSession.Writeln("")
	}
	doRepeatEditIndex(c, "setting index", "Set Index No:", len(enum), func(n int) bool {
		edit.value.Set(reflect.ValueOf(enum[n].Value))
		c.Session.TermSession.Writeln(zstr.EscGreen+edit.field.Name, zstr.EscNoColor+"set to", enum[n].Name)
		s.callUpdate()
		return true
	})
}

func editLocalEnumIndicator(s *StructCommander, c *zcommands.CommandInfo, edit editField) {
	zlog.Info("editLocalEnumIndicator:", edit.value.Interface(), zlog.Full(*edit.field))
	ei, findex := zfields.FindLocalFieldWithFieldName(s.StructurePointer, edit.field.LocalEnum)
	zlog.Assert(findex != -1, edit.field.Name, edit.field.LocalEnum)
	enum := ei.Interface().(zdict.ItemsGetter).GetItems()
	for i, e := range enum {
		c.Session.TermSession.Write(i+1, ") ", e.Name)
		if edit.value.Equal(reflect.ValueOf(e.Value)) {
			c.Session.TermSession.Write(" " + checkedString)
		}
		c.Session.TermSession.Writeln("")
	}
	doRepeatEditIndex(c, "setting index", "Set Index No:", len(enum), func(n int) bool {
		edit.value.Set(reflect.ValueOf(enum[n].Value))
		c.Session.TermSession.Writeln(zstr.EscGreen+edit.field.Name, zstr.EscNoColor+"set to", enum[n].Name)
		s.callUpdate()
		return true
	})
}

func editSliceIndicator(s *StructCommander, c *zcommands.CommandInfo, edit editField) bool {
	sliceVal := edit.value
	if sliceVal.Kind() == reflect.Pointer {
		sliceVal = sliceVal.Elem()
	}
	length := sliceVal.Len()
	zlog.Assert(length != 0)
	indicatorName := zfields.FindIndicatorOfSlice(sliceVal.Interface())
	c.Session.TermSession.Writeln("Set", zstr.EscCyan+edit.field.Name+zstr.EscNoColor, "index:")
	lastUsedID, _ := zkeyvalue.DefaultStore.GetString(edit.key)
	var ids, titles []string
	for j := 0; j < length; j++ {
		a := sliceVal.Index(j).Addr().Interface()
		id := zstr.GetIDFromAnySliceItemWithIndex(a, j)
		fval, _, findex := zreflect.FieldForName(a, true, indicatorName)
		var title string
		if findex == -1 {
			title = fmt.Sprint(j + 1)
		} else {
			title = fmt.Sprint(fval.Interface())
		}
		ids = append(ids, id)
		titles = append(titles, title)
		var current string
		if id == lastUsedID || lastUsedID == "" && j == 0 {
			current = checkedString
		}
		c.Session.TermSession.Write(j+1, ") ", title, current, "\n")
	}
	doRepeatEditIndex(c, "setting index", "Set Index No:", length, func(n int) bool {
		zlog.Info("SetIndexForSlice:", ids[n], n)
		zkeyvalue.DefaultStore.SetString(ids[n], edit.key, true)
		c.Session.TermSession.Writeln("Set index for", zstr.EscCyan+edit.field.Name+zstr.EscNoColor, "to:", titles[n])
		s.callUpdate()
		return true
	})
	return true
}

func outputFields(s *StructCommander, c *zcommands.CommandInfo, path string, structurePtr any, i int, indent int, inStatic bool, edits *[]editField) int {
	zfields.ForEachField(structurePtr, s.Parameters, []zfields.Field{}, func(index int, f *zfields.Field, val reflect.Value, sf reflect.StructField) bool {
		// col := indentColors[indent%len(indentColors)]
		if f.Flags&zfields.FlagIsButton != 0 || f.Flags&zfields.FlagIsImage != 0 {
			if !f.IsImageToggle() {
				return true // skip
			}
		}
		kind := zreflect.KindFromReflectKindAndType(val.Kind(), val.Type())
		sval, skip := getValueString(val, f, sf, 3000, false)
		if skip {
			return true
		}
		sindent := strings.Repeat(" ", 2*indent)
		var readOnlyStruct bool
		if val.Kind() == reflect.Struct {
			if sval == "" {
				c.Session.TermSession.Writeln(sindent + zstr.EscMagenta + f.Name + zstr.EscNoColor)
				// zlog.Info("AddSubStructEdit:", reflect.TypeOf(structurePtr), f.Name, val.Type(), val.CanAddr())
				i = outputFields(s, c, path+"/"+sf.Name, val.Addr().Interface(), i, indent+1, f.IsStatic(), edits)
				return true
			}
			if f.LocalEnum == "" {
				_, got := val.Interface().(zfields.UISetStringer)
				readOnlyStruct = !got
			}
		}
		var pre string
		fpath := path + "/" + sf.Name
		if kind == zreflect.KindSlice && val.Type().Elem().Kind() == reflect.Struct {
			sliceVal := val
			if val.CanAddr() {
				sliceVal = val.Addr()
			}
			i = outputSlice(s, c, pre, fpath, sliceVal, f, i, indent, edits)
			return true
		}
		if !f.IsStatic() && !inStatic && !readOnlyStruct && s.UpdateFunc != nil {
			pre = fmt.Sprintf("%d) ", i+1)
			zlog.Info("AddEdit:", reflect.TypeOf(structurePtr), f.Name, val.Type(), val.CanAddr())
			*edits = append(*edits, editField{value: val, field: f})
			i++
		}
		c.Session.TermSession.Write(sindent, zstr.EscGreen, pre, f.Name, zstr.EscNoColor, " ", sval, "\n")
		return true
	})
	return i
}

func outputSlice(s *StructCommander, c *zcommands.CommandInfo, pre, path string, sliceVal reflect.Value, f *zfields.Field, i int, indent int, edits *[]editField) int {
	sindent := strings.Repeat(" ", 2*indent)
	if sliceVal.Kind() == reflect.Pointer {
		sliceVal = sliceVal.Elem()
	}
	length := sliceVal.Len()
	indicatorName := zfields.FindIndicatorOfSlice(sliceVal.Interface())
	if length == 0 {
		c.Session.TermSession.Write(sindent, zstr.EscCyan, pre, f.Name, zstr.EscNoColor, " [empty]", "\n")
		return i
	}
	pre = fmt.Sprintf("%d) ", i+1)
	key := fmt.Sprintf("ztermfields.StructCommander/SliceIdentifier/userid:%d%s", c.Session.TermSession.UserID(), c.Session.Path()+path) // path starts with /, so append without /
	*edits = append(*edits, editField{value: sliceVal, field: f, key: key})
	i++
	lastUsedID, _ := zkeyvalue.DefaultStore.GetString(key)
	for j := 0; j < length; j++ {
		a := sliceVal.Index(j).Addr().Interface()
		fval, _, indicatorIndex := zreflect.FieldForName(a, true, indicatorName)
		id := zstr.GetIDFromAnySliceItemWithIndex(a, j)
		if lastUsedID == id || lastUsedID == "" || j == length-1 {
			title := fmt.Sprint(fval)
			if indicatorIndex == -1 || lastUsedID == "" {
				title = fmt.Sprint(j + 1)
			}
			c.Session.TermSession.Write(sindent, zstr.EscCyan, pre, f.Name, zstr.EscNoColor, " [", title, "/", length, "]\n")
			i = outputFields(s, c, path, a, i, indent+1, f.IsStatic(), edits)
			break
		}
	}
	return i
}
