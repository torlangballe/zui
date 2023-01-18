package zfields

import (
	"strings"

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

func (v *FieldView) callTriggerHandler(f *Field, action ActionType, value any, view *zview.View) bool {
	if v.params.triggerHandlers != nil {
		t := trigger{id: f.FieldName, action: action}
		function := v.params.triggerHandlers[t]
		if function != nil {
			if function(v, f, value, view) {
				return true
			}
		}
		for t, function := range v.params.triggerHandlers {
			if t.action == action || !strings.Contains(t.id, "*") {
				continue
			}
			path := v.id + "/" + f.FieldName
			// zlog.Info("CallTrig:", v.params.triggerHandlers, t, path, id, action, value)
			if zstr.MatchWildcard(t.id, path) {
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
