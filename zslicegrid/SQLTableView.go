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
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/zsql"
	"github.com/torlangballe/zutil/zstr"
)

type SQLOwner[S zstr.StrIDer] struct {
	Grid *SQLTableView[S]
	// slice         *[]S // we need to store slice when grid is nil
	TableName     string
	rpcCallerName string
	DeleteQuery   string
	IsSqlite      bool
	IsQuoteIDs    bool
	Constraints   string
	limit         int
	offset        int
	slicePage     *[]S
	searchFields  []string
}

type SQLTableView[S zstr.StrIDer] struct {
	TableView[S]
	owner        *SQLOwner[S]
	searchString string
	// selectMethod  string
	// skipFields   []string
	showID int64
}

func (o *SQLOwner[S]) Init(slice *[]S, tableName, rpcCallerName string, limit int) {
	o.slicePage = slice
	o.TableName = tableName
	o.rpcCallerName = rpcCallerName
	o.limit = limit
}

func (o *SQLOwner[S]) NewTable(options OptionType) (sv *SQLTableView[S]) {
	v := &SQLTableView[S]{}
	v.StructName = o.TableName
	v.Init(v, o, options)
	o.Grid = v
	return v
}

func (v *SQLTableView[S]) Init(view zview.View, owner *SQLOwner[S], options OptionType) {
	if v.Header != nil {
		v.Header.SortingPressedFunc = func() {
			go v.owner.UpdateSlice()
		}
	}
	v.owner = owner
	v.SortFunc = nil
	v.TableView.Init(v, v.owner.slicePage, "ztable."+v.owner.TableName, options)
	v.StoreChangedItemsFunc = v.owner.PushRowsToServer
	v.DeleteItemsFunc = v.deleteItems
	if v.options&AddHeader != 0 {
		v.addActionButton()
	}
	go owner.UpdateSlice()
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
				idup := zmenu.MenuedSCFuncAction("Duplcate "+noItems, 'D', 0, func() {
					v.doEdit(ids, true, true, true)
				})
				items = append(items, idup)
			}
			if v.options&AllowEdit != 0 {
				idup := zmenu.MenuedSCFuncAction("Edit "+noItems, 'E', 0, func() {
					v.doEdit(ids, false, false, false)
				})
				items = append(items, idup)
			}
		}
		if v.options&AllowNew != 0 {
			inew := zmenu.MenuedSCFuncAction("New "+v.StructName, 'N', 0, func() {
				var s S
				zfields.CallStructInitializer(&s)
				v.editRows([]S{s}, true)
			})
			items = append(items, inew)
		}
		return items
	}
}

func (v *SQLTableView[S]) doEdit(ids []string, clearPrimary, initStruct, insert bool) {
	var rows []S
	for _, sid := range ids {
		s := *v.StructForID(sid)
		if clearPrimary {
			zsql.ForEachColumn(&s, nil, "", func(each zsql.ColumnInfo) bool {
				// zlog.Info("Column:", column, primary, dbTags)
				if each.IsPrimary {
					each.ReflectValue.Set(reflect.Zero(each.ReflectValue.Type()))
				}
				return true
			})
		}
		rows = append(rows, s)
	}
	v.editRows(rows, insert)
}

func (v *SQLTableView[S]) editRows(rows []S, insert bool) {
	zfields.PresentOKCancelStructSlice(&rows, v.EditParameters, "Edit "+v.StructName, zpresent.AttributesNew(), func(ok bool) bool {
		// zlog.Info("Edited items:", ok, v.StoreChangedItemsFunc != nil)
		if !ok {
			return true
		}
		if insert {
			go v.insertRows(rows)
		} else {
			v.owner.PushRowsToServer(rows)
		}
		return true
	})
}

func (v *SQLTableView[S]) insertRows(s []S) {
	err := zrpc.MainClient.Call(v.owner.rpcCallerName+".InsertRows", s, nil)
	if err != nil {
		zalert.ShowError(err, "inserting")
		return
	}
}

func (v *SQLTableView[S]) deleteItems(ids []string) {
	var affected int64
	if v.owner.IsQuoteIDs {
		for i := range ids {
			ids[i] = zsql.QuoteString(ids[i])
		}
	}
	query := "DELETE FROM " + v.owner.TableName + " WHERE id IN (" + strings.Join(ids, ",") + ")"
	err := zrpc.MainClient.Call("SQLCalls.ExecuteQuery", query, &affected)
	if err != nil {
		zalert.ShowError(err, "updating")
	}
	v.RemoveItemsFromSlice(ids)
	v.updateView()
}

func (o *SQLOwner[S]) createConstraints() string {
	var order string
	// zlog.Info("createConstraints", v.Header != nil, v.Header.SortOrder)
	if o.Grid != nil && o.Grid.Header != nil {
		var s S
		fieldColMap, primary := zsql.FieldNamesToColumnFromStruct(s, nil, "")
		var orders []string
		for _, s := range o.Grid.Header.SortOrder {
			column := fieldColMap[s.FieldName] + " "
			if column == primary {
				continue
			}
			o := column + " "
			if s.SmallFirst {
				o += "ASC"
			} else {
				o += "DESC"
			}
			orders = append(orders, o)
		}
		order = strings.Join(orders, ",")
	}
	cons := o.Constraints
	if o.Grid != nil && cons == "" && o.Grid.searchString != "" && len(o.searchFields) > 0 {
		var wheres []string
		for _, s := range o.searchFields {
			w := s + `ILIKE '%` + zsql.SanitizeString(o.Grid.searchString) + `%'`
			wheres = append(wheres, w)
		}
		cons = "WHERE (" + strings.Join(wheres, " OR ") + ")"
	}
	if order != "" {
		cons += " ORDER BY " + order
	}
	cons += fmt.Sprintf(" LIMIT %d OFFSET %d", o.limit, o.offset)
	// zlog.Info("createConstraints:", cons)
	return cons
}

// func (v *SQLTableView[S]) SetConstraints(constraints string) {
// 	v.owner.Constraints = constraints
// }

func (o *SQLOwner[S]) UpdateRows(rows []S) {
	UpdateRows[S](rows, o.Grid, o.slicePage)
	o.PushRowsToServer(rows)
}

func (o *SQLOwner[S]) UpdateSlice() {
	var slice []S
	var q zsql.QueryBase

	q.Table = o.TableName
	q.Constraints = o.createConstraints()
	err := zrpc.MainClient.Call(o.rpcCallerName+".Select", q, &slice)
	if err != nil {
		zlog.Error(err, "select", q.Constraints, o.limit, o.offset)
		return
	}
	if o.Grid != nil {
		o.Grid.UpdateSlice(slice)
	}
}

func (o *SQLOwner[S]) PushRowsToServer(items []S) {
	// zlog.Info("UpdateItems:", zlog.Full(items))
	// v.SetItemsInSlice(items)
	// v.UpdateViewFunc() // here we call UpdateViewFunc and not updateView, as just sorted in line above
	err := zrpc.MainClient.Call(o.rpcCallerName+".UpdateRows", items, nil)
	if err != nil {
		zalert.ShowError(err, "updating")
		return
	}
}
