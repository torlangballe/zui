package zui

import "github.com/torlangballe/zutil/zgeo"

//  Created by Tor Langballe on /13/11/15.

type ScrollView struct {
	CustomView
	Margin       zgeo.Rect
	HandleScroll func(pos zgeo.Pos)
	child        View
}

func ScrollViewNew() *ScrollView {
	v := &ScrollView{}
	v.init(v, "scrollview")
	return v
}

func (v *ScrollView) AddChild(child View, index int) {
	v.child = child
	v.CustomView.AddChild(child, index)
}

func (v *ScrollView) GetChildren() []View {
	if v.child != nil {
		return []View{v.child}
	}
	return []View{}
}

func (v *ScrollView) ArrangeChildren(onlyChild *View) {
	if v.child != nil {
		ct, got := v.child.(ContainerType)
		if got {
			ct.ArrangeChildren(onlyChild)
		}
	}
}

func (v *ScrollView) GetCalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.minSize
	if v.child != nil {
		cs := v.child.GetCalculatedSize(total)
		s.W = cs.W
	}
	return s
}

func (v *ScrollView) SetContentOffset(offset zgeo.Pos, animated bool) {
}

func (v *ScrollView) SetRect(rect zgeo.Rect) View {
	v.CustomView.SetRect(rect)
	if v.child != nil {
		ls := rect.Size
		ls.H = 20000
		cs := v.child.GetCalculatedSize(ls)
		cs.W = ls.W
		r := zgeo.Rect{Size: cs}
		r.Add(v.Margin)
		v.child.SetRect(r)
	}
	return v
}

func (v *ScrollView) drawIfExposed() {
	v.CustomView.drawIfExposed()
	if v.child != nil {
		et, got := v.child.(ExposableType)
		if got {
			et.drawIfExposed()
		}
	}
}

func ScrollViewToMakeItVisible(view View) {
	// var s:UIView? = view.View()
	// while s != nil {
	//     s = s!.superview
	//     if s != nil {
	//         if let sv = s! as? ZScrollView {
	//             if Double(sv.frame.size.height) - sv.margin.size.h < Double(sv.contentSize.height) {
	//                 let y = float64(view.View().convert(view.View().bounds, to:sv.View()).origin.y)
	//                 sv.SetContentOffset(ZPos(0, y - 40))
	//             }
	//             break
	//         }
	//     }
	// }
}
