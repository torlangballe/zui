package zgo

import "fmt"

//  Created by Tor Langballe on /13/11/15.

type ScrollView struct {
	CustomView
	Margin       Rect
	HandleScroll func(pos Pos)
	child        View
}

func ScrollViewNew() *ScrollView {
	v := &ScrollView{}
	v.init(v, "scrollview")
	return v
}

func (v *ScrollView) SetChild(child View) {
	v.child = child
	v.AddChild(child, -1)
}

func (v *ScrollView) GetCalculatedSize(total Size) Size {
	s := v.minSize
	if v.child != nil {
		cs := v.child.GetCalculatedSize(total)
		s.W = cs.W
	}
	fmt.Println("ScrollView.GetCalculatedSize:", s)
	return s
}

func (v *ScrollView) SetContentOffset(offset Pos, animated bool) {
}

func (v *ScrollView) Rect(rect Rect) View {
	v.CustomView.Rect(rect)
	if v.child != nil {
		ls := rect.Size
		ls.H = 20000
		cs := v.child.GetCalculatedSize(ls)
		cs.W = ls.W
		r := Rect{Size: cs}
		r.Add(v.Margin)
		v.child.Rect(r)
		fmt.Println("scrollview rect:", r)
	}
	return v
}

func (v *ScrollView) drawIfExposed() {
	v.CustomView.drawIfExposed()
	if v.child != nil {
		et, _ := v.child.(ExposableType)
		if et != nil {
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
