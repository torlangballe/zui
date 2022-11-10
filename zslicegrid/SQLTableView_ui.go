//go:build zui

package zslicegrid

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zrpc2"
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
	Where        string
	skipFields   []string
	searchFields []string
	equalFields  map[string]string
	setFields    map[string]string
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
	var s S
	v.tableName = tableName
	v.selectMethod = selectMethod
	v.limit = limit
	if v.Header != nil {
		v.Header.SortingPressedFunc = func() {
			go v.fillPage()
		}
	}
	v.TableView.Init(v, &v.slicePage, "ztable."+tableName, options)
	v.StoreChangedItemsFunc = v.updateForIDs
	v.equalFields = map[string]string{}
	v.setFields = map[string]string{}
	zreflect.ForEachField(s, func(index int, val reflect.Value, sf reflect.StructField) {
		var column string
		tags := zreflect.GetTagAsMap(string(sf.Tag))
		dbTags := tags["db"]
		if len(dbTags) == 0 {
			return // next ForEachField
		}
		column = dbTags[0]
		primary := zstr.StringsContain(dbTags, "primary")
		// zlog.Info("Column:", column, primary, dbTags)
		if primary {
			v.equalFields[sf.Name] = column
		}
		for _, part := range tags["zui"] {
			if part == "-" {
				v.skipFields = append(v.skipFields, column)
				break
			}
			if !primary {
				v.setFields[sf.Name] = column
			}
			if part == "search" {
				v.searchFields = append(v.searchFields, column)
			}
		}
	})
	go v.fillPage()
}

func (v *SQLTableView[S]) updateForIDs(items []S) {
	var info zsql.UpdateInfoSend
	info.Rows = items
	info.TableName = v.tableName
	info.SetColumns = v.setFields
	info.EqualColumns = v.equalFields
	err := zrpc2.MainClient.Call("SQLCalls.UpdateStructs", info, nil)
	zlog.Info("updateForIDs", len(items), err)
	if err != nil {
		zalert.ShowError(err, "updating")
	}
}

func (v *SQLTableView[S]) createSelect() string {
	var s S
	var order string
	if v.Header != nil {
		var orders []string
		for _, s := range v.Header.SortOrder {
			o := s.ID + " "
			if s.SmallFirst {
				o += "ASC"
			} else {
				o += "DESC"
			}
			orders = append(orders, o)
		}
		order = strings.Join(orders, ",")
	}

	where := v.Where
	if v.Where == "" && v.searchString != "" && len(v.searchFields) > 0 {
		var wheres []string
		for _, s := range v.searchFields {
			w := s + `ILIKE '%` + zsql.SanitizeString(v.searchString) + `%'`
			wheres = append(wheres, w)
		}
		where = "(" + strings.Join(wheres, " OR ") + ")"
	}
	if where != "" {
		where = "WHERE " + where
	}
	fields := zsql.FieldNamesStringFromStruct(&s, v.skipFields, "")
	query := fmt.Sprintf("SELECT %s FROM %s %s LIMIT %d OFFSET %d", fields, v.tableName, v.Where, v.limit, v.offset)
	if order != "" {
		query += " ORDER BY " + order
	}
	zlog.Info("Query:", query)
	return query
}

func (v *SQLTableView[S]) createQueryTrailer() string {
	var order string
	if v.Header != nil {
		var orders []string
		for _, s := range v.Header.SortOrder {
			o := s.ID + " "
			if s.SmallFirst {
				o += "ASC"
			} else {
				o += "DESC"
			}
			orders = append(orders, o)
		}
		order = strings.Join(orders, ",")
	}

	where := v.Where
	if v.Where == "" && v.searchString != "" && len(v.searchFields) > 0 {
		var wheres []string
		for _, s := range v.searchFields {
			w := s + `ILIKE '%` + zsql.SanitizeString(v.searchString) + `%'`
			wheres = append(wheres, w)
		}
		where = "(" + strings.Join(wheres, " OR ") + ")"
	}
	if where != "" {
		where = "WHERE " + where
	}
	trailer := fmt.Sprintf("%s LIMIT %d OFFSET %d", v.Where, v.limit, v.offset)
	if order != "" {
		trailer += " ORDER BY " + order
	}
	zlog.Info("Trailer:", trailer)
	return trailer
}

func (v *SQLTableView[S]) SetWhere(where string) {
	v.Where = where
}

func (v *SQLTableView[S]) fillPage() {
	var slice []S
	var q SQLQuery

	q.Query = v.createSelect()
	q.SkipFields = v.skipFields
	err := zrpc2.MainClient.Call(v.selectMethod, q, &slice)
	if err != nil {
		zlog.Error(err, "select", q.Query, v.limit, v.offset)
		return
	}
	v.UpdateSlice(slice)
}

func (v *SQLTableView[S]) fillPageBad() {
	var slice [][]any
	var sr zsql.SelectInfo
	var s S
	sr.Trailer = v.createQueryTrailer()
	sr.GetColumns = zsql.FieldNamesFromStruct(&s, v.skipFields, "")
	sr.TableName = v.tableName
	err := zrpc2.MainClient.Call("SQLCalls.SelectStructs", sr, &slice)
	if err != nil {
		zlog.Error(err, "select", sr.Trailer, v.limit, v.offset)
		return
	}
	// v.UpdateSlice(slice)
}
