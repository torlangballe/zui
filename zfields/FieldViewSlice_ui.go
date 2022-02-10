//go:build zui
// +build zui

package zfields

import (
	"fmt"
	"reflect"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
)

func makeCircledButton() *zui.ShapeView {
	v := zui.ShapeViewNew(zui.ShapeViewTypeCircle, zgeo.Size{30, 30})
	v.SetColor(zui.StyleGray(0.8, 0.2))
	return v
}

func makeCircledTextButton(text string, f *Field) *zui.ShapeView {
	v := makeCircledButton()
	w := zgeo.FontDefaultSize + 6
	font := zgeo.FontNice(w, zgeo.FontStyleNormal)
	f.SetFont(v, font)
	v.SetText(text)
	v.SetTextColor(zui.StyleGray(0.2, 0.8))
	v.SetTextAlignment(zgeo.Center)
	return v
}

func makeCircledImageButton(iname string) *zui.ShapeView {
	v := makeCircledButton()
	v.ImageMaxSize = zgeo.Size{20, 20}
	v.SetImage(nil, "images/"+iname+".png", nil)
	return v
}

func makeCircledTrashButton() *zui.ShapeView {
	trash := makeCircledImageButton("trash")
	trash.SetColor(zui.StyleCol(zgeo.ColorNew(0.5, 0.8, 1, 1), zgeo.ColorNew(0.4, 0.1, 0.1, 1)))
	return trash
}

func (v *FieldView) updateSliceValue(structure interface{}, stack *zui.StackView, vertical, showStatic bool, f *Field, sendEdited bool) zui.View {
	ct := stack.Parent().View.(zui.ContainerType)
	newStack := v.buildStackFromSlice(structure, vertical, showStatic, f)
	ct.ReplaceChild(stack, newStack)
	ns := zui.ViewGetNative(newStack)
	ctp := ns.Parent().Parent().View.(zui.ContainerType)
	ctp.ArrangeChildren()
	zui.PresentViewCallReady(newStack, false)
	if sendEdited {
		fh, _ := structure.(ActionHandler)
		// zlog.Info("updateSliceValue:", fh != nil, f.Name, fh)
		if fh != nil {
			fh.HandleAction(f, EditedAction, &newStack)
		}
	}
	return newStack
}

func (v *FieldView) makeNamedSelectionKey(f *Field) string {
	// zlog.Info("makeNamedSelectKey:", v.id, f.FieldName)
	return v.id + "." + f.FieldName + ".NamedSelectionIndex"
}

func (v *FieldView) changeNamedSelectionIndex(i int, f *Field) {
	key := v.makeNamedSelectionKey(f)
	zui.DefaultLocalKeyValueStore.SetInt(i, key, true)
}

