//go:build zui

package zfields

import (
	"strings"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zstr"
)

type TriggerDataChangedTriggerer interface {
	HandleDataChange(fv *FieldView, f *Field, value any, view *zview.View) bool
}

type trigger struct {
	id     string
	action ActionType
}

// if item.Interface.
// func (s Status) HandleDataChange(fv *FieldView, f Field, value any, view *zview.View) bool {

func (f *FieldViewParameters) AddTrigger(id string, action ActionType, function func(fv *FieldView, f *Field, value any, view *zview.View) bool) {
	if f.triggerHandlers == nil {
		f.triggerHandlers = map[trigger]func(fv *FieldView, f *Field, value any, view *zview.View) bool{}
	}
	t := trigger{id: id, action: action}
	f.triggerHandlers[t] = function
}

func (v *FieldView) reloadFieldViewIfUseInValueChanged(f *Field) {
	if f.Flags&FlagIsUseInValue != 0 {
		if v.parent != nil {
			sv, _ := v.parent.View.(*FieldSliceView)
			if sv != nil {
				sv.UpdateSlice(nil)
				zcontainer.ArrangeChildrenAtRootContainer(v.parent)
				return
			}
		}
		v.Rebuild()
	}
}

func (v *FieldView) callTriggerHandler(f *Field, action ActionType, value any, view *zview.View) bool {
	if action == EditedAction {
		// zlog.Info("callTriggerHandler edit:", f.Name)
		defer v.reloadFieldViewIfUseInValueChanged(f)
	}
	if v.params.triggerHandlers != nil {
		t := trigger{id: f.FieldName, action: action}
		function := v.params.triggerHandlers[t]
		if function != nil {
			if function(v, f, value, view) {
				return true
			}
		}
		for t, function := range v.params.triggerHandlers {
			if t.action != action || !strings.Contains(t.id, "*") {
				continue
			}
			path := v.ID + "/" + f.FieldName
			// zlog.Info("CallTrig:", t.id, path, f.FieldName, action, value)
			if zstr.MatchWildcard(t.id, path) {
				// zlog.Info("callTriggerHandler3", f.FieldName, t.action, t.id, path)
				if function(v, f, value, view) {
					return true
				}
			}
		}
	}
	if v.parent != nil {
		return v.parent.callTriggerHandler(f, action, value, view)
	}
	return false
}
