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

type SQLTableView[S zstr.StrIDer] struct {
	TableView[S]
	searchString string
	tableName    string
	selectMethod string
	DeleteQuery  string
	IsSqlite     bool
	IsQuoteIDs   bool
	Constraints  string
	CallerName   string
	// skipFields   []string
	searchFields []string
	showID       int64
	slicePage    []S
	limit        int
	offset       int
}

func NewSQLView[S zstr.StrIDer](tableName, selectMethod string, limit int, options OptionType) (sv *SQLTableView[S]) {
	v := &SQLTableView[S]{}
	v.Init(v, tableName, selectMethod, limit, options)
	return v
}

func (v *SQLTableView[S]) Init(view zview.View, tableName, selectMethod string, limit int, options OptionType) {
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
	v.StoreChangedItemsFunc = v.UpdateItems
	v.DeleteItemsFunc = v.deleteItems
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
		zsql.ForEachColumn(&s, nil, "", func(each zsql.ColumnInfo) bool {
			// zlog.Info("Column:", column, primary, dbTags)
			if each.IsPrimary {
				each.ReflectValue.Set(reflect.Zero(each.ReflectValue.Type()))
			}
			return true
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
	err := zrpc.MainClient.Call(v.CallerName+".InsertRows", []S{s}, nil)
	if err != nil {
		zalert.ShowError(err, "inserting")
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

func (v *SQLTableView[S]) UpdateItems(items []S) {
	// zlog.Info("UpdateItems:", zlog.Full(items))
	v.SetItemsInSlice(items)
	v.UpdateViewFunc() // here we call UpdateViewFunc and not updateView, as just sorted in line above
	err := zrpc.MainClient.Call(v.CallerName+".UpdateRows", items, nil)
	if err != nil {
		zalert.ShowError(err, "updating")
		return
	}
}

func (v *SQLTableView[S]) createConstraints() string {
	var order string
	// zlog.Info("createConstraints", v.Header != nil, v.Header.SortOrder)
	if v.Header != nil {
		var s S
		fieldColMap, primary := zsql.FieldNamesToColumnFromStruct(s, nil, "")
		var orders []string
		for _, s := range v.Header.SortOrder {
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
	err := zrpc.MainClient.Call(v.selectMethod, q, &slice)
	if err != nil {
		zlog.Error(err, "select", q.Constraints, v.limit, v.offset)
		return
	}
	v.UpdateSlice(slice)
}
