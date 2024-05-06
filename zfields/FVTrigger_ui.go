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

func (f *FieldViewParameters) AddTrigger(onFVID string, action ActionType, function func(ap ActionPack) bool) {
	if f.triggerHandlers == nil {
		f.triggerHandlers = map[trigger]func(ap ActionPack) bool{}
	}
	t := trigger{id: onFVID, action: action}
	f.triggerHandlers[t] = function
}

func (v *FieldView) rebuildFieldViewIfUseInValueChangedOrIsRebuild(f *Field) {
	if f.Flags&FlagIsRebuildAllOnChange != 0 {
		v.Rebuild()
		return
	}
	if f.Flags&FlagIsUseInValue != 0 {
		if v.ParentFV != nil {
			sv, _ := v.ParentFV.View.(*FieldSliceView)
			if sv != nil {
				sv.UpdateSlice(f, nil)
				zcontainer.ArrangeChildrenAtRootContainer(v.ParentFV)
				return
			}
		}
		v.Rebuild()
	}
}

func (v *FieldView) callTriggerHandler(ap ActionPack) bool {
	if ap.FieldView == nil {
		ap.FieldView = v
	}
	// zlog.Info("callTriggerHandler", v.Hierarchy(), ap.Field.Name, ap.Action)
	if ap.Action == EditedAction {
		defer v.rebuildFieldViewIfUseInValueChangedOrIsRebuild(ap.Field)
	}
	if v.params.triggerHandlers != nil {
		t := trigger{id: ap.Field.FieldName, action: ap.Action}
		function := v.params.triggerHandlers[t]
		if function != nil {
			if function(ap) {
				return true
			}
		}
		for t, function := range v.params.triggerHandlers {
			// zlog.Info(v.Hierarchy(), "callTrig2?", f.Name, t.action, t.id)
			if ap.Action != NoAction && t.action != ap.Action || !strings.Contains(t.id, "*") {
				continue
			}
			path := v.ID + "/" + ap.Field.FieldName
			// zlog.Info("CallTrig:", t.id, path, f.FieldName, action, value)
			if zstr.MatchWildcard(t.id, path) {
				// zlog.Info("callTriggerHandler3", f.FieldName, t.action, t.id, path)
				if function(ap) {
					return true
				}
			}
		}
	}
	if v.ParentFV != nil {
		return v.ParentFV.callTriggerHandler(ap)
	}
	return false
}
