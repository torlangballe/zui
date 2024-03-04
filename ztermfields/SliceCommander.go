package ztermfields

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zutil/zcommands"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
)

type SliceCommander struct {
	RowParameters    zfields.FieldParameters
	EditParameters   zfields.FieldParameters
	SlicePointerFunc func() any
	UpdateFunc       func(rowPtr any)
	diffMatch        *diffmatchpatch.DiffMatchPatch
}

const RowUseZTermSliceName = "$zterm"

func (s *SliceCommander) callUpdate(rowPtr any) {
	if s.UpdateFunc != nil {
		s.UpdateFunc(rowPtr)
	}
}

func makeRowParameters() zfields.FieldParameters {
	var p zfields.FieldParameters

	p.UseInValues = []string{RowUseZTermSliceName}
	return p
}

func (s *SliceCommander) In(c *zcommands.CommandInfo, a struct {
	Index int `zui:"desc:index of rows to go into."`
}) string {
	if c.Type == zcommands.CommandHelp {
		return "Enters row <index> as listed with 'rows'.\rShows hierarchy, and then allows editing, and more showing.\rType cd .. to exit."
	}
	if c.Type == zcommands.CommandExpand {
		return ""
	}
	showSliceStruct(s, c, a.Index-1, true)
	return ""
}

func (s *SliceCommander) Rows(c *zcommands.CommandInfo, a struct {
	WildCard string `zui:"desc:Use wildcard to match only text in some rows.,allowempty"`
}) string {
	switch c.Type {
	case zcommands.CommandExpand:
		return ""
	case zcommands.CommandHelp:
		return "lists rows in the table, indexing each row.\rUse the show <index> command to alter one."
	}
	lastFieldValues := map[int]string{}
	outputRows(s, c, "", s.SlicePointerFunc(), a.WildCard, lastFieldValues) //, &edits)
	return ""
}

func (s *SliceCommander) Show(c *zcommands.CommandInfo, a struct {
	Index int `zui:"desc:Index from 'rows' to show fields of."`
}) string {
	switch c.Type {
	case zcommands.CommandExpand:
		return ""
	case zcommands.CommandHelp:
		return "Show fields of a row."
	}
	showSliceStruct(s, c, a.Index-1, false)
	return ""
}

func showSliceStruct(s *SliceCommander, c *zcommands.CommandInfo, index int, goIn bool) {
	sval := reflect.ValueOf(s.SlicePointerFunc()).Elem()
	if index < 0 || index >= sval.Len() {
		c.Session.TermSession.Writeln("Index outside of table length")
		return
	}
	structCommander := &StructCommander{}
	structCommander.StructurePointerFunc = func() any {
		return sval.Index(index).Addr().Interface()
	}
	structCommander.Parameters = s.EditParameters
	structCommander.UpdateFunc = s.UpdateFunc
	if goIn {
		dir := fmt.Sprint(index + 1)
		rval, f, got := zfields.FindIndicatorRValOfStruct(structCommander.StructurePointerFunc())
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
		c.Session.GotoChildNode(dir, structCommander)
	}
	structCommander.Show(c)
}

func outputRows(s *SliceCommander, c *zcommands.CommandInfo, path string, slicePtr any, wildCard string, lastFieldValues map[int]string) { //, edits *[]editRow) {
	s.RowParameters = makeRowParameters()
	sval := reflect.ValueOf(slicePtr).Elem()
	//	tabs := tabwriter.NewWriter(c.Session.TermSession.Writer(), 5, 0, 3, ' ', 0)
	tabs := zstr.NewTabWriter(c.Session.TermSession.Writer())
	tabs.RighAdjustedColumns[0] = true // right-justify index
	tabs.MaxColumnWidth = 60
	for i := 0; i < sval.Len(); i++ {
		rval := sval.Index(i).Addr()
		// if i == 48 || i == 49 {
		s.outputRow(c, tabs, rval, i, wildCard, lastFieldValues)
		// }
	}
	tabs.Flush()
}