func (v *FieldView) buildStackFromSlice(structure interface{}, vertical, showStatic bool, f *Field) zui.View {
	sliceVal, _ := zreflect.FindFieldWithNameInStruct(f.FieldName, structure, true)
	// zlog.Info("buildStackFromSlice:", v.ObjectName(), sliceVal.Len())
	var bar *zui.StackView
	stack := zui.StackViewNew(vertical, f.ID)
	if f != nil && f.Spacing != 0 {
		stack.SetSpacing(f.Spacing)
	}
	key := v.makeNamedSelectionKey(f)
	var selectedIndex int
	single := (f.Flags&flagIsNamedSelection != 0)
	// zlog.Info("buildStackFromSlice:", f.FieldName, vertical, single)
	var fieldView *FieldView
	// zlog.Info("buildStackFromSlice:", vertical, f.ID, val.Len())
	if single {
		selectedIndex, _ = zui.DefaultLocalKeyValueStore.GetInt(key, 0)
		// zlog.Info("buildStackFromSlice:", key, selectedIndex, vertical, f.ID)
		zint.Minimize(&selectedIndex, sliceVal.Len()-1)
		zint.Maximize(&selectedIndex, 0)
		stack.SetMargin(zgeo.RectFromXY2(8, 6, -8, -10))
		stack.SetCorner(zui.GroupingStrokeCorner)
		stack.SetStroke(zui.GroupingStrokeWidth, zui.GroupingStrokeColor)
		label := zui.LabelNew(f.Name)
		label.SetColor(zgeo.ColorNewGray(0, 1))
		font := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleBoldItalic)
		f.SetFont(label, font)
		stack.Add(label, zgeo.TopLeft)
	}
	// zlog.Info("buildStackFromSlice:", f.ID, sliceVal.Len())

	for n := 0; n < sliceVal.Len(); n++ {
		var view zui.View
		a := zgeo.Center
		if vertical {
			a = zgeo.TopLeft
		}
		if f.WidgetName != "" {
			w := widgeters[f.WidgetName]
			if w != nil {
				view = w.Create(f)
				stack.Add(view, a)
			}
		}
		nval := sliceVal.Index(n)
		h, _ := nval.Interface().(ActionFieldHandler)
		if view == nil && h != nil {
			if h.HandleFieldAction(f, CreateFieldViewAction, &view) {
				stack.Add(view, a)
				// fmt.Println("buildStackFromSlice element:", f.FieldName)
			}
		}
		if view == nil {
			childStruct := nval.Addr().Interface()
			vert := !vertical
			if !f.Vertical.IsUndetermined() {
				vert = f.Vertical.Bool()
			}
			if f.LabelizeWidth != 0 {
				vert = true
			}
			// fmt.Printf("buildStackFromSlice element: %s %p\n", f.FieldName, childStruct)
			params := FieldViewParametersDefault()
			fieldView = fieldViewNew(f.ID, vert, childStruct, params, zgeo.Size{}, v)
			view = fieldView
			fieldView.parentField = f
			a := zgeo.Left //| zgeo.HorExpand
			if fieldView.Vertical {
				a |= zgeo.Top
			} else {
				a |= zgeo.VertCenter
			}
			fieldView.buildStack(f.ID, a, showStatic, zgeo.Size{}, true, 5)
			if !f.IsStatic() && !single {
				trash := makeCircledTrashButton()
				fieldView.Add(trash, zgeo.CenterLeft)
				index := n
				trash.SetPressedHandler(func() {
					val, _ := zreflect.FindFieldWithNameInStruct(f.FieldName, structure, true)
					zslice.RemoveAt(val.Addr().Interface(), index)
					// zlog.Info("newlen:", index, val.Len())
					v.updateSliceValue(structure, stack, vertical, showStatic, f, true)
				})
			}
			stack.Add(fieldView, zgeo.TopLeft|zgeo.HorExpand)
		}
		collapse := single && n != selectedIndex
		stack.CollapseChild(view, collapse, false)
		zlog.Assert(view != nil)
	}
	if f.MinWidth == 0 && f.Size.W != 0 {
		flen := float64(sliceVal.Len())
		f.MinWidth = f.Size.W*flen + f.Spacing*(flen-1)
	}
	if single {
		zlog.Assert(!f.IsStatic())
		bar = zui.StackViewHor(f.ID + ".bar")
		stack.Add(bar, zgeo.TopLeft|zgeo.HorExpand)
	}
	if !f.IsStatic() {
		plus := makeCircledImageButton("plus")
		plus.SetPressedHandler(func() {
			val, _ := zreflect.FindFieldWithNameInStruct(f.FieldName, structure, true)
			a := reflect.New(val.Type().Elem()).Elem()
			nv := reflect.Append(val, a)
			if fieldView != nil {
				fieldView.structure = nv.Interface()
			}
			val.Set(nv)
			a = val.Index(val.Len() - 1) // we re-set a, as it is now a new value at the end of slice
			if single {
				v.changeNamedSelectionIndex(val.Len()-1, f)
			}
			// fmt.Printf("SLICER + Pressed: %p %p\n", val.Index(val.Len()-1).Addr().Interface(), a.Addr().Interface())
			fhItem, _ := a.Addr().Interface().(ActionHandler)
			if fhItem != nil {
				fhItem.HandleAction(f, NewStructAction, nil)
			}
			v.updateSliceValue(structure, stack, vertical, showStatic, f, true)
			//			stack.CustomView.PressedHandler()()
		})
		if bar != nil {
			bar.Add(plus, zgeo.CenterRight)
		} else {
			stack.Add(plus, zgeo.TopLeft)
		}
	}
	if single {
		shape := makeCircledTrashButton()
		bar.Add(shape, zgeo.CenterRight)
		shape.SetPressedHandler(func() {
			zui.AlertAsk("Delete this entry?", func(ok bool) {
				if ok {
					val, _ := zreflect.FindFieldWithNameInStruct(f.FieldName, structure, true)
					zslice.RemoveAt(val.Addr().Interface(), selectedIndex)
					// zlog.Info("newlen:", index, val.Len())
					v.updateSliceValue(structure, stack, vertical, showStatic, f, true)
				}
			})
		})
		shape.SetUsable(sliceVal.Len() > 0)
		// zlog.Info("Make Slice thing:", key, selectedIndex, val.Len())

		shape = makeCircledTextButton("⇦", f)
		bar.Add(shape, zgeo.CenterLeft)
		shape.SetPressedHandler(func() {
			v.changeNamedSelectionIndex(selectedIndex-1, f)
			v.updateSliceValue(structure, stack, vertical, showStatic, f, false)
		})
		shape.SetUsable(selectedIndex > 0)

		str := "0"
		if sliceVal.Len() > 0 {
			str = fmt.Sprintf("%d of %d", selectedIndex+1, sliceVal.Len())
		}
		label := zui.LabelNew(str)
		label.SetColor(zgeo.ColorNewGray(0, 1))
		f.SetFont(label, nil)
		bar.Add(label, zgeo.CenterLeft)

		shape = makeCircledTextButton("⇨", f)
		bar.Add(shape, zgeo.CenterLeft)
		shape.SetPressedHandler(func() {
			v.changeNamedSelectionIndex(selectedIndex+1, f)
			v.updateSliceValue(structure, stack, vertical, showStatic, f, false)
		})
		shape.SetUsable(selectedIndex < sliceVal.Len()-1)
	}
	return stack
}

