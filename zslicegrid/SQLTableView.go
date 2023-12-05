//go:build zui

package zslicegrid

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zmenu"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/zsql"
	"github.com/torlangballe/zutil/zstr"
)

type SQLTableView[S zstr.StrIDer] struct {
	TableView[S]
	searchString  string
	tableName     string
	selectMethod  string
	DeleteQuery   string
	IsSqlite      bool
	IsQuoteIDs    bool
	Constraints   string
	skipFields    []string
	searchFields  []string
	showID        int64
	equalFields   map[string]string
	setFields     map[string]string
	fieldIsString map[string]bool
	slicePage     []S
	limit         int
	offset        int
}

func NewSQLView[S zstr.StrIDer](tableName, selectMethod string, limit int, options OptionType) (sv *SQLTableView[S]) {
	v := &SQLTableView[S]{}
	v.Init(v, tableName, selectMethod, limit, options)
	return v
}

func (v *SQLTableView[S]) Init(view zview.View, tableName, selectMethod string, limit int, options OptionType) {
	var s S
	v.tableName = tableName
	v.selectMethod = selectMethod
	v.limit = limit
	if v.Header != nil {
		v.Header.SortingPressedFunc = func() {
			go v.FillPage()
		}
	}
	v.SortFunc = nil
	v.TableView.Init(v, &v.slicePage, "ztable."+tableName, options)
	v.StoreChangedItemsFunc = v.updateForIDs
	v.DeleteItemsFunc = v.deleteItems
	v.equalFields = map[string]string{}
	v.setFields = map[string]string{}
	v.fieldIsString = map[string]bool{}
	zsql.ForEachColumn(s, nil, "", func(rval reflect.Value, sf reflect.StructField, column string, primary bool) {
		zlog.Info("Column:", column, primary)
		dbTags := zreflect.GetTagAsMap(string(sf.Tag))["db"]
		if primary {
			v.equalFields[sf.Name] = column
		}
		if primary { // part == "-" ||
			v.EditParameters.SkipFieldNames = append(v.EditParameters.SkipFieldNames, sf.Name)
			return
		}
		if rval.Kind() == reflect.String {
			v.fieldIsString[sf.Name] = true
		}
		if !primary {
			v.setFields[sf.Name] = column
		}
		if zstr.StringsContain(dbTags, "search") {
			v.searchFields = append(v.searchFields, column)
		}
		return
	})
	if v.options&AddHeader != 0 {
		v.addActionButton()
	}
	go v.FillPage()
}

func (v *SQLTableView[S]) addActionButton() {
	v.ActionMenu.CreateItemsFunc = func() []zmenu.MenuedOItem {
		var items []zmenu.MenuedOItem
		ids := v.Grid.SelectedIDs()
		noItems := v.NameOfXItemsFunc(ids, true)
		if len(ids) > 0 {
			if v.options&AllowDelete != 0 {
				idel := zmenu.MenuedSCFuncAction("Delete "+noItems+"…", zkeyboard.KeyBackspace, 0, func() {
					v.handleDeleteKey(true)
				})
				items = append(items, idel)
			}
			if v.options&AllowDuplicate != 0 {
				idup := zmenu.MenuedFuncAction("Duplcate "+noItems, func() {
					v.addNewAction(true)
				})
				items = append(items, idup)
			}
		}
		if v.options&AllowNew != 0 {
			inew := zmenu.MenuedFuncAction("New "+v.StructName, func() {
				v.addNewAction(false)
			})
			items = append(items, inew)
		}
		return items
	}
}

func (v *SQLTableView[S]) addNewAction(duplicate bool) {
	var s S
	if duplicate {
		sid := v.Grid.SelectedIDs()[0]
		s = *v.StructForID(sid)
		zsql.ForEachColumn(&s, nil, "", func(val reflect.Value, sf reflect.StructField, column string, primary bool) {
			// zlog.Info("Column:", column, primary, dbTags)
			if primary {
				val.Set(reflect.Zero(val.Type()))
			}
		})
	}
	zfields.PresentOKCancelStruct(&s, v.EditParameters, "Edit "+v.StructName, zpresent.AttributesNew(), func(ok bool) bool {
		// zlog.Info("Edited items:", ok, v.StoreChangedItemsFunc != nil)
		if !ok {
			return true
		}
		go v.insertRow(s)
		return true
	})
}