func MakeColoredDiff(dmp *diffmatchpatch.DiffMatchPatch, a, b string, w int) string {
	diffs := dmp.DiffMain(a, b, false)
	var count, whiteCount, whites int
	for _, d := range diffs {
		count += len(d.Text)
		if d.Type == diffmatchpatch.DiffEqual {
			whiteCount += len(d.Text)
			whites++
		}
	}
	if count-whiteCount > w/2 {
		return zstr.EscMagenta + zstr.TruncatedFromEnd(b, w, "…") + zstr.EscNoColor
	}
	// zlog.Info("MakeColoredDiff1:", w, dmp.DiffCleanupSemantic(diffs))
	var str string
	var chars int
	spaceForWhite := w - (count - whiteCount)
	for i, d := range diffs {
		var col, add string
		switch d.Type {
		case diffmatchpatch.DiffDelete:
			col = zstr.EscRed
			add = d.Text
		case diffmatchpatch.DiffEqual:
			if spaceForWhite > 0 {
				columns := (spaceForWhite + whites - 1) / whites
				columns = zint.Max(columns, 15)
				col = zstr.EscWhite
				if i == 0 {
					add = zstr.TruncatedFromStart(d.Text, columns, "…")
				} else if i == len(diffs)-1 {
					add = zstr.TruncatedFromEnd(d.Text, columns, "…")
				} else {
					add = zstr.TruncatedMiddle(d.Text, columns, "…")
				}
				zlog.Info("MakeColoredDiff:", len(d.Text), spaceForWhite, whites, add)
				spaceForWhite -= len(add)
				whites--
			}
		case diffmatchpatch.DiffInsert:
			col = zstr.EscGreen
			add = d.Text
		}
		if chars+len(add) >= w {
			str += col + zstr.TruncatedFromEnd(add, w-chars, "…")
			break
		}
		str += col + add
		chars += len(add)
	}
	str += zstr.EscNoColor
	// zlog.Info("MakeColoredDiff:", str, chars, len(str))

	return str
}

func (s *SliceCommander) outputRow(c *zcommands.CommandInfo, tabs *zstr.TabWriter, val reflect.Value, i int, wildCard string, lastFieldValues map[int]string) {
	var line string
	if i == 0 {
		fmt.Fprint(tabs, "\t")
	}
	var tabColumn int = 1 // we start at one, since we output index as well
	keep := (wildCard == "")

	zfields.ForEachField(val.Interface(), s.RowParameters, []zfields.Field{}, func(each zfields.FieldInfo) bool {
		if each.Field.Flags&zfields.FlagIsButton != 0 || each.Field.Flags&zfields.FlagIsImage != 0 {
			if !each.Field.IsImageToggle() {
				return true // skip
			}
		}
		if i == 0 {
			title := each.Field.Name
			if each.Field.Title != "" {
				title = each.Field.Title
			}
			title = strings.Replace(title, " ", "", -1)
			fmt.Fprint(tabs, zstr.EscGreen, title, "\t")
		}
		cols := 60
		if each.Field.Columns != 0 {
			cols = each.Field.Columns
		}
		tabs.MaxColumnWidths[tabColumn] = cols
		sval, skip := getValueString(each.ReflectValue, each.Field, each.StructField, cols, true)
		if skip {
			return true
		}
		replacer := strings.NewReplacer(
			"\n", " • ",
			"\t", " ",
		)
		sval = replacer.Replace(sval)
		if each.Field.Justify&zgeo.Right != 0 {
			tabs.RighAdjustedColumns[tabColumn] = true
		}
		if !keep && wildCard != "" && zstr.MatchWildcard(wildCard, sval) {
			keep = true
		}
		tabColumn++
		var diffed bool
		_, got := each.Field.CustomFields["diff"]
		if got && len(sval) > cols {
			if s.diffMatch == nil {
				s.diffMatch = diffmatchpatch.New()
			}
			last := lastFieldValues[each.FieldIndex]
			lastFieldValues[each.FieldIndex] = sval
			if len(last) > cols && last != sval {
				sval = MakeColoredDiff(s.diffMatch, last, sval, cols)
				diffed = true
			}
		}
		if !diffed {
			sval = zstr.TruncatedFromEnd(sval, cols, "…")
		}
		line += zstr.EscWhite + sval + "\t"
		return true
	})
	if i == 0 {
		fmt.Fprintln(tabs, zstr.EscNoColor) // end header
	}
	if !keep {
		return
	}
	fmt.Fprint(tabs, i+1, ")\t", line, zstr.EscNoColor, "\n")
}