func updateSliceFieldView(view zui.View, selectedIndex int, item zreflect.Item, f *Field, dontOverwriteEdited bool) {
	// zlog.Info("updateSliceFieldView:", view.ObjectName(), item.FieldName, f.Name)
	children := (view.(zui.ContainerType)).GetChildren(false)
	n := 0
	subViewCount := len(children)
	single := (f.Flags&flagIsNamedSelection != 0)
	if single {
		subViewCount -= 2
	}
	// if subViewCount != item.Value.Len() {
	// 	zlog.Info("SLICE VIEW: length changed!!!", subViewCount, item.Value.Len())
	// }
	for _, c := range children {
		// zlog.Info("Update Sub", c.ObjectName())
		if n >= item.Value.Len() {
			break
		}
		if single && n != selectedIndex {
			continue
		}
		val := item.Value.Index(n)
		w := widgeters[f.WidgetName]
		if w != nil {
			w.SetValue(c, val.Interface())
			n++
			continue
		}
		fv, _ := c.(*FieldView)
		if fv == nil {
			ah, _ := val.Interface().(ActionFieldHandler)
			// zlog.Info("Update Sub Slice field fv == nil:", n, ah != nil)
			if ah != nil {
				cview := c
				ah.HandleFieldAction(f, DataChangedAction, &cview)
			}
		} else {
			fv.structure = val.Addr().Interface()
			fv.Update(dontOverwriteEdited)
		}
		n++
		// }
		// zlog.Info("struct make field view:", f.Name, f.Kind, exp)
	}
	// if updateStackFromActionFieldHandlerSlice(view, &item, f) {
	// 	continue
	// }
}
