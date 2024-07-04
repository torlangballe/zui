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
	"github.com/torlangballe/zutil/ztimer"
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
	HandleGot     func()
}

type SQLTableView[S zstr.StrIDer] struct {
	TableView[S]
	Owner        *SQLOwner[S]
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
	ztimer.StartIn(0.05, func() { // we give it time to set o.HandleGot etc
		o.GetAndUpdate()
	})
}

func (o *SQLOwner[S]) NewTable(structName string, options OptionType) (sv *SQLTableView[S]) {
	v := &SQLTableView[S]{}
	v.Init(v, o, options)
	v.StructName = structName
	o.Grid = v
	return v
}

func (v *SQLTableView[S]) Init(view zview.View, owner *SQLOwner[S], options OptionType) {
	if v.Header != nil {
		v.Header.SortingPressedFunc = func() {
			go v.Owner.GetAndUpdate()
		}
	}
	v.Owner = owner
	v.SortFunc = nil
	v.TableView.Init(v, v.Owner.slicePage, "ztable."+v.Owner.TableName, options)
	v.StoreChangedItemsFunc = v.Owner.PushRowsToServer
	v.DeleteItemsFunc = v.deleteItems
	if v.options&AddHeader != 0 {
		v.addActionButton()
	}
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
				iedit := zmenu.MenuedSCFuncAction("Edit "+noItems, 'E', 0, func() {
					v.doEdit(ids, false, false, false)
				})
				items = append(items, iedit)
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
		if !ok {
			return true
		}
		if insert {
			go v.Owner.InsertRows(rows)
		} else {
			v.Owner.PushRowsToServer(rows)
		}
		return true
	})
}

func (o *SQLOwner[S]) InsertRows(slice any) {
	err := zrpc.MainClient.Call(o.rpcCallerName+".InsertRows", slice, nil)
	if err != nil {
		zalert.ShowError(err, "inserting")
		return
	}
}

func (v *SQLTableView[S]) deleteItems(ids []string) {
	var affected int64
	if v.Owner.IsQuoteIDs {
		for i := range ids {
			ids[i] = zsql.QuoteString(ids[i])
		}
	}
	zrpc.MainClient.Call(v.Owner.rpcCallerName+".PreDeleteRows", ids, nil)

	query := "DELETE FROM " + v.Owner.TableName + " WHERE id IN (" + strings.Join(ids, ",") + ")"
	err := zrpc.MainClient.Call("SQLCalls.ExecuteQuery", query, &affected)
	if err != nil {
		zalert.ShowError(err, "updating")
	}
	v.RemoveItemsFromSlice(ids)
	v.UpdateViewFunc(true)
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
	if o.Grid != nil {
		UpdateRows[S](rows, o.Grid, o.slicePage)
	}
	o.PushRowsToServer(rows)
}

func (o *SQLOwner[S]) GetAndUpdate() {
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
		o.Grid.UpdateSlice(slice, true)
	} else {
		*o.slicePage = slice
	}
	if o.HandleGot != nil {
		o.HandleGot()
	}
}

func (o *SQLOwner[S]) PushRowsToServer(items []S) {
	// zlog.Info("UpdateItems:", o.TableName, zlog.Full(items))
	// v.SetItemsInSlice(items)
	// v.UpdateViewFunc() // here we call UpdateViewFunc and not updateView, as just sorted in line above
	err := zrpc.MainClient.Call(o.rpcCallerName+".UpdateRows", items, nil)
	if err != nil {
		zalert.ShowError(err, "updating")
		return
	}
}

func (o *SQLOwner[S]) PushRowsToServerWithAnySlice(slice any) {
	// v.SetItemsInSlice(items)
	// v.UpdateViewFunc() // here we call UpdateViewFunc and not updateView, as just sorted in line above
	err := zrpc.MainClient.Call(o.rpcCallerName+".UpdateRows", slice, nil)
	zlog.Info("PushRowsToServerWithAnySlice:", o.TableName, reflect.TypeOf(slice), err, zlog.Full(slice))
	if err != nil {
		zalert.ShowError(err, "updating")
		return
	}
}
