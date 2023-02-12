package ztermfields

import (
	"fmt"
	"reflect"
	"text/tabwriter"

	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zutil/zcommands"
	"github.com/torlangballe/zutil/zstr"
)

type SliceCommander struct {
	RowParameters  zfields.FieldParameters
	EditParameters zfields.FieldParameters
	SlicePointer   any
	UpdateFunc     func(rowPtr any)
}

func (s *SliceCommander) callUpdate(rowPtr any) {
	if s.UpdateFunc != nil {
		s.UpdateFunc(rowPtr)
	}
}

func (s *SliceCommander) In(c *zcommands.CommandInfo, index int) string {
	if c.Type == zcommands.CommandHelp {
		return "<index>\tEnters row <index> as listed with show.\nShows hierarchy, and then allows editing, and more showing.\nType cd .. to exit."
	}
	if c.Type == zcommands.CommandExpand {
		return ""
	}
	s.showSliceStruct(c, index-1, true)
	return ""
}

func (s *SliceCommander) Show(c *zcommands.CommandInfo, index *int) string {
	if c.Type == zcommands.CommandHelp {
		return "[index]\tWith no arguments, shows all rows in the table, indexing each row.\nUse edit <index> command to alter a row.\nSpecify an index to show field hierarchy of a row."
	}
	if c.Type == zcommands.CommandExpand {
		return ""
	}
	if index != nil {
		s.showSliceStruct(c, *index-1, false)
		return ""
	}
	// var edits []editRow
	s.outputRows(c, "", s.SlicePointer) //, &edits)
	// s.lastEdits = edits
	return ""
}

func (s *SliceCommander) showSliceStruct(c *zcommands.CommandInfo, index int, goIn bool) {
	sval := reflect.ValueOf(s.SlicePointer).Elem()
	if index < 0 || index >= sval.Len() {
		c.Session.TermSession.Writeln("Index outside of table length")
		return
	}
	com := &StructCommander{}
	com.StructurePointer = sval.Index(index).Addr().Interface()
	com.Parameters = s.EditParameters
	com.UpdateFunc = s.UpdateFunc
	if goIn {
		dir := fmt.Sprint(index + 1)
		rval, f, got := zfields.FindIndicatorRValOfStruct(com.StructurePointer)
		if got {
			dir = fmt.Sprint(rval)
			if f.Enum != "" {
				enum := zfields.GetEnum(f.Enum)
				for _, e := range enum {
					if rval.Equal(reflect.ValueOf(e.Value)) {
						dir = e.Name
					}
				}
			}
		}
		c.Session.GotoChildNode(dir, com)
	}
	com.Show(c)
}

func (s *SliceCommander) outputRows(c *zcommands.CommandInfo, path string, slicePtr any) { //, edits *[]editRow) {
	sval := reflect.ValueOf(slicePtr).Elem()
	tabs := tabwriter.NewWriter(c.Session.TermSession.Writer(), 5, 0, 3, ' ', 0)
	for i := 0; i < sval.Len(); i++ {
		rval := sval.Index(i).Addr()
		s.outputRow(c, tabs, rval, i)
		// *edits = append(*edits, editRow{value: rval, index: i})
	}
	tabs.Flush()
}

func (s *SliceCommander) outputRow(c *zcommands.CommandInfo, tabs *tabwriter.Writer, val reflect.Value, i int) {
	var line string
	if i == 0 {
		fmt.Fprint(tabs, "\t")
	}
	zfields.ForEachField(val.Interface(), s.RowParameters, []zfields.Field{}, func(index int, f *zfields.Field, val reflect.Value, sf reflect.StructField) {
		if f.Flags&zfields.FlagIsButton != 0 || f.Flags&zfields.FlagIsImage != 0 {
			if !f.IsImageToggle() {
				return // skip
			}
		}
		if i == 0 {
			title := f.Name
			if f.Title != "" {
				title = f.Title
			}
			fmt.Fprint(tabs, zstr.EscGreen, title, "\t")
		}
		sval, skip := getValueString(val, f, sf, 60, true)
		if skip {
			return
		}
		line += zstr.EscWhite + sval + "\t"
	})
	if i == 0 {
		fmt.Fprintln(tabs, zstr.EscNoColor) // end header
	}
	fmt.Fprint(tabs, i+1, ")\t", line, zstr.EscNoColor, "\n")
}