func (v *SQLTableView[S]) insertRow(s S) {
	var info zsql.UpsertInfo
	var offset int64
	info.Rows = []S{s}
	info.TableName = v.tableName
	info.SetColumns = v.setFields
	info.EqualColumns = v.equalFields

	first := v.setFields[v.Header.SortOrder[0].FieldName]
	val, _, findex := zreflect.FieldForName(&s, zfields.FlattenIfAnonymousOrZUITag, first)
	if zlog.ErrorIf(findex == -1, first) {
		return
	}
	sval := fmt.Sprint(val)
	if v.fieldIsString[first] {
		sval = zsql.QuoteString(sval)
	}
	info.OffsetQuery = fmt.Sprintf("SELECT COUNT(*) FROM ", v.tableName, " WHERE ", first, "<", sval)
	zlog.Info("InserOffQ:", info.OffsetQuery)
	err := zrpc.MainClient.Call("SQLCalls.InsertRows", info, &offset)
	zlog.Info("insert", err)
	if err != nil {
		zalert.ShowError(err, "updating")
		return
	}
}

func (v *SQLTableView[S]) deleteItems(ids []string) {
	var affected int64
	if v.IsQuoteIDs {
		for i := range ids {
			ids[i] = zsql.QuoteString(ids[i])
		}
	}
	query := "DELETE FROM " + v.tableName + " WHERE id IN (" + strings.Join(ids, ",") + ")"
	err := zrpc.MainClient.Call("SQLCalls.ExecuteQuery", query, &affected)
	if err != nil {
		zalert.ShowError(err, "updating")
	}
	v.RemoveItemsFromSlice(ids)
	v.updateView()
}

func (v *SQLTableView[S]) updateForIDs(items []S) {
	var info zsql.UpsertInfo
	info.Rows = items
	info.TableName = v.tableName
	info.SetColumns = v.setFields
	info.EqualColumns = v.equalFields
	zlog.Info("updateForIDs1", zlog.Full(items))
	err := zrpc.MainClient.Call("SQLCalls.UpdateRows", info, nil)
	zlog.Info("updateForIDs", len(items), err)
	if err != nil {
		zalert.ShowError(err, "updating")
	}
}

func (v *SQLTableView[S]) createConstraints() string {
	// var s S
	var order string
	// zlog.Info("createConstraints", v.Header != nil, v.Header.SortOrder)
	if v.Header != nil {
		var orders []string
		for _, s := range v.Header.SortOrder {
			o := v.setFields[s.FieldName] + " "
			if s.SmallFirst {
				o += "ASC"
			} else {
				o += "DESC"
			}
			orders = append(orders, o)
		}
		order = strings.Join(orders, ",")
	}
	cons := v.Constraints
	if cons == "" && v.searchString != "" && len(v.searchFields) > 0 {
		var wheres []string
		for _, s := range v.searchFields {
			w := s + `ILIKE '%` + zsql.SanitizeString(v.searchString) + `%'`
			wheres = append(wheres, w)
		}
		cons = "WHERE (" + strings.Join(wheres, " OR ") + ")"
	}
	if order != "" {
		cons += " ORDER BY " + order
	}
	cons += fmt.Sprintf(" LIMIT %d OFFSET %d", v.limit, v.offset)
	// zlog.Info("createConstraints:", cons)
	return cons
}

func (v *SQLTableView[S]) SetConstraints(constraints string) {
	v.Constraints = constraints
}

func (v *SQLTableView[S]) FillPage() {
	var slice []S
	var q zsql.QueryBase

	q.Table = v.tableName
	q.Constraints = v.createConstraints()
	q.SkipFields = v.skipFields
	err := zrpc.MainClient.Call(v.selectMethod, q, &slice)
	if err != nil {
		zlog.Error(err, "select", q.Constraints, v.limit, v.offset)
		return
	}
	v.UpdateSlice(slice)
}
